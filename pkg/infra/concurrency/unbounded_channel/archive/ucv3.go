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
// 创建日期:2025/6/3
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
// 在v2版本的基础上提升性能:
// 1.环形队列替换为切片；
// 2.checklist不再固定时间等待，而是等待信号唤醒
type UnboundedChannelV3 struct {
	mutex     sync.Mutex
	buffer    queue.IQueue[int]
	bufferLen atomic.Int32
	closed    atomic.Bool
	done      chan struct{}
	channel   chan int // 通道
	cond      *sync.Cond
	// 测试用字段
	seq int
}

func NewUnboundedChannelV3() *UnboundedChannelV3 {
	uc := &UnboundedChannelV3{
		buffer:    queue.NewQueueWithSize[int](100),
		bufferLen: atomic.Int32{},
		channel:   make(chan int, 10),
		done:      make(chan struct{}),
		closed:    atomic.Bool{},
	}
	uc.cond = sync.NewCond(&uc.mutex)
	uc.listCheck()
	return uc
}

// 往通道发消息(保证消息的先后顺序)
func (uc *UnboundedChannelV3) Send(msg int) {
	assert.Assert(uc.channel != nil)
	assert.Assert(uc.buffer != nil)
	if uc.closed.Load() {
		panic("通道已关闭，不能再投递数据")
	}
	
	uc.mutex.Lock()
	// TODO 测试代码 将msg强行设置，保证入队的数值按单调递增，便于验证FIFO性质 start
	uc.seq++
	msg = uc.seq
	// TODO 测试代码 将msg强行设置，保证入队的数值按单调递增，便于验证FIFO性质 end
	if len(uc.channel) == cap(uc.channel) {
		// 通道满了，先暂存list
		uc.buffer.Enqueue(msg)
		uc.mutex.Unlock()
		// 发送信号 有新消息到缓存列表
		uc.cond.Signal()
		return
	}
	// 现在channel未满，但是不能直接将msg投递到channel. v2版本要保证消息的先后顺序,需要先检查list中是否有消息.
	// 如果list里面还有消息，则需要优先从list取消息投递到channel. 那么当前msg呢？则必须放入list
	if uc.buffer.Len() != 0 {
		// 列表非空,当前消息进入列表
		uc.buffer.Enqueue(msg)
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

func (uc *UnboundedChannelV3) channelSend(msg int) {
	// 复制消息到通道
	select {
	case uc.channel <- msg:
		// fmt.Printf("发送消息:%d到通道\n", msg)
	default:
		panic(fmt.Sprintf("丢弃消息:%d\n", msg))
	}
}

func (uc *UnboundedChannelV3) transfer() {
	// 尝试数据转移
	for len(uc.channel) < cap(uc.channel) && uc.buffer.Len() != 0 {
		val := uc.buffer.Dequeue()
		uc.channelSend(val)
	}
}

func (uc *UnboundedChannelV3) listCheck() {
	go func() {
		for {
			uc.mutex.Lock()
			for uc.buffer.Len() == 0 || len(uc.channel) == cap(uc.channel) {
				// 阻塞等待信号
				uc.cond.Wait()
			}
			// 转移数据
			uc.transfer()
			// 如果关闭标志已设置，现在需要检查列表中的元素数量;当channel和列表中都没有元素后，才通知.
			if uc.closed.Load() {
				if uc.buffer.Len() == 0 && len(uc.channel) == 0 {
					// 通知真正关闭
					close(uc.done)
					close(uc.channel)
					uc.mutex.Unlock()
					return
				}
			}
			uc.mutex.Unlock()
		}
	}()
}

// 从通道接收消息
func (uc *UnboundedChannelV3) Receive() int {
	// 不用再加锁，直接从channel里取数据
	// 因为这里没有竞争
	var res int
	select {
	case res = <-uc.channel:
		// 收到通道的消息
	case <-uc.done:
		return 0
	}
	// 发信号，通道有空位了
	uc.cond.Signal()
	return res
}

func (uc *UnboundedChannelV3) Close() {
	uc.closed.Store(true)
}

func (uc *UnboundedChannelV3) PrintList() {
	// uc.mutex.Lock()
	// uc.buffer.Range(func(i int) {
	//     fmt.Printf("%d\t", i)
	// })
	// fmt.Println()
	// uc.mutex.Unlock()
}