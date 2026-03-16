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
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/cpu"
	"github.com/motocat46/yytools/pkg/ds/queue"
)

// UnboundedChannelV53 在 V51 基础上去掉指数退避，改为固定 1ms sleep。
//
// listCheck 策略对比：
//   - V51：自旋10次 + 指数退避（1ms→2ms→...→100ms）
//   - V53：自旋10次 + 固定 1ms sleep（无退避）
//
// 去掉退避的目的：消除"channel 满导致搬运失败→退避积累→积压加剧"的问题。
// 代价：空闲时定时器唤醒频率固定在 1ms，比退避版略高，但 CPU 影响可忽略。
type UnboundedChannelV53 struct {
	// ── 快速路径热字段：独占 cache line ──────────────────────────────────
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

func NewUnboundedChannelV53(chanSize int, limit int) *UnboundedChannelV53 {
	assert.Assert(chanSize > 0, "chanSize should be > 0", chanSize)
	uc := &UnboundedChannelV53{
		buffer:  queue.NewQueueWithSize[any](100),
		channel: make(chan any, chanSize),
		done:    make(chan struct{}),
		limit:   int32(limit),
	}
	uc.condNotFull = sync.NewCond(&uc.mutex)
	uc.listCheck()
	return uc
}

func (uc *UnboundedChannelV53) sendSlow(msg any) {
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
func (uc *UnboundedChannelV53) Send(msg any) bool {
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
	if uc.bufferLen.Load() == 0 && len(uc.channel) < cap(uc.channel) {
		select {
		case uc.channel <- msg:
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

func (uc *UnboundedChannelV53) channelSend(msg any) {
	select {
	case uc.channel <- msg:
	default:
		// 通道满了，先暂存list
		uc.buffer.Enqueue(msg)
		uc.bufferLen.Add(1)
	}
}

func (uc *UnboundedChannelV53) transfer() {
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
		}
	}
	
	if uc.waiters.Load() == 0 {
		return
	}
	res := float32(movedCount) / float32(uc.limit) * 100
	if res < 50 {
		for i := 0; i < max(1, int(res)); i++ {
			uc.condNotFull.Signal()
		}
	} else {
		uc.condNotFull.Broadcast()
	}
}

func (uc *UnboundedChannelV53) notifyClose() bool {
	if uc.closed.Load() &&
		uc.buffer.Len() == 0 && len(uc.channel) == 0 {
		close(uc.done)
		close(uc.channel)
		return true
	}
	return false
}

// listCheck V53 策略：自旋 10 次（mutex + Gosched）+ 固定 1ms sleep（无退避）。
// 相比 V51：去掉指数退避，无论是否有数据、搬运是否成功，sleep 固定为 1ms。
// 避免"channel 满→退避积累→积压加剧"的问题，但空闲时唤醒频率固定（影响极小）。
func (uc *UnboundedChannelV53) listCheck() {
	go func() {
		for {
			uc.mutex.Lock()
			if uc.notifyClose() {
				uc.mutex.Unlock()
				return
			}
			uc.mutex.Unlock()
			
			flag := false
			// 自旋判断和等待：探测"buffer 非空且 channel 有空位"的条件是否成立
			for range 10 {
				uc.mutex.Lock()
				if uc.buffer.Len() != 0 && len(uc.channel) < cap(uc.channel) {
					flag = true
					uc.mutex.Unlock()
					break
				}
				uc.mutex.Unlock()
				runtime.Gosched()
			}
			
			if !flag {
				// 固定 1ms sleep，不退避
				time.Sleep(time.Millisecond)
				continue
			}
			
			uc.mutex.Lock()
			// 1.缓冲队列为空；或者2.通道已满
			// 1或2任意满足一个条件就应该等待(否则唤醒也没什么意义)
			if uc.buffer.Len() == 0 || len(uc.channel) == cap(uc.channel) {
				uc.mutex.Unlock()
				// 搬运条件不满足（channel 仍满），固定 1ms 后重试
				time.Sleep(time.Millisecond)
				continue
			}
			// 转移数据
			uc.transfer()
			uc.mutex.Unlock()
			// 搬运成功后不 sleep，立即进入下一轮：buffer 可能还有数据
		}
	}()
}

// 从通道接收消息
func (uc *UnboundedChannelV53) Receive() (any, bool) {
	select {
	case res := <-uc.channel:
		return res, true
	case <-uc.done:
		return nil, false
	}
}

func (uc *UnboundedChannelV53) Close() {
	uc.closed.Store(true)
}

func (uc *UnboundedChannelV53) Out() <-chan any {
	return uc.channel
}