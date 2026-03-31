// Package unbounded_channel.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2025/6/12
package unbounded_channel

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/cpu"
	"github.com/motocat46/yytools/pkg/ds/queue"
)

// UnboundedChannelV6 在V5基础上改进后台worker的触发机制。
//
// V5的listCheck采用"自旋(10次Gosched+mutex)+指数退避sleep"：
//   - 空闲时存在无效的mutex争用（自旋阶段）
//   - 退避积累时存在不必要的搬运延迟
//
// V6改为事件驱动+固定ticker兜底：
//  1. 消费者调用Receive()后，若buffer非空则主动发信号
//  2. 关闭时Close()发信号，确保worker及时感知
//  3. 1ms固定ticker兜底，覆盖直接使用Out()的场景
//
// worker空闲时阻塞在select，零CPU消耗；有事件时精准唤醒。
type UnboundedChannelV6[T any] struct {
	// ── 快速路径热字段：独占 cache line ──────────────────────────────────
	// bufferLen 在每次 Send() 的快速路径中都会被读取。
	// 慢路径加锁时 mutex.state 会 invalidate 整个 cache line，
	// 用 CacheLinePad 将其隔离，避免伪共享（false sharing）。
	bufferLen atomic.Int32
	_         cpu.CacheLinePad

	// ── 其余字段（慢路径 + 控制字段）────────────────────────────────────
	// activeSenders 记录当前正在执行 Send() 的 goroutine 数量。
	// 语义上等价于 RWMutex 的读计数：Send() 持有"读锁"，worker 关闭通道前必须等待计数归零。
	// 保证：channel 被关闭时 activeSenders 必然为 0；activeSenders > 0 时 channel 必然未关闭。
	// 每次 Send() 调用都会写两次（Add(1)+defer Add(-1)），高并发下写操作频繁；
	// 用 CacheLinePad 隔离，防止与 mutex 产生 false sharing。
	activeSenders  atomic.Int32
	_              cpu.CacheLinePad
	mutex          sync.Mutex
	buffer         queue.IQueue[T]
	closed         atomic.Bool
	channel        chan T // 通道
	condSendWaiter *sync.Cond
	limit          int32         // 缓冲区数据数量上限
	notify         chan struct{} // 事件信号通道，cap=1，多余信号自动合并
	workerDone     chan struct{} // worker goroutine 完全退出后关闭
}

// NewUnboundedChannelV6 创建一个无界通道。
//
// 注意：内部 worker goroutine 在 Close() 被调用且所有消息消费完毕后才会退出。
// 若调用方不调用 Close()，worker goroutine 将永久存活，造成 goroutine 泄漏。
// 调用方应在使用完毕后确保调用 Close()，例如通过 defer uc.Close()。
func NewUnboundedChannelV6[T any](chanSize int, limit int) *UnboundedChannelV6[T] {
	assert.Assert(chanSize > 0, "chanSize should be > 0", chanSize)
	assert.Assert(limit > 0, "limit should be > 0", limit)
	uc := &UnboundedChannelV6[T]{
		buffer:     queue.NewQueueWithSize[T](32), // 设置为2的幂，方便go runtime的内存分配器分配大小合适的内存
		channel:    make(chan T, chanSize),
		notify:     make(chan struct{}, 1),
		limit:      int32(limit),
		workerDone: make(chan struct{}),
	}
	uc.condSendWaiter = sync.NewCond(&uc.mutex)
	uc.worker()
	return uc
}

// bufferEnqueue 将消息入队到 buffer，并更新 bufferLen 代理计数。
//
// 顺序约束：bufferLen.Add(1) 必须在 buffer.Enqueue() 之前执行。
// 原因：bufferLen 是快速路径的无锁代理，快速路径依赖 bufferLen==0 判断 buffer 为空。
// 若先 Enqueue 再 Add(1)，窗口期内 buffer 非空但 bufferLen==0，
// 快速路径会误判"buffer 为空"而绕过 buffer，破坏 FIFO 顺序。
// 先 Add(1) 保证 bufferLen 始终对 buffer.Len() 保持高估，快速路径只会误判为"非空"（走慢路径），
// 不会误判为"空"（绕过 buffer），代价是偶尔多走一次慢路径，但正确性得到保证。
//
// 必须在 mutex 保护下调用。
func (uc *UnboundedChannelV6[T]) bufferEnqueue(msg T) {
	uc.bufferLen.Add(1)    // 先标记：bufferLen 高估，快速路径立即看到"buffer 非空"
	uc.buffer.Enqueue(msg) // 再入队：物理写入
}

// bufferDequeue 将消息从 buffer 出队，并更新 bufferLen 代理计数。
//
// 顺序约束：buffer.Dequeue() 必须在 bufferLen.Add(-1) 之前执行。
// 原因：先减 bufferLen 再 Dequeue，窗口期内 bufferLen==0 但 buffer 仍有数据，
// 快速路径误判为"buffer 为空"而绕过，破坏 FIFO 顺序（与 bufferEnqueue 的约束对称）。
// 先 Dequeue 保证 bufferLen 高估不变量在减计数前始终成立。
//
// 必须在 mutex 保护下调用。
func (uc *UnboundedChannelV6[T]) bufferDequeue() {
	uc.buffer.Dequeue()  // 先出队：物理移除
	uc.bufferLen.Add(-1) // 再标记：bufferLen 降低，快速路径才可能看到"buffer 为空"
}

// signal 非阻塞发信号，多余的信号自动合并，不会阻塞调用方。
// 当notify已有信号时走default分支，是一次atomic.Load（~0.3ns），几乎零开销。
func (uc *UnboundedChannelV6[T]) signal() {
	select {
	case uc.notify <- struct{}{}:
	default:
	}
}

// sendSlow 在加锁保护下处理慢路径投递，分两个阶段：
//  1. 路由决策：根据 channel 是否满、buffer 是否有积压，决定消息去向
//  2. 执行投递：bufferEnqueue / transfer / channelSend
//
// 注意：sendSlow 本身不发 signal。
// - channel 满时：依靠消费者 Receive() 触发信号，worker 继续搬运
// - buffer 非空时：inline transfer 内联搬运，channel 满后 Receive() 兜底
func (uc *UnboundedChannelV6[T]) sendSlow(msg T) {
	uc.mutex.Lock()
	// 必须在发送到通道后解锁，因为发送数据到通道后，通道就可能满
	// 必须要保证上面的是否满的判断和添加元素的操作为原子性操作。
	defer uc.mutex.Unlock()

	// ── 阶段一：channel 是否已满 ──────────────────────────────────────
	if len(uc.channel) == cap(uc.channel) {
		// 通道满了，暂存 buffer，直接返回。
		// transfer() 此时无意义（channel 满无法搬运）。
		// 等消费者 Receive() 读走数据后会主动 signal，触发 worker 搬运；Out() 消费者有 1ms ticker 兜底。
		uc.bufferEnqueue(msg)
		return
	}

	// ── 阶段二：channel 有空位，检查 buffer 积压 ─────────────────────
	// 若 buffer 非空，当前 msg 必须排在 buffer 末尾，保证 FIFO 顺序。
	if uc.buffer.Len() != 0 {
		// buffer非空,当前消息进入buffer
		uc.bufferEnqueue(msg)
		// 内联搬运：生产者顺带推进buffer→channel，保证FIFO顺序
		// transfer 可能因 channel 中途被填满而未搬完 buffer，但此时 channel 是满的，
		// 发 signal 也无意义——worker 来了同样无法转移。
		// 等消费者读走数据时，Receive() 会在 bufferLen>0 时主动发 signal，触发 worker 继续搬运。
		// 对于使用 Out() 的消费者，1ms ticker 兜底。
		uc.transfer()
	} else {
		// buffer是空的，那么直接投递消息到通道
		uc.channelSend(msg)
	}
}

// senderShouldWait 判断当前 Send() 是否应因背压而阻塞。
//
// 在 Send() 中使用双重检查模式（double-check）：
//  1. 无锁预检（lock-free pre-check）：快速过滤大多数不需要等待的情况，避免无谓加锁
//  2. 加锁后 for 循环守卫（locked for-loop guard）：在 condSendWaiter.Wait() 前再次验证条件，
//     防止虚假唤醒（spurious wakeup）导致过早退出等待
//
// 之所以用 for 循环而非 if：sync.Cond.Wait() 可能因调度原因提前返回，
// 必须在 Wait() 返回后重新检查条件，否则会绕过背压保护。
func (uc *UnboundedChannelV6[T]) senderShouldWait() bool {
	return uc.bufferLen.Load() > uc.limit && !uc.closed.Load()
}

// Send 往通道发消息（保证FIFO顺序）。
// 返回 true 表示消息已投递；返回 false 表示通道已关闭，消息被丢弃。
//
// 并发安全关闭设计：
//   - activeSenders.Add(1) 是第一个操作，相当于持有"读锁"
//   - worker 关闭底层 channel 前必须等待 activeSenders 归零（canClose 条件）
//   - 这保证了：只要 Send() 持有活跃计数，channel 就不会被关闭，快速路径的 select 不会 panic
func (uc *UnboundedChannelV6[T]) Send(msg T) bool {
	// 第一步：注册活跃发送者（必须在 closed 检查之前）
	// 这是正确性的关键：先 Add(1) 再检查 closed，建立 happens-before 链。
	// 若 worker 已读到 activeSenders=0 并关闭了 channel，则 closed 此时必然为 true，
	// 后续的 closed.Load() 检查保证我们不会触碰已关闭的 channel。
	uc.activeSenders.Add(1)
	defer uc.activeSenders.Add(-1)

	// 第二步：检查关闭状态（必须在 activeSenders.Add(1) 之后）
	if uc.closed.Load() {
		return false
	}

	// 第三步：背压检测 buffer 超过理论上允许的上限限则阻塞，避免内存无限增长
	if uc.senderShouldWait() {
		uc.mutex.Lock()
		for uc.senderShouldWait() {
			uc.condSendWaiter.Wait()
		}
		uc.mutex.Unlock()
		// 被唤醒后再次检查 closed（可能是被 Close() 广播唤醒）
		if uc.closed.Load() {
			return false
		}
	}

	// 最后投递数据
	// 快速路径.负载较低时，缓冲区大概率没数据
	// 1.缓冲区无数据；且
	// 2.通道未满(当前时刻状态判断，只是一个snapshot)
	// 此时 activeSenders>0，canClose() 必然返回 false，channel 不会被关闭，select 安全。
	if uc.bufferLen.Load() == 0 && len(uc.channel) < cap(uc.channel) {
		select {
		case uc.channel <- msg:
			return true
		default:
			// 投递失败：与其他 goroutine 竞争失败。channel 已满，进入慢路径。
		}
	}
	// 慢路径
	uc.sendSlow(msg)
	return true
}

// channelSend 尝试直接将 msg 投递到 channel；channel 满时改存 buffer。
// 必须在 mutex 保护下调用。
func (uc *UnboundedChannelV6[T]) channelSend(msg T) {
	// 投递消息到通道
	select {
	case uc.channel <- msg:
		// channel 内部计数由 runtime 维护，无需手动更新
	default:
		// 通道满了，先暂存buffer
		// 当然实际执行到这里时，也有极端情况:刚访问完通道，通道就立刻被消费者消费了数据，变成不满
		// 但也不影响，数据入队buffer即可
		uc.bufferEnqueue(msg)
		// 通道已满，触发了也没用，大概率无法迁移数据
		// uc.signal()
	}
}

// transfer 将 buffer 中的消息依次搬运到 channel，直到 channel 满或 buffer 空；
// buffer 排空后广播唤醒阻塞生产者，否则按搬运量比例保守唤醒。
// 必须在 mutex 保护下调用。
func (uc *UnboundedChannelV6[T]) transfer() {
	movedCount := 0
	for len(uc.channel) < cap(uc.channel) && uc.buffer.Len() != 0 {
		val := uc.buffer.Peek() // 不出队，先看
		select {
		case uc.channel <- val:
			// 进入通道成功才从缓冲队列中移除
			uc.bufferDequeue()
			movedCount++
		default:
			// channel 满了，停止，避免乱序。
			// select 内的 break 只退出 select，不退出 for。
			// 但 channel 满后 for 条件 len < cap 为 false，循环自然退出，无需显式 break。
		}
	}
	// movedCount 可能为 0：调用时 channel 看似有空位，但快速路径的并发 Send() 可能在
	// transfer() 执行前将 channel 填满，导致循环条件一开始就不成立。
	// 此时 percent=0，wakenCount=1，发出一次虚假唤醒，无害。

	// ── 死锁防护：buffer 已完全排空 ──────────────────────────────────────
	// 若 buffer 排空后仍有生产者在 condSendWaiter 上阻塞，此后不会再有任何唤醒源：
	//   - buffer.Len()==0 → Receive() 不发 signal → worker 不再调用 transfer()
	// 因此必须在此广播，确保所有等待中的生产者都能退出 Wait()，继续投递消息。
	if uc.buffer.Len() == 0 {
		uc.condSendWaiter.Broadcast()
		return
	}

	// buffer 未排空（channel 已满）：按搬运量占 limit 的百分比保守唤醒，避免雪崩效应。
	// 1% 对应唤醒 1 个阻塞的生产者；≥50% 时广播（所有生产者均可继续）。
	// 注意：当 limit 较大时（如 100万）,百分比通常远低于 1%，实际唤醒数为 1，Broadcast 分支仅对极小 limit 生效。
	// 这也是期望的表现，如果buffer累计的数据过多，唤醒过多的消费者goroutine也是没有意义的。
	// 就像路口本来就堵车，稍微缓解了一点；如果又放很多车进入路口，会加剧拥堵。
	percent := float32(movedCount) / float32(uc.limit) * 100
	// 50%是个经验值
	if percent > 50 {
		uc.condSendWaiter.Broadcast()
		return
	}

	// 按百分比计算需要唤醒的goroutine数量
	wakenCount := max(1, int(percent))
	for i := 0; i < wakenCount; i++ {
		uc.condSendWaiter.Signal()
	}
}

// canClose 判断是否满足关闭底层 channel 的四个条件：
// 关闭标志已设置、无活跃发送者、buffer 为空、channel 为空。
func (uc *UnboundedChannelV6[T]) canClose() bool {
	// 1.关闭标志已设置; 且
	// 2.没有正在执行的 Send()（activeSenders==0）; 且
	// 3.缓冲区无数据; 且
	// 4.chan中没有数据
	return uc.closed.Load() && uc.activeSenders.Load() == 0 && uc.buffer.Len() == 0 && len(uc.channel) == 0
}

// skipCheck 是 worker 的前置过滤，避免无意义的 mutex 加锁。
// 两次原子读之间存在窗口（non-atomic double-check），可能产生虚假跳过，这是有意为之：
// 虚假跳过的代价仅是"本次不检查"，下一次 signal 或 ticker 会补救，不影响正确性。
func (uc *UnboundedChannelV6[T]) skipCheck() bool {
	// 1.未关闭且2.缓冲区长度为0
	return !uc.closed.Load() && uc.bufferLen.Load() == 0
}

// worker 替换V5的listCheck。
//
// 核心差异：
//   - V5：自旋10次（每次加锁/解锁+Gosched）+ 指数退避sleep
//   - V6：阻塞在select等待信号，空闲时零CPU；ticker作为安全兜底
func (uc *UnboundedChannelV6[T]) worker() {
	go func() {
		defer close(uc.workerDone) // worker 完全退出时通知等待方
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-uc.notify:
				// 前置检查：buffer 为空且未关闭，inline transfer 已处理完，跳过 mutex
				if uc.skipCheck() {
					continue
				}
			case <-ticker.C:
				// 兜底检查；buffer为空且未关闭则跳过，避免无效加锁
				if uc.skipCheck() {
					continue
				}
			}

			uc.mutex.Lock()
			if uc.canClose() {
				// 真正关闭底层通道
				close(uc.channel)
				uc.mutex.Unlock()
				return
			}
			// 缓冲区长度大于0 且 chan未满，进行数据转移
			if uc.buffer.Len() > 0 && len(uc.channel) < cap(uc.channel) {
				// 转移数据
				uc.transfer()
			}
			uc.mutex.Unlock()
		}
	}()
}

// Receive 从通道接收消息。
// 不用再加锁，直接从channel里取数据，因为这里和消费者没有额外的竞争。
// 相比V5：读取后若buffer非空，主动发信号触发worker搬运（消费者侧触发）
func (uc *UnboundedChannelV6[T]) Receive() (T, bool) {
	res, ok := <-uc.channel
	if ok && uc.bufferLen.Load() > 0 {
		// 收到通道的消息，channel腾出了空位
		// 若buffer有数据，通知worker从buffer搬运，避免等待下次ticker
		uc.signal()
	}
	return res, ok
}

// Out 返回只读channel，供消费者在select中直接使用
func (uc *UnboundedChannelV6[T]) Out() <-chan T {
	return uc.channel
}

// Close 标记通道为关闭状态，唤醒所有阻塞的生产者并通知 worker 尽快感知。
// 幂等：多次调用安全；Close 返回后通道已标记关闭，但 worker goroutine 会在消息排空后才退出。
// 如需等待 worker 完全退出，请调用 WaitDone。
func (uc *UnboundedChannelV6[T]) Close() {
	uc.closed.Store(true)
	// 确保worker能及时感知关闭，不必等待下次ticker
	uc.signal()
	// 唤醒所有阻塞的生产者(调用是不需要持有对应的锁)
	// go源码注释原文：
	// Broadcast wakes all goroutines waiting on c.
	// It is allowed but not required for the caller to hold c.L
	// during the call.
	uc.condSendWaiter.Broadcast()
}

// WaitDone 阻塞直到内部 worker goroutine 完全退出。
// 通常在 Close() 之后调用，用于需要 goroutine 零泄漏验证的场景（如 goleak）。
func (uc *UnboundedChannelV6[T]) WaitDone() {
	<-uc.workerDone
}
