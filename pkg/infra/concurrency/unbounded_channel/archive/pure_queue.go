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
// 创建日期:2025/6/13
package unbounded_channel

import (
	"sync"
	
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/ds/queue"
)

// PureQueue 纯互斥锁+条件变量实现的动态有界FIFO队列。
//
// 与 V6 的语义对齐：
//   - 动态 buffer（无固定容量上限），Send 正常不阻塞
//   - buffer 超过 limit 时 Send 阻塞（背压，防止内存无限增长）
//   - Receive 为空时阻塞
//
// 与 V6 的核心差异：
//   - 无内部 channel，无法用于 select / range（不兼容 Go channel 语义）
//   - 无后台 goroutine：消费者直接从 buffer 出队，零搬运延迟
//   - 代价：失去 select 兼容性
//
// 对比目的：量化"channel 包装层 + 搬运 goroutine"在 V6 中的固定开销。
type PureQueue struct {
	mu       sync.Mutex
	buf      queue.IQueue[any]
	notEmpty *sync.Cond
	notFull  *sync.Cond
	limit    int
	closed   bool
}

func NewPureQueue(limit int) *PureQueue {
	assert.Assert(limit > 0, "limit should be > 0", limit)
	q := &PureQueue{
		buf:   queue.NewQueueWithSize[any](100),
		limit: limit,
	}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

// Send 入队。buffer 超过 limit 时阻塞（背压）。
func (q *PureQueue) Send(msg any) {
	assert.Assert(msg != nil)
	q.mu.Lock()
	for q.buf.Len() >= q.limit && !q.closed {
		q.notFull.Wait()
	}
	if !q.closed {
		q.buf.Enqueue(msg)
		q.notEmpty.Signal()
	}
	q.mu.Unlock()
}

// Receive 出队。buffer 为空时阻塞直到有消息或队列关闭。
// 队列关闭且已排空时返回 nil。
func (q *PureQueue) Receive() any {
	q.mu.Lock()
	for q.buf.Len() == 0 && !q.closed {
		q.notEmpty.Wait()
	}
	if q.buf.Len() == 0 {
		q.mu.Unlock()
		return nil
	}
	msg := q.buf.Dequeue()
	q.notFull.Signal()
	q.mu.Unlock()
	return msg
}

// Close 关闭队列，广播唤醒所有阻塞的 Send/Receive。
func (q *PureQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
	q.mu.Unlock()
}