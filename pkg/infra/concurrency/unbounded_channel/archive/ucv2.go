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
// 创建日期:2025/6/2
package main

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/motocat46/yytools/pkg/common/assert"
)

// Target 实现一个简单的
// 1.保证消息顺序的;
// 2.size无限制的;
// 通道。
type UnboundedChannelV2 struct {
	mutex sync.Mutex
	list  *list.List // buffer
	// listLen atomic.Int32
	closed  atomic.Bool
	done    chan struct{}
	channel chan int // 通道
	// 测试用字段
	seq int
}

func NewUnboundedChannelV2() *UnboundedChannelV2 {
	uc := &UnboundedChannelV2{
		list: list.New(),
		// listLen: atomic.Int32{},
		channel: make(chan int, 10),
		done:    make(chan struct{}),
		closed:  atomic.Bool{},
	}
	uc.listCheck()
	return uc
}

// 往通道发消息(保证消息的先后顺序)
func (uc *UnboundedChannelV2) Send(msg int) {
	assert.Assert(uc.channel != nil)
	assert.Assert(uc.list != nil)
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
		uc.list.PushFront(msg)
		uc.mutex.Unlock()
		return
	}
	// 现在channel未满，但是不能直接将msg投递到channel. v2版本要保证消息的先后顺序,需要先检查list中是否有消息.
	// 如果list里面还有消息，则需要优先从list取消息投递到channel. 那么当前msg呢？则必须放入list
	if uc.list.Len() != 0 {
		// 列表非空,当前消息进入列表
		uc.list.PushFront(msg)
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

func (uc *UnboundedChannelV2) channelSend(msg int) {
	// 复制消息到通道
	select {
	case uc.channel <- msg:
		// fmt.Printf("发送消息:%d到通道\n", msg)
	default:
		panic(fmt.Sprintf("丢弃消息:%d\n", msg))
	}
}

func (uc *UnboundedChannelV2) transfer() {
	// 尝试数据转移
	for len(uc.channel) < cap(uc.channel) && uc.list.Len() != 0 {
		val := uc.list.Remove(uc.list.Back())
		uc.channelSend(val.(int))
	}
}

func (uc *UnboundedChannelV2) listCheck() {
	go func() {
		for {
			// 快速路径判断列表中是否有数据(go不推荐这样的写法，要用原子变量来当做快速路径判断依据)
			// 这里干脆先不加快速路径判断
			// if uc.list.Len() != 0 {
			uc.mutex.Lock()
			// 获取锁后再判断一次
			if uc.list.Len() != 0 {
				uc.transfer()
			}
			uc.mutex.Unlock()
			// }
			
			// 关闭标志已设置，现在需要检查列表中的元素数量;当channel和列表中都没有元素后，才通知.
			if uc.closed.Load() {
				uc.mutex.Lock()
				if uc.list.Len() == 0 && len(uc.channel) == 0 {
					// 通知真正关闭
					close(uc.done)
					close(uc.channel)
					uc.mutex.Unlock()
					return
				} else {
					uc.mutex.Unlock()
				}
			}
			
			// TODO 先简单睡眠10毫秒
			time.Sleep(10 * time.Millisecond)
		}
		
	}()
}

// 从通道接收消息
func (uc *UnboundedChannelV2) Receive() int {
	// 不用再加锁，直接从channel里取数据
	// 因为这里没有竞争
	select {
	case msg := <-uc.channel:
		return msg
	case <-uc.done:
		return 0
	}
}

func (uc *UnboundedChannelV2) Close() {
	uc.closed.Store(true)
}

func (uc *UnboundedChannelV2) PrintList() {
	uc.mutex.Lock()
	for e := uc.list.Front(); e != nil; e = e.Next() {
		fmt.Printf("%d\t", e.Value)
	}
	fmt.Println()
	uc.mutex.Unlock()
}