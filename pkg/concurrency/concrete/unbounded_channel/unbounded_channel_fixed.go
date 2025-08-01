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
	Channel    chan *Msg     // fast路径
	List       list.List     // 慢路径
	HasSlowMsg atomic.Bool   // 原子量标志
	mutex      sync.Mutex    // 锁
	transferCh chan struct{} // 用于触发transfer的通道
	closed     atomic.Bool   // 关闭标志
}

func NewUnboundedChannel() *UnboundedChannel {
	uc := &UnboundedChannel{
		Channel:    make(chan *Msg, 1000),
		List:       *list.New(),
		HasSlowMsg: atomic.Bool{},
		mutex:      sync.Mutex{},
		closed:     atomic.Bool{},
		transferCh: make(chan struct{}, 1), // 缓冲为1，避免阻塞
	}
	uc.periodTryTranser()
	return uc
}

func (uc *UnboundedChannel) periodTryTranser() {
	go func() {
		ticker := time.NewTicker(7 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-uc.transferCh:
				// 收到transfer信号
				uc.transfer()
			case <-ticker.C:
				// 定时检查
				if uc.HasSlowMsg.Load() {
					uc.transfer()
				}
			}

			// 检查是否已关闭
			if uc.closed.Load() {
				return
			}
		}
	}()
}

func (uc *UnboundedChannel) transfer() {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	// 批量投递，一次最多投递1000条
	transferred := 0
	for i := 0; i < 1000 && len(uc.Channel) < cap(uc.Channel); i++ {
		// 取消息
		ele := uc.List.Front()
		if ele == nil {
			// List为空，重置标志
			uc.HasSlowMsg.Store(false)
			break
		}

		// 移除消息
		msg := uc.List.Remove(ele)
		// 投递到通道
		select {
		case uc.Channel <- msg.(*Msg):
			transferred++
		default:
			// 通道满了，把消息放回List头部
			uc.List.PushFront(msg)
			return
		}
	}

	// 如果List为空，重置标志
	if uc.List.Len() == 0 {
		uc.HasSlowMsg.Store(false)
	}
}

func (uc *UnboundedChannel) tryTransfer() {
	if !uc.HasSlowMsg.Load() {
		return
	}

	// 非阻塞发送transfer信号
	select {
	case uc.transferCh <- struct{}{}:
	default:
		// 通道已满，说明transfer正在进行中
	}
}

func (uc *UnboundedChannel) SendMsg(msg *Msg) error {
	if uc.closed.Load() {
		return fmt.Errorf("channel is closed")
	}

	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	if uc.HasSlowMsg.Load() {
		// 直接进入缓冲列表
		uc.mutex.Lock()
		uc.List.PushBack(msg)
		uc.mutex.Unlock()

		uc.tryTransfer()
		return nil
	}

	uc.mutex.Lock()
	// 二次检查
	if uc.HasSlowMsg.Load() {
		// 直接进入缓冲列表
		uc.List.PushBack(msg)
		uc.mutex.Unlock()
		uc.tryTransfer()
		return nil
	}

	// HasSlowMsg==false,才会走到这里
	select {
	case uc.Channel <- msg:
		// 投递到快速通道成功
		uc.mutex.Unlock()
		return nil
	default:
		// 通道满了，说明投递失败，则进入慢路径
		uc.List.PushBack(msg)
		uc.HasSlowMsg.Store(true)
		uc.mutex.Unlock()
		uc.tryTransfer()
		return nil
	}
}

func (uc *UnboundedChannel) Close() {
	if uc.closed.CompareAndSwap(false, true) {
		close(uc.Channel)
		close(uc.transferCh)
	}
}

func (uc *UnboundedChannel) IsClosed() bool {
	return uc.closed.Load()
}

func (uc *UnboundedChannel) Len() int {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()
	return uc.List.Len()
}

func (uc *UnboundedChannel) ChannelLen() int {
	return len(uc.Channel)
}

func Consume(uc *UnboundedChannel) {
	for msg := range uc.Channel {
		// 使用消息的数据
		fmt.Printf("Consumed msg ID: %d, data: %+v\n", msg.ID, msg.Data)
	}
}