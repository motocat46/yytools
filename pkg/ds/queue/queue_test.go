package queue

import (
	"fmt"
	"testing"
)

func TestNewQueue(t *testing.T) {
	queue := NewQueue[int]()
	if queue == nil {
		t.Fatal("NewQueue() 返回了 nil")
	}
	if queue.Len() != 0 {
		t.Errorf("新队列的长度应该是 0，实际是 %d", queue.Len())
	}
	if !queue.Empty() {
		t.Error("新队列应该是空的")
	}
}

func TestNewQueueWithSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"正常大小", 10, 0},
		{"大尺寸", 1000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewQueueWithSize[int](tt.size)
			if queue == nil {
				t.Fatal("NewQueueWithSize() 返回了 nil")
			}
			if queue.Len() != tt.expected {
				t.Errorf("期望长度 %d，实际长度 %d", tt.expected, queue.Len())
			}
		})
	}
}

func TestQueue_Enqueue(t *testing.T) {
	queue := NewQueue[string]()

	// 测试基本入队操作
	items := []string{"hello", "world", "test", "queue"}

	for i, item := range items {
		queue.Enqueue(item)
		if queue.Len() != i+1 {
			t.Errorf("入队后长度应该是 %d，实际是 %d", i+1, queue.Len())
		}
		if queue.Empty() {
			t.Error("队列不应该为空")
		}
	}
}

func TestQueue_Dequeue(t *testing.T) {
	queue := NewQueue[int]()
	items := []int{1, 2, 3, 4, 5}

	// 先入队所有元素
	for _, item := range items {
		queue.Enqueue(item)
	}

	// 测试出队操作（先进先出）
	for i, item := range items {
		dequeued := queue.Dequeue()
		if dequeued != item {
			t.Errorf("期望出队 %v，实际出队 %v", item, dequeued)
		}
		if queue.Len() != len(items)-i-1 {
			t.Errorf("出队后长度应该是 %d，实际是 %d", len(items)-i-1, queue.Len())
		}
	}
}

func TestQueue_Peek_Empty(t *testing.T) {
	queue := NewQueue[int]()
	defer func() {
		if r := recover(); r == nil {
			t.Error("空队列 Peek 应 panic，但没有 panic")
		}
	}()
	queue.Peek()
}

func TestQueue_Dequeue_Empty(t *testing.T) {
	queue := NewQueue[int]()
	defer func() {
		if r := recover(); r == nil {
			t.Error("空队列 Dequeue 应 panic，但没有 panic")
		}
	}()
	queue.Dequeue()
}

func TestQueue_PeekWithItems(t *testing.T) {
	queue := NewQueue[float64]()
	items := []float64{1.1, 2.2, 3.3}

	for _, item := range items {
		queue.Enqueue(item)
		peek := queue.Peek()
		if peek != items[0] {
			t.Errorf("期望队首元素 %v，实际是 %v", items[0], peek)
		}
	}
}

func TestQueue_Empty(t *testing.T) {
	queue := NewQueue[string]()

	if !queue.Empty() {
		t.Error("新队列应该是空的")
	}

	queue.Enqueue("test")
	if queue.Empty() {
		t.Error("有元素的队列不应该为空")
	}

	queue.Dequeue()
	if !queue.Empty() {
		t.Error("出队所有元素后队列应该是空的")
	}
}

func TestQueue_Length(t *testing.T) {
	queue := NewQueue[int]()

	if queue.Len() != 0 {
		t.Errorf("新队列长度应该是 0，实际是 %d", queue.Len())
	}

	// 入队元素
	for i := 1; i <= 5; i++ {
		queue.Enqueue(i)
		if queue.Len() != i {
			t.Errorf("入队 %d 个元素后长度应该是 %d，实际是 %d", i, i, queue.Len())
		}
	}

	// 出队元素
	for i := 4; i >= 0; i-- {
		queue.Dequeue()
		if queue.Len() != i {
			t.Errorf("出队后长度应该是 %d，实际是 %d", i, queue.Len())
		}
	}
}

func TestQueue_Capacity(t *testing.T) {
	queue := NewQueueWithSize[int](10)

	if queue.Capacity() != 10 {
		t.Errorf("期望容量 10，实际容量 %d", queue.Capacity())
	}

	// 测试扩容
	for i := 0; i < 20; i++ {
		queue.Enqueue(i)
	}

	if queue.Capacity() <= 10 {
		t.Error("队列应该已经扩容")
	}
}

func TestQueue_Range(t *testing.T) {
	queue := NewQueue[string]()
	items := []string{"hello", "world", "test", "queue"}

	// 入队元素
	for _, item := range items {
		queue.Enqueue(item)
	}

	// 测试遍历
	visited := make([]string, 0)
	queue.Range(func(item string) {
		visited = append(visited, item)
	})

	if len(visited) != len(items) {
		t.Errorf("期望遍历 %d 个元素，实际遍历 %d 个", len(items), len(visited))
	}

	for i, item := range items {
		if visited[i] != item {
			t.Errorf("期望遍历元素 %v，实际是 %v", item, visited[i])
		}
	}
}

func TestQueue_EmptyRange(t *testing.T) {
	queue := NewQueue[int]()

	// 测试空队列的遍历
	count := 0
	queue.Range(func(item int) {
		count++
	})

	if count != 0 {
		t.Errorf("空队列遍历应该不执行任何操作，实际执行了 %d 次", count)
	}
}

func TestQueue_ExpandAndShrink(t *testing.T) {
	queue := NewQueueWithSize[int](4)

	// 测试扩容：入队25个元素，队列会从4扩容至32
	for i := 0; i < 25; i++ {
		queue.Enqueue(i)
	}

	initialCap := queue.Capacity()

	// 测试缩容：出队20个元素，剩余5个 < initialCap/4，触发缩容
	for i := 0; i < 20; i++ {
		queue.Dequeue()
	}

	// 验证缩容是否生效
	if queue.Capacity() >= initialCap {
		t.Error("队列应该已经缩容")
	}
}

// Queue 非并发安全，不提供 ConcurrentOperations 测试。
// 需要并发访问时，调用方负责加锁。

func TestQueue_WithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	queue := NewQueue[Person]()

	persons := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 35},
	}

	// 入队结构体
	for _, person := range persons {
		queue.Enqueue(person)
	}

	// 出队结构体
	for i, person := range persons {
		dequeued := queue.Dequeue()
		if dequeued != person {
			t.Errorf("期望出队 %v，实际出队 %v", person, dequeued)
		}
		if queue.Len() != len(persons)-i-1 {
			t.Errorf("出队后长度应该是 %d，实际是 %d", len(persons)-i-1, queue.Len())
		}
	}
}

func TestQueue_WithPointer(t *testing.T) {
	queue := NewQueue[*int]()

	values := []int{1, 2, 3, 4, 5}

	// 入队指针
	for _, v := range values {
		val := v
		queue.Enqueue(&val)
	}

	// 出队指针
	for _, v := range values {
		dequeued := queue.Dequeue()
		if *dequeued != v {
			t.Errorf("期望出队 %d，实际出队 %d", v, *dequeued)
		}
	}
}

var queueBenchSizes = []int{100, 1_000, 10_000, 100_000}

// BenchmarkQueue_Enqueue 入队基准，集合维持在 size 规模（Enqueue+Dequeue 配对）。
func BenchmarkQueue_Enqueue(b *testing.B) {
	for _, size := range queueBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			q := NewQueue[int]()
			for i := range size {
				q.Enqueue(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			for b.Loop() {
				q.Enqueue(i)
				q.Dequeue() // 保持规模稳定
				i++
			}
		})
	}
}

// BenchmarkQueue_Dequeue 出队基准，集合维持在 size 规模（Dequeue+Enqueue 配对）。
func BenchmarkQueue_Dequeue(b *testing.B) {
	for _, size := range queueBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			q := NewQueue[int]()
			for i := range size {
				q.Enqueue(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			for b.Loop() {
				q.Dequeue()
				q.Enqueue(i) // 保持规模稳定
				i++
			}
		})
	}
}

// BenchmarkQueue_Peek 队首窥视基准（只读，无副作用）。
func BenchmarkQueue_Peek(b *testing.B) {
	for _, size := range queueBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			q := NewQueue[int]()
			for i := range size {
				q.Enqueue(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				q.Peek()
			}
		})
	}
}

// BenchmarkQueue_Range 遍历基准，固定 size 规模。
func BenchmarkQueue_Range(b *testing.B) {
	for _, size := range queueBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			q := NewQueue[int]()
			for i := range size {
				q.Enqueue(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				q.Range(func(item int) {
					_ = item
				})
			}
		})
	}
}

// BenchmarkQueue_Mixed 混合负载基准：50% Enqueue + 50% Dequeue，模拟稳态生产消费。
func BenchmarkQueue_Mixed(b *testing.B) {
	for _, size := range queueBenchSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			q := NewQueue[int]()
			for i := range size {
				q.Enqueue(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := size
			op := 0
			for b.Loop() {
				if op%2 == 0 {
					q.Enqueue(i)
					i++
				} else {
					q.Dequeue()
					q.Enqueue(i) // 防止队列耗尽
					i++
				}
				op++
			}
		})
	}
}
