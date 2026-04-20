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
	"math/rand/v2"
	"testing"

	ft "github.com/motocat46/yytools/pkg/ds/fenwicktree"
)

// refFenwick 是 FenwickTree 的参考模型，用朴素 []int 全量扫描实现，用于随机测试对比。
type refFenwick struct {
	data []int
}

func newRefFenwick(n int) *refFenwick { return &refFenwick{data: make([]int, n)} }

func (r *refFenwick) add(i, delta int) { r.data[i] += delta }

func (r *refFenwick) prefixSum(i int) int {
	s := 0
	for j := 0; j <= i; j++ {
		s += r.data[j]
	}
	return s
}

func (r *refFenwick) rangeSum(l, right int) int {
	s := 0
	for j := l; j <= right; j++ {
		s += r.data[j]
	}
	return s
}

// TestCorrectness_RandomOps 在 n=200 上执行 100k 随机操作，与参考模型逐步对比。
func TestCorrectness_RandomOps(t *testing.T) {
	const (
		n   = 200
		ops = 100_000
	)
	rng := rand.New(rand.NewPCG(42, 0))
	f := ft.New[int](n)
	ref := newRefFenwick(n)

	for i := range ops {
		switch rng.IntN(3) {
		case 0:
			idx := rng.IntN(n)
			delta := rng.IntN(201) - 100
			f.Add(idx, delta)
			ref.add(idx, delta)
		case 1:
			idx := rng.IntN(n)
			got, want := f.PrefixSum(idx), ref.prefixSum(idx)
			if got != want {
				t.Fatalf("op %d PrefixSum(%d): got %d, want %d", i, idx, got, want)
			}
		case 2:
			a, b := rng.IntN(n), rng.IntN(n)
			if a > b {
				a, b = b, a
			}
			got, want := f.RangeSum(a, b), ref.rangeSum(a, b)
			if got != want {
				t.Fatalf("op %d RangeSum(%d,%d): got %d, want %d", i, a, b, got, want)
			}
		}
	}
}

// TestCorrectness_AllSameIndex 在同一下标反复 Add，验证前缀和累加正确。
func TestCorrectness_AllSameIndex(t *testing.T) {
	const n = 10
	f := ft.New[int](n)
	ref := newRefFenwick(n)
	rng := rand.New(rand.NewPCG(99, 0))
	for range 1000 {
		delta := rng.IntN(21) - 10
		f.Add(5, delta)
		ref.add(5, delta)
	}
	for i := range n {
		got, want := f.PrefixSum(i), ref.prefixSum(i)
		if got != want {
			t.Errorf("PrefixSum(%d): got %d, want %d", i, got, want)
		}
	}
}

// TestCorrectness_BuildThenRandomOps 先从初始数组建树，再执行随机操作与参考模型对比。
func TestCorrectness_BuildThenRandomOps(t *testing.T) {
	const (
		n   = 256
		ops = 100_000
	)
	rng := rand.New(rand.NewPCG(2026, 0))
	initial := make([]int, n)
	ref := newRefFenwick(n)
	for i := range n {
		v := rng.IntN(201) - 100
		initial[i] = v
		ref.add(i, v)
	}
	f := ft.Build(initial)

	for i := range ops {
		switch rng.IntN(3) {
		case 0:
			idx := rng.IntN(n)
			delta := rng.IntN(201) - 100
			f.Add(idx, delta)
			ref.add(idx, delta)
		case 1:
			idx := rng.IntN(n)
			if got, want := f.PrefixSum(idx), ref.prefixSum(idx); got != want {
				t.Fatalf("op %d PrefixSum(%d): got %d, want %d", i, idx, got, want)
			}
		case 2:
			a, b := rng.IntN(n), rng.IntN(n)
			if a > b {
				a, b = b, a
			}
			if got, want := f.RangeSum(a, b), ref.rangeSum(a, b); got != want {
				t.Fatalf("op %d RangeSum(%d,%d): got %d, want %d", i, a, b, got, want)
			}
		}
	}
}

// TestCorrectness_Stress 在 n=1_000_000 上执行 100k 操作，验证百万规模下的正确性。
func TestCorrectness_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过百万规模压力测试")
	}
	const (
		n   = 1_000_000
		ops = 100_000
	)
	rng := rand.New(rand.NewPCG(777, 0))
	f := ft.New[int](n)
	ref := newRefFenwick(n)

	for i := range ops {
		switch rng.IntN(3) {
		case 0:
			idx := rng.IntN(n)
			delta := rng.IntN(201) - 100
			f.Add(idx, delta)
			ref.add(idx, delta)
		case 1:
			idx := rng.IntN(n)
			if got, want := f.PrefixSum(idx), ref.prefixSum(idx); got != want {
				t.Fatalf("op %d PrefixSum(%d): got %d, want %d (n=%d)", i, idx, got, want, n)
			}
		case 2:
			a, b := rng.IntN(n), rng.IntN(n)
			if a > b {
				a, b = b, a
			}
			if got, want := f.RangeSum(a, b), ref.rangeSum(a, b); got != want {
				t.Fatalf("op %d RangeSum(%d,%d): got %d, want %d (n=%d)", i, a, b, got, want, n)
			}
		}
	}
}

// TestCorrectness_LargeRandom 在不同规模下执行 100k 操作，验证不变量。
func TestCorrectness_LargeRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模正确性测试")
	}
	for _, n := range []int{100, 1_000, 10_000} {
		n := n
		t.Run(itoa(n), func(t *testing.T) {
			const ops = 100_000
			rng := rand.New(rand.NewPCG(uint64(n), 0))
			f := ft.New[int](n)
			ref := newRefFenwick(n)
			for i := range ops {
				switch rng.IntN(3) {
				case 0:
					idx := rng.IntN(n)
					delta := rng.IntN(201) - 100
					f.Add(idx, delta)
					ref.add(idx, delta)
				case 1:
					idx := rng.IntN(n)
					if got, want := f.PrefixSum(idx), ref.prefixSum(idx); got != want {
						t.Fatalf("op %d PrefixSum(%d): got %d, want %d (n=%d)", i, idx, got, want, n)
					}
				case 2:
					a, b := rng.IntN(n), rng.IntN(n)
					if a > b {
						a, b = b, a
					}
					if got, want := f.RangeSum(a, b), ref.rangeSum(a, b); got != want {
						t.Fatalf("op %d RangeSum(%d,%d): got %d, want %d (n=%d)", i, a, b, got, want, n)
					}
				}
			}
		})
	}
}
