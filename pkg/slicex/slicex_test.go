// Package slicex.

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

package slicex

import (
	"testing"
)

// ---------- MinInSliceOK ----------

func TestMinInSliceOK(t *testing.T) {
	t.Run("空切片返回ok=false", func(t *testing.T) {
		idx, v, ok := MinInSliceOK[int]([]int{})
		if ok {
			t.Fatal("空切片应返回 ok=false")
		}
		if idx != 0 || v != 0 {
			t.Errorf("空切片应返回 (0, 0, false)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("单元素", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]int{42})
		if !ok || idx != 0 || v != 42 {
			t.Errorf("期望 (0, 42, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("最小值在头部", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]int{1, 3, 5, 7})
		if !ok || idx != 0 || v != 1 {
			t.Errorf("期望 (0, 1, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("最小值在中部", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]int{5, 1, 7, 3})
		if !ok || idx != 1 || v != 1 {
			t.Errorf("期望 (1, 1, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("最小值在尾部", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]int{9, 7, 3, 1})
		if !ok || idx != 3 || v != 1 {
			t.Errorf("期望 (3, 1, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("全相同元素稳定返回下标0", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]int{5, 5, 5, 5})
		if !ok || idx != 0 || v != 5 {
			t.Errorf("期望 (0, 5, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("有重复最小值返回第一个", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]int{3, 1, 7, 1, 9})
		if !ok || idx != 1 || v != 1 {
			t.Errorf("期望 (1, 1, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("float64类型", func(t *testing.T) {
		idx, v, ok := MinInSliceOK([]float64{3.5, 1.1, 2.2})
		if !ok || idx != 1 || v != 1.1 {
			t.Errorf("期望 (1, 1.1, true)，实际 (%d, %f, %v)", idx, v, ok)
		}
	})
}

// ---------- MaxInSliceOK ----------

func TestMaxInSliceOK(t *testing.T) {
	t.Run("空切片返回ok=false", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK[int]([]int{})
		if ok {
			t.Fatal("空切片应返回 ok=false")
		}
		if idx != 0 || v != 0 {
			t.Errorf("空切片应返回 (0, 0, false)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("单元素", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]int{42})
		if !ok || idx != 0 || v != 42 {
			t.Errorf("期望 (0, 42, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("最大值在头部", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]int{9, 5, 3, 1})
		if !ok || idx != 0 || v != 9 {
			t.Errorf("期望 (0, 9, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("最大值在中部", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]int{3, 9, 1, 5})
		if !ok || idx != 1 || v != 9 {
			t.Errorf("期望 (1, 9, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("最大值在尾部", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]int{1, 3, 5, 9})
		if !ok || idx != 3 || v != 9 {
			t.Errorf("期望 (3, 9, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("全相同元素稳定返回下标0", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]int{5, 5, 5, 5})
		if !ok || idx != 0 || v != 5 {
			t.Errorf("期望 (0, 5, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("有重复最大值返回第一个", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]int{3, 9, 7, 9, 1})
		if !ok || idx != 1 || v != 9 {
			t.Errorf("期望 (1, 9, true)，实际 (%d, %d, %v)", idx, v, ok)
		}
	})

	t.Run("float64类型", func(t *testing.T) {
		idx, v, ok := MaxInSliceOK([]float64{1.1, 3.3, 2.2})
		if !ok || idx != 1 || v != 3.3 {
			t.Errorf("期望 (1, 3.3, true)，实际 (%d, %f, %v)", idx, v, ok)
		}
	})
}

// ---------- MinInSlice / MaxInSlice (assert 版本) ----------

func TestMinInSlice(t *testing.T) {
	t.Run("正常路径", func(t *testing.T) {
		idx, v := MinInSlice([]int{5, 1, 3})
		if idx != 1 || v != 1 {
			t.Errorf("期望 (1, 1)，实际 (%d, %d)", idx, v)
		}
	})

	t.Run("空切片触发panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("空切片应该触发 panic")
			}
		}()
		MinInSlice[int]([]int{})
	})
}

func TestMaxInSlice(t *testing.T) {
	t.Run("正常路径", func(t *testing.T) {
		idx, v := MaxInSlice([]int{5, 9, 3})
		if idx != 1 || v != 9 {
			t.Errorf("期望 (1, 9)，实际 (%d, %d)", idx, v)
		}
	})

	t.Run("空切片触发panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("空切片应该触发 panic")
			}
		}()
		MaxInSlice[int]([]int{})
	})
}

// ---------- MinByOK / MaxByOK ----------

type person struct {
	Name  string
	Score int
}

func TestMinByOK(t *testing.T) {
	lessScore := func(a, b person) bool { return a.Score < b.Score }

	t.Run("空切片返回ok=false", func(t *testing.T) {
		idx, v, ok := MinByOK([]person{}, lessScore)
		if ok {
			t.Fatal("空切片应返回 ok=false")
		}
		var zero person
		if idx != 0 || v != zero {
			t.Errorf("空切片应返回零值，实际 (%d, %v)", idx, v)
		}
	})

	t.Run("正常找到最小Score", func(t *testing.T) {
		s := []person{
			{"Alice", 80},
			{"Bob", 60},
			{"Charlie", 90},
		}
		idx, v, ok := MinByOK(s, lessScore)
		if !ok || idx != 1 || v.Name != "Bob" {
			t.Errorf("期望 (1, Bob, true)，实际 (%d, %v, %v)", idx, v, ok)
		}
	})

	t.Run("并列最优返回第一个", func(t *testing.T) {
		s := []person{
			{"Alice", 60},
			{"Bob", 60},
			{"Charlie", 90},
		}
		idx, v, ok := MinByOK(s, lessScore)
		if !ok || idx != 0 || v.Name != "Alice" {
			t.Errorf("期望 (0, Alice, true)，实际 (%d, %v, %v)", idx, v, ok)
		}
	})
}

func TestMaxByOK(t *testing.T) {
	greaterScore := func(a, b person) bool { return a.Score > b.Score }

	t.Run("空切片返回ok=false", func(t *testing.T) {
		idx, v, ok := MaxByOK([]person{}, greaterScore)
		if ok {
			t.Fatal("空切片应返回 ok=false")
		}
		var zero person
		if idx != 0 || v != zero {
			t.Errorf("空切片应返回零值，实际 (%d, %v)", idx, v)
		}
	})

	t.Run("正常找到最大Score", func(t *testing.T) {
		s := []person{
			{"Alice", 80},
			{"Bob", 60},
			{"Charlie", 90},
		}
		idx, v, ok := MaxByOK(s, greaterScore)
		if !ok || idx != 2 || v.Name != "Charlie" {
			t.Errorf("期望 (2, Charlie, true)，实际 (%d, %v, %v)", idx, v, ok)
		}
	})

	t.Run("并列最优返回第一个", func(t *testing.T) {
		s := []person{
			{"Alice", 90},
			{"Bob", 90},
			{"Charlie", 60},
		}
		idx, v, ok := MaxByOK(s, greaterScore)
		if !ok || idx != 0 || v.Name != "Alice" {
			t.Errorf("期望 (0, Alice, true)，实际 (%d, %v, %v)", idx, v, ok)
		}
	})
}

// ---------- MinBy / MaxBy (assert 版本) ----------

func TestMinBy(t *testing.T) {
	lessScore := func(a, b person) bool { return a.Score < b.Score }

	t.Run("正常路径", func(t *testing.T) {
		s := []person{{"Alice", 80}, {"Bob", 60}}
		idx, v := MinBy(s, lessScore)
		if idx != 1 || v.Name != "Bob" {
			t.Errorf("期望 (1, Bob)，实际 (%d, %v)", idx, v)
		}
	})

	t.Run("空切片触发panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("空切片应该触发 panic")
			}
		}()
		MinBy([]person{}, lessScore)
	})
}

func TestMaxBy(t *testing.T) {
	greaterScore := func(a, b person) bool { return a.Score > b.Score }

	t.Run("正常路径", func(t *testing.T) {
		s := []person{{"Alice", 80}, {"Bob", 90}}
		idx, v := MaxBy(s, greaterScore)
		if idx != 1 || v.Name != "Bob" {
			t.Errorf("期望 (1, Bob)，实际 (%d, %v)", idx, v)
		}
	})

	t.Run("空切片触发panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("空切片应该触发 panic")
			}
		}()
		MaxBy([]person{}, greaterScore)
	})
}

// ---------- Benchmark ----------

func BenchmarkMinInSlice(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = 1000 - i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MinInSlice(s)
	}
}

func BenchmarkMaxInSlice(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MaxInSlice(s)
	}
}
