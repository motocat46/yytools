// 版权所有(Copyright)[yangyuan]
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

// 作者:  yangyuan
// 创建日期:2026/4/15
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pd "github.com/motocat46/yytools/pkg/algorithms/mathx/probability_distribution"
)

const pdTrials = 100_000

func handleAlgoProbDist(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：命中率准确性验证
	// 权重 [1,2,4,8,16]，总权重 31，理论命中率分别为 3.2%/6.5%/12.9%/25.8%/51.6%
	weights1 := []int{1, 2, 4, 8, 16}
	total1 := 0
	for _, w := range weights1 {
		total1 += w
	}

	nm1 := pd.NewNormalMethod(weights1)
	actualHits := make([]int64, len(weights1))
	for range pdTrials {
		actualHits[nm1.Generate()]++
	}

	xLabels1 := make([]string, len(weights1))
	theoreticalHits := make([]int64, len(weights1))
	for i, wt := range weights1 {
		xLabels1[i] = fmt.Sprintf("w=%d(%.1f%%)", wt, float64(wt)*100/float64(total1))
		theoreticalHits[i] = int64(wt) * pdTrials / int64(total1)
	}

	// Chart 2：三种实现生成耗时 vs 权重项数
	const opsPerMeasure = 10_000
	sizes := []int{10, 50, 100, 500, 1000, 5000, 10000}
	xLabels2 := make([]string, len(sizes))
	linearNs := make([]int64, len(sizes))  // CalcIndexByWeight O(n)
	logNs := make([]int64, len(sizes))     // NormalMethod O(log n)
	constNs := make([]int64, len(sizes))   // VoseAliasMethod O(1)

	for i, n := range sizes {
		xLabels2[i] = fmt.Sprintf("%d", n)

		weights := make([]int, n)
		total := 0
		for j := range n {
			weights[j] = j + 1
			total += j + 1
		}

		// O(n)：CalcIndexByWeight 线性遍历
		start := time.Now()
		for range opsPerMeasure {
			pd.CalcIndexByWeight(weights, total)
		}
		linearNs[i] = time.Since(start).Nanoseconds() / opsPerMeasure

		// O(log n)：NormalMethod 前缀和 + 二分搜索
		nm := pd.NewNormalMethod(weights)
		start = time.Now()
		for range opsPerMeasure {
			nm.Generate()
		}
		logNs[i] = time.Since(start).Nanoseconds() / opsPerMeasure

		// O(1)：VoseAliasMethod 别名方法
		vm := pd.NewVoseAliasMethod(weights)
		start = time.Now()
		for range opsPerMeasure {
			vm.Generate()
		}
		constNs[i] = time.Since(start).Nanoseconds() / opsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "概率分布 — 准确性与性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("命中率准确性验证（权重 %v，总权重 %d，%d 万次）", weights1, total1, pdTrials/10000),
				XAxis:     xLabels1,
				XAxisName: "权重项",
				YAxisName: "命中次数",
				Series: []chartSeries{
					{Name: "理论命中次数", Data: theoreticalHits},
					{Name: "实际命中次数（NormalMethod）", Data: actualHits},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("三种实现生成耗时 vs 权重项数（每规模 %d 次均值）", opsPerMeasure),
				XAxis:     xLabels2,
				XAxisName: "权重项数",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "CalcIndexByWeight（O(n) 线性遍历）", Data: linearNs},
					{Name: "NormalMethod（O(log n) 前缀和+二分）", Data: logNs},
					{Name: "VoseAliasMethod（O(1) 别名法）", Data: constNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/probability_distribution/", Title: "概率分布 — 准确性与性能",
		Desc: "命中率准确性验证（10万次）；O(n)/O(log n)/O(1) 三种实现生成耗时 vs 权重项数对比",
		Path: "/api/algo/prob_dist", DataHandler: handleAlgoProbDist,
	})
}
