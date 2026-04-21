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

package segtree_test

import (
	"testing"

	st "github.com/motocat46/yytools/pkg/ds/segtree"
)

// newRangeAddSum 创建区间加法 + 区间求和线段树。
func newRangeAddSum(n int) *st.SegTree[int, int] {
	return st.New[int, int](n, 0,
		func(a, b int) int { return a + b },
		0,
		func(val, lazy, size int) int { return val + lazy*size },
		func(newL, oldL int) int { return newL + oldL },
	)
}

// itoa 将 int 转字符串，用于子测试命名，避免引入 strconv。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func TestNew_Len(t *testing.T) {
	s := newRangeAddSum(8)
	if s.Len() != 8 {
		t.Errorf("Len(): got %d, want 8", s.Len())
	}
}

func TestNew_ZeroPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("New(0) 应触发 panic")
		}
	}()
	newRangeAddSum(0)
}

func TestNew_NegativePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("New(-1) 应触发 panic")
		}
	}()
	newRangeAddSum(-1)
}

func TestNew_InitAllIdentity(t *testing.T) {
	s := newRangeAddSum(5)
	if got := s.QueryAll(); got != 0 {
		t.Errorf("QueryAll() 初始应为 0，got %d", got)
	}
}

func TestSet_Basic(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(2, 10)
	// Set 后 Query 还不存在，先用 QueryAll 验证
	if got := s.QueryAll(); got != 10 {
		t.Errorf("QueryAll() after Set(2,10): got %d, want 10", got)
	}
}

func TestSet_OverwritesSameIndex(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(2, 10)
	s.Set(2, 3)
	if got := s.QueryAll(); got != 3 {
		t.Errorf("QueryAll() after Set(2,10) then Set(2,3): got %d, want 3", got)
	}
}

func TestSet_SumsDifferentIndexes(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(1, 4)
	s.Set(3, 6)
	if got := s.QueryAll(); got != 10 {
		t.Errorf("QueryAll() after Set(1,4) then Set(3,6): got %d, want 10", got)
	}
}

func TestSet_BoundaryFirst(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(0, 42)
	if got := s.QueryAll(); got != 42 {
		t.Errorf("QueryAll() after Set(0,42): got %d, want 42", got)
	}
}

func TestSet_BoundaryLast(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(4, 99)
	if got := s.QueryAll(); got != 99 {
		t.Errorf("QueryAll() after Set(4,99): got %d, want 99", got)
	}
}

func TestSet_OutOfBoundsPanics(t *testing.T) {
	s := newRangeAddSum(5)
	for _, i := range []int{-1, 5} {
		i := i
		t.Run(itoa(i), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Set(%d, 0) 应触发 panic", i)
				}
			}()
			s.Set(i, 0)
		})
	}
}

// newRangeAssignMin 创建区间赋值 + 区间最小值线段树。
type assignLazy struct {
	val    int
	hasVal bool
}

func newRangeAssignMin(n int) *st.SegTree[int, assignLazy] {
	return st.New[int, assignLazy](n, 1<<62,
		func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		assignLazy{},
		func(val int, lazy assignLazy, _ int) int {
			if lazy.hasVal {
				return lazy.val
			}
			return val
		},
		func(newL, oldL assignLazy) assignLazy {
			if newL.hasVal {
				return newL
			}
			return oldL
		},
	)
}

// newRangeAddMax 创建区间加法 + 区间最大值线段树。
func newRangeAddMax(n int) *st.SegTree[int, int] {
	return st.New[int, int](n, -1<<62,
		func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		0,
		func(val, lazy, _ int) int { return val + lazy },
		func(newL, oldL int) int { return newL + oldL },
	)
}

func TestQuery_Basic(t *testing.T) {
	s := newRangeAddSum(5)
	for i, v := range []int{1, 2, 3, 4, 5} {
		s.Set(i, v)
	}
	cases := []struct{ l, r, want int }{
		{0, 4, 15},
		{1, 3, 9},
		{0, 0, 1},
		{4, 4, 5},
		{2, 2, 3},
	}
	for _, tc := range cases {
		if got := s.Query(tc.l, tc.r); got != tc.want {
			t.Errorf("Query(%d,%d): got %d, want %d", tc.l, tc.r, got, tc.want)
		}
	}
}

func TestQuery_LEqualsZero(t *testing.T) {
	s := newRangeAddSum(3)
	for i, v := range []int{3, 5, 7} {
		s.Set(i, v)
	}
	if got := s.Query(0, 2); got != 15 {
		t.Errorf("Query(0,2): got %d, want 15", got)
	}
}

func TestQuery_LEqualsR(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(3, 77)
	if got := s.Query(3, 3); got != 77 {
		t.Errorf("Query(3,3): got %d, want 77", got)
	}
}

func TestQueryAll_SingleElement(t *testing.T) {
	s := newRangeAddSum(1)
	s.Set(0, 42)
	if got := s.QueryAll(); got != 42 {
		t.Errorf("QueryAll() single element: got %d, want 42", got)
	}
}

func TestApply_RangeAddSum(t *testing.T) {
	s := newRangeAddSum(5)
	for i := range 5 {
		s.Set(i, 1)
	}
	s.Apply(1, 3, 10)
	cases := []struct{ l, r, want int }{
		{0, 4, 35},
		{1, 3, 33},
		{0, 0, 1},
		{4, 4, 1},
		{2, 2, 11},
	}
	for _, tc := range cases {
		if got := s.Query(tc.l, tc.r); got != tc.want {
			t.Errorf("Query(%d,%d): got %d, want %d", tc.l, tc.r, got, tc.want)
		}
	}
}

func TestApply_FullRange(t *testing.T) {
	s := newRangeAddSum(4)
	for i := range 4 {
		s.Set(i, 1)
	}
	s.Apply(0, 3, 5)
	if got := s.QueryAll(); got != 24 {
		t.Errorf("QueryAll() after Apply(0,3,5): got %d, want 24", got)
	}
}

func TestApply_LEqualsR(t *testing.T) {
	s := newRangeAddSum(5)
	s.Set(2, 10)
	s.Apply(2, 2, 5)
	if got := s.Query(2, 2); got != 15 {
		t.Errorf("Query(2,2) after Apply(2,2,5): got %d, want 15", got)
	}
}

func TestApply_RangeAssignMin(t *testing.T) {
	s := newRangeAssignMin(5)
	for i, v := range []int{5, 3, 8, 1, 7} {
		s.Set(i, v)
	}
	s.Apply(1, 3, assignLazy{val: 4, hasVal: true})
	cases := []struct{ l, r, want int }{
		{0, 4, 4},
		{1, 3, 4},
		{0, 0, 5},
		{4, 4, 7},
	}
	for _, tc := range cases {
		if got := s.Query(tc.l, tc.r); got != tc.want {
			t.Errorf("Query(%d,%d): got %d, want %d", tc.l, tc.r, got, tc.want)
		}
	}
}

func TestApply_RangeAddMax(t *testing.T) {
	s := newRangeAddMax(5)
	for i, v := range []int{3, 1, 4, 1, 5} {
		s.Set(i, v)
	}
	s.Apply(1, 3, 2)
	cases := []struct{ l, r, want int }{
		{0, 4, 6},
		{1, 3, 6},
		{0, 0, 3},
		{4, 4, 5},
	}
	for _, tc := range cases {
		if got := s.Query(tc.l, tc.r); got != tc.want {
			t.Errorf("Query(%d,%d): got %d, want %d", tc.l, tc.r, got, tc.want)
		}
	}
}

func TestSet_AfterApply(t *testing.T) {
	s := newRangeAddSum(5)
	for i := range 5 {
		s.Set(i, 1)
	}
	s.Apply(0, 4, 10)
	s.Set(2, 100)
	if got := s.Query(2, 2); got != 100 {
		t.Errorf("Query(2,2) after Set over Apply: got %d, want 100", got)
	}
	if got := s.Query(0, 1); got != 22 {
		t.Errorf("Query(0,1): got %d, want 22", got)
	}
}

func TestApply_MultipleOverlapping(t *testing.T) {
	s := newRangeAddSum(5)
	for i := range 5 {
		s.Set(i, 0)
	}
	s.Apply(0, 4, 1)
	s.Apply(1, 3, 2)
	s.Apply(2, 2, 10)
	if got := s.Query(0, 4); got != 21 {
		t.Errorf("Query(0,4): got %d, want 21", got)
	}
	if got := s.Query(2, 2); got != 13 {
		t.Errorf("Query(2,2): got %d, want 13", got)
	}
}

func TestQuery_InvalidPanics(t *testing.T) {
	s := newRangeAddSum(5)
	cases := []struct{ l, r int }{{-1, 2}, {0, 5}, {3, 1}}
	for _, tc := range cases {
		tc := tc
		t.Run(itoa(tc.l)+"_"+itoa(tc.r), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Query(%d,%d) 应触发 panic", tc.l, tc.r)
				}
			}()
			s.Query(tc.l, tc.r)
		})
	}
}

func TestApply_InvalidPanics(t *testing.T) {
	s := newRangeAddSum(5)
	cases := []struct{ l, r int }{{-1, 2}, {0, 5}, {3, 1}}
	for _, tc := range cases {
		tc := tc
		t.Run(itoa(tc.l)+"_"+itoa(tc.r), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Apply(%d,%d,...) 应触发 panic", tc.l, tc.r)
				}
			}()
			s.Apply(tc.l, tc.r, 1)
		})
	}
}

func TestSet_NoAllocs(t *testing.T) {
	s := newRangeAddSum(1024)
	allocs := testing.AllocsPerRun(1000, func() {
		s.Set(512, 1)
	})
	if allocs != 0 {
		t.Fatalf("Set() allocs = %v, want 0", allocs)
	}
}

func TestApply_NoAllocs(t *testing.T) {
	s := newRangeAddSum(1024)
	allocs := testing.AllocsPerRun(1000, func() {
		s.Apply(100, 900, 1)
	})
	if allocs != 0 {
		t.Fatalf("Apply() allocs = %v, want 0", allocs)
	}
}

func TestQuery_NoAllocs(t *testing.T) {
	s := newRangeAddSum(1024)
	for i := range 1024 {
		s.Set(i, i+1)
	}
	allocs := testing.AllocsPerRun(1000, func() {
		_ = s.Query(100, 900)
	})
	if allocs != 0 {
		t.Fatalf("Query() allocs = %v, want 0", allocs)
	}
}
