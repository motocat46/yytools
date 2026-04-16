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
// 创建日期:2026/4/16
package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"slices"
	"time"

	"github.com/motocat46/yytools/pkg/slicex"
)

const slicexOpsPerMeasure = 100

// makeRandSlice 生成 n 个随机 int 组成的切片，使用固定种子保证可复现。
func makeRandSlice(n int) []int {
	rng := rand.New(rand.NewPCG(42, 0))
	s := make([]int, n)
	for i := range n {
		s[i] = rng.IntN(1_000_000_000)
	}
	return s
}

func handleSlicex(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 100000, 1000000, 5000000, 10000000}
	xLabels := make([]string, len(sizes))
	minInSliceNs := make([]int64, len(sizes))
	maxInSliceNs := make([]int64, len(sizes))
	minByNs := make([]int64, len(sizes))
	slicesMinNs := make([]int64, len(sizes))

	for i, n := range sizes {
		if n >= 1_000_000 {
			xLabels[i] = fmt.Sprintf("%d万", n/10000)
		} else {
			xLabels[i] = fmt.Sprintf("%d", n)
		}
		data := makeRandSlice(n)

		start := time.Now()
		for range slicexOpsPerMeasure {
			slicex.MinInSlice(data)
		}
		minInSliceNs[i] = time.Since(start).Nanoseconds() / slicexOpsPerMeasure

		start = time.Now()
		for range slicexOpsPerMeasure {
			slicex.MaxInSlice(data)
		}
		maxInSliceNs[i] = time.Since(start).Nanoseconds() / slicexOpsPerMeasure

		start = time.Now()
		for range slicexOpsPerMeasure {
			slicex.MinBy(data, func(a, b int) bool { return a < b })
		}
		minByNs[i] = time.Since(start).Nanoseconds() / slicexOpsPerMeasure

		// 标准库对比：slices.Min 只返回值，不返回下标
		start = time.Now()
		for range slicexOpsPerMeasure {
			slices.Min(data)
		}
		slicesMinNs[i] = time.Since(start).Nanoseconds() / slicexOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "slicex 切片工具性能",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("MinInSlice / MaxInSlice / MinBy vs slices.Min 耗时 vs 切片规模（每规模 %d 次均值）", slicexOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "切片规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "MinInSlice（返回下标+值）", Data: minInSliceNs},
				{Name: "MaxInSlice（返回下标+值）", Data: maxInSliceNs},
				{Name: "MinBy（自定义比较器）", Data: minByNs},
				{Name: "slices.Min（仅返回值，标准库）", Data: slicesMinNs},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/slicex", SubPkg: "", Title: "slicex 切片工具性能",
		Desc: "MinInSlice / MaxInSlice / MinBy 与标准库 slices.Min 耗时对比（1万~1000万，100次均值）",
		Path: "/api/slicex", DataHandler: handleSlicex,
	})
}
