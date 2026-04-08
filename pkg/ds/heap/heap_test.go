package heap

import (
	"fmt"
	"testing"
)

func TestNewHeap(t *testing.T) {
	heap := NewHeap[int]()
	if heap == nil {
		t.Fatal("NewHeap() 返回了 nil")
	}
	if heap.Length() != 0 {
		t.Errorf("新堆的长度应该是 0，实际是 %d", heap.Length())
	}
}

func TestHeap_PushItem(t *testing.T) {
	heap := NewHeap[int]()

	// 测试基本推入操作
	items := []*Item[int]{
		{Data: 1, Weight: 3},
		{Data: 2, Weight: 1},
		{Data: 3, Weight: 2},
		{Data: 4, Weight: 0},
	}

	for i, item := range items {
		heap.PushItem(item)
		if heap.Length() != i+1 {
			t.Errorf("推入后长度应该是 %d，实际是 %d", i+1, heap.Length())
		}
	}
}

func TestHeap_PopItem(t *testing.T) {
	heap := NewHeap[int]()
	items := []*Item[int]{
		{Data: 1, Weight: 3},
		{Data: 2, Weight: 1},
		{Data: 3, Weight: 2},
		{Data: 4, Weight: 0},
	}

	// 先推入所有元素
	for _, item := range items {
		heap.PushItem(item)
	}

	// 测试弹出操作（最小堆，权重最小的先出）
	expectedOrder := []int{4, 2, 3, 1} // 按权重排序
	for i, expectedData := range expectedOrder {
		popped := heap.PopItem()
		if popped.Data != expectedData {
			t.Errorf("期望弹出数据 %d，实际弹出 %d", expectedData, popped.Data)
		}
		if heap.Length() != len(items)-i-1 {
			t.Errorf("弹出后长度应该是 %d，实际是 %d", len(items)-i-1, heap.Length())
		}
	}
}

func TestHeap_PeekItem_Empty(t *testing.T) {
	h := NewHeap[string]()
	// 空堆返回 nil，不 panic
	if got := h.PeekItem(); got != nil {
		t.Errorf("空堆 PeekItem 应返回 nil，实际返回 %v", got)
	}
}

func TestHeap_PopItem_Empty(t *testing.T) {
	h := NewHeap[int]()
	// 空堆返回 nil，不 panic
	if got := h.PopItem(); got != nil {
		t.Errorf("空堆 PopItem 应返回 nil，实际返回 %v", got)
	}
}

func TestHeap_PeekItemWithItems(t *testing.T) {
	heap := NewHeap[int]()
	items := []*Item[int]{
		{Data: 1, Weight: 3},
		{Data: 2, Weight: 1},
		{Data: 3, Weight: 2},
	}

	minWeight := items[0].Weight
	for _, item := range items {
		heap.PushItem(item)
		if item.Weight < minWeight {
			minWeight = item.Weight
		}
		peek := heap.PeekItem()
		// 应该始终是已入堆元素中权重最小的元素
		if peek.Weight != minWeight {
			t.Errorf("期望最小权重 %d，实际是 %d", minWeight, peek.Weight)
		}
	}
}

func TestHeap_Length(t *testing.T) {
	heap := NewHeap[string]()

	if heap.Length() != 0 {
		t.Errorf("新堆长度应该是 0，实际是 %d", heap.Length())
	}

	// 推入元素
	for i := 1; i <= 5; i++ {
		heap.PushItem(&Item[string]{Data: "test", Weight: i})
		if heap.Length() != i {
			t.Errorf("推入 %d 个元素后长度应该是 %d，实际是 %d", i, i, heap.Length())
		}
	}

	// 弹出元素
	for i := 4; i >= 0; i-- {
		heap.PopItem()
		if heap.Length() != i {
			t.Errorf("弹出后长度应该是 %d，实际是 %d", i, heap.Length())
		}
	}
}

func TestHeap_WithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	heap := NewHeap[Person]()

	persons := []*Item[Person]{
		{Data: Person{"Alice", 25}, Weight: 3},
		{Data: Person{"Bob", 30}, Weight: 1},
		{Data: Person{"Charlie", 35}, Weight: 2},
	}

	// 推入结构体
	for _, person := range persons {
		heap.PushItem(person)
	}

	// 弹出结构体（按权重排序）
	expectedOrder := []string{"Bob", "Charlie", "Alice"}
	for i, expectedName := range expectedOrder {
		popped := heap.PopItem()
		if popped.Data.Name != expectedName {
			t.Errorf("期望弹出 %s，实际弹出 %s", expectedName, popped.Data.Name)
		}
		if heap.Length() != len(persons)-i-1 {
			t.Errorf("弹出后长度应该是 %d，实际是 %d", len(persons)-i-1, heap.Length())
		}
	}
}

func TestHeap_WithPointer(t *testing.T) {
	heap := NewHeap[*int]()

	values := []int{1, 2, 3, 4, 5}
	weights := []int{3, 1, 2, 0, 4}

	// 推入指针
	for i, v := range values {
		val := v
		heap.PushItem(&Item[*int]{Data: &val, Weight: weights[i]})
	}

	// 弹出指针（按权重排序）
	expectedOrder := []int{4, 2, 3, 1, 5}
	for _, expectedData := range expectedOrder {
		popped := heap.PopItem()
		if *popped.Data != expectedData {
			t.Errorf("期望弹出 %d，实际弹出 %d", expectedData, *popped.Data)
		}
	}
}

// Heap 非并发安全，不提供 ConcurrentOperations 测试。
// 需要并发访问时，调用方负责加锁。

func TestHeap_LargeDataset(t *testing.T) {
	heap := NewHeap[int]()

	// 推入大量数据
	for i := 0; i < 10000; i++ {
		heap.PushItem(&Item[int]{Data: i, Weight: i})
	}

	// 验证堆的性质：每次弹出的都是当前最小值
	lastWeight := -1
	for heap.Length() > 0 {
		popped := heap.PopItem()
		if lastWeight != -1 && popped.Weight < lastWeight {
			t.Errorf("堆性质被破坏：当前权重 %d 小于前一个权重 %d", popped.Weight, lastWeight)
		}
		lastWeight = popped.Weight
	}
}

var heapBenchSizes = []int{100, 1_000, 10_000, 100_000}

// BenchmarkHeap_Push 最小堆 Push 基准，集合维持在 size 规模（Push+Pop 配对）。
func BenchmarkHeap_Push(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			for b.Loop() {
				h.PushItem(&Item[int]{Data: i, Weight: i})
				h.PopItem() // 保持集合规模稳定
				i++
			}
		})
	}
}

// BenchmarkHeap_Pop 最小堆 Pop 基准，集合维持在 size 规模（Pop+Push 配对）。
func BenchmarkHeap_Pop(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			for b.Loop() {
				h.PopItem()
				h.PushItem(&Item[int]{Data: i, Weight: i}) // 保持集合规模稳定
				i++
			}
		})
	}
}

// BenchmarkHeap_Peek 最小堆 Peek 基准（只读，无副作用）。
func BenchmarkHeap_Peek(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				h.PeekItem()
			}
		})
	}
}

// BenchmarkHeap_Mixed 混合负载基准：70% Push + 30% Pop，模拟优先级队列真实使用比例。
func BenchmarkHeap_Mixed(b *testing.B) {
	for _, size := range heapBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			h := NewHeap[int]()
			for i := range size {
				h.PushItem(&Item[int]{Data: i, Weight: i})
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			op := 0
			for b.Loop() {
				if op%10 < 7 { // 70% Push
					h.PushItem(&Item[int]{Data: i, Weight: i})
					i++
				} else { // 30% Pop（若非空）
					if h.Length() > 0 {
						h.PopItem()
					}
				}
				op++
			}
		})
	}
}
