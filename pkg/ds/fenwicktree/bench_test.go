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
	"fmt"
	"math/rand/v2"
	"testing"

	ft "github.com/motocat46/yytools/pkg/ds/fenwicktree"
)

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// BenchmarkFenwickTree_Build 测量 Build 的 O(n) 建树代价。
func BenchmarkFenwickTree_Build(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			nums := make([]int, n)
			rng := rand.New(rand.NewPCG(42, 0))
			for i := range n {
				nums[i] = rng.IntN(1000) + 1
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = ft.Build(nums)
			}
		})
	}
}

// BenchmarkFenwickTree_Add 测量 Add 的 O(log n) 代价。
func BenchmarkFenwickTree_Add(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			f := ft.New[int](n)
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				f.Add(rng.IntN(n), 1)
			}
		})
	}
}

// BenchmarkFenwickTree_PrefixSum 测量 PrefixSum 的 O(log n) 代价。
func BenchmarkFenwickTree_PrefixSum(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			f := ft.New[int](n)
			for i := range n {
				f.Add(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = f.PrefixSum(rng.IntN(n))
			}
		})
	}
}

// BenchmarkFenwickTree_RangeSum 测量 RangeSum 的 O(log n) 代价。
func BenchmarkFenwickTree_RangeSum(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			f := ft.New[int](n)
			for i := range n {
				f.Add(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				a, b2 := rng.IntN(n), rng.IntN(n)
				if a > b2 {
					a, b2 = b2, a
				}
				_ = f.RangeSum(a, b2)
			}
		})
	}
}

// BenchmarkFenwickTree_Mixed 混合负载：1/3 Add + 1/3 PrefixSum + 1/3 RangeSum。
func BenchmarkFenwickTree_Mixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			f := ft.New[int](n)
			for i := range n {
				f.Add(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				switch rng.IntN(3) {
				case 0:
					f.Add(rng.IntN(n), 1)
				case 1:
					_ = f.PrefixSum(rng.IntN(n))
				case 2:
					a, b2 := rng.IntN(n), rng.IntN(n)
					if a > b2 {
						a, b2 = b2, a
					}
					_ = f.RangeSum(a, b2)
				}
			}
		})
	}
}
