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

package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	st "github.com/motocat46/yytools/pkg/ds/sparsetable"
)

const spOpsPerMeasure = 10_000

func spInitRepeats(n int) int {
	repeats := 500_000 / n
	if repeats < 3 {
		return 3
	}
	if repeats > 500 {
		return 500
	}
	return repeats
}

func handleDsSparseTable(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{100, 1_000, 10_000, 100_000, 1_000_000}
	xLabels := make([]string, len(sizes))
	buildNs := make([]int64, len(sizes))
	queryNs := make([]int64, len(sizes))
	naiveNs := make([]int64, len(sizes))

	minFn := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d", n)
		rng := rand.New(rand.NewPCG(42, 0))
		data := make([]int, n)
		for j := range data {
			data[j] = rng.IntN(1_000_000)
		}

		initRepeats := spInitRepeats(n)
		start := time.Now()
		for range initRepeats {
			_ = st.New(data, minFn)
		}
		buildNs[i] = time.Since(start).Nanoseconds() / int64(initRepeats)

		table := st.New(data, minFn)

		start = time.Now()
		for range spOpsPerMeasure {
			l := rng.IntN(n)
			r := l + rng.IntN(n-l)
			_ = table.Query(l, r)
		}
		queryNs[i] = time.Since(start).Nanoseconds() / spOpsPerMeasure

		start = time.Now()
		for range spOpsPerMeasure {
			l := rng.IntN(n)
			r := l + rng.IntN(n-l)
			res := data[l]
			for k := l + 1; k <= r; k++ {
				if data[k] < res {
					res = data[k]
				}
			}
			_ = res
		}
		naiveNs[i] = time.Since(start).Nanoseconds() / spOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "SparseTable 性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     "Build 耗时 vs 规模（每规模自适应次数，O(n log n)）",
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Build（O(n log n)）", Data: buildNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("Query O(1) vs 朴素 O(n) 扫描（每规模 %d 次）", spOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "SparseTable.Query()（O(1)）", Data: queryNs},
					{Name: "朴素全数组扫描（O(n)，线性增长）", Data: naiveNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg:         "pkg/ds",
		SubPkg:      "sparsetable/",
		Title:       "SparseTable 性能",
		Desc:        "Build O(n log n) 耗时 vs 规模；Query O(1) vs 朴素 O(n) 扫描对比",
		Path:        "/api/ds/sparsetable",
		DataHandler: handleDsSparseTable,
	})
}
