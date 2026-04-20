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

	sw "github.com/motocat46/yytools/pkg/ds/slidingwindow"
)

const swOpsPerMeasure = 1000

func handleDsSlidingWindow(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：Add 均摊耗时 vs 窗口大小（稳态：窗口已满）
	windowSizes := []int{10, 100, 1000, 10000, 100000}
	xLabels := make([]string, len(windowSizes))
	addNs := make([]int64, len(windowSizes))

	for i, n := range windowSizes {
		xLabels[i] = fmt.Sprintf("%d", n)
		win := sw.New[int](n)
		for j := range n {
			win.Add(j) // 预填充
		}
		start := time.Now()
		for j := range swOpsPerMeasure {
			win.Add(n + j)
		}
		addNs[i] = time.Since(start).Nanoseconds() / swOpsPerMeasure
	}

	// Chart 2：Max vs 朴素 O(N) 扫描耗时 vs 窗口大小
	maxNs := make([]int64, len(windowSizes))
	naiveMaxNs := make([]int64, len(windowSizes))

	for i, n := range windowSizes {
		rng := rand.New(rand.NewPCG(42, 0))
		win := sw.New[int](n)
		vals := make([]int, n)
		for j := range n {
			v := rng.IntN(1_000_000)
			vals[j] = v
			win.Add(v)
		}

		// SlidingWindow Max O(1)
		start := time.Now()
		for range swOpsPerMeasure {
			_ = win.Max()
		}
		maxNs[i] = time.Since(start).Nanoseconds() / swOpsPerMeasure

		// 朴素 O(N) 全窗口扫描
		start = time.Now()
		for range swOpsPerMeasure {
			m := vals[0]
			for _, v := range vals[1:] {
				if v > m {
					m = v
				}
			}
			_ = m
		}
		naiveMaxNs[i] = time.Since(start).Nanoseconds() / swOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "SlidingWindow 性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Add 均摊耗时 vs 窗口大小（稳态满窗口，每规模 %d 次）", swOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "窗口大小",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Add（O(1) 均摊，与规模无关）", Data: addNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("Max 查询耗时 vs 窗口大小（每规模 %d 次）", swOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "窗口大小",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "SlidingWindow.Max()（O(1) 均摊）", Data: maxNs},
					{Name: "朴素全窗口扫描（O(N)，线性增长）", Data: naiveMaxNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "slidingwindow/", Title: "SlidingWindow 性能",
		Desc:        "Add 均摊耗时 vs 窗口大小（O(1) 稳态）；Max O(1) vs 朴素 O(N) 全窗口扫描对比",
		Path:        "/api/ds/slidingwindow",
		DataHandler: handleDsSlidingWindow,
	})
}
