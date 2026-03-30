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
package timingwheel

import (
	"sync"
	"sync/atomic"
)

// bucket 是时间轮的一个槽，双向循环链表（带哨兵节点 root）。
// 实现 delayqueue.Item 接口（ExpireAt），可直接入队 DelayQueue。
type bucket struct {
	// 量化到期时间（ms）：-1 表示空（未入 DelayQueue）
	// 通过 CAS(-1, bucketExp) 控制首次入队；Flush 后重置为 -1
	expireAt atomic.Int64
	root     Timer      // 哨兵节点，root.next = 链表头，root.prev = 链表尾
	mu       sync.Mutex // 保护链表操作（Cancel 绕过顶层锁直接摘链）
}

// newBucket 创建空 bucket，expireAt=-1，链表为空循环链表。
func newBucket() *bucket {
	b := &bucket{}
	b.expireAt.Store(-1)
	b.root.next = &b.root
	b.root.prev = &b.root
	return b
}

// ExpireAt 实现 delayqueue.Item 接口，返回量化到期时间。
func (b *bucket) ExpireAt() int64 {
	return b.expireAt.Load()
}

// Add 将 timer 插入链表末尾，设置 timer.bucket。
// 返回 true 表示这是加入空 bucket 的第一个 timer（调用方应 CAS 设置 expireAt 并 Offer 入队）。
func (b *bucket) Add(t *Timer) (isFirst bool) {
	b.mu.Lock()
	// 插入到 root.prev（末尾）
	t.prev = b.root.prev
	t.next = &b.root
	b.root.prev.next = t
	b.root.prev = t
	t.bucket.Store(b)
	isFirst = t.prev == &b.root // 插入前链表为空
	b.mu.Unlock()
	return
}

// detach 从链表摘除 t（调用方必须持有 b.mu）。
func (b *bucket) detach(t *Timer) {
	t.prev.next = t.next
	t.next.prev = t.prev
	t.prev = nil
	t.next = nil
}

// Flush 清空 bucket，对每个 timer 调用 fn（在 bucket.mu 外、writeLock 保护下调用）。
// 执行顺序：
//  1. 在 mu 内：清除每个 timer.bucket 指针并摘链（与 Cancel 互斥）
//  2. 在 mu 内：重置 expireAt=-1（与下一轮 CAS 互斥）
//  3. 在 mu 外：对每个 timer 调用 fn（fn 内可能触发 DelayQueue.Offer，避免持锁 O(log n) 操作）
func (b *bucket) Flush(fn func(*Timer)) {
	var timers []*Timer

	b.mu.Lock()
	for t := b.root.next; t != &b.root; t = t.next {
		t.bucket.Store(nil) // 清除 bucket 指针（与 Cancel 的 double-check 互斥）
	}
	for b.root.next != &b.root {
		t := b.root.next
		b.detach(t)
		timers = append(timers, t)
	}
	b.expireAt.Store(-1) // 重置，允许下一个 timer 重新入队
	b.mu.Unlock()

	for _, t := range timers {
		fn(t)
	}
}
