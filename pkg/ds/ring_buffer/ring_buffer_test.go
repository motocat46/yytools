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
package ring_buffer_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	rb "github.com/motocat46/yytools/pkg/ds/ring_buffer"
)

func TestRingBuffer_NewRingBuffer_非法容量Panic(t *testing.T) {
	for _, c := range []int{0, -1} {
		t.Run(fmt.Sprintf("cap=%d", c), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("NewRingBuffer(%d) 应 panic", c)
				}
			}()
			rb.NewRingBuffer[int](c)
		})
	}
}

// --- 单方法测试 ---

func TestRingBuffer_Enqueue_Dequeue_正常(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	r.Enqueue(1)
	r.Enqueue(2)
	r.Enqueue(3)

	for i, want := range []int{1, 2, 3} {
		got := r.Dequeue()
		if got != want {
			t.Errorf("第 %d 次 Dequeue: got %d, want %d", i+1, got, want)
		}
	}
	if !r.Empty() {
		t.Error("全部 Dequeue 后应为空")
	}
}

func TestRingBuffer_Enqueue_满时覆盖(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	r.Enqueue(1)
	r.Enqueue(2)
	r.Enqueue(3)
	r.Enqueue(4) // 覆盖 1

	// 覆盖后 Len 和 Cap 必须不变（API 契约）
	if r.Len() != 3 {
		t.Errorf("覆盖后 Len 应为 3，got %d", r.Len())
	}
	if r.Cap() != 3 {
		t.Errorf("覆盖后 Cap 应为 3，got %d", r.Cap())
	}
	if got := r.Dequeue(); got != 2 {
		t.Errorf("覆盖后队首应为 2，got %d", got)
	}
}

func TestRingBuffer_Peek_不移除(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	r.Enqueue(42)
	if got := r.Peek(); got != 42 {
		t.Errorf("Peek got %d, want 42", got)
	}
	if r.Len() != 1 {
		t.Errorf("Peek 后 Len 应为 1，got %d", r.Len())
	}
}

func TestRingBuffer_Dequeue_空时Panic(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	defer func() {
		if recover() == nil {
			t.Error("空缓冲区 Dequeue 应 panic")
		}
	}()
	r.Dequeue()
}

func TestRingBuffer_Peek_空时Panic(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	defer func() {
		if recover() == nil {
			t.Error("空缓冲区 Peek 应 panic")
		}
	}()
	r.Peek()
}

func TestRingBuffer_Full_Empty(t *testing.T) {
	r := rb.NewRingBuffer[int](2)
	if !r.Empty() {
		t.Error("初始应为空")
	}
	r.Enqueue(1)
	r.Enqueue(2)
	if !r.Full() {
		t.Error("写满后应为 full")
	}
	r.Dequeue()
	if r.Full() {
		t.Error("Dequeue 后不应为 full")
	}
}

// --- 边界测试 ---

func TestRingBuffer_Capacity1(t *testing.T) {
	r := rb.NewRingBuffer[int](1)
	r.Enqueue(1)
	if !r.Full() {
		t.Error("capacity=1 写入后应满")
	}
	r.Enqueue(2) // 覆盖
	if got := r.Dequeue(); got != 2 {
		t.Errorf("覆盖后应读到 2，got %d", got)
	}
	if !r.Empty() {
		t.Error("Dequeue 后应为空")
	}
}

func TestRingBuffer_反复写满读空(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	for round := range 5 {
		for i := range 3 {
			r.Enqueue(round*3 + i)
		}
		for i := range 3 {
			want := round*3 + i
			if got := r.Dequeue(); got != want {
				t.Errorf("round=%d i=%d: got %d, want %d", round, i, got, want)
			}
		}
	}
}

func TestRingBuffer_多轮覆盖(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	// 写入 9 个元素（3 轮覆盖），最终应保留最后 3 个
	for i := 1; i <= 9; i++ {
		r.Enqueue(i)
	}
	for _, want := range []int{7, 8, 9} {
		if got := r.Dequeue(); got != want {
			t.Errorf("多轮覆盖后: got %d, want %d", got, want)
		}
	}
}

// --- 集成测试：Range 顺序 ---

func TestRingBuffer_Range_顺序正确(t *testing.T) {
	r := rb.NewRingBuffer[int](5)
	for i := 1; i <= 5; i++ {
		r.Enqueue(i)
	}
	var got []int
	r.Range(func(v int) bool { got = append(got, v); return true })
	for i, want := range []int{1, 2, 3, 4, 5} {
		if got[i] != want {
			t.Errorf("Range[%d]: got %d, want %d", i, got[i], want)
		}
	}
}

func TestRingBuffer_Range_覆盖后顺序正确(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	for i := 1; i <= 5; i++ {
		r.Enqueue(i) // 最终保留 3,4,5
	}
	var got []int
	r.Range(func(v int) bool { got = append(got, v); return true })
	for i, want := range []int{3, 4, 5} {
		if got[i] != want {
			t.Errorf("Range[%d]: got %d, want %d", i, got[i], want)
		}
	}
}

func TestRingBuffer_Range_提前终止(t *testing.T) {
	r := rb.NewRingBuffer[int](5)
	for i := 1; i <= 5; i++ {
		r.Enqueue(i)
	}
	var got []int
	r.Range(func(v int) bool {
		got = append(got, v)
		return v != 3 // 遍历到 3 时停止
	})
	if len(got) != 3 {
		t.Errorf("提前终止后应收到 3 个元素，got %d", len(got))
	}
	if got[2] != 3 {
		t.Errorf("最后一个元素应为 3，got %d", got[2])
	}
}

func TestRingBuffer_Range_空缓冲区(t *testing.T) {
	r := rb.NewRingBuffer[int](3)
	called := false
	r.Range(func(int) bool { called = true; return true })
	if called {
		t.Error("空缓冲区 Range 不应调用 f")
	}
}

// --- 随机混合测试（参考模型对比）---

// refModel 用 slice 模拟 ring buffer 语义，作为参考模型
type refModel[T any] struct {
	items    []T
	capacity int
}

func newRefModel[T any](capacity int) *refModel[T] {
	return &refModel[T]{capacity: capacity}
}

func (m *refModel[T]) enqueue(item T) {
	if len(m.items) == m.capacity {
		// copy 而非切片，避免底层数组随操作次数持续增长
		copy(m.items, m.items[1:])
		m.items[m.capacity-1] = item
		return
	}
	m.items = append(m.items, item)
}

func (m *refModel[T]) dequeue() T {
	item := m.items[0]
	// copy 而非切片，与 enqueue() 策略一致，避免底层数组持续增长
	copy(m.items, m.items[1:])
	var zero T
	m.items[len(m.items)-1] = zero // 避免内存泄漏
	m.items = m.items[:len(m.items)-1]
	return item
}

func (m *refModel[T]) peek() T         { return m.items[0] }
func (m *refModel[T]) len() int        { return len(m.items) }
func (m *refModel[T]) isEmpty() bool   { return len(m.items) == 0 }
func (m *refModel[T]) isFull() bool    { return len(m.items) == m.capacity }

func TestRingBuffer_随机混合_参考模型对比(t *testing.T) {
	const (
		capacity = 50
		ops      = 100_000
	)
	rng := rand.New(rand.NewPCG(42, 0))
	r := rb.NewRingBuffer[int](capacity)
	ref := newRefModel[int](capacity)

	checkInvariants := func(i int) {
		t.Helper()
		if r.Len() != ref.len() {
			t.Fatalf("op=%d: Len 不一致: got %d, want %d", i, r.Len(), ref.len())
		}
		if r.Empty() != ref.isEmpty() {
			t.Fatalf("op=%d: Empty got %v, want %v", i, r.Empty(), ref.isEmpty())
		}
		if r.Full() != ref.isFull() {
			t.Fatalf("op=%d: Full got %v, want %v", i, r.Full(), ref.isFull())
		}
		// 验证 Range 遍历元素值与顺序均与参考模型一致（同时验证 tail==(head+length)%cap）
		var got []int
		r.Range(func(v int) bool { got = append(got, v); return true })
		if len(got) != len(ref.items) {
			t.Fatalf("op=%d: Range 长度 %d 与 ref %d 不一致", i, len(got), len(ref.items))
		}
		for j, want := range ref.items {
			if got[j] != want {
				t.Fatalf("op=%d: Range[%d] got %d, want %d", i, j, got[j], want)
			}
		}
	}

	// 操作各占 1/3，但 Dequeue/Peek 在空时跳过，实际入队比例偏高；
	// 缓冲区会较快填满，持续覆盖路径得到充分覆盖。
	for i := range ops {
		switch rng.IntN(3) {
		case 0: // Enqueue
			val := rng.Int()
			r.Enqueue(val)
			ref.enqueue(val)
		case 1: // Dequeue（非空时）
			if !r.Empty() {
				got := r.Dequeue()
				want := ref.dequeue()
				if got != want {
					t.Fatalf("op=%d Dequeue: got %d, want %d", i, got, want)
				}
			}
		case 2: // Peek（非空时）
			if !r.Empty() {
				got := r.Peek()
				want := ref.peek()
				if got != want {
					t.Fatalf("op=%d Peek: got %d, want %d", i, got, want)
				}
			}
		}
		checkInvariants(i)
	}
}

func TestRingBuffer_压力_百万(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过百万级压力测试")
	}
	const (
		capacity = 1000
		ops      = 1_000_000
	)
	rng := rand.New(rand.NewPCG(99, 0))
	r := rb.NewRingBuffer[int](capacity)
	ref := newRefModel[int](capacity)

	for i := range ops {
		if r.Len() != ref.len() {
			t.Fatalf("op=%d: Len 不一致: got %d, want %d", i, r.Len(), ref.len())
		}
		switch rng.IntN(3) {
		case 0:
			val := rng.Int()
			r.Enqueue(val)
			ref.enqueue(val)
		case 1:
			if !r.Empty() {
				got := r.Dequeue()
				want := ref.dequeue()
				if got != want {
					t.Fatalf("op=%d Dequeue: got %d, want %d", i, got, want)
				}
			}
		case 2:
			if !r.Empty() {
				got := r.Peek()
				want := ref.peek()
				if got != want {
					t.Fatalf("op=%d Peek: got %d, want %d", i, got, want)
				}
			}
		}
		// 每 1000 步验证一次 Range（避免每步都调用拖慢压力测试）
		if i%1000 == 0 {
			var got []int
			r.Range(func(v int) bool { got = append(got, v); return true })
			if len(got) != ref.len() {
				t.Fatalf("op=%d Range 长度: got %d, want %d", i, len(got), ref.len())
			}
			for j, want := range ref.items {
				if got[j] != want {
					t.Fatalf("op=%d Range[%d]: got %d, want %d", i, j, got[j], want)
				}
			}
		}
	}
}
