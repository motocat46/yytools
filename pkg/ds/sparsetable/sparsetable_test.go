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

package sparsetable_test

import (
	"testing"

	"github.com/motocat46/yytools/pkg/ds/sparsetable"
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func gcd(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("%s: 期望触发 panic，但实际没有 panic", name)
		}
	}()
	fn()
}

func TestNew_EmptyPanic(t *testing.T) {
	mustPanic(t, "New 空切片", func() {
		sparsetable.New[int](nil, minInt)
	})
}

func TestLen(t *testing.T) {
	s := sparsetable.New([]int{5, 2, 7, 1}, minInt)
	if got, want := s.Len(), 4; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
}

func TestQuery_SingleElement(t *testing.T) {
	s := sparsetable.New([]int{42}, minInt)
	if got, want := s.Query(0, 0), 42; got != want {
		t.Fatalf("单元素 Query(0, 0) = %d, want %d", got, want)
	}
}

func TestQuery_Min(t *testing.T) {
	data := []int{7, 3, 8, 2, 9, 1, 6}
	s := sparsetable.New(data, minInt)
	if got, want := s.Query(1, 5), 1; got != want {
		t.Fatalf("最小值 Query(1, 5) = %d, want %d, data=%v", got, want, data)
	}
}

func TestQuery_Max(t *testing.T) {
	data := []int{7, 3, 8, 2, 9, 1, 6}
	s := sparsetable.New(data, maxInt)
	if got, want := s.Query(0, 4), 9; got != want {
		t.Fatalf("最大值 Query(0, 4) = %d, want %d, data=%v", got, want, data)
	}
}

func TestQuery_GCD(t *testing.T) {
	data := []int{18, 24, 30, 42, 60}
	s := sparsetable.New(data, gcd)
	if got, want := s.Query(0, 4), 6; got != want {
		t.Fatalf("GCD Query(0, 4) = %d, want %d, data=%v", got, want, data)
	}
}

func TestQuery_AllSame(t *testing.T) {
	data := []int{5, 5, 5, 5, 5, 5}
	s := sparsetable.New(data, minInt)
	if got, want := s.Query(1, 4), 5; got != want {
		t.Fatalf("全相同数据 Query(1, 4) = %d, want %d, data=%v", got, want, data)
	}
}

func TestQuery_Ascending(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7}
	s := sparsetable.New(data, minInt)
	if got, want := s.Query(2, 6), 3; got != want {
		t.Fatalf("递增数据 Query(2, 6) = %d, want %d, data=%v", got, want, data)
	}
}

func TestQuery_Descending(t *testing.T) {
	data := []int{9, 8, 7, 6, 5, 4, 3}
	s := sparsetable.New(data, maxInt)
	if got, want := s.Query(1, 5), 8; got != want {
		t.Fatalf("递减数据 Query(1, 5) = %d, want %d, data=%v", got, want, data)
	}
}

func TestQuery_PanicOutOfBounds(t *testing.T) {
	s := sparsetable.New([]int{4, 2, 7}, minInt)

	mustPanic(t, "Query 负下标", func() {
		s.Query(-1, 1)
	})
	mustPanic(t, "Query 右边界越界", func() {
		s.Query(0, 3)
	})
	mustPanic(t, "Query 左大于右", func() {
		s.Query(2, 1)
	})
}

func TestQuery_NoAllocOnSuccess(t *testing.T) {
	s := sparsetable.New([]int{7, 3, 8, 2, 9, 1, 6}, minInt)
	ranges := [][2]int{{0, 0}, {1, 5}, {2, 6}, {0, 3}}
	idx := 0

	allocs := testing.AllocsPerRun(1_000, func() {
		q := ranges[idx]
		_ = s.Query(q[0], q[1])
		idx++
		if idx == len(ranges) {
			idx = 0
		}
	})
	if allocs != 0 {
		t.Fatalf("Query 成功路径 allocs/run = %.2f, want 0", allocs)
	}
}
