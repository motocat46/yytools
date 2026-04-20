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

package fenwicktree_test

import (
	"testing"

	ft "github.com/motocat46/yytools/pkg/ds/fenwicktree"
)

func TestNew_Len(t *testing.T) {
	f := ft.New[int](5)
	if f.Len() != 5 {
		t.Errorf("Len(): got %d, want 5", f.Len())
	}
}

func TestNew_ZeroPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("New(0) 应触发 panic")
		}
	}()
	ft.New[int](0)
}

func TestNew_NegativePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("New(-1) 应触发 panic")
		}
	}()
	ft.New[int](-1)
}

func TestAdd_PrefixSum_Single(t *testing.T) {
	f := ft.New[int](5)
	f.Add(2, 10)
	if got := f.PrefixSum(1); got != 0 {
		t.Errorf("PrefixSum(1): got %d, want 0", got)
	}
	if got := f.PrefixSum(2); got != 10 {
		t.Errorf("PrefixSum(2): got %d, want 10", got)
	}
	if got := f.PrefixSum(4); got != 10 {
		t.Errorf("PrefixSum(4): got %d, want 10", got)
	}
}

func TestAdd_PrefixSum_Accumulate(t *testing.T) {
	f := ft.New[int](5)
	for i, v := range []int{1, 2, 3, 4, 5} {
		f.Add(i, v)
	}
	cases := []struct{ i, want int }{
		{0, 1},
		{1, 3},
		{2, 6},
		{3, 10},
		{4, 15},
	}
	for _, tc := range cases {
		if got := f.PrefixSum(tc.i); got != tc.want {
			t.Errorf("PrefixSum(%d): got %d, want %d", tc.i, got, tc.want)
		}
	}
}

func TestAdd_NegativeDelta(t *testing.T) {
	f := ft.New[int](3)
	f.Add(1, 10)
	f.Add(1, -3)
	if got := f.PrefixSum(1); got != 7 {
		t.Errorf("PrefixSum(1): got %d, want 7", got)
	}
}

func TestAdd_BoundaryFirst(t *testing.T) {
	f := ft.New[int](5)
	f.Add(0, 42)
	if got := f.PrefixSum(0); got != 42 {
		t.Errorf("PrefixSum(0): got %d, want 42", got)
	}
}

func TestAdd_BoundaryLast(t *testing.T) {
	f := ft.New[int](5)
	f.Add(4, 99)
	if got := f.PrefixSum(3); got != 0 {
		t.Errorf("PrefixSum(3) should be 0 before Add position: got %d", got)
	}
	if got := f.PrefixSum(4); got != 99 {
		t.Errorf("PrefixSum(4): got %d, want 99", got)
	}
}

func TestAdd_OutOfBoundsPanics(t *testing.T) {
	f := ft.New[int](5)
	for _, i := range []int{-1, 5} {
		i := i
		t.Run(itoa(i), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Add(%d, 0) 应触发 panic", i)
				}
			}()
			f.Add(i, 0)
		})
	}
}

func TestPrefixSum_OutOfBoundsPanics(t *testing.T) {
	f := ft.New[int](5)
	for _, i := range []int{-1, 5} {
		i := i
		t.Run(itoa(i), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("PrefixSum(%d) 应触发 panic", i)
				}
			}()
			f.PrefixSum(i)
		})
	}
}

// itoa 将 int 转为字符串，用于子测试命名，避免引入 strconv。
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

func TestRangeSum_Basic(t *testing.T) {
	f := ft.New[int](5)
	for i, v := range []int{1, 2, 3, 4, 5} {
		f.Add(i, v)
	}
	cases := []struct{ l, r, want int }{
		{0, 0, 1},
		{0, 4, 15},
		{1, 3, 9},
		{2, 2, 3},
		{0, 2, 6},
	}
	for _, tc := range cases {
		if got := f.RangeSum(tc.l, tc.r); got != tc.want {
			t.Errorf("RangeSum(%d,%d): got %d, want %d", tc.l, tc.r, got, tc.want)
		}
	}
}

func TestRangeSum_LEqualsZero(t *testing.T) {
	f := ft.New[int](3)
	f.Add(0, 5)
	f.Add(1, 3)
	f.Add(2, 2)
	if got := f.RangeSum(0, 2); got != 10 {
		t.Errorf("RangeSum(0,2): got %d, want 10", got)
	}
}

func TestRangeSum_InvalidPanics(t *testing.T) {
	f := ft.New[int](5)
	cases := []struct{ l, r int }{
		{-1, 2},
		{2, 5},
		{3, 1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(itoa(tc.l)+"_"+itoa(tc.r), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("RangeSum(%d,%d) 应触发 panic", tc.l, tc.r)
				}
			}()
			f.RangeSum(tc.l, tc.r)
		})
	}
}

func TestFloatType(t *testing.T) {
	f := ft.New[float64](3)
	f.Add(0, 1.5)
	f.Add(1, 2.5)
	f.Add(2, 3.0)
	if got := f.PrefixSum(2); got != 7.0 {
		t.Errorf("PrefixSum(2): got %v, want 7.0", got)
	}
	if got := f.RangeSum(1, 2); got != 5.5 {
		t.Errorf("RangeSum(1,2): got %v, want 5.5", got)
	}
}

func TestBuild_Basic(t *testing.T) {
	f := ft.Build([]int{1, 2, 3, 4, 5})
	if f.Len() != 5 {
		t.Errorf("Len(): got %d, want 5", f.Len())
	}
	cases := []struct{ i, want int }{
		{0, 1},
		{1, 3},
		{2, 6},
		{3, 10},
		{4, 15},
	}
	for _, tc := range cases {
		if got := f.PrefixSum(tc.i); got != tc.want {
			t.Errorf("PrefixSum(%d): got %d, want %d", tc.i, got, tc.want)
		}
	}
	if got := f.RangeSum(1, 3); got != 9 {
		t.Errorf("RangeSum(1,3): got %d, want 9", got)
	}
}

func TestBuild_Single(t *testing.T) {
	f := ft.Build([]int{42})
	if f.Len() != 1 {
		t.Errorf("Len(): got %d, want 1", f.Len())
	}
	if got := f.PrefixSum(0); got != 42 {
		t.Errorf("PrefixSum(0): got %d, want 42", got)
	}
	if got := f.RangeSum(0, 0); got != 42 {
		t.Errorf("RangeSum(0,0): got %d, want 42", got)
	}
}

func TestBuild_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Build([]int{}) 应触发 panic")
		}
	}()
	ft.Build([]int{})
}

func TestBuild_DoesNotAliasInput(t *testing.T) {
	nums := []int{1, 2, 3}
	f := ft.Build(nums)
	nums[0] = 999
	if got := f.PrefixSum(2); got != 6 {
		t.Errorf("PrefixSum(2): got %d, want 6", got)
	}
}

func TestBuild_FloatType(t *testing.T) {
	f := ft.Build([]float64{1.5, 2.5, 3.0})
	if got := f.PrefixSum(2); got != 7.0 {
		t.Errorf("PrefixSum(2): got %v, want 7.0", got)
	}
	if got := f.RangeSum(1, 2); got != 5.5 {
		t.Errorf("RangeSum(1,2): got %v, want 5.5", got)
	}
}

func TestAdd_NoAllocs(t *testing.T) {
	f := ft.New[int](1024)
	allocs := testing.AllocsPerRun(1000, func() {
		f.Add(512, 1)
	})
	if allocs != 0 {
		t.Fatalf("Add() allocs = %v, want 0", allocs)
	}
}

func TestPrefixSum_NoAllocs(t *testing.T) {
	f := ft.Build([]int{1, 2, 3, 4, 5, 6, 7, 8})
	allocs := testing.AllocsPerRun(1000, func() {
		_ = f.PrefixSum(7)
	})
	if allocs != 0 {
		t.Fatalf("PrefixSum() allocs = %v, want 0", allocs)
	}
}

func TestRangeSum_NoAllocs(t *testing.T) {
	f := ft.Build([]int{1, 2, 3, 4, 5, 6, 7, 8})
	allocs := testing.AllocsPerRun(1000, func() {
		_ = f.RangeSum(2, 6)
	})
	if allocs != 0 {
		t.Fatalf("RangeSum() allocs = %v, want 0", allocs)
	}
}
