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
	"fmt"
	"math/rand/v2"
	"testing"

	st "github.com/motocat46/yytools/pkg/ds/segtree"
)

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// newBenchTree 创建区间加法 + 区间求和线段树，用于基准测试。
func newBenchTree(n int) *st.SegTree[int, int] {
	return st.New[int, int](n, 0,
		func(a, b int) int { return a + b },
		0,
		func(val, lazy, size int) int { return val + lazy*size },
		func(newL, oldL int) int { return newL + oldL },
	)
}

// BenchmarkSegTree_Set 测量单点赋值的 O(log n) 代价。
func BenchmarkSegTree_Set(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			s := newBenchTree(n)
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				s.Set(rng.IntN(n), 1)
			}
		})
	}
}

// BenchmarkSegTree_Apply 测量区间 lazy 更新的 O(log n) 代价。
func BenchmarkSegTree_Apply(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			s := newBenchTree(n)
			for i := range n {
				s.Set(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				l, r := rng.IntN(n), rng.IntN(n)
				if l > r {
					l, r = r, l
				}
				s.Apply(l, r, 1)
			}
		})
	}
}

// BenchmarkSegTree_Query 测量区间查询的 O(log n) 代价。
func BenchmarkSegTree_Query(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			s := newBenchTree(n)
			for i := range n {
				s.Set(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				l, r := rng.IntN(n), rng.IntN(n)
				if l > r {
					l, r = r, l
				}
				_ = s.Query(l, r)
			}
		})
	}
}

// BenchmarkSegTree_Mixed 混合负载：1/3 Set + 1/3 Apply + 1/3 Query。
// 模拟游戏帧循环：伤害计算（Set）+ 范围 buff（Apply）+ 排行榜区间查询（Query）各占 1/3。
func BenchmarkSegTree_Mixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			s := newBenchTree(n)
			for i := range n {
				s.Set(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				switch rng.IntN(3) {
				case 0:
					s.Set(rng.IntN(n), 1)
				case 1:
					l, r := rng.IntN(n), rng.IntN(n)
					if l > r {
						l, r = r, l
					}
					s.Apply(l, r, 1)
				case 2:
					l, r := rng.IntN(n), rng.IntN(n)
					if l > r {
						l, r = r, l
					}
					_ = s.Query(l, r)
				}
			}
		})
	}
}
