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
// 创建日期:2026/4/20
package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	fenwick "github.com/motocat46/yytools/pkg/ds/fenwicktree"
)

const fenwickOpsPerMeasure = 10_000

func fenwickInitRepeats(n int) int {
	repeats := 2_000_000 / n
	if repeats < 5 {
		return 5
	}
	if repeats > 1000 {
		return 1000
	}
	return repeats
}

func handleDsFenwickTree(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{100, 1_000, 10_000, 100_000, 1_000_000}
	xLabels := make([]string, len(sizes))
	addNs := make([]int64, len(sizes))
	prefixNs := make([]int64, len(sizes))
	rangeNs := make([]int64, len(sizes))
	naiveNs := make([]int64, len(sizes))
	buildNs := make([]int64, len(sizes))
	newAddNs := make([]int64, len(sizes))

	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d", n)
		rng := rand.New(rand.NewPCG(42, 0))
		vals := make([]int, n)
		for j := range n {
			vals[j] = rng.IntN(100) + 1
		}
		f := fenwick.Build(vals)

		initRepeats := fenwickInitRepeats(n)
		start := time.Now()
		for range initRepeats {
			_ = fenwick.Build(vals)
		}
		buildNs[i] = time.Since(start).Nanoseconds() / int64(initRepeats)

		start = time.Now()
		for range initRepeats {
			f2 := fenwick.New[int](n)
			for j, v := range vals {
				f2.Add(j, v)
			}
		}
		newAddNs[i] = time.Since(start).Nanoseconds() / int64(initRepeats)

		start = time.Now()
		for range fenwickOpsPerMeasure {
			f.Add(rng.IntN(n), 1)
		}
		addNs[i] = time.Since(start).Nanoseconds() / fenwickOpsPerMeasure

		start = time.Now()
		for range fenwickOpsPerMeasure {
			_ = f.PrefixSum(rng.IntN(n))
		}
		prefixNs[i] = time.Since(start).Nanoseconds() / fenwickOpsPerMeasure

		start = time.Now()
		for range fenwickOpsPerMeasure {
			a, b := rng.IntN(n), rng.IntN(n)
			if a > b {
				a, b = b, a
			}
			_ = f.RangeSum(a, b)
		}
		rangeNs[i] = time.Since(start).Nanoseconds() / fenwickOpsPerMeasure

		start = time.Now()
		for range fenwickOpsPerMeasure {
			idx := rng.IntN(n)
			s := 0
			for j := 0; j <= idx; j++ {
				s += vals[j]
			}
			_ = s
		}
		naiveNs[i] = time.Since(start).Nanoseconds() / fenwickOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "FenwickTree 性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Add / PrefixSum / RangeSum 耗时 vs 规模（每规模 %d 次）", fenwickOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Add（O(log n)）", Data: addNs},
					{Name: "PrefixSum（O(log n)）", Data: prefixNs},
					{Name: "RangeSum（O(log n)）", Data: rangeNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("PrefixSum O(log n) vs 朴素 O(N) 扫描（每规模 %d 次）", fenwickOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "FenwickTree.PrefixSum()（O(log n)）", Data: prefixNs},
					{Name: "朴素全数组扫描（O(N)，线性增长）", Data: naiveNs},
				},
			},
			{
				Type:      "line",
				Title:     "初始化耗时：Build O(n) vs New+Add O(n log n)（每规模自适应次数）",
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Build(nums)（O(n)）", Data: buildNs},
					{Name: "New()+逐个 Add（O(n log n)）", Data: newAddNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "fenwicktree/", Title: "FenwickTree 性能",
		Desc:        "Add/PrefixSum/RangeSum 耗时 vs 规模；O(log n) vs 朴素 O(N) 前缀扫描对比",
		Path:        "/api/ds/fenwicktree",
		DataHandler: handleDsFenwickTree,
	})
}
