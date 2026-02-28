// Package sampling.

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

// 作者:  yangyuan
// 创建日期:2026/2/28
package sampling

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

// 用于验证派生类型能正确派发
type myInt32 int32

func newRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 0))
}

// allDistinct 检查切片中所有元素是否两两不同
func allDistinct[T comparable](s []T) bool {
	seen := make(map[T]struct{}, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			return false
		}
		seen[v] = struct{}{}
	}
	return true
}

// =================================================================
// SampleKDistinctFloyd
// =================================================================

// --- 返回长度 ---

func TestSampleKDistinctFloyd_Length(t *testing.T) {
	r := newRand(1)
	cases := []struct {
		name string
		k    int
	}{
		{"k=0", 0},
		{"k=1", 1},
		{"k=5", 5},
		{"k=m全区间", 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SampleKDistinctFloyd(0, 9, tc.k, r) // m=10
			if len(got) != tc.k {
				t.Fatalf("期望长度 %d，实际 %d", tc.k, len(got))
			}
		})
	}
}

// --- 值域范围 ---

func TestSampleKDistinctFloyd_Range(t *testing.T) {
	r := newRand(2)
	for i := range 200 {
		_ = i
		got := SampleKDistinctFloyd(3, 20, 5, r)
		for _, v := range got {
			if v < 3 || v > 20 {
				t.Fatalf("值 %d 超出 [3, 20]", v)
			}
		}
	}
}

// --- 不重复性 ---

func TestSampleKDistinctFloyd_Distinct(t *testing.T) {
	r := newRand(3)
	for range 200 {
		got := SampleKDistinctFloyd(0, 99, 10, r)
		if !allDistinct(got) {
			t.Fatalf("采样结果存在重复值: %v", got)
		}
	}
}

// --- k=m 全区间：必须恰好覆盖 [lo, hi] 全部值 ---

func TestSampleKDistinctFloyd_FullRange(t *testing.T) {
	r := newRand(4)
	for range 50 {
		got := SampleKDistinctFloyd(5, 14, 10, r) // [5..14], m=k=10
		slices.Sort(got)
		for j, v := range got {
			if v != 5+j {
				t.Fatalf("k=m 时排序后第 %d 个元素期望 %d，实际 %d", j, 5+j, v)
			}
		}
	}
}

// --- lo=hi 单元素区间 ---

func TestSampleKDistinctFloyd_SingleElement(t *testing.T) {
	r := newRand(5)
	for range 50 {
		got := SampleKDistinctFloyd(7, 7, 1, r)
		if len(got) != 1 || got[0] != 7 {
			t.Fatalf("单元素区间期望 [7]，实际 %v", got)
		}
	}
}

// --- 负数范围 ---

func TestSampleKDistinctFloyd_NegativeRange(t *testing.T) {
	r := newRand(6)
	for range 200 {
		got := SampleKDistinctFloyd(-50, -1, 5, r)
		for _, v := range got {
			if v < -50 || v > -1 {
				t.Fatalf("值 %d 超出 [-50, -1]", v)
			}
		}
		if !allDistinct(got) {
			t.Fatal("采样结果存在重复值")
		}
	}
}

// --- 跨零范围 ---

func TestSampleKDistinctFloyd_CrossZero(t *testing.T) {
	r := newRand(7)
	for range 200 {
		got := SampleKDistinctFloyd(-10, 10, 5, r)
		for _, v := range got {
			if v < -10 || v > 10 {
				t.Fatalf("值 %d 超出 [-10, 10]", v)
			}
		}
		if !allDistinct(got) {
			t.Fatal("采样结果存在重复值")
		}
	}
}

// --- 多类型覆盖 ---

func TestSampleKDistinctFloyd_Types(t *testing.T) {
	r := newRand(8)
	t.Run("int32", func(t *testing.T) {
		for range 100 {
			got := SampleKDistinctFloyd[int32](0, 99, 5, r)
			for _, v := range got {
				if v < 0 || v > 99 {
					t.Fatalf("int32 值 %d 超出 [0, 99]", v)
				}
			}
			if !allDistinct(got) {
				t.Fatal("int32 结果存在重复值")
			}
		}
	})
	t.Run("int64", func(t *testing.T) {
		for range 100 {
			got := SampleKDistinctFloyd[int64](1000, 2000, 10, r)
			for _, v := range got {
				if v < 1000 || v > 2000 {
					t.Fatalf("int64 值 %d 超出 [1000, 2000]", v)
				}
			}
		}
	})
	t.Run("int16", func(t *testing.T) {
		for range 100 {
			got := SampleKDistinctFloyd[int16](-100, 100, 8, r)
			for _, v := range got {
				if v < -100 || v > 100 {
					t.Fatalf("int16 值 %d 超出 [-100, 100]", v)
				}
			}
		}
	})
}

// --- 派生类型 ---

func TestSampleKDistinctFloyd_DerivedType(t *testing.T) {
	r := newRand(9)
	for range 100 {
		got := SampleKDistinctFloyd[myInt32](0, 99, 5, r)
		for _, v := range got {
			if v < 0 || v > 99 {
				t.Fatalf("myInt32 值 %d 超出 [0, 99]", v)
			}
		}
	}
}

// --- 相同种子产生相同序列 ---

func TestSampleKDistinctFloyd_Replay(t *testing.T) {
	r1 := newRand(42)
	r2 := newRand(42)
	for i := range 50 {
		a := SampleKDistinctFloyd(0, 999, 10, r1)
		b := SampleKDistinctFloyd(0, 999, 10, r2)
		for j := range a {
			if a[j] != b[j] {
				t.Fatalf("第 %d 次第 %d 个元素不同: %d vs %d", i, j, a[j], b[j])
			}
		}
	}
}

// --- 均匀性：k=1 时每个值出现频率接近 1/m ---

func TestSampleKDistinctFloyd_Uniform(t *testing.T) {
	const m = 10
	const n = 10000
	const tolerance = 0.15
	r := newRand(99)
	counts := make([]int, m)
	for range n {
		got := SampleKDistinctFloyd(0, m-1, 1, r)
		counts[got[0]]++
	}
	expected := float64(n) / m
	for i, cnt := range counts {
		dev := math.Abs(float64(cnt)-expected) / expected
		if dev > tolerance {
			t.Errorf("值 %d 命中 %d 次，期望约 %.0f（偏差 %.1f%%，允许 %.0f%%）",
				i, cnt, expected, dev*100, tolerance*100)
		}
	}
}

// --- assert/panic 触发 ---

func TestSampleKDistinctFloyd_AssertPanic(t *testing.T) {
	r := newRand(0)
	t.Run("lo>hi触发panic", func(t *testing.T) {
		defer func() {
			if rv := recover(); rv == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		SampleKDistinctFloyd(10, 5, 1, r)
	})
	t.Run("k<0触发panic", func(t *testing.T) {
		defer func() {
			if rv := recover(); rv == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		SampleKDistinctFloyd(0, 9, -1, r)
	})
	t.Run("k>m触发panic", func(t *testing.T) {
		defer func() {
			if rv := recover(); rv == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		SampleKDistinctFloyd(0, 4, 6, r) // m=5, k=6
	})
}

// =================================================================
// SampleWithMinGap
// =================================================================

// --- k<=0 返回 nil ---

func TestSampleWithMinGap_KZero(t *testing.T) {
	r := newRand(1)
	t.Run("k=0返回nil", func(t *testing.T) {
		if got := SampleWithMinGap(0, 100, 0, 3, r); got != nil {
			t.Fatalf("k=0 应返回 nil，实际 %v", got)
		}
	})
	t.Run("k<0返回nil", func(t *testing.T) {
		if got := SampleWithMinGap(0, 100, -1, 3, r); got != nil {
			t.Fatalf("k<0 应返回 nil，实际 %v", got)
		}
	})
}

// --- 返回长度 ---

func TestSampleWithMinGap_Length(t *testing.T) {
	r := newRand(2)
	for range 200 {
		got := SampleWithMinGap(0, 99, 5, 3, r)
		if len(got) != 5 {
			t.Fatalf("期望长度 5，实际 %d", len(got))
		}
	}
}

// --- 结果有序 ---

func TestSampleWithMinGap_Sorted(t *testing.T) {
	r := newRand(3)
	for range 200 {
		got := SampleWithMinGap(0, 99, 6, 4, r)
		for j := 1; j < len(got); j++ {
			if got[j] <= got[j-1] {
				t.Fatalf("结果未严格升序: [%d]=%d <= [%d]=%d", j, got[j], j-1, got[j-1])
			}
		}
	}
}

// --- 值域范围 ---

func TestSampleWithMinGap_InRange(t *testing.T) {
	r := newRand(4)
	for range 200 {
		got := SampleWithMinGap(10, 50, 4, 3, r)
		for _, v := range got {
			if v < 10 || v > 50 {
				t.Fatalf("值 %d 超出 [10, 50]", v)
			}
		}
	}
}

// --- 相邻间隔 >= gap ---

func TestSampleWithMinGap_MinGapEnforced(t *testing.T) {
	const gap = 5
	r := newRand(5)
	for range 200 {
		got := SampleWithMinGap(0, 99, 6, gap, r)
		for j := 1; j < len(got); j++ {
			diff := int(got[j]) - int(got[j-1])
			if diff < gap {
				t.Fatalf("[%d]-[%d] 差值 %d 小于 gap=%d；结果 %v", j, j-1, diff, gap, got)
			}
		}
	}
}

// --- k=1 时无相邻约束，值在 [L, R] 内即可 ---

func TestSampleWithMinGap_K1(t *testing.T) {
	r := newRand(6)
	for range 200 {
		got := SampleWithMinGap(20, 80, 1, 10, r)
		if len(got) != 1 || got[0] < 20 || got[0] > 80 {
			t.Fatalf("k=1 结果异常: %v", got)
		}
	}
}

// --- gap=0 退化为不重复采样，结果有序 ---

func TestSampleWithMinGap_Gap0(t *testing.T) {
	r := newRand(7)
	for range 200 {
		got := SampleWithMinGap(0, 49, 5, 0, r)
		if len(got) != 5 {
			t.Fatalf("期望长度 5，实际 %d", len(got))
		}
		for j := 1; j < len(got); j++ {
			if got[j] <= got[j-1] {
				t.Fatalf("gap=0 结果应严格升序: [%d]=%d <= [%d]=%d", j, got[j], j-1, got[j-1])
			}
		}
		for _, v := range got {
			if v < 0 || v > 49 {
				t.Fatalf("值 %d 超出 [0, 49]", v)
			}
		}
	}
}

// --- 最紧约束（M == k）：仅有唯一选法，结果确定 ---
// L=0, R=6, k=3, gap=2 → M=7-4=3=k，压缩范围 [0,2]，唯一选法 {0,1,2}
// 还原后 a = [0, 1+2, 2+4] = [0, 3, 6]
func TestSampleWithMinGap_TightConstraint(t *testing.T) {
	r := newRand(8)
	want := []int{0, 3, 6}
	for range 20 {
		got := SampleWithMinGap(0, 6, 3, 2, r)
		if len(got) != 3 {
			t.Fatalf("期望长度 3，实际 %d", len(got))
		}
		for j, v := range got {
			if v != want[j] {
				t.Fatalf("最紧约束结果错误: got[%d]=%d，want %d", j, v, want[j])
			}
		}
	}
}

// --- 负数范围 ---

func TestSampleWithMinGap_NegativeRange(t *testing.T) {
	const gap = 3
	r := newRand(9)
	for range 200 {
		got := SampleWithMinGap(-50, -1, 4, gap, r)
		for j, v := range got {
			if v < -50 || v > -1 {
				t.Fatalf("值 %d 超出 [-50, -1]", v)
			}
			if j > 0 && int(got[j])-int(got[j-1]) < gap {
				t.Fatalf("相邻间隔 %d 小于 gap=%d；结果 %v", int(got[j])-int(got[j-1]), gap, got)
			}
		}
	}
}

// --- 多类型覆盖 ---

func TestSampleWithMinGap_Types(t *testing.T) {
	r := newRand(10)
	t.Run("int32", func(t *testing.T) {
		for range 100 {
			got := SampleWithMinGap[int32](0, 99, 5, 4, r)
			for j := 1; j < len(got); j++ {
				if got[j]-got[j-1] < 4 {
					t.Fatalf("int32 间隔不足: %d", got[j]-got[j-1])
				}
			}
		}
	})
	t.Run("int64", func(t *testing.T) {
		for range 100 {
			got := SampleWithMinGap[int64](1000, 9999, 5, 100, r)
			for j := 1; j < len(got); j++ {
				if got[j]-got[j-1] < 100 {
					t.Fatalf("int64 间隔不足: %d", got[j]-got[j-1])
				}
			}
		}
	})
}

// --- 相同种子产生相同序列 ---

func TestSampleWithMinGap_Replay(t *testing.T) {
	r1 := newRand(42)
	r2 := newRand(42)
	for i := range 50 {
		a := SampleWithMinGap(0, 999, 5, 10, r1)
		b := SampleWithMinGap(0, 999, 5, 10, r2)
		for j := range a {
			if a[j] != b[j] {
				t.Fatalf("第 %d 次第 %d 个元素不同: %d vs %d", i, j, a[j], b[j])
			}
		}
	}
}

// --- assert/panic 触发 ---

func TestSampleWithMinGap_AssertPanic(t *testing.T) {
	r := newRand(0)
	t.Run("gap<0触发panic", func(t *testing.T) {
		defer func() {
			if rv := recover(); rv == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		SampleWithMinGap(0, 99, 5, -1, r)
	})
	t.Run("约束不可满足触发panic", func(t *testing.T) {
		// L=0, R=5, k=3, gap=3 → M=6-6=0 < k=3
		defer func() {
			if rv := recover(); rv == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		SampleWithMinGap(0, 5, 3, 3, r)
	})
	t.Run("无效区间触发panic", func(t *testing.T) {
		// L > R → N <= 0
		defer func() {
			if rv := recover(); rv == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		SampleWithMinGap(10, 5, 3, 1, r)
	})
}

// =================================================================
// 基准测试
// =================================================================

var benchSinkFloyd []int
var benchSinkGap []int32

func BenchmarkSampleKDistinctFloyd(b *testing.B) {
	r := newRand(1)
	b.ReportAllocs()
	for b.Loop() {
		benchSinkFloyd = SampleKDistinctFloyd(0, 9999, 20, r)
	}
}

func BenchmarkSampleWithMinGap(b *testing.B) {
	r := newRand(1)
	b.ReportAllocs()
	for b.Loop() {
		benchSinkGap = SampleWithMinGap[int32](0, 9999, 20, 10, r)
	}
}
