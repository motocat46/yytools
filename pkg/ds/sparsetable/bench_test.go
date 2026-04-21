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

package sparsetable_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	st "github.com/motocat46/yytools/pkg/ds/sparsetable"
)

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

type benchRange struct {
	l int
	r int
}

func makeBenchRanges(n, count int, seedA, seedB uint64) []benchRange {
	rng := rand.New(rand.NewPCG(seedA, seedB))
	ranges := make([]benchRange, count)
	for i := range ranges {
		l := rng.IntN(n)
		r := l + rng.IntN(n-l)
		ranges[i] = benchRange{l: l, r: r}
	}
	return ranges
}

// BenchmarkSparseTable_Build 测量 SparseTable 的 O(n log n) 预处理代价。
func BenchmarkSparseTable_Build(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			rng := rand.New(rand.NewPCG(42, 0))
			data := make([]int, n)
			for i := range data {
				data[i] = rng.IntN(1_000_000)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = st.New(data, minInt)
			}
		})
	}
}

// BenchmarkSparseTable_Query 测量 Query 的 O(1) 查询代价。
func BenchmarkSparseTable_Query(b *testing.B) {
	const queryCount = 1 << 16

	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			rng := rand.New(rand.NewPCG(42, 0))
			data := make([]int, n)
			for i := range data {
				data[i] = rng.IntN(1_000_000)
			}
			table := st.New(data, minInt)
			ranges := makeBenchRanges(n, queryCount, 43, uint64(n))

			b.ResetTimer()
			b.ReportAllocs()
			idx := 0
			for b.Loop() {
				q := ranges[idx]
				_ = table.Query(q.l, q.r)
				idx++
				if idx == len(ranges) {
					idx = 0
				}
			}
		})
	}
}

// BenchmarkSparseTable_QueryVsNaive 对比 O(1) Query 与 O(n) 朴素扫描。
func BenchmarkSparseTable_QueryVsNaive(b *testing.B) {
	const n = 100_000
	const queryCount = 1 << 16

	rng := rand.New(rand.NewPCG(42, 0))
	data := make([]int, n)
	for i := range data {
		data[i] = rng.IntN(1_000_000)
	}
	table := st.New(data, minInt)
	ranges := makeBenchRanges(n, queryCount, 43, 0)

	b.Run("SparseTable_O1", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		idx := 0
		for b.Loop() {
			q := ranges[idx]
			_ = table.Query(q.l, q.r)
			idx++
			if idx == len(ranges) {
				idx = 0
			}
		}
	})

	b.Run("Naive_On", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		idx := 0
		for b.Loop() {
			q := ranges[idx]
			_ = refQuery(data, minInt, q.l, q.r)
			idx++
			if idx == len(ranges) {
				idx = 0
			}
		}
	})
}
