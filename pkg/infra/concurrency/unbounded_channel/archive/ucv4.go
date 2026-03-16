//go:build ignore

// Package main.

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
// 创建日期:2025/6/4
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/ds/queue"
)

// 实现
// 1.保证消息顺序的;
// 2.size无限制的;
// 通道。
// 在v3版本的基础上提升性能:
// 1.加入快速路径判断；
// 注意:这个基于信号量的版本还有死锁的风险。
type UnboundedChannelV4 struct {
	mutex       sync.Mutex
	buffer      queue.IQueue[any]
	bufferLen   atomic.Int32
	closed      atomic.Bool
	done        chan struct{}
	channel     chan any // 通道
	chanLen     atomic.Int32
	cond        *sync.Cond
	condNotFull *sync.Cond
	// mutexNotFull sync.Mutex
}

func NewUnboundedChannelV4(chanSize int) *UnboundedChannelV4 {
	assert.Assert(chanSize > 0, "chanSize should be > 0", chanSize)
	uc := &UnboundedChannelV4{
		buffer:    queue.NewQueueWithSize[any](100),
		bufferLen: atomic.Int32{},
		channel:   make(chan any, chanSize),
		chanLen:   atomic.Int32{},
		done:      make(chan struct{}),
		closed:    atomic.Bool{},
	}
	uc.cond = sync.NewCond(&uc.mutex)
	uc.condNotFull = sync.NewCond(&uc.mutex)
	uc.listCheck()
	return uc
}

func (uc *UnboundedChannelV4) sendSlow(msg any) {
	uc.mutex.Lock()
	// 在加锁保护下，直接判断通道是否已满
	if len(uc.channel) == cap(uc.channel) {
		// 通道满了，先暂存list
		uc.buffer.Enqueue(msg)
		uc.bufferLen.Add(1)
		// 发送信号 有新消息到缓存列表
		uc.cond.Signal()
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
func (uc *UnboundedChannelV4) Send(msg any) {
	assert.Assert(msg != nil)
	assert.Assert(!uc.closed.Load(), "通道已关闭，不能再投递数据,msg:", msg)
	// 上限判断
	if uc.bufferLen.Load() > 1e6 {
		uc.mutex.Lock()
		for uc.bufferLen.Load() > 1e6 {
			uc.condNotFull.Wait()
		}
		uc.mutex.Unlock()
	}
	
	// 快速路径判断 (负载较低时)缓冲区(大概率)没数据
	// 1.缓冲区无数据；且2.通道未满(当前时刻状态判断，只是一个snapshot)
	if uc.bufferLen.Load() == 0 &&
		int(uc.chanLen.Load()) < cap(uc.channel) {
		select {
		case uc.channel <- msg:
			// 假如某个goroutine执行到这被中断(chanLen就尚未更新)
			uc.chanLen.Add(1)
			// fmt.Printf("发送消息:%d到通道\n", msg)
			return
		default:
			// 投递失败.
			// 可能是和其他goroutine竞争失败，其他goroutine投递消息让channel满了.(本质上通道也会加锁)
			// 进入后续流程尝试将消息缓存到buffer.
		}
	}
	// 到这里也不能一定说channel就满了，因为还有消费者不停地取数据，所以必须进入慢路径进行检查和处理
	// 慢路径
	uc.sendSlow(msg)
}

func (uc *UnboundedChannelV4) channelSend(msg any) {
	// 复制消息到通道
	select {
	case uc.channel <- msg:
		uc.chanLen.Add(1)
		// fmt.Printf("发送消息:%d到通道\n", msg)
	default:
		// 通道满了，先暂存list
		uc.buffer.Enqueue(msg)
		uc.bufferLen.Add(1)
		// 发送信号 有新消息到缓存列表
		uc.cond.Signal()
	}
}

func (uc *UnboundedChannelV4) transfer() {
	movedCount := 0
	for len(uc.channel) < cap(uc.channel) && uc.buffer.Len() != 0 {
		val := uc.buffer.Peek() // 不出队，先看
		select {
		case uc.channel <- val:
			// 进入通道成功才从缓冲队列中移除
			uc.buffer.Dequeue()
			uc.bufferLen.Add(-1)
			uc.chanLen.Add(1)
			movedCount++
		default:
			// channel 满了，停止，避免乱序
			break
		}
	}
	
	// 这里极端情况下可能存在虚假唤醒，但是仍然在这里发信号(避免在循环里不停发信号)
	// 转移数据占缓冲区大小的百分比
	res := float32(movedCount) / 1e6 * 100
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

func (uc *UnboundedChannelV4) notifyClose() bool {
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

func (uc *UnboundedChannelV4) listCheck() {
	go func() {
		for {
			uc.mutex.Lock()
			if uc.notifyClose() {
				uc.mutex.Unlock()
				fmt.Printf("listCheck 退出...\n")
				return
			}
			// 1.缓冲队列为空；或者2.通道已满
			// 1或2任意满足一个条件就应该等待(否则唤醒也没什么意义)
			for uc.buffer.Len() == 0 || len(uc.channel) == cap(uc.channel) {
				if uc.notifyClose() {
					uc.mutex.Unlock()
					fmt.Printf("listCheck 退出...\n")
					return
				}
				// 阻塞等待信号
				uc.cond.Wait()
			}
			// 转移数据
			uc.transfer()
			uc.mutex.Unlock()
		}
	}()
}

// 从通道接收消息
func (uc *UnboundedChannelV4) Receive() any {
	// 不用再加锁，直接从channel里取数据
	// 因为这里和消费者没有额外的竞争
	select {
	case res := <-uc.channel:
		uc.chanLen.Add(-1)
		// 收到通道的消息
		// 发信号，通道有空位了
		uc.mutex.Lock()
		uc.cond.Signal()
		uc.mutex.Unlock()
		// uc.condNotFull.Signal()
		return res
	case <-uc.done:
		return nil
	}
}

func (uc *UnboundedChannelV4) Close() {
	uc.closed.Store(true)
}