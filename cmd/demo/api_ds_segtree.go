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

	seg "github.com/motocat46/yytools/pkg/ds/segtree"
)

const segOpsPerMeasure = 10_000

func handleDsSegTree(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{100, 1_000, 10_000, 100_000, 1_000_000}
	xLabels := make([]string, len(sizes))
	setNs := make([]int64, len(sizes))
	applyNs := make([]int64, len(sizes))
	queryNs := make([]int64, len(sizes))
	naiveNs := make([]int64, len(sizes))

	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d", n)
		rng := rand.New(rand.NewPCG(42, 0))

		s := seg.New[int, int](n, 0,
			func(a, b int) int { return a + b },
			0,
			func(val, lazy, size int) int { return val + lazy*size },
			func(newL, oldL int) int { return newL + oldL },
		)
		vals := make([]int, n)
		for j := range n {
			v := rng.IntN(100) + 1
			vals[j] = v
			s.Set(j, v)
		}

		start := time.Now()
		for range segOpsPerMeasure {
			s.Set(rng.IntN(n), 1)
		}
		setNs[i] = time.Since(start).Nanoseconds() / segOpsPerMeasure

		start = time.Now()
		for range segOpsPerMeasure {
			l, r := rng.IntN(n), rng.IntN(n)
			if l > r {
				l, r = r, l
			}
			s.Apply(l, r, 1)
		}
		applyNs[i] = time.Since(start).Nanoseconds() / segOpsPerMeasure

		start = time.Now()
		for range segOpsPerMeasure {
			l, r := rng.IntN(n), rng.IntN(n)
			if l > r {
				l, r = r, l
			}
			_ = s.Query(l, r)
		}
		queryNs[i] = time.Since(start).Nanoseconds() / segOpsPerMeasure

		start = time.Now()
		for range segOpsPerMeasure {
			l, r := rng.IntN(n), rng.IntN(n)
			if l > r {
				l, r = r, l
			}
			sum := 0
			for j := l; j <= r; j++ {
				sum += vals[j]
			}
			_ = sum
		}
		naiveNs[i] = time.Since(start).Nanoseconds() / segOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "SegTree 性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Set / Apply / Query 耗时 vs 规模（每规模 %d 次）", segOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Set（O(log n)）", Data: setNs},
					{Name: "Apply（O(log n)）", Data: applyNs},
					{Name: "Query（O(log n)）", Data: queryNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("Query O(log n) vs 朴素 O(N) 扫描（每规模 %d 次）", segOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "SegTree.Query()（O(log n)）", Data: queryNs},
					{Name: "朴素全数组扫描（O(N)，线性增长）", Data: naiveNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "segtree/", Title: "SegTree 性能",
		Desc:        "Set/Apply/Query 耗时 vs 规模；O(log n) vs 朴素 O(N) 区间扫描对比",
		Path:        "/api/ds/segtree",
		DataHandler: handleDsSegTree,
	})
}
