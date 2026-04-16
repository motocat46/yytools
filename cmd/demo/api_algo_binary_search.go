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
	"net/http"
	"time"

	bs "github.com/motocat46/yytools/pkg/algorithms/binary_search"
)

// linearSearch 在切片中线性查找 target，返回第一个匹配的下标，不存在返回 -1。
func linearSearch(nums []int, target int) int {
	for i, v := range nums {
		if v == target {
			return i
		}
	}
	return -1
}

// makeSortedSlice 生成 n 个偶数组成的升序切片：[0, 2, 4, ..., 2*(n-1)]。
func makeSortedSlice(n int) []int {
	s := make([]int, n)
	for i := range n {
		s[i] = i * 2
	}
	return s
}

const bsOpsPerMeasure = 10_000

func handleAlgoBinarySearch(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：BinarySearch vs 线性搜索（小规模，直观对比 O(log n) vs O(n)）
	compSizes := []int{1000, 5000, 10000, 50000, 100000}
	xLabels1 := make([]string, len(compSizes))
	bsCompNs := make([]int64, len(compSizes))
	linearNs := make([]int64, len(compSizes))

	const compOps = 1000

	for i, n := range compSizes {
		xLabels1[i] = fmt.Sprintf("%d", n)
		nums := makeSortedSlice(n)
		target := nums[n/2] // 中间元素：线性搜索平均情况，二分搜索 O(log n)

		start := time.Now()
		for range compOps {
			bs.BinarySearch(nums, target)
		}
		bsCompNs[i] = time.Since(start).Nanoseconds() / compOps

		start = time.Now()
		for range compOps {
			linearSearch(nums, target)
		}
		linearNs[i] = time.Since(start).Nanoseconds() / compOps
	}

	// Chart 2：BinarySearch / LeftBound / RightBound 大规模性能对比
	variantSizes := []int{10000, 100000, 1000000, 5000000, 10000000}
	xLabels2 := make([]string, len(variantSizes))
	bsNs := make([]int64, len(variantSizes))
	leftNs := make([]int64, len(variantSizes))
	rightNs := make([]int64, len(variantSizes))

	for i, n := range variantSizes {
		if n >= 1_000_000 {
			xLabels2[i] = fmt.Sprintf("%d万", n/10000)
		} else {
			xLabels2[i] = fmt.Sprintf("%d", n)
		}
		nums := makeSortedSlice(n)
		target := nums[n/2]

		start := time.Now()
		for range bsOpsPerMeasure {
			bs.BinarySearch(nums, target)
		}
		bsNs[i] = time.Since(start).Nanoseconds() / bsOpsPerMeasure

		start = time.Now()
		for range bsOpsPerMeasure {
			bs.LeftBound(nums, target)
		}
		leftNs[i] = time.Since(start).Nanoseconds() / bsOpsPerMeasure

		start = time.Now()
		for range bsOpsPerMeasure {
			bs.RightBound(nums, target)
		}
		rightNs[i] = time.Since(start).Nanoseconds() / bsOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "二分搜索性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("BinarySearch vs 线性搜索耗时（搜索中间元素，每规模 %d 次均值）", compOps),
				XAxis:     xLabels1,
				XAxisName: "数组规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "BinarySearch（O(log n)）", Data: bsCompNs},
					{Name: "线性搜索（O(n) 平均情况）", Data: linearNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("BinarySearch / LeftBound / RightBound 耗时 vs 规模（每规模 %d 次均值）", bsOpsPerMeasure),
				XAxis:     xLabels2,
				XAxisName: "数组规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "BinarySearch（返回任意匹配）", Data: bsNs},
					{Name: "LeftBound（最左边界）", Data: leftNs},
					{Name: "RightBound（最右边界）", Data: rightNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "binary_search/", Title: "二分搜索性能",
		Desc: "BinarySearch vs 线性搜索（1K~100K）；三种变体 O(log n) 耗时 vs 规模（1万~1000万）",
		Path: "/api/algo/binary_search", DataHandler: handleAlgoBinarySearch,
	})
}
