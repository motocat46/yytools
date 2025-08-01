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
	closed     atomic.Bool   // 关闭标志
	transferCh chan struct{} // 用于触发transfer的通道

	// 自适应参数
	lastTransferTime    atomic.Int64 // 上次transfer时间戳
	transferCount       atomic.Int64 // transfer次数统计
	avgTransferInterval atomic.Int64 // 平均transfer间隔
}

const (
	// 基础检查间隔
	BaseCheckInterval = 1 * time.Millisecond

	// 最大检查间隔
	MaxCheckInterval = 50 * time.Millisecond

	// 最小检查间隔
	MinCheckInterval = 100 * time.Microsecond

	// 自适应权重
	AdaptiveWeight = 0.8 // 新值的权重

	// 批量转移阈值
	BatchTransferThreshold = 10 // 当List长度超过此值时，减少检查间隔
)

func NewUnboundedChannel() *UnboundedChannel {
	uc := &UnboundedChannel{
		Channel:    make(chan *Msg, 1000),
		List:       *list.New(),
		HasSlowMsg: atomic.Bool{},
		mutex:      sync.Mutex{},
		closed:     atomic.Bool{},
		transferCh: make(chan struct{}, 1),
	}
	uc.periodTryTranser()
	return uc
}

func (uc *UnboundedChannel) periodTryTranser() {
	go func() {
		// 初始检查间隔
		checkInterval := BaseCheckInterval
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-uc.transferCh:
				// 收到transfer信号，立即处理
				uc.transfer()
				// 重置为最小间隔，快速响应
				checkInterval = MinCheckInterval
				ticker.Reset(checkInterval)

			case <-ticker.C:
				// 定时检查
				if uc.HasSlowMsg.Load() {
					uc.transfer()
					// 根据List长度动态调整间隔
					checkInterval = uc.calculateCheckInterval()
					ticker.Reset(checkInterval)
				} else {
					// 没有慢路径消息，逐渐增加间隔
					checkInterval = uc.increaseCheckInterval(checkInterval)
					ticker.Reset(checkInterval)
				}
			}

			// 检查是否已关闭
			if uc.closed.Load() {
				return
			}
		}
	}()
}

// calculateCheckInterval 根据当前状态计算检查间隔
func (uc *UnboundedChannel) calculateCheckInterval() time.Duration {
	uc.mutex.Lock()
	listLen := uc.List.Len()
	uc.mutex.Unlock()

	// 根据List长度动态调整
	if listLen > BatchTransferThreshold {
		// List较长，需要快速处理
		return MinCheckInterval
	} else if listLen > 0 {
		// List有数据但不多，使用基础间隔
		return BaseCheckInterval
	} else {
		// List为空，使用最大间隔
		return MaxCheckInterval
	}
}

// increaseCheckInterval 当没有慢路径消息时，逐渐增加检查间隔
func (uc *UnboundedChannel) increaseCheckInterval(current time.Duration) time.Duration {
	// 指数退避，但不超过最大间隔
	newInterval := current * 2
	if newInterval > MaxCheckInterval {
		newInterval = MaxCheckInterval
	}
	return newInterval
}

func (uc *UnboundedChannel) transfer() {
	startTime := time.Now()

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
			goto transferComplete
		}
	}

transferComplete:
	// 如果List为空，重置标志
	if uc.List.Len() == 0 {
		uc.HasSlowMsg.Store(false)
	}

	// 更新统计信息
	if transferred > 0 {
		uc.updateTransferStats(startTime)
	}
}

// updateTransferStats 更新transfer统计信息
func (uc *UnboundedChannel) updateTransferStats(startTime time.Time) {
	duration := time.Since(startTime).Microseconds()

	// 更新transfer次数
	uc.transferCount.Add(1)

	// 更新平均间隔（使用指数移动平均）
	oldAvg := uc.avgTransferInterval.Load()
	if oldAvg == 0 {
		uc.avgTransferInterval.Store(duration)
	} else {
		newAvg := int64(float64(oldAvg)*(1-AdaptiveWeight) + float64(duration)*AdaptiveWeight)
		uc.avgTransferInterval.Store(newAvg)
	}

	// 更新最后transfer时间
	uc.lastTransferTime.Store(time.Now().UnixNano())
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

// GetStats 获取统计信息
func (uc *UnboundedChannel) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"list_len":                 uc.Len(),
		"channel_len":              uc.ChannelLen(),
		"transfer_count":           uc.transferCount.Load(),
		"avg_transfer_interval_us": uc.avgTransferInterval.Load(),
		"last_transfer_time":       uc.lastTransferTime.Load(),
		"has_slow_msg":             uc.HasSlowMsg.Load(),
		"is_closed":                uc.IsClosed(),
	}
}

func Consume(uc *UnboundedChannel) {
	for msg := range uc.Channel {
		// 使用消息的数据
		fmt.Printf("Consumed msg ID: %d, data: %+v\n", msg.ID, msg.Data)
	}
}
