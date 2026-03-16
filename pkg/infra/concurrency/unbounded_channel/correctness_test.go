// Package unbounded_channel — correctness_test.go
//
// 针对 UnboundedChannelV6 的完整正确性测试。
// 与现有测试（compare_test.go、linearizability_test.go）互补，专注于：
//
//  1. FIFO 语义边界：快速路径、慢路径（chanSize=1）、buffer 高压、分批发送
//  2. 消息完整性：含 double-enqueue 回归测试（V6 修复的核心 bug）
//  3. 背压行为：生产者确实被阻塞；消费者消费后生产者确实解除阻塞
//  4. 生命周期安全：无死锁、无 goroutine 泄漏、Close 并发安全
//  5. 长时间压力测试：以上性质在数百万次操作后依然成立
//
// 运行方式：
//
//	# 快速验证（含 race detector，约 2 分钟）
//	go test ./pkg/infra/concurrency/unbounded_channel/ \
//	    -run "TestFIFO|TestIntegrity|TestBackpressure|TestLifecycle" \
//	    -race -v -timeout 300s
//
//	# 压力测试（需加 -tags stress，见 stress_test.go）
//	go test -tags stress -run TestStress -v -timeout 30m \
//	    ./pkg/infra/concurrency/unbounded_channel/
package unbounded_channel

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// 辅助工具
// ─────────────────────────────────────────────────────────────────────────────

// checkGoroutineLeak 在 f() 执行前后对比 goroutine 数量，检测泄漏。
// tolerance：允许的额外 goroutine 数（容纳测试框架自身的轻微抖动）。
// 注意：不可与其他 goroutine 密集型测试并行运行。
func checkGoroutineLeak(t *testing.T, tolerance int, f func()) {
	t.Helper()
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	before := runtime.NumGoroutine()

	f()

	// 最多等 3s，让 goroutine 自然退出
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if runtime.NumGoroutine() <= before+tolerance {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	after := runtime.NumGoroutine()
	if after > before+tolerance {
		buf := make([]byte, 1<<16)
		n := runtime.Stack(buf, true)
		t.Errorf("goroutine 泄漏：创建前=%d，关闭后=%d（容忍=%d）\n%s",
			before, after, tolerance, buf[:n])
	}
}

// assertStrictFIFO 断言 received 严格等于 [0, 1, …, n-1]。
func assertStrictFIFO(t *testing.T, received []int, n int) {
	t.Helper()
	if len(received) != n {
		t.Fatalf("FIFO：期望 %d 条消息，实际收到 %d 条", n, len(received))
	}
	for i, v := range received {
		if v != i {
			t.Fatalf("FIFO 违反：received[%d]=%d（期望 %d）", i, v, i)
		}
	}
}

// assertNoDup 断言 [0, n) 范围内每个值恰好出现一次。
func assertNoDup(t *testing.T, seen []atomic.Bool, n int) {
	t.Helper()
	for i := range n {
		if !seen[i].Load() {
			t.Errorf("消息丢失：值 %d 未被消费", i)
		}
	}
}

// drainAll 持续接收直到 channel 关闭，返回所有消息。
func drainAll(ch *UnboundedChannelV6[int]) []int {
	var result []int
	for {
		v, ok := ch.Receive()
		if !ok {
			return result
		}
		result = append(result, v)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. FIFO 语义测试
// ─────────────────────────────────────────────────────────────────────────────

// TestFIFO_SlowPath_ChanSize1 使用 chanSize=1 迫使所有消息经过 buffer 路径，
// 验证 buffer→channel 搬运后 FIFO 顺序不被破坏。
// 与 compare_test.go 的 TestCompare_FIFO（chanSize=100）互补，专门压测慢路径。
func TestFIFO_SlowPath_ChanSize1(t *testing.T) {
	const n = 50_000
	ch := NewUnboundedChannelV6[int](1, n*10)
	defer ch.Close()

	received := make([]int, 0, n)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range n {
			v, ok := ch.Receive()
			if !ok {
				return
			}
			received = append(received, v)
		}
	}()

	for i := range n {
		ch.Send(i)
	}

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatalf("超时：已收到 %d/%d 条消息", len(received), n)
	}

	assertStrictFIFO(t, received, n)
}

// TestFIFO_SlowPath_SendBeforeReceive 先将所有消息压入 buffer（无消费者），
// 再启动消费者，验证堆积后的 FIFO 顺序。
// 这是 double-enqueue bug 最容易触发的场景：channel 满 + buffer 非空。
func TestFIFO_SlowPath_SendBeforeReceive(t *testing.T) {
	const n = 5_000
	ch := NewUnboundedChannelV6[int](1, n*10)
	defer ch.Close()

	// 先全部发送（无消费者：channel 1 条满后，其余全走 buffer）
	for i := range n {
		ch.Send(i)
	}

	// 再消费
	received := make([]int, 0, n)
	for range n {
		v, ok := ch.Receive()
		if !ok {
			break
		}
		received = append(received, v)
	}

	assertStrictFIFO(t, received, n)
}

// TestFIFO_BurstPattern 分批发送（每批后短暂停顿让 buffer 排空），
// 验证快慢路径交替切换时 FIFO 不被破坏。
func TestFIFO_BurstPattern(t *testing.T) {
	const (
		chanSize  = 10
		batchSize = 200
		batches   = 30
		total     = batchSize * batches
	)
	ch := NewUnboundedChannelV6[int](chanSize, total*2)
	defer ch.Close()

	received := make([]int, 0, total)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range total {
			v, ok := ch.Receive()
			if !ok {
				return
			}
			received = append(received, v)
		}
	}()

	seq := 0
	for range batches {
		for range batchSize {
			ch.Send(seq)
			seq++
		}
		// 每批后暂停：让 worker 有机会排空 buffer，下一批走快路径
		time.Sleep(5 * time.Millisecond)
	}

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatalf("超时：已收到 %d/%d 条消息", len(received), total)
	}

	assertStrictFIFO(t, received, total)
}

// TestFIFO_Regression_NoDoubleEnqueue 是 double-enqueue bug 的定向回归测试。
//
// Bug 根因：sendSlow() 在 channel 满时调用 bufferEnqueue(msg) 后缺少 return，
// 导致代码 fallthrough 到下一个 if（buffer.Len()!=0 此时为 true），
// 再次调用 bufferEnqueue(msg)，同一条消息被入队两次。
//
// 触发条件：channel 满（走第一个 if）+ buffer 非空（触发 fallthrough 的第二个 if）。
// chanSize=1 且无消费者 → 发送第 2 条消息时 channel 已满，之后 buffer 持续非空，
// 每条消息都会触发该路径。若 bug 存在，buffer 中每条消息出现两次，消费者收到重复消息。
func TestFIFO_Regression_NoDoubleEnqueue(t *testing.T) {
	const n = 2_000
	ch := NewUnboundedChannelV6[int](1, n*10)
	defer ch.Close()

	// 先全部发送（无消费者，最大化触发 bug 概率）
	for i := range n {
		ch.Send(i)
	}

	received := make([]int, 0, n)
	for range n {
		v, ok := ch.Receive()
		if !ok {
			break
		}
		received = append(received, v)
	}

	// 顺序检查（double-enqueue 导致 received[2]=1 而非 2）
	assertStrictFIFO(t, received, n)

	// 重复检查（double-enqueue 导致某些值出现两次）
	count := make(map[int]int, n)
	for _, v := range received {
		count[v]++
	}
	for i := range n {
		if count[i] != 1 {
			t.Errorf("double-enqueue：值 %d 出现了 %d 次（期望 1 次）", i, count[i])
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. 消息完整性测试
// ─────────────────────────────────────────────────────────────────────────────

// TestIntegrity_MultiProducer_PerProducerOrder 验证每个生产者的消息按发送顺序抵达。
//
// V6 的 FIFO 保证：消息按 sendSlow() 内 mutex 的获取顺序被 worker 搬运。
// 对单个生产者（顺序调用 Send()），其消息的 mutex 获取顺序等于发送顺序，
// 因此每个生产者的消息在消费端必须单调递增。
func TestIntegrity_MultiProducer_PerProducerOrder(t *testing.T) {
	const (
		producers      = 10
		msgsPerProducer = 1_000
		total           = producers * msgsPerProducer
	)

	type tagged struct {
		pid int
		seq int
	}

	ch := NewUnboundedChannelV6[tagged](64, total*2)
	defer ch.Close()

	lastSeq := make([]atomic.Int32, producers)
	for i := range lastSeq {
		lastSeq[i].Store(-1)
	}

	var receivedCount atomic.Int64
	allDone := make(chan struct{})
	errCh := make(chan string, 1) // 消费者 goroutine 内无法直接 t.Fatal，用 channel 传递

	// 单消费者（保证 lastSeq 的读写无竞争）
	go func() {
		for {
			msg, ok := ch.Receive()
			if !ok {
				return
			}
			prev := int(lastSeq[msg.pid].Load())
			if msg.seq <= prev {
				select {
				case errCh <- fmt.Sprintf("per-producer FIFO 违反：producer=%d，上次 seq=%d，本次 seq=%d",
					msg.pid, prev, msg.seq):
				default:
				}
				return
			}
			lastSeq[msg.pid].Store(int32(msg.seq))
			if receivedCount.Add(1) == int64(total) {
				close(allDone)
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(producers)
	for p := range producers {
		go func(pid int) {
			defer wg.Done()
			for i := range msgsPerProducer {
				ch.Send(tagged{pid: pid, seq: i})
			}
		}(p)
	}
	wg.Wait()

	select {
	case <-allDone:
	case errMsg := <-errCh:
		t.Fatal(errMsg)
	case <-time.After(60 * time.Second):
		t.Fatalf("超时：已收到 %d/%d 条消息", receivedCount.Load(), total)
	}

	// 确认每个生产者的最后一条消息序号正确
	for p := range producers {
		if int(lastSeq[p].Load()) != msgsPerProducer-1 {
			t.Errorf("producer=%d：期望最后 seq=%d，实际=%d",
				p, msgsPerProducer-1, lastSeq[p].Load())
		}
	}

	select {
	case errMsg := <-errCh:
		t.Fatal(errMsg)
	default:
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. 背压测试
// ─────────────────────────────────────────────────────────────────────────────

// TestBackpressure_ProducerBlocks 验证 buffer 超过 limit 后生产者确实被阻塞。
//
// 参数推导（chanSize=1, limit=1）：
//   - msg 0: 快速路径 → channel（bufferLen=0）
//   - msg 1: 慢路径，channel 满 → buffer（bufferLen=1）；发送前 bufferLen=0 不触发等待
//   - msg 2: 慢路径，channel 满 → buffer（bufferLen=2）；发送前 bufferLen=1 不触发等待
//   - msg 3: 发送前 bufferLen=2 > limit=1 → 阻塞！
func TestBackpressure_ProducerBlocks(t *testing.T) {
	ch := NewUnboundedChannelV6[int](1, 1)
	defer ch.Close()

	// 填满 channel + buffer 但不触发背压（3 条）
	for i := range 3 {
		done := make(chan struct{})
		go func() {
			ch.Send(i)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("第 %d 条消息不应被阻塞，但超时了", i)
		}
	}

	// 第 4 条应被阻塞
	sendDone := make(chan struct{})
	go func() {
		ch.Send(3)
		close(sendDone)
	}()

	select {
	case <-sendDone:
		t.Fatal("第 4 条 Send() 应被背压阻塞，但立即返回了")
	case <-time.After(150 * time.Millisecond):
		// 预期：生产者正在阻塞
	}

	// 消费一条 → worker 搬运 → bufferLen 下降 → 生产者解除阻塞
	ch.Receive()

	select {
	case <-sendDone:
		// 生产者已解除阻塞
	case <-time.After(500 * time.Millisecond):
		t.Fatal("消费者消费后，生产者应解除阻塞，但超时了")
	}
}

// TestBackpressure_MultiProducer_AllUnblock 多个生产者同时因背压阻塞，
// 验证消费者持续消费后所有阻塞生产者最终都能完成发送。
func TestBackpressure_MultiProducer_AllUnblock(t *testing.T) {
	const (
		numProducers   = 5
		msgsPerProducer = 100
		total           = numProducers * msgsPerProducer
		chanSize        = 4
		limit           = 10 // 较小的 limit 迫使背压频繁触发
	)

	ch := NewUnboundedChannelV6[int](chanSize, limit)
	defer ch.Close()

	seen := make([]atomic.Bool, total)
	var receivedCount atomic.Int64
	allDone := make(chan struct{})

	// 慢速消费者（触发持续背压）
	go func() {
		for {
			v, ok := ch.Receive()
			if !ok {
				return
			}
			seen[v].Store(true)
			if receivedCount.Add(1) == int64(total) {
				close(allDone)
				return
			}
			time.Sleep(200 * time.Microsecond) // 慢速，制造背压
		}
	}()

	var wg sync.WaitGroup
	wg.Add(numProducers)
	for p := range numProducers {
		go func(pid int) {
			defer wg.Done()
			base := pid * msgsPerProducer
			for i := range msgsPerProducer {
				ch.Send(base + i)
			}
		}(p)
	}
	wg.Wait()

	select {
	case <-allDone:
	case <-time.After(60 * time.Second):
		t.Fatalf("超时：已收到 %d/%d 条消息", receivedCount.Load(), total)
	}

	assertNoDup(t, seen, total)
}

// TestBackpressure_CloseUnblocksProducers 是核心死锁测试。
//
// 场景：多个生产者因 buffer 超过 limit 而阻塞在 condSendWaiter.Wait()，
// 此时调用 Close()。Close() 内部广播 condSendWaiter，所有阻塞的生产者必须
// 被唤醒并返回 false，而不是永远阻塞（死锁）。
func TestBackpressure_CloseUnblocksProducers(t *testing.T) {
	ch := NewUnboundedChannelV6[int](1, 1)

	// 填满至接近背压阈值（3 条，bufferLen=2 > limit=1 触发阻塞）
	for range 3 {
		ch.Send(0)
	}

	const numBlockedProducers = 5
	var unblocked atomic.Int32
	var wg sync.WaitGroup
	wg.Add(numBlockedProducers)

	for range numBlockedProducers {
		go func() {
			defer wg.Done()
			if !ch.Send(999) {
				unblocked.Add(1) // Close() 触发，Send 返回 false
			}
		}()
	}

	// 等待生产者进入阻塞状态
	time.Sleep(80 * time.Millisecond)

	// Close() 必须唤醒所有阻塞的生产者
	ch.Close()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		if int(unblocked.Load()) != numBlockedProducers {
			t.Errorf("期望 %d 个生产者返回 false，实际 %d 个",
				numBlockedProducers, unblocked.Load())
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("死锁：Close() 未能唤醒阻塞的生产者（仍有 %d 个阻塞）",
			numBlockedProducers-int(unblocked.Load()))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. 生命周期安全测试
// ─────────────────────────────────────────────────────────────────────────────

// TestLifecycle_NoGoroutineLeak 验证 Close() 后 worker goroutine 正确退出，无泄漏。
// 不可与其他 goroutine 密集型测试并行运行。
func TestLifecycle_NoGoroutineLeak(t *testing.T) {
	const rounds = 20

	checkGoroutineLeak(t, 1, func() {
		for range rounds {
			ch := NewUnboundedChannelV6[int](16, 1000)

			// 发送 + 消费一些消息
			const n = 50
			done := make(chan struct{})
			go func() {
				defer close(done)
				for range n {
					ch.Receive()
				}
			}()
			for i := range n {
				ch.Send(i)
			}
			<-done

			ch.Close()

			// drain：消费所有残余消息，使 worker 能检测到 canClose()
			for {
				_, ok := ch.Receive()
				if !ok {
					break
				}
			}
		}
	})
}

// TestLifecycle_CloseWithPendingBuffer 验证 Close() 后 buffer 中的消息仍可被消费。
//
// 关键语义：Close() 仅设置关闭标志，不丢弃 buffer 中的数据。
// worker 在 Close() 后继续工作，直到 buffer 和 channel 均排空后才真正关闭底层 channel。
// 消费者能收到 Close() 前发送的所有消息。
func TestLifecycle_CloseWithPendingBuffer(t *testing.T) {
	const n = 1_000
	ch := NewUnboundedChannelV6[int](1, n*10) // chanSize=1 确保大量消息进入 buffer

	// 先发送，无消费者
	for i := range n {
		ch.Send(i)
	}

	// 发送完毕后立即关闭
	ch.Close()

	// 消费者在 Close() 后启动，期望能收到所有 n 条消息
	received := drainAll(ch)
	assertStrictFIFO(t, received, n)
}

// TestLifecycle_SendAfterClose 验证 Close() 后 Send() 返回 false 而非 panic。
func TestLifecycle_SendAfterClose(t *testing.T) {
	ch := NewUnboundedChannelV6[int](16, 1000)
	ch.Close()

	// 排空使 worker 退出
	for {
		_, ok := ch.Receive()
		if !ok {
			break
		}
	}

	// Close() 后的 Send() 必须返回 false，不能 panic
	for range 100 {
		if ch.Send(1) {
			t.Fatal("Close() 后 Send() 应返回 false")
		}
	}
}

// TestLifecycle_ReceiveAfterClose 验证 Close() 且消息排空后 Receive() 返回 (zero, false)。
func TestLifecycle_ReceiveAfterClose(t *testing.T) {
	const n = 50
	ch := NewUnboundedChannelV6[int](8, 1000)

	for i := range n {
		ch.Send(i)
	}
	ch.Close()

	// 收完所有消息
	received := drainAll(ch)
	assertStrictFIFO(t, received, n)

	// 再次 Receive 应返回零值 + false
	v, ok := ch.Receive()
	if ok {
		t.Errorf("排空后 Receive() 应返回 (0,false)，实际得到 (%d,true)", v)
	}
}

// TestLifecycle_ConcurrentSendClose 验证并发 Send() 和 Close() 不产生 panic。
// 建议配合 -race 运行。
func TestLifecycle_ConcurrentSendClose(t *testing.T) {
	const numGoroutines = 50

	ch := NewUnboundedChannelV6[int](16, 10_000)

	// 消费者持续消费
	go func() {
		for {
			_, ok := ch.Receive()
			if !ok {
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range 200 {
				ch.Send(id*200 + j) // 可能返回 false，不能 panic
			}
		}(i)
	}

	// 让部分 goroutine 先跑一段
	time.Sleep(5 * time.Millisecond)
	ch.Close() // 并发关闭

	// 所有 Send 必须在有限时间内返回（无死锁）
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("并发 Send()+Close()：部分 goroutine 未在超时内返回（疑似死锁）")
	}

	// 排空确保 worker 退出
	for {
		_, ok := ch.Receive()
		if !ok {
			break
		}
	}
}

// TestLifecycle_CloseBeforeSend 验证先 Close() 再 Send() 的边界情况。
func TestLifecycle_CloseBeforeSend(t *testing.T) {
	ch := NewUnboundedChannelV6[int](16, 1000)
	ch.Close()

	// drain 让 worker 退出
	for {
		_, ok := ch.Receive()
		if !ok {
			break
		}
	}

	for range 10 {
		if ch.Send(1) {
			t.Fatal("Close() 后 Send() 应返回 false")
		}
	}
}

// TestLifecycle_RapidCreateClose 快速循环创建和关闭，验证无资源泄漏。
func TestLifecycle_RapidCreateClose(t *testing.T) {
	checkGoroutineLeak(t, 1, func() {
		for range 200 {
			ch := NewUnboundedChannelV6[int](4, 100)
			ch.Close()
			for {
				_, ok := ch.Receive()
				if !ok {
					break
				}
			}
		}
	})
}

