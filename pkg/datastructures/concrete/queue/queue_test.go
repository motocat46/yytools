package queue

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	queue := NewQueue()
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
		{"零大小", 0, 0},
		{"大尺寸", 1000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewQueueWithSize(tt.size)
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
	queue := NewQueue()

	// 测试基本入队操作
	items := []interface{}{1, "hello", 3.14, true, nil}

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
	queue := NewQueue()
	items := []interface{}{1, "hello", 3.14}

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

func TestQueue_Peek(t *testing.T) {
	queue := NewQueue()

	// 测试空队列的 Peek 操作
	defer func() {
		if r := recover(); r == nil {
			t.Error("空队列的 Peek 操作应该 panic")
		}
	}()
	queue.Peek()
}

func TestQueue_PeekWithItems(t *testing.T) {
	queue := NewQueue()
	items := []interface{}{1, "hello", 3.14}

	for _, item := range items {
		queue.Enqueue(item)
		peek := queue.Peek()
		if peek != items[0] {
			t.Errorf("期望队首元素 %v，实际是 %v", items[0], peek)
		}
	}
}

func TestQueue_Empty(t *testing.T) {
	queue := NewQueue()

	if !queue.Empty() {
		t.Error("新队列应该是空的")
	}

	queue.Enqueue(1)
	if queue.Empty() {
		t.Error("有元素的队列不应该为空")
	}

	queue.Dequeue()
	if !queue.Empty() {
		t.Error("出队所有元素后队列应该是空的")
	}
}

func TestQueue_Length(t *testing.T) {
	queue := NewQueue()

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
	queue := NewQueueWithSize(10)

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
	queue := NewQueue()
	items := []interface{}{1, "hello", 3.14, true}

	// 入队元素
	for _, item := range items {
		queue.Enqueue(item)
	}

	// 测试遍历
	visited := make([]interface{}, 0)
	queue.Range(func(item interface{}) {
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
	queue := NewQueue()

	// 测试空队列的遍历
	count := 0
	queue.Range(func(item interface{}) {
		count++
	})

	if count != 0 {
		t.Errorf("空队列遍历应该不执行任何操作，实际执行了 %d 次", count)
	}
}

func TestQueue_ExpandAndShrink(t *testing.T) {
	queue := NewQueueWithSize(4)

	// 测试扩容
	for i := 0; i < 10; i++ {
		queue.Enqueue(i)
	}

	initialCap := queue.Capacity()

	// 测试缩容
	for i := 0; i < 8; i++ {
		queue.Dequeue()
	}

	// 验证缩容是否生效
	if queue.Capacity() >= initialCap {
		t.Error("队列应该已经缩容")
	}
}

func TestQueue_ConcurrentOperations(t *testing.T) {
	queue := NewQueue()
	done := make(chan bool, 2)

	// 并发入队
	go func() {
		for i := 0; i < 1000; i++ {
			queue.Enqueue(i)
		}
		done <- true
	}()

	// 并发出队
	go func() {
		for i := 0; i < 1000; i++ {
			if !queue.Empty() {
				queue.Dequeue()
			}
		}
		done <- true
	}()

	<-done
	<-done
}

func BenchmarkQueue_Enqueue(b *testing.B) {
	queue := NewQueue()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		queue.Enqueue(i)
	}
}

func BenchmarkQueue_Dequeue(b *testing.B) {
	queue := NewQueue()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !queue.Empty() {
			queue.Dequeue()
		}
	}
}

func BenchmarkQueue_Peek(b *testing.B) {
	queue := NewQueue()
	queue.Enqueue(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Peek()
	}
}

func BenchmarkQueue_Range(b *testing.B) {
	queue := NewQueue()
	for i := 0; i < 100; i++ {
		queue.Enqueue(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Range(func(item interface{}) {
			_ = item
		})
	}
}
