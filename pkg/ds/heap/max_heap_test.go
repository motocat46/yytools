// Package heap.

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

package heap

import (
	"fmt"
	"testing"
)

func TestNewMaxHeap(t *testing.T) {
	h := NewMaxHeap[int]()
	if h == nil {
		t.Fatal("NewMaxHeap() 返回了 nil")
	}
	if h.Length() != 0 {
		t.Errorf("新最大堆的长度应该是 0，实际是 %d", h.Length())
	}
}

func TestMaxHeap_PushItem(t *testing.T) {
	h := NewMaxHeap[int]()
	items := []*Item[int]{
		{Data: 1, Weight: 3},
		{Data: 2, Weight: 1},
		{Data: 3, Weight: 5},
		{Data: 4, Weight: 0},
	}

	for i, item := range items {
		h.PushItem(item)
		if h.Length() != i+1 {
			t.Errorf("推入后长度应该是 %d，实际是 %d", i+1, h.Length())
		}
	}
}

// TestMaxHeap_PopOrder 验证最大堆弹出顺序为降序（权重最大的先出）
func TestMaxHeap_PopOrder(t *testing.T) {
	h := NewMaxHeap[int]()
	// 乱序插入10个元素
	weights := []int{3, 7, 1, 9, 5, 2, 8, 4, 6, 0}
	for i, w := range weights {
		h.PushItem(&Item[int]{Data: i, Weight: w})
	}

	prev := h.PopItem().Weight
	for h.Length() > 0 {
		cur := h.PopItem().Weight
		if cur > prev {
			t.Errorf("最大堆顺序被破坏：当前权重 %d 大于前一个权重 %d", cur, prev)
		}
		prev = cur
	}
}

func TestMaxHeap_PeekItem(t *testing.T) {
	t.Run("Peek不改变长度", func(t *testing.T) {
		h := NewMaxHeap[int]()
		h.PushItem(&Item[int]{Data: 1, Weight: 10})
		h.PushItem(&Item[int]{Data: 2, Weight: 5})
		h.PushItem(&Item[int]{Data: 3, Weight: 20})

		lengthBefore := h.Length()
		peek := h.PeekItem()
		if h.Length() != lengthBefore {
			t.Errorf("Peek 不应改变长度，期望 %d，实际 %d", lengthBefore, h.Length())
		}
		// Peek 应当返回权重最大的元素
		if peek.Weight != 20 {
			t.Errorf("Peek 应返回权重最大的元素，期望 20，实际 %d", peek.Weight)
		}
	})

	t.Run("空堆Peek返回nil", func(t *testing.T) {
		h := NewMaxHeap[int]()
		if got := h.PeekItem(); got != nil {
			t.Errorf("空最大堆 PeekItem 应返回 nil，实际返回 %v", got)
		}
	})
}

func TestMaxHeap_PopItem_Empty(t *testing.T) {
	h := NewMaxHeap[int]()
	if got := h.PopItem(); got != nil {
		t.Errorf("空最大堆 PopItem 应返回 nil，实际返回 %v", got)
	}
}

func TestMaxHeap_WithStruct(t *testing.T) {
	type Task struct {
		Name     string
		Priority int
	}

	h := NewMaxHeap[Task]()
	tasks := []*Item[Task]{
		{Data: Task{"低优先级任务", 1}, Weight: 1},
		{Data: Task{"高优先级任务", 10}, Weight: 10},
		{Data: Task{"中优先级任务", 5}, Weight: 5},
	}
	for _, task := range tasks {
		h.PushItem(task)
	}

	// 弹出顺序：按权重降序
	expectedWeights := []int{10, 5, 1}
	for i, expected := range expectedWeights {
		popped := h.PopItem()
		if popped.Weight != expected {
			t.Errorf("第 %d 次弹出：期望权重 %d，实际 %d", i+1, expected, popped.Weight)
		}
	}
}

func TestMaxHeap_LargeDataset(t *testing.T) {
	h := NewMaxHeap[int]()
	n := 1000
	for i := 0; i < n; i++ {
		h.PushItem(&Item[int]{Data: i, Weight: i})
	}

	if h.Length() != n {
		t.Errorf("期望长度 %d，实际 %d", n, h.Length())
	}

	// 验证每次弹出的权重 >= 下一次（降序）
	lastWeight := h.PopItem().Weight
	for h.Length() > 0 {
		cur := h.PopItem().Weight
		if cur > lastWeight {
			t.Errorf("最大堆顺序被破坏：当前权重 %d 大于前一个权重 %d", cur, lastWeight)
		}
		lastWeight = cur
	}
}

// BenchmarkMaxHeap_Push 最大堆 Push 基准，集合维持在 size 规模（Push+Pop 配对）。
func BenchmarkMaxHeap_Push(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewMaxHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			for b.Loop() {
				h.PushItem(&Item[int]{Data: i, Weight: i})
				h.PopItem()
				i++
			}
		})
	}
}

// BenchmarkMaxHeap_Pop 最大堆 Pop 基准，集合维持在 size 规模（Pop+Push 配对）。
func BenchmarkMaxHeap_Pop(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewMaxHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			for b.Loop() {
				h.PopItem()
				h.PushItem(&Item[int]{Data: i, Weight: i})
				i++
			}
		})
	}
}

// BenchmarkMaxHeap_Mixed 最大堆混合负载基准：70% Push + 30% Pop。
func BenchmarkMaxHeap_Mixed(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewMaxHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			op := 0
			for b.Loop() {
				if op%10 < 7 {
					h.PushItem(&Item[int]{Data: i, Weight: i})
					i++
				} else {
					if h.Length() > 0 {
						h.PopItem()
					}
				}
				op++
			}
		})
	}
}
