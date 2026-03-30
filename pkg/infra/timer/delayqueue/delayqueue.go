// Package delayqueue 提供并发安全的延迟队列，元素按到期时间排序。
//
// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// 作者: yangyuan
// 创建日期: 2026-03-28
package delayqueue

import (
	"container/heap"
	"context"
	"math"
	"sync"
	"time"
)

// Item 是 DelayQueue 元素必须实现的接口。
// ExpireAt 返回到期时间（ms，与 clock 函数同一时间域）。
type Item interface {
	ExpireAt() int64
}

// DelayQueue 是并发安全的延迟队列，基于 min-heap，元素按到期时间升序出队。
// clock 函数提供当前时间（ms），与元素的 ExpireAt() 同一时间域。
type DelayQueue[T Item] struct {
	mu    sync.Mutex
	items []T
	clock func() int64
	wakeC chan struct{} // 缓冲为 1，Offer 加入更早元素时通知 Poll 重新计算等待时间
}

// New 创建 DelayQueue。clock 返回当前时间（ms），需与 Item.ExpireAt() 同一时间域。
func New[T Item](clock func() int64) *DelayQueue[T] {
	return &DelayQueue[T]{
		clock: clock,
		wakeC: make(chan struct{}, 1),
	}
}

// --- heap.Interface 实现（mu 已持有时调用）---

func (q *DelayQueue[T]) Len() int           { return len(q.items) }
func (q *DelayQueue[T]) Less(i, j int) bool { return q.items[i].ExpireAt() < q.items[j].ExpireAt() }
func (q *DelayQueue[T]) Swap(i, j int)      { q.items[i], q.items[j] = q.items[j], q.items[i] }
func (q *DelayQueue[T]) Push(x any)         { q.items = append(q.items, x.(T)) }
func (q *DelayQueue[T]) Pop() any {
	n := len(q.items)
	x := q.items[n-1]
	var zero T
	q.items[n-1] = zero // 避免 GC 泄漏
	q.items = q.items[:n-1]
	return x
}

// Offer 将 item 加入队列。若 item 比当前最早元素更早到期，唤醒阻塞中的 Poll。
func (q *DelayQueue[T]) Offer(item T) {
	q.mu.Lock()
	prevMin := int64(math.MaxInt64)
	if len(q.items) > 0 {
		prevMin = q.items[0].ExpireAt()
	}
	heap.Push(q, item)
	isNewMin := item.ExpireAt() < prevMin
	q.mu.Unlock()

	if isNewMin {
		select {
		case q.wakeC <- struct{}{}:
		default:
		}
	}
}

// TryPoll 非阻塞尝试取出最早到期元素。队列为空或堆顶未到期时返回 (zero, false)。
func (q *DelayQueue[T]) TryPoll() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := q.clock()
	if len(q.items) > 0 && q.items[0].ExpireAt() <= now {
		return heap.Pop(q).(T), true
	}
	var zero T
	return zero, false
}

// Poll 阻塞直到有元素到期或 ctx 被取消。
// ctx 取消时返回 (zero, false)；取到元素时返回 (item, true)。
func (q *DelayQueue[T]) Poll(ctx context.Context) (T, bool) {
	for {
		q.mu.Lock()
		now := q.clock()
		if len(q.items) > 0 && q.items[0].ExpireAt() <= now {
			item := heap.Pop(q).(T)
			q.mu.Unlock()
			return item, true
		}
		// 计算到下一个元素到期的等待时间（队列为空时 sleepC==nil，永久阻塞于 select）。
		// 使用 time.NewTimer 而非 time.After：在 wakeC 或 ctx.Done() 先触发时显式 Stop()，
		// 避免 timer 在 d 到期前长期占用（高频循环时减少 GC 压力）。
		var sleepC <-chan time.Time
		var sleepTimer *time.Timer
		if len(q.items) > 0 {
			d := time.Duration(q.items[0].ExpireAt()-now) * time.Millisecond
			sleepTimer = time.NewTimer(d)
			sleepC = sleepTimer.C
		}
		q.mu.Unlock()

		select {
		case <-ctx.Done():
			if sleepTimer != nil {
				sleepTimer.Stop()
			}
			var zero T
			return zero, false
		case <-sleepC: // 到期，重新检查
		case <-q.wakeC: // 有更早的元素入队，重新计算
			if sleepTimer != nil {
				sleepTimer.Stop()
			}
		}
	}
}
