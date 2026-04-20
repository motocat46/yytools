// Copyright [yangyuan]
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

// Package slidingwindow 实现固定容量的泛型滑动窗口，提供 O(1) 的 Sum/Avg/Max/Min 统计。
// Max/Min 通过单调双端队列实现 O(1) 均摊，优于朴素 O(N) 全窗口扫描。
// 非并发安全，并发访问由调用方负责加锁。
package slidingwindow

import (
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/base"
)

// Window 是固定容量的滑动窗口，维护增量统计量。
// 零值不可用，必须通过 New 创建；非并发安全。
type Window[T base.Number] struct {
	size  int
	buf   []T // 环形缓冲区
	wHead int // 下一个写入位置（物理索引）
	count int // 当前元素数
	sum   T   // 增量维护的总和
	seq   int // 下一个逻辑序号（已分配序号 = seq-1）

	// 单调递减队列（队首 = 当前最大值的逻辑序号）
	maxBuf  []int
	maxHead int
	maxTail int
	maxLen  int

	// 单调递增队列（队首 = 当前最小值的逻辑序号）
	minBuf  []int
	minHead int
	minTail int
	minLen  int
}

// New 创建容量为 size 的滑动窗口，size 须 > 0，否则 panic。
func New[T base.Number](size int) *Window[T] {
	assert.Assert(size > 0, "slidingwindow: size 须 > 0，实际值:", size)
	return &Window[T]{
		size:   size,
		buf:    make([]T, size),
		maxBuf: make([]int, size),
		minBuf: make([]int, size),
	}
}

// Len 返回当前窗口元素数量（0 ≤ Len ≤ size）。
func (w *Window[T]) Len() int { return w.count }

// Empty 报告窗口是否为空。
func (w *Window[T]) Empty() bool { return w.count == 0 }

// Full 报告窗口是否已满（Len() == size）。
func (w *Window[T]) Full() bool { return w.count == w.size }

// bufAt 返回逻辑序号 s 对应的 buf 元素值。
func (w *Window[T]) bufAt(s int) T {
	idx := ((w.wHead - (w.seq - s)) % w.size + w.size) % w.size
	return w.buf[idx]
}

// maxBack 返回 maxBuf 队尾的逻辑序号。
func (w *Window[T]) maxBack() int {
	return w.maxBuf[(w.maxTail-1+w.size)%w.size]
}

// minBack 返回 minBuf 队尾的逻辑序号。
func (w *Window[T]) minBack() int {
	return w.minBuf[(w.minTail-1+w.size)%w.size]
}

// Add 向窗口追加一个值；窗口满时自动淘汰最旧元素。O(1) 均摊。
func (w *Window[T]) Add(v T) {
	newSeq := w.seq
	w.seq++
	writePos := w.wHead
	w.wHead = (w.wHead + 1) % w.size

	if w.count == w.size {
		// 窗口已满：淘汰最旧元素（位于 writePos）
		evictedSeq := newSeq - w.size
		w.sum -= w.buf[writePos]
		if w.maxLen > 0 && w.maxBuf[w.maxHead] == evictedSeq {
			w.maxHead = (w.maxHead + 1) % w.size
			w.maxLen--
		}
		if w.minLen > 0 && w.minBuf[w.minHead] == evictedSeq {
			w.minHead = (w.minHead + 1) % w.size
			w.minLen--
		}
	} else {
		w.count++
	}

	w.buf[writePos] = v
	w.sum += v

	// 维护单调递减队列（Max）：弹出队尾所有 <= v 的序号
	for w.maxLen > 0 && w.bufAt(w.maxBack()) <= v {
		w.maxTail = (w.maxTail - 1 + w.size) % w.size
		w.maxLen--
	}
	w.maxBuf[w.maxTail] = newSeq
	w.maxTail = (w.maxTail + 1) % w.size
	w.maxLen++

	// 维护单调递增队列（Min）：弹出队尾所有 >= v 的序号
	for w.minLen > 0 && w.bufAt(w.minBack()) >= v {
		w.minTail = (w.minTail - 1 + w.size) % w.size
		w.minLen--
	}
	w.minBuf[w.minTail] = newSeq
	w.minTail = (w.minTail + 1) % w.size
	w.minLen++
}

// Sum 返回当前窗口内所有元素之和。O(1)。
func (w *Window[T]) Sum() T { return w.sum }

// Avg 返回当前窗口均值（float64）。O(1)。
// 窗口为空时 panic；调用前用 Empty() 检查。
func (w *Window[T]) Avg() float64 {
	assert.Assert(!w.Empty(), "slidingwindow: Avg() 在空窗口上调用")
	return float64(w.sum) / float64(w.count)
}

// Max 返回当前窗口最大值。O(1) 均摊（单调双端队列）。
// 窗口为空时 panic；调用前用 Empty() 检查。
func (w *Window[T]) Max() T {
	assert.Assert(!w.Empty(), "slidingwindow: Max() 在空窗口上调用")
	return w.bufAt(w.maxBuf[w.maxHead])
}

// Min 返回当前窗口最小值。O(1) 均摊（单调双端队列）。
// 窗口为空时 panic；调用前用 Empty() 检查。
func (w *Window[T]) Min() T {
	assert.Assert(!w.Empty(), "slidingwindow: Min() 在空窗口上调用")
	return w.bufAt(w.minBuf[w.minHead])
}
