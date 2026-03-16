//go:build ignore

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
// 创建日期:2025/6/11
package unbounded_channel

import (
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/cpu"
	"github.com/motocat46/yytools/pkg/ds/queue"
)

// 实现
// 1.保证消息顺序的;
// 2.size无限制的;
// 通道。
// 在v3版本的基础上提升性能:
// 1.加入快速路径判断；
type UnboundedChannelV5 struct {
	// ── 快速路径热字段：独占 cache line ──────────────────────────────────
	// bufferLen 在每次 Send() 的快速路径中都会被读取。
	// 慢路径加锁时 mutex.state 会 invalidate 整个 cache line，
	// 用 CacheLinePad 将其隔离，避免伪共享（false sharing）。
	bufferLen atomic.Int32
	_         cpu.CacheLinePad
	
	// ── 其余字段（慢路径 + 控制字段）────────────────────────────────────
	mutex       sync.Mutex
	buffer      queue.IQueue[any]
	closed      atomic.Bool
	done        chan struct{}
	channel     chan any // 通道
	condNotFull *sync.Cond
	limit       int32
	waiters     atomic.Int32 // 正在 condSendWaiter.Wait() 的生产者数量，用于跳过无等待者时的无效Signal
}

func NewUnboundedChannelV5(chanSize int, limit int) *UnboundedChannelV5 {
	assert.Assert(chanSize > 0, "chanSize should be > 0", chanSize)
	uc := &UnboundedChannelV5{
		buffer:  queue.NewQueueWithSize[any](100),
		channel: make(chan any, chanSize),
		done:    make(chan struct{}),
		limit:   int32(limit),
	}
	uc.condNotFull = sync.NewCond(&uc.mutex)
	uc.listCheck()
	return uc
}

func (uc *UnboundedChannelV5) sendSlow(msg any) {
	uc.mutex.Lock()
	// 在加锁保护下，直接判断通道是否已满
	if len(uc.channel) == cap(uc.channel) {
		// 通道满了，先暂存list
		uc.buffer.Enqueue(msg)
		uc.bufferLen.Add(1)
		// 发送信号 有新消息到缓存列表
		uc.mutex.Unlock()
		return
	}
	// 现在channel未满，但是不能直接将msg投递到channel. 要保证消息的先后顺序,需要先检查list中是否有消息.
	// 如果list里面还有消息，则需要优先从list取消息投递到channel. 那么当前msg呢？则必须放入list
	if uc.buffer.Len() != 0 {
		// 列表非空,当前消息进入列表
		uc.buffer.Enqueue(msg)
		uc.bufferLen.Add(1)
		// 将列表中的数据放入通道 FIFO顺序
		uc.transfer()
	} else {
		// 列表是空的，那么直接投递消息到通道
		uc.channelSend(msg)
	}
	// 必须在发送到通道后解锁，因为发送数据到通道后，通道就可能满
	// 必须要保证上面的是否满的判断和添加元素的操作为原子性操作。
	uc.mutex.Unlock()
}

// 往通道发消息(保证消息的先后顺序)
func (uc *UnboundedChannelV5) Send(msg any) bool {
	assert.Assert(msg != nil)
	assert.Assert(!uc.closed.Load(), "通道已关闭，不能再投递数据,msg:", msg)
	// 上限判断：buffer超过上限则阻塞等待，避免内存无限增长（兜底背压）
	if uc.bufferLen.Load() > uc.limit {
		uc.mutex.Lock()
		for uc.bufferLen.Load() > uc.limit {
			uc.waiters.Add(1)
			uc.condNotFull.Wait()
			uc.waiters.Add(-1)
		}
		uc.mutex.Unlock()
	}
	
	// 快速路径判断 (负载较低时)缓冲区(大概率)没数据
	// 1.缓冲区无数据；且2.通道未满(当前时刻状态判断，只是一个snapshot)
	// len(uc.channel) 直接读取 Go runtime 维护的 channel 内部计数，与原 chanLen 精度相同，
	// 但无需额外的 atomic.Add，消除了高并发下最热的原子写争用。
	if uc.bufferLen.Load() == 0 && len(uc.channel) < cap(uc.channel) {
		select {
		case uc.channel <- msg:
			// 假如某个goroutine执行到这被中断，channel 内部计数由 runtime 维护，无需手动更新
			return true
		default:
			// 投递失败.
			// 可能是和其他goroutine竞争失败，其他goroutine投递消息让channel满了.(本质上通道也会加锁)
			// 进入后续流程尝试将消息缓存到buffer.
		}
	}
	// 到这里也不能一定说channel就满了，因为还有消费者不停地取数据，所以必须进入慢路径进行检查和处理
	// 慢路径
	uc.sendSlow(msg)
	return true
}

func (uc *UnboundedChannelV5) channelSend(msg any) {
	// 复制消息到通道
	select {
	case uc.channel <- msg:
		// channel 内部计数由 runtime 维护，无需手动更新
	default:
		// 通道满了，先暂存list
		uc.buffer.Enqueue(msg)
		uc.bufferLen.Add(1)
		// 发送信号 有新消息到缓存列表
	}
}

func (uc *UnboundedChannelV5) transfer() {
	movedCount := 0
	for len(uc.channel) < cap(uc.channel) && uc.buffer.Len() != 0 {
		val := uc.buffer.Peek() // 不出队，先看
		select {
		case uc.channel <- val:
			// 进入通道成功才从缓冲队列中移除
			uc.buffer.Dequeue()
			uc.bufferLen.Add(-1)
			movedCount++
		default:
			// channel 满了，停止，避免乱序。
			// select 内的 break 只退出 select，不退出 for。
			// 但 channel 满后 for 条件 len < cap 为 false，循环自然退出，无需显式 break。
		}
	}
	
	// 这里极端情况下可能存在虚假唤醒，但是仍然在这里发信号(避免在循环里不停发信号)
	// 只有确实有goroutine在等待时才发信号，避免无意义的Signal调用开销（Signal内部也需要加锁）
	if uc.waiters.Load() == 0 {
		return
	}
	// 转移数据占缓冲区大小的百分比
	res := float32(movedCount) / float32(uc.limit) * 100
	if res < 50 {
		// 空闲占比小于50%，则根据百分比唤醒对应数量的goroutine
		for i := 0; i < max(1, int(res)); i++ {
			uc.condNotFull.Signal()
		}
	} else {
		// 超过一半的缓冲区空闲，则直接进行广播
		uc.condNotFull.Broadcast()
	}
}

func (uc *UnboundedChannelV5) notifyClose() bool {
	// 如果关闭标志已设置，现在需要检查列表中的元素数量;当channel和列表中都没有元素后，才通知.
	if uc.closed.Load() &&
		uc.buffer.Len() == 0 && len(uc.channel) == 0 {
		// 通知真正关闭
		close(uc.done)
		close(uc.channel)
		return true
	}
	return false
}

func (uc *UnboundedChannelV5) listCheck() {
	go func() {
		// 初始退避 1ms，与 OS 定时器精度一致（sleep 更短无实际意义）
		backoff := time.Millisecond
		for {
			// 原子预检：buffer 为空时直接退避，完全跳过 mutex，避免干扰快速路径
			// bufferLen 独占 cache line，此读取代价极低（~0.3ns）
			if uc.bufferLen.Load() == 0 {
				if uc.closed.Load() {
					uc.mutex.Lock()
					uc.notifyClose()
					uc.mutex.Unlock()
					return
				}
				time.Sleep(backoff)
				// 指数退避：空闲时逐步降低检查频率，减少不必要的定时器唤醒
				backoff = min(backoff*2, 100*time.Millisecond)
				continue
			}
			
			// buffer 非空，加锁尝试搬运
			uc.mutex.Lock()
			if uc.notifyClose() {
				uc.mutex.Unlock()
				return
			}
			// 1.缓冲队列为空；或者2.通道已满
			// 1或2任意满足一个条件就应该等待(否则唤醒也没什么意义)
			if uc.buffer.Len() > 0 && len(uc.channel) < cap(uc.channel) {
				// 转移数据
				uc.transfer()
				uc.mutex.Unlock()
				// 搬运后立即重置退避并进入下一轮：buffer 可能还有数据
				backoff = time.Millisecond
				continue
			}
			uc.mutex.Unlock()
			
			// channel 满或 buffer 已空（被其他路径消费），退避等待
			time.Sleep(backoff)
			backoff = min(backoff*2, 100*time.Millisecond)
		}
	}()
}

// 从通道接收消息
func (uc *UnboundedChannelV5) Receive() (any, bool) {
	// 不用再加锁，直接从channel里取数据
	// 因为这里和消费者没有额外的竞争
	select {
	case res := <-uc.channel:
		// 收到通道的消息，channel 内部计数由 runtime 自动维护，无需手动更新
		return res, true
	case <-uc.done:
		return nil, false
	}
}

func (uc *UnboundedChannelV5) Close() {
	uc.closed.Store(true)
}

func (uc *UnboundedChannelV5) Out() <-chan any {
	return uc.channel
}