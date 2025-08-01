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
// 创建日期:2025/7/18
package unbounded_channel

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Msg struct {
	ID   int64  // 唯一ID
	Data []byte // 模拟数据
}

type UnboundedChannel struct {
	Channel    chan *Msg   // fast路径
	List       list.List   // 慢路径
	HasSlowMsg atomic.Bool // 原子量标志
	mutex      sync.Mutex  // 锁
}

func NewUnboundedChannel() *UnboundedChannel {
	uc := &UnboundedChannel{
		Channel:    make(chan *Msg, 1000),
		List:       *list.New(),
		HasSlowMsg: atomic.Bool{},
		mutex:      sync.Mutex{},
	}
	uc.periodTryTranser()
	return uc
}

func (this *UnboundedChannel) periodTryTranser() {
	go func() {
		for {
			if this.HasSlowMsg.Load() {
				this.transfer()
			}
			// TODO 先简单实现为，睡眠7毫秒
			time.Sleep(7 * time.Millisecond)
		}
	}()
}

func (this *UnboundedChannel) transfer() {
	// TODO 考虑这个地方加锁的性能
	// 应该考虑直接将批量数据直接取出来再投递
	this.mutex.Lock()
	defer this.mutex.Unlock()
	// 批量投递，一次最多投递1000条
	for i := 0; i < 1000 && len(this.Channel) < cap(this.Channel); i++ {
		// 取消息
		ele := this.List.Front()
		// 移除消息
		msg := this.List.Remove(ele)
		// 投递到通道
		this.Channel <- msg.(*Msg)
	}
}

func (this *UnboundedChannel) tryTransfer() {
	if !this.HasSlowMsg.Load() {
		return
	}
	go func() {
		this.transfer()
	}()
}

func (this *UnboundedChannel) SendMsg(msg *Msg) {
	if msg == nil {
		return
	}
	if this.HasSlowMsg.Load() {
		// 直接进入缓冲列表
		this.mutex.Lock()
		this.List.PushBack(msg)
		this.mutex.Unlock()

		this.tryTransfer()
		return
	}

	this.mutex.Lock()
	// 二次检查
	// 有可能在获取原子值和加锁操作之间，Flag发生了改变
	// 有可能其他gourinte设置了这个值
	if this.HasSlowMsg.Load() {
		// 直接进入缓冲列表
		this.List.PushBack(msg)
		// 解锁路径1
		this.mutex.Unlock()

		this.tryTransfer()
		return
	}

	// HasSlowMsg==false,才会走到这里
	select {
	case this.Channel <- msg:
		// 投递到快速通道成功
		// 改变标志
		this.HasSlowMsg.Store(false)
		// 解锁路径2
		this.mutex.Unlock()
		fmt.Printf("send msg to channel %d\n", msg.ID)
	default:
		// 通道满了，说明投递失败，则进入慢路径
		// 直接进入缓冲列表
		this.List.PushBack(msg)
		this.HasSlowMsg.Store(true)
		// 解锁路劲3
		this.mutex.Unlock()

		this.tryTransfer()
	}
}

func Consume(uc *UnboundedChannel) {
	for msg := range uc.Channel {
		// 使用消息的数据
		fmt.Printf("msg data:%+v\n", msg.Data)
	}
}