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
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue[string]()
	if pq == nil {
		t.Fatal("NewPriorityQueue() 返回了 nil")
	}
	if pq.Length() != 0 {
		t.Errorf("新优先级队列的长度应该是 0，实际是 %d", pq.Length())
	}
}

func TestPriorityQueue_PushPopOrder(t *testing.T) {
	pq := NewPriorityQueue[string]()
	items := []*PriorityItem[string]{
		{Data: "低", Priority: 1},
		{Data: "高", Priority: 10},
		{Data: "中", Priority: 5},
		{Data: "最高", Priority: 20},
	}
	for _, item := range items {
		pq.PushItem(item)
	}

	// 优先级越高的越先出（降序）
	expectedPriorities := []int{20, 10, 5, 1}
	for i, expected := range expectedPriorities {
		popped := pq.PopItem()
		if popped.Priority != expected {
			t.Errorf("第 %d 次弹出：期望优先级 %d，实际 %d", i+1, expected, popped.Priority)
		}
		if pq.Length() != len(items)-i-1 {
			t.Errorf("弹出后长度应该是 %d，实际是 %d", len(items)-i-1, pq.Length())
		}
	}
}

func TestPriorityQueue_PeekItem(t *testing.T) {
	t.Run("Peek返回最高优先级不弹出", func(t *testing.T) {
		pq := NewPriorityQueue[int]()
		pq.PushItem(&PriorityItem[int]{Data: 1, Priority: 5})
		pq.PushItem(&PriorityItem[int]{Data: 2, Priority: 15})
		pq.PushItem(&PriorityItem[int]{Data: 3, Priority: 10})

		lengthBefore := pq.Length()
		peek := pq.PeekItem()
		if pq.Length() != lengthBefore {
			t.Errorf("Peek 不应改变长度，期望 %d，实际 %d", lengthBefore, pq.Length())
		}
		if peek.Priority != 15 {
			t.Errorf("Peek 应返回优先级最高的元素，期望 15，实际 %d", peek.Priority)
		}
	})
}

func TestPriorityQueue_UpdatePriority(t *testing.T) {
	pq := NewPriorityQueue[string]()
	low := &PriorityItem[string]{Data: "低", Priority: 1}
	mid := &PriorityItem[string]{Data: "中", Priority: 5}
	high := &PriorityItem[string]{Data: "高", Priority: 10}

	pq.PushItem(low)
	pq.PushItem(mid)
	pq.PushItem(high)

	// 将"低"优先级提升到最高
	pq.UpdatePriority(low, 100)

	// 现在"低"（改名后是优先级100）应该最先弹出
	first := pq.PopItem()
	if first.Data != "低" || first.Priority != 100 {
		t.Errorf("UpdatePriority 后期望先弹出 '低'(100)，实际弹出 '%s'(%d)", first.Data, first.Priority)
	}

	// 再将"中"降低到最低
	pq.UpdatePriority(mid, 0)
	second := pq.PopItem()
	if second.Data != "高" {
		t.Errorf("期望第二个弹出 '高'，实际弹出 '%s'", second.Data)
	}
	third := pq.PopItem()
	if third.Data != "中" || third.Priority != 0 {
		t.Errorf("期望最后弹出 '中'(0)，实际弹出 '%s'(%d)", third.Data, third.Priority)
	}
}

func TestPriorityQueue_PopItem_Empty(t *testing.T) {
	pq := NewPriorityQueue[int]()
	if got := pq.PopItem(); got != nil {
		t.Errorf("空优先级队列 PopItem 应返回 nil，实际返回 %v", got)
	}
}

func TestPriorityQueue_PeekItem_Empty(t *testing.T) {
	pq := NewPriorityQueue[int]()
	if got := pq.PeekItem(); got != nil {
		t.Errorf("空优先级队列 PeekItem 应返回 nil，实际返回 %v", got)
	}
}

func TestPriorityQueue_SamePriority(t *testing.T) {
	pq := NewPriorityQueue[int]()
	// 推入相同优先级的多个元素
	for i := 0; i < 5; i++ {
		pq.PushItem(&PriorityItem[int]{Data: i, Priority: 10})
	}
	if pq.Length() != 5 {
		t.Errorf("期望长度 5，实际 %d", pq.Length())
	}
	// 弹出时所有优先级都为10
	for pq.Length() > 0 {
		popped := pq.PopItem()
		if popped.Priority != 10 {
			t.Errorf("期望优先级 10，实际 %d", popped.Priority)
		}
	}
}

func TestPriorityQueue_IndexTracking(t *testing.T) {
	// 验证 Index 字段在 Push/Pop 后被正确维护
	pq := NewPriorityQueue[int]()
	items := make([]*PriorityItem[int], 5)
	for i := 0; i < 5; i++ {
		items[i] = &PriorityItem[int]{Data: i, Priority: i * 2}
		pq.PushItem(items[i])
	}
	// 弹出后，Index 应为 -1
	popped := pq.PopItem()
	if popped.Index != -1 {
		t.Errorf("弹出的元素 Index 应为 -1，实际 %d", popped.Index)
	}
}

func BenchmarkPriorityQueuePush(b *testing.B) {
	pq := NewPriorityQueue[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.PushItem(&PriorityItem[int]{Data: i, Priority: i % 100})
	}
}

func BenchmarkPriorityQueuePop(b *testing.B) {
	pq := NewPriorityQueue[int]()
	for i := 0; i < b.N; i++ {
		pq.PushItem(&PriorityItem[int]{Data: i, Priority: i % 100})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if pq.Length() > 0 {
			pq.PopItem()
		}
	}
}
