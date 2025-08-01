package stack

import (
	"testing"
)

func TestNewStack(t *testing.T) {
	stack := NewStack[int]()
	if stack == nil {
		t.Fatal("NewStack() 返回了 nil")
	}
	if stack.Length() != 0 {
		t.Errorf("新栈的长度应该是 0，实际是 %d", stack.Length())
	}
	if !stack.Empty() {
		t.Error("新栈应该是空的")
	}
}

func TestNewStackWithSize(t *testing.T) {
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
			stack := NewStackWithSize[int](tt.size)
			if stack == nil {
				t.Fatal("NewStackWithSize() 返回了 nil")
			}
			if stack.Length() != tt.expected {
				t.Errorf("期望长度 %d，实际长度 %d", tt.expected, stack.Length())
			}
		})
	}
}

func TestStack_Push(t *testing.T) {
	stack := NewStack[int]()

	// 测试基本推入操作
	items := []int{1, 2, 3, 4, 5}

	for i, item := range items {
		stack.Push(item)
		if stack.Length() != i+1 {
			t.Errorf("推入后长度应该是 %d，实际是 %d", i+1, stack.Length())
		}
		if stack.Empty() {
			t.Error("栈不应该为空")
		}
	}
}

func TestStack_Pop(t *testing.T) {
	stack := NewStack[string]()
	items := []string{"hello", "world", "test"}

	// 先推入所有元素
	for _, item := range items {
		stack.Push(item)
	}

	// 测试弹出操作（后进先出）
	for i := len(items) - 1; i >= 0; i-- {
		popped := stack.Pop()
		if popped != items[i] {
			t.Errorf("期望弹出 %v，实际弹出 %v", items[i], popped)
		}
		if stack.Length() != i {
			t.Errorf("弹出后长度应该是 %d，实际是 %d", i, stack.Length())
		}
	}
}

func TestStack_Top(t *testing.T) {
	stack := NewStack[int]()

	// 测试空栈的 Top 操作
	defer func() {
		if r := recover(); r == nil {
			t.Error("空栈的 Top 操作应该 panic")
		}
	}()
	stack.Top()
}

func TestStack_TopWithItems(t *testing.T) {
	stack := NewStack[float64]()
	items := []float64{1.1, 2.2, 3.3}

	for _, item := range items {
		stack.Push(item)
		top := stack.Top()
		if top != item {
			t.Errorf("期望顶部元素 %v，实际是 %v", item, top)
		}
	}
}

func TestStack_Empty(t *testing.T) {
	stack := NewStack[int]()

	if !stack.Empty() {
		t.Error("新栈应该是空的")
	}

	stack.Push(1)
	if stack.Empty() {
		t.Error("有元素的栈不应该为空")
	}

	stack.Pop()
	if !stack.Empty() {
		t.Error("弹出所有元素后栈应该是空的")
	}
}

func TestStack_Length(t *testing.T) {
	stack := NewStack[string]()

	if stack.Length() != 0 {
		t.Errorf("新栈长度应该是 0，实际是 %d", stack.Length())
	}

	// 推入元素
	for i := 1; i <= 5; i++ {
		stack.Push("item")
		if stack.Length() != i {
			t.Errorf("推入 %d 个元素后长度应该是 %d，实际是 %d", i, i, stack.Length())
		}
	}

	// 弹出元素
	for i := 4; i >= 0; i-- {
		stack.Pop()
		if stack.Length() != i {
			t.Errorf("弹出后长度应该是 %d，实际是 %d", i, stack.Length())
		}
	}
}

func TestStack_TryShrink(t *testing.T) {
	stack := NewStackWithSize[int](16)

	// 推入大量元素
	for i := 0; i < 100; i++ {
		stack.Push(i)
	}

	initialCap := cap(stack.Items)

	// 弹出大部分元素，触发缩容
	for i := 0; i < 90; i++ {
		stack.Pop()
	}

	// 验证缩容是否生效
	if cap(stack.Items) >= initialCap {
		t.Error("栈应该已经缩容")
	}
}

func TestStack_ConcurrentOperations(t *testing.T) {
	stack := NewStack[int]()
	done := make(chan bool, 2)

	// 并发推入
	go func() {
		for i := 0; i < 1000; i++ {
			stack.Push(i)
		}
		done <- true
	}()

	// 并发弹出
	go func() {
		for i := 0; i < 1000; i++ {
			if !stack.Empty() {
				stack.Pop()
			}
		}
		done <- true
	}()

	<-done
	<-done
}

func TestStack_WithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	stack := NewStack[Person]()

	persons := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 35},
	}

	// 推入结构体
	for _, person := range persons {
		stack.Push(person)
	}

	// 弹出结构体
	for i := len(persons) - 1; i >= 0; i-- {
		popped := stack.Pop()
		if popped != persons[i] {
			t.Errorf("期望弹出 %v，实际弹出 %v", persons[i], popped)
		}
	}
}

func TestStack_WithPointer(t *testing.T) {
	stack := NewStack[*int]()

	values := []int{1, 2, 3, 4, 5}

	// 推入指针
	for _, v := range values {
		val := v
		stack.Push(&val)
	}

	// 弹出指针
	for i := len(values) - 1; i >= 0; i-- {
		popped := stack.Pop()
		if *popped != values[i] {
			t.Errorf("期望弹出 %d，实际弹出 %d", values[i], *popped)
		}
	}
}

func BenchmarkStack_Push(b *testing.B) {
	stack := NewStack[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stack.Push(i)
	}
}

func BenchmarkStack_Pop(b *testing.B) {
	stack := NewStack[int]()
	for i := 0; i < b.N; i++ {
		stack.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !stack.Empty() {
			stack.Pop()
		}
	}
}

func BenchmarkStack_Top(b *testing.B) {
	stack := NewStack[int]()
	stack.Push(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.Top()
	}
}
