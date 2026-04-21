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
	"math/rand/v2"
	"testing"

	st "github.com/motocat46/yytools/pkg/ds/segtree"
)

// refSegTree 是 SegTree 的参考模型：朴素 []T + O(N) 操作，用于随机测试对比。
type refSegTree[T, L any] struct {
	data     []T
	identity T
	merge    func(T, T) T
	applyFn  func(T, L, int) T
}

func newRefSegTree[T, L any](n int, identity T, merge func(T, T) T, applyFn func(T, L, int) T) *refSegTree[T, L] {
	data := make([]T, n)
	for i := range data {
		data[i] = identity
	}
	return &refSegTree[T, L]{data: data, identity: identity, merge: merge, applyFn: applyFn}
}

func (r *refSegTree[T, L]) set(i int, val T) { r.data[i] = val }

func (r *refSegTree[T, L]) applyRange(l, right int, lazy L) {
	for i := l; i <= right; i++ {
		r.data[i] = r.applyFn(r.data[i], lazy, 1)
	}
}

func (r *refSegTree[T, L]) query(l, right int) T {
	res := r.identity
	for i := l; i <= right; i++ {
		res = r.merge(res, r.data[i])
	}
	return res
}

// sortedPair 生成 [0,n) 内的有序对 (l, r)，l <= r。
func sortedPair(rng *rand.Rand, n int) (int, int) {
	a, b := rng.IntN(n), rng.IntN(n)
	if a > b {
		a, b = b, a
	}
	return a, b
}

// TestCorrectness_RangeAddSum 在 n=200 上执行 100k 随机 Set/Apply/Query，与参考模型对比。
// Monoid：区间加法 + 区间求和。
func TestCorrectness_RangeAddSum(t *testing.T) {
	const (
		n   = 200
		ops = 100_000
	)
	mergeFn := func(a, b int) int { return a + b }
	applyFn := func(val, lazy, size int) int { return val + lazy*size }
	composeFn := func(newL, oldL int) int { return newL + oldL }

	rng := rand.New(rand.NewPCG(42, 0))
	s := st.New[int, int](n, 0, mergeFn, 0, applyFn, composeFn)
	ref := newRefSegTree[int, int](n, 0, mergeFn, applyFn)

	for i := range ops {
		switch rng.IntN(3) {
		case 0:
			idx := rng.IntN(n)
			val := rng.IntN(201) - 100
			s.Set(idx, val)
			ref.set(idx, val)
		case 1:
			l, r := sortedPair(rng, n)
			lazy := rng.IntN(21) - 10
			s.Apply(l, r, lazy)
			ref.applyRange(l, r, lazy)
		case 2:
			l, r := sortedPair(rng, n)
			got, want := s.Query(l, r), ref.query(l, r)
			if got != want {
				t.Fatalf("op %d Query(%d,%d): got %d, want %d", i, l, r, got, want)
			}
		}
	}
}

// TestCorrectness_RangeAssignMin 在 n=200 上执行 100k 随机操作，与参考模型对比。
// Monoid：区间赋值 + 区间最小值。assignLazy 定义在 segtree_test.go。
func TestCorrectness_RangeAssignMin(t *testing.T) {
	const (
		n   = 200
		ops = 100_000
	)
	mergeFn := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	applyFn := func(val int, lazy assignLazy, _ int) int {
		if lazy.hasVal {
			return lazy.val
		}
		return val
	}
	composeFn := func(newL, oldL assignLazy) assignLazy {
		if newL.hasVal {
			return newL
		}
		return oldL
	}

	rng := rand.New(rand.NewPCG(123, 0))
	s := st.New[int, assignLazy](n, 1<<62, mergeFn, assignLazy{}, applyFn, composeFn)
	ref := newRefSegTree[int, assignLazy](n, 1<<62, mergeFn, applyFn)

	for i := range ops {
		switch rng.IntN(3) {
		case 0:
			idx := rng.IntN(n)
			val := rng.IntN(201) - 100
			s.Set(idx, val)
			ref.set(idx, val)
		case 1:
			l, r := sortedPair(rng, n)
			var lazy assignLazy
			if rng.IntN(2) == 0 {
				lazy = assignLazy{val: rng.IntN(201) - 100, hasVal: true}
			}
			s.Apply(l, r, lazy)
			ref.applyRange(l, r, lazy)
		case 2:
			l, r := sortedPair(rng, n)
			got, want := s.Query(l, r), ref.query(l, r)
			if got != want {
				t.Fatalf("op %d Query(%d,%d): got %d, want %d", i, l, r, got, want)
			}
		}
	}
}

// TestCorrectness_LargeRandom 在不同规模下执行 100k 操作，验证不变量（-short 跳过）。
func TestCorrectness_LargeRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模正确性测试")
	}
	mergeFn := func(a, b int) int { return a + b }
	applyFn := func(val, lazy, size int) int { return val + lazy*size }
	composeFn := func(newL, oldL int) int { return newL + oldL }

	for _, n := range []int{100, 1_000, 10_000} {
		n := n
		t.Run(itoa(n), func(t *testing.T) {
			const ops = 100_000
			rng := rand.New(rand.NewPCG(uint64(n), 0))
			s := st.New[int, int](n, 0, mergeFn, 0, applyFn, composeFn)
			ref := newRefSegTree[int, int](n, 0, mergeFn, applyFn)

			for i := range ops {
				switch rng.IntN(3) {
				case 0:
					idx := rng.IntN(n)
					val := rng.IntN(201) - 100
					s.Set(idx, val)
					ref.set(idx, val)
				case 1:
					l, r := sortedPair(rng, n)
					lazy := rng.IntN(21) - 10
					s.Apply(l, r, lazy)
					ref.applyRange(l, r, lazy)
				case 2:
					l, r := sortedPair(rng, n)
					got, want := s.Query(l, r), ref.query(l, r)
					if got != want {
						t.Fatalf("op %d Query(%d,%d): got %d, want %d (n=%d)", i, l, r, got, want, n)
					}
				}
			}
		})
	}
}

// TestCorrectness_Stress 在 n=1_000_000 上执行 100k 操作，验证百万规模下的正确性（-short 跳过）。
func TestCorrectness_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过百万规模压力测试")
	}
	const (
		n   = 1_000_000
		ops = 100_000
	)
	mergeFn := func(a, b int) int { return a + b }
	applyFn := func(val, lazy, size int) int { return val + lazy*size }
	composeFn := func(newL, oldL int) int { return newL + oldL }

	rng := rand.New(rand.NewPCG(777, 0))
	s := st.New[int, int](n, 0, mergeFn, 0, applyFn, composeFn)
	ref := newRefSegTree[int, int](n, 0, mergeFn, applyFn)

	for i := range ops {
		switch rng.IntN(3) {
		case 0:
			idx := rng.IntN(n)
			val := rng.IntN(201) - 100
			s.Set(idx, val)
			ref.set(idx, val)
		case 1:
			l, r := sortedPair(rng, n)
			lazy := rng.IntN(21) - 10
			s.Apply(l, r, lazy)
			ref.applyRange(l, r, lazy)
		case 2:
			l, r := sortedPair(rng, n)
			got, want := s.Query(l, r), ref.query(l, r)
			if got != want {
				t.Fatalf("op %d Query(%d,%d): got %d, want %d (n=%d)", i, l, r, got, want, n)
			}
		}
	}
}
