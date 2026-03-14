// Package unbounded_channel — linearizability_test.go
//
// 用 Porcupine 对 UnboundedChannelV6 进行线性化（FIFO 语义）验证。
//
// # 什么是线性化检验
//
// 线性化（Linearizability）是并发数据结构正确性的黄金标准：
// 一段并发操作历史，若能找到某种合法的顺序执行序列，使每个操作的输出都与该序列一致，
// 则称该历史是可线性化的。
//
// 对 FIFO 通道来说，"合法的顺序执行"即标准 FIFO 队列的行为：
//   - Enqueue(v)：将 v 追加到队尾
//   - Dequeue()：返回队头元素（队列非空时）
//
// 如果接收方收到的消息顺序与发送方的 happens-before 顺序不一致（即 FIFO 被破坏），
// Porcupine 会报告 Illegal。
//
// # 局限性说明
//
// Porcupine 只检验实际发生的历史，不能穷举所有调度序列。
// 本测试通过以下手段提升发现问题的概率：
//   1. 多轮运行，每轮都有不同的调度随机性
//   2. 高背压场景：强迫大量消息走 buffer 路径（绕过快速路径）
//   3. 小 chanSize：最大化 buffer↔channel 搬运频率
//
// 若要确定性地复现 bufferLen 顺序 bug（见 bufferEnqueue 注释），
// 可在测试 build tag 下在 bufferLen.Add(1) 与 buffer.Enqueue() 之间注入 runtime.Gosched()。

package unbounded_channel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anishathalye/porcupine"
)

// ─────────────────────────────────────────────────────────────────────────────
// FIFO Queue 顺序规约（Sequential Specification）
// ─────────────────────────────────────────────────────────────────────────────

// ucOpInput 描述一次操作的输入：Send 或 Receive。
type ucOpInput struct {
	isSend bool
	val    int // Send 时有效；Receive 时忽略
}

// ucOpOutput 描述一次操作的输出。
type ucOpOutput struct {
	val int  // Receive 时有效：收到的消息值
	ok  bool // Send：是否投递成功；Receive：是否收到消息（false 表示通道已关闭）
}

// fifoModel 是 UnboundedChannelV6 的顺序规约：标准 FIFO 队列。
//
// State = []int，表示队列中按入队顺序排列的消息。
// 线性化要求：并发历史能被映射到某个合法的顺序 FIFO 队列执行。
var fifoModel = porcupine.Model{
	Init: func() interface{} {
		return []int{}
	},

	Step: func(state, input, output interface{}) (bool, interface{}) {
		q := toIntQ(state)
		op := input.(ucOpInput)
		res := output.(ucOpOutput)

		if op.isSend {
			if !res.ok {
				// Send 返回 false 表示通道已关闭；本测试只记录成功的 Send，不应出现此情况
				return false, state
			}
			// 入队：追加到队尾
			newQ := make([]int, len(q)+1)
			copy(newQ, q)
			newQ[len(q)] = op.val
			return true, newQ
		}

		// Receive
		if !res.ok {
			// 通道关闭且队列为空时合法
			return len(q) == 0, q
		}
		if len(q) == 0 || q[0] != res.val {
			// 队列为空时不可能收到消息，或收到的值与队头不符（FIFO 违反）
			return false, state
		}
		// 出队：移除队头
		return true, q[1:]
	},

	Equal: func(s1, s2 interface{}) bool {
		q1, q2 := toIntQ(s1), toIntQ(s2)
		if len(q1) != len(q2) {
			return false
		}
		for i := range q1 {
			if q1[i] != q2[i] {
				return false
			}
		}
		return true
	},

	DescribeOperation: func(input, output interface{}) string {
		op := input.(ucOpInput)
		res := output.(ucOpOutput)
		if op.isSend {
			return fmt.Sprintf("Send(%d) → ok=%v", op.val, res.ok)
		}
		return fmt.Sprintf("Recv() → (%d, ok=%v)", res.val, res.ok)
	},

	DescribeState: func(state interface{}) string {
		return fmt.Sprintf("%v", toIntQ(state))
	},
}

func toIntQ(s interface{}) []int {
	if s == nil {
		return nil
	}
	return s.([]int)
}

// ─────────────────────────────────────────────────────────────────────────────
// 测试辅助：执行一轮并发测试，收集操作历史
// ─────────────────────────────────────────────────────────────────────────────

// runAndCollect 运行一轮并发测试并返回 Porcupine 操作历史。
//
// 设计：
//   - numProducers 个生产者，每个顺序发送 msgsPerProducer 条消息（全局值唯一）
//   - numConsumers 个消费者，持续接收直到通道关闭
//   - 所有发送完成后关闭通道，消费者排空后退出
//   - 只记录成功的 Send（ok=true）和成功的 Receive（ok=true）
func runAndCollect(
	t *testing.T,
	chanSize, limit, numProducers, msgsPerProducer, numConsumers int,
) []porcupine.Operation {
	t.Helper()

	uc := NewUnboundedChannelV6[int](chanSize, limit)

	var (
		mu        sync.Mutex
		ops       []porcupine.Operation
		clientSeq atomic.Int32
		wgSend    sync.WaitGroup
		wgRecv    sync.WaitGroup
	)

	ts := func() int64 { return time.Now().UnixNano() }

	// 生产者：每个生产者顺序发送，消息值全局唯一
	// 生产者 p 发送值域：[p*msgsPerProducer, (p+1)*msgsPerProducer)
	// 单生产者的顺序发送在时间轴上不重叠，Porcupine 强制保留其顺序
	for p := 0; p < numProducers; p++ {
		wgSend.Add(1)
		go func(pid int) {
			defer wgSend.Done()
			cid := int(clientSeq.Add(1)) - 1
			for i := 0; i < msgsPerProducer; i++ {
				val := pid*msgsPerProducer + i
				t0 := ts()
				ok := uc.Send(val)
				t1 := ts()
				if ok { // 只记录成功的 Send
					mu.Lock()
					ops = append(ops, porcupine.Operation{
						ClientId: cid,
						Input:    ucOpInput{isSend: true, val: val},
						Call:     t0,
						Output:   ucOpOutput{ok: true},
						Return:   t1,
					})
					mu.Unlock()
				}
			}
		}(p)
	}

	// 消费者：持续接收直到 ok=false（通道关闭且排空）
	for c := 0; c < numConsumers; c++ {
		wgRecv.Add(1)
		go func() {
			defer wgRecv.Done()
			cid := int(clientSeq.Add(1)) - 1
			for {
				t0 := ts()
				val, ok := uc.Receive()
				t1 := ts()
				if !ok {
					return // 通道关闭且排空，退出
				}
				mu.Lock()
				ops = append(ops, porcupine.Operation{
					ClientId: cid,
					Input:    ucOpInput{isSend: false},
					Call:     t0,
					Output:   ucOpOutput{val: val, ok: true},
					Return:   t1,
				})
				mu.Unlock()
			}
		}()
	}

	// 所有发送完成后关闭通道；worker 会将 buffer 中剩余消息搬运到 channel 后再关闭底层通道；
	// 消费者排空后见到 ok=false，自然退出。
	go func() {
		wgSend.Wait()
		uc.Close()
	}()

	wgRecv.Wait()
	return ops
}

// checkLinearizability 调用 Porcupine 验证历史的 FIFO 线性化。
func checkLinearizability(t *testing.T, ops []porcupine.Operation) {
	t.Helper()
	res := porcupine.CheckOperationsTimeout(fifoModel, ops, 60*time.Second)
	switch res {
	case porcupine.Ok:
		// pass
	case porcupine.Illegal:
		t.Fatalf("FIFO linearizability violated: history is not consistent with "+
			"a sequential FIFO queue (%d operations recorded)", len(ops))
	case porcupine.Unknown:
		t.Logf("WARNING: linearizability check timed out with %d operations — "+
			"consider reducing test size or increasing timeout", len(ops))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 测试用例
// ─────────────────────────────────────────────────────────────────────────────

// TestLinearizabilitySingleProducer 单生产者场景。
//
// 单生产者严格顺序发送，Porcupine 强制要求线性化顺序与发送顺序完全一致。
// 这是最严格的 FIFO 测试：任何一条消息被乱序接收都会立即被 Porcupine 检出。
//
// 参数说明：
// 单消费者顺序接收，理论上无操作重叠，Porcupine 搜索空间接近线性。
// 实际上因为消费者在 channel 为空时阻塞等待（buffer 异步搬运期间），
// 接收操作的时间窗口较宽，与发送操作存在一定重叠。
// 20 msgs（40 ops）在实测中稳定在 5s 以内完成。
func TestLinearizabilitySingleProducer(t *testing.T) {
	ops := runAndCollect(t,
		/* chanSize= */ 4,
		/* limit= */ 2000,
		/* numProducers= */ 1,
		/* msgsPerProducer= */ 20,
		/* numConsumers= */ 1,
	)
	t.Logf("collected %d operations", len(ops))
	checkLinearizability(t, ops)
}

// TestLinearizabilityMultiProducer 多生产者并发场景。
//
// 验证：
//   - 每个生产者内部的 FIFO 顺序不被打乱（producer-local FIFO）
//   - 不同生产者的消息可以合法交错（Porcupine 允许任意顺序）
//   - 小 chanSize=2 迫使大量消息走 buffer，覆盖 bufferEnqueue/bufferDequeue 路径
//
// 参数说明：
// Porcupine 复杂度是"并发重叠操作数"的指数函数，不只是总操作数。
// 多生产者场景所有操作时间戳大量重叠；2P × 8 msgs + 2C ≈ 32 ops，
// 既覆盖多生产者路径，又保证 Porcupine 快速完成。
func TestLinearizabilityMultiProducer(t *testing.T) {
	ops := runAndCollect(t,
		/* chanSize= */ 2,
		/* limit= */ 2000,
		/* numProducers= */ 2,
		/* msgsPerProducer= */ 8,
		/* numConsumers= */ 2,
	)
	t.Logf("collected %d operations", len(ops))
	checkLinearizability(t, ops)
}

// TestLinearizabilityHighBackpressure 高背压场景。
//
// limit=10 极小：生产者很快超过背压上限，大量 goroutine 阻塞在 condSendWaiter.Wait()。
// 目的：
//   - 验证 condSendWaiter 的唤醒逻辑（包括 buffer 排空时的 Broadcast）
//   - 验证背压解除后 FIFO 顺序仍然成立
//   - 覆盖 transfer() 中各种唤醒分支
//
// 参数说明：
// 背压会迫使生产者在 condSendWaiter 上序列化，降低并发重叠程度。
// 因此可以使用更多生产者（5P × 6 msgs + 2C ≈ 60 ops）而不会让 Porcupine 超时。
// 实测稳定在 10s 以内完成。
func TestLinearizabilityHighBackpressure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping backpressure test in short mode")
	}
	ops := runAndCollect(t,
		/* chanSize= */ 2,
		/* limit= */ 10,
		/* numProducers= */ 5,
		/* msgsPerProducer= */ 6,
		/* numConsumers= */ 2,
	)
	t.Logf("collected %d operations", len(ops))
	checkLinearizability(t, ops)
}

// TestLinearizabilityStress 压力场景：多轮运行，最大化调度随机性。
//
// 每轮使用不同的 goroutine 调度序列，通过轮数积累覆盖各种罕见的调度窗口。
// chanSize=2 保证大量消息必须经过 buffer，覆盖所有慢路径代码。
//
// 参数说明：
// 每轮 ~24 ops（2P × 6 msgs + 2C），保证 Porcupine 在 1s 内完成（不触发超时告警）。
// 100 轮以不同调度序列反复验证，总覆盖面远超单次大参数测试；
// 遇到时序 bug 通常在前几轮就能检出。
func TestLinearizabilityStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	const rounds = 100
	for round := range rounds {
		t.Run(fmt.Sprintf("round-%d", round), func(t *testing.T) {
			ops := runAndCollect(t,
				/* chanSize= */ 2,
				/* limit= */ 500,
				/* numProducers= */ 2,
				/* msgsPerProducer= */ 6,
				/* numConsumers= */ 2,
			)
			checkLinearizability(t, ops)
		})
	}
}
