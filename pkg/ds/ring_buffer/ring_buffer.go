// Package ring_buffer 提供固定容量的泛型环形缓冲区。
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
// 作者:  yangyuan
package ring_buffer

import (
	"github.com/motocat46/yytools/pkg/common/assert"
)

// RingBuffer 是固定容量的泛型环形缓冲区。
// 写满时覆盖最旧的元素，不扩容。非并发安全，并发访问由调用方负责。
// Dequeue/Peek 在缓冲区为空时 panic，调用前应先检查 Empty()。
//
// 内部不变量（任何公开方法执行前后均须成立）：
//
//	0 <= head < cap(items)
//	0 <= tail < cap(items)
//	0 <= length <= cap(items)
//	tail == (head + length) % cap(items)
type RingBuffer[T any] struct {
	items  []T
	head   int // 最旧元素的下标
	tail   int // 下一个写入位置的下标；满时 tail == head
	length int // 当前元素数量
}

// NewRingBuffer 创建容量为 capacity 的空环形缓冲区，capacity 必须 > 0，否则 panic。
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	assert.Assert(capacity > 0, "capacity 必须 > 0，实际值:", capacity)
	return &RingBuffer[T]{
		items: make([]T, capacity),
	}
}

// newIndex 返回 index 在容量为 n 的环形数组中前进一步后的下标。
func newIndex(index int, n int) int {
	return (index + 1) % n
}

// Enqueue 将 item 入队。
// 缓冲区未满时直接写入；已满时覆盖最旧的元素，Len() 和 Cap() 保持不变。
func (r *RingBuffer[T]) Enqueue(item T) {
	n := r.Cap()
	r.items[r.tail] = item
	r.tail = newIndex(r.tail, n)
	if r.length == n {
		// 已满：tail 已前进，head 同步前进，length 不变，维持不变量
		r.head = newIndex(r.head, n)
	} else {
		r.length++
	}
}

// Dequeue 返回并移除最旧的元素（队首）。
// 缓冲区为空时 panic；调用前应先检查 Empty()。
func (r *RingBuffer[T]) Dequeue() T {
	assert.Assert(!r.Empty(), "缓冲区为空，无法 Dequeue!")
	item := r.items[r.head]
	var zero T
	r.items[r.head] = zero // 避免内存泄漏
	r.head = newIndex(r.head, r.Cap())
	r.length--
	return item
}

// Peek 返回队首元素但不移除。
// 缓冲区为空时 panic；调用前应先检查 Empty()。
func (r *RingBuffer[T]) Peek() T {
	assert.Assert(!r.Empty(), "缓冲区为空，无法 Peek!")
	return r.items[r.head]
}

// Len 返回当前元素数量。
func (r *RingBuffer[T]) Len() int {
	return r.length
}

// Cap 返回缓冲区容量（固定值）。
func (r *RingBuffer[T]) Cap() int {
	return len(r.items)
}

// Empty 返回缓冲区是否为空。
func (r *RingBuffer[T]) Empty() bool {
	return r.length == 0
}

// Full 返回缓冲区是否已满。
func (r *RingBuffer[T]) Full() bool {
	return r.length == r.Cap()
}

// Range 从最旧到最新依次对每个元素调用 f，缓冲区为空时直接返回。
// f 返回 false 时立即停止遍历；返回 true 时继续。
// f 为 nil 时 panic。
func (r *RingBuffer[T]) Range(f func(T) bool) {
	assert.Assert(f != nil, "f 不能为 nil")
	if r.Empty() {
		return
	}
	if r.tail > r.head {
		// 无环绕：数据在 [head, tail)
		for i := r.head; i < r.tail; i++ {
			if !f(r.items[i]) {
				return
			}
		}
	} else {
		// tail <= head（含满时 tail==head）：先遍历 [head, cap)，再遍历 [0, tail)
		for i := r.head; i < r.Cap(); i++ {
			if !f(r.items[i]) {
				return
			}
		}
		for i := 0; i < r.tail; i++ {
			if !f(r.items[i]) {
				return
			}
		}
	}
}