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
