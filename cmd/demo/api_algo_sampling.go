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
	"math/rand/v2"
	"net/http"
	"time"

	sampling_pkg "github.com/motocat46/yytools/pkg/algorithms/mathx/sampling"
)

func handleAlgoSampling(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：均匀性验证
	// range [0,19]，k=10，重复 50000 次，统计各值命中频率
	const (
		uniformLo    = 0
		uniformHi    = 19
		uniformK     = 10
		uniformTrials = 50_000
	)
	hitCount := make([]int64, uniformHi-uniformLo+1)
	rng1 := rand.New(rand.NewPCG(42, 0))
	for range uniformTrials {
		for _, v := range sampling_pkg.SampleKDistinctFloyd(uniformLo, uniformHi, uniformK, rng1) {
			hitCount[v-uniformLo]++
		}
	}
	xLabels1 := make([]string, uniformHi-uniformLo+1)
	for i := range xLabels1 {
		xLabels1[i] = fmt.Sprintf("%d", uniformLo+i)
	}

	// Chart 2：耗时 vs N（固定 k=1000，N 从 1000 到 10 亿）
	// 展示 O(k) 特性：时间只取决于 k，与范围大小 N 无关
	const fixedK = 1000
	const samplingOpsPerMeasure = 200
	rangeSizes := []int{1_000, 10_000, 100_000, 1_000_000, 10_000_000, 100_000_000, 1_000_000_000}
	xLabels2 := []string{"1千", "1万", "10万", "100万", "1千万", "1亿", "10亿"}
	timeNs := make([]int64, len(rangeSizes))

	rng2 := rand.New(rand.NewPCG(99, 0))
	for i, n := range rangeSizes {
		start := time.Now()
		for range samplingOpsPerMeasure {
			sampling_pkg.SampleKDistinctFloyd(0, n-1, fixedK, rng2)
		}
		timeNs[i] = time.Since(start).Nanoseconds() / samplingOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Sampling 采样均匀性与 O(k) 特性",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("采样均匀性验证：range [%d,%d]，k=%d，重复 %d 次各值命中次数（期望 %d）", uniformLo, uniformHi, uniformK, uniformTrials, int64(uniformTrials)*uniformK/(uniformHi-uniformLo+1)),
				XAxis:     xLabels1,
				XAxisName: "采样值",
				YAxisName: "命中次数",
				Series: []chartSeries{
					{Name: "命中次数", Data: hitCount},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("SampleKDistinctFloyd 耗时 vs 范围大小 N（固定 k=%d，每点 %d 次均值）", fixedK, samplingOpsPerMeasure),
				XAxis:     xLabels2,
				XAxisName: "范围大小 N",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: fmt.Sprintf("SampleKDistinctFloyd（k=%d）", fixedK), Data: timeNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/sampling/", Title: "Sampling 均匀性与 O(k) 特性",
		Desc: "Floyd 采样均匀性验证（50000次）；耗时 vs 范围大小 N（k=1000，N 从 1千到 10亿）",
		Path: "/api/algo/sampling", DataHandler: handleAlgoSampling,
	})
}
