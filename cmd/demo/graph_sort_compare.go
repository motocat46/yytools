// Package main.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/3/6
package main

import (
	"fmt"
	"slices"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
	sort2 "github.com/motocat46/yytools/pkg/algorithms/sort"
)

// ---- 数据生成 ----

func genRandom(n int) []int32 {
	arr := make([]int32, n)
	for i := range arr {
		arr[i] = random.RandInt[int32](1, 100000)
	}
	return arr
}

func genSorted(n int) []int32 {
	arr := genRandom(n)
	slices.Sort(arr)
	return arr
}

func genReverse(n int) []int32 {
	arr := genSorted(n)
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// genNearlySorted 有序数组随机交换 1% 的元素对
func genNearlySorted(n int) []int32 {
	arr := genSorted(n)
	for range n / 100 {
		i := int(random.RandInt[int32](0, int32(n-1)))
		j := int(random.RandInt[int32](0, int32(n-1)))
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// genManyDuplicates 只有 10 种不同的值
func genManyDuplicates(n int) []int32 {
	arr := make([]int32, n)
	for i := range arr {
		arr[i] = random.RandInt[int32](1, 10)
	}
	return arr
}

// ---- 图表构造 ----

type sortAlgo struct {
	name   string
	sortFn func([]int32) // nil 表示 slices.Sort
}

var compareAlgos = []sortAlgo{
	{"QuickSort 递归", sort2.QuickSort[int32]},
	{"QuickSort 遍历", sort2.QuickSortTraversal[int32]},
	{"slices.Sort (pdqsort)", nil},
}

// createSortCompareChart 针对单种输入场景，对比三种算法的性能曲线
func createSortCompareChart(title string, genData func(int) []int32) *charts.Line {
	costs := make([][]int64, len(compareAlgos))
	for i := range costs {
		costs[i] = make([]int64, 0, 10)
	}

	xLabels := make([]string, 10)
	for i := range xLabels {
		xLabels[i] = fmt.Sprintf("%d万", (i+1)*10)
		n := int(1e5) * (i + 1)
		arr := genData(n)
		for j, algo := range compareAlgos {
			var ms int64
			if algo.sortFn == nil {
				ms = measureGoSort(arr)
			} else {
				ms = measureSort(arr, algo.sortFn)
			}
			costs[j] = append(costs[j], ms)
		}
	}

	line := newSortLine(title)
	line.SetXAxis(xLabels)
	for i, algo := range compareAlgos {
		line.AddSeries(algo.name, generateLineData(costs[i]))
	}
	return line
}

// createSortCompareCharts 生成全部场景的图表列表
func createSortCompareCharts() []*charts.Line {
	return []*charts.Line{
		createSortCompareChart("随机数据（基准）", genRandom),
		createSortCompareChart("近乎有序（1% 随机交换）", genNearlySorted),
		createSortCompareChart("完全有序（升序）", genSorted),
		createSortCompareChart("逆序", genReverse),
		createSortCompareChart("大量重复（10 种值）", genManyDuplicates),
	}
}
