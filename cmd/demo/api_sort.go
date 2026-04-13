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
// 创建日期:2026/4/10
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
	sort2 "github.com/motocat46/yytools/pkg/algorithms/sort"
)

// ---- 通用数据结构 ----

type chartSeries struct {
	Name string  `json:"name"`
	Data []int64 `json:"data"`
}

type chartData struct {
	Type      string        `json:"type"`
	Title     string        `json:"title"`
	XAxis     []string      `json:"xAxis"`
	XAxisName string        `json:"xAxisName,omitempty"`
	YAxisName string        `json:"yAxisName,omitempty"`
	Series    []chartSeries `json:"series"`
}

type pageData struct {
	Title  string      `json:"title"`
	Charts []chartData `json:"charts"`
}

// ---- 计算辅助 ----

// measureSort 复制数组后执行 sortFn，返回耗时毫秒数
func measureSort(arr []int32, sortFn func([]int32)) int64 {
	aux := make([]int32, len(arr))
	copy(aux, arr)
	start := time.Now()
	sortFn(aux)
	return time.Since(start).Milliseconds()
}

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

func genNearlySorted(n int) []int32 {
	arr := genSorted(n)
	for range n / 100 {
		i := int(random.RandInt[int32](0, int32(n-1)))
		j := int(random.RandInt[int32](0, int32(n-1)))
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func genManyDuplicates(n int) []int32 {
	arr := make([]int32, n)
	for i := range arr {
		arr[i] = random.RandInt[int32](1, 10)
	}
	return arr
}

// ---- Handlers ----

func handleSortEfficient(w http.ResponseWriter, _ *http.Request) {
	series := []*sortAlgoEntry{
		{name: "Quick Sort", sortFn: sort2.QuickSort[int32]},
		{name: "Quick Sort (Traversal)", sortFn: sort2.QuickSortTraversal[int32]},
		{name: "Counting Sort", sortFn: sort2.CountingSort[int32]},
		{name: "Radix Sort", sortFn: sort2.RadixSort[int32]},
		{name: "Go slices.Sort", sortFn: slices.Sort[[]int32]},
	}
	xLabels := make([]string, 10)
	for i := range xLabels {
		xLabels[i] = fmt.Sprintf("%d万", (i+1)*10)
		n := int(1e5) * (i + 1)
		arr := genRandom(n)
		for _, s := range series {
			s.costs = append(s.costs, measureSort(arr, s.sortFn))
		}
	}
	ss := make([]chartSeries, len(series))
	for i, s := range series {
		ss[i] = chartSeries{Name: s.name, Data: s.costs}
	}
	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "高效排序对比",
		Charts: []chartData{{
			Type: "line", Title: "高效排序对比（10万~100万元素）",
			XAxis: xLabels, XAxisName: "元素数量", YAxisName: "耗时(ms)", Series: ss,
		}},
	})
}

func handleSortSimple(w http.ResponseWriter, _ *http.Request) {
	series := []*sortAlgoEntry{
		{name: "Selection Sort", sortFn: sort2.SelectionSort[int32]},
		{name: "Insertion Sort", sortFn: sort2.InsertionSort[int32]},
		{name: "Quick Sort (参照)", sortFn: sort2.QuickSort[int32]},
	}
	xLabels := make([]string, 10)
	for i := range xLabels {
		n := 2000 * (i + 1)
		xLabels[i] = fmt.Sprintf("%d", n)
		arr := genRandom(n)
		for _, s := range series {
			s.costs = append(s.costs, measureSort(arr, s.sortFn))
		}
	}
	ss := make([]chartSeries, len(series))
	for i, s := range series {
		ss[i] = chartSeries{Name: s.name, Data: s.costs}
	}
	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "简单排序对比",
		Charts: []chartData{{
			Type: "line", Title: "简单排序对比（2千~2万元素）",
			XAxis: xLabels, XAxisName: "元素数量", YAxisName: "耗时(ms)", Series: ss,
		}},
	})
}

type sortAlgoEntry struct {
	name   string
	sortFn func([]int32)
	costs  []int64
}

var compareAlgos = []sortAlgoEntry{
	{name: "QuickSort 递归", sortFn: sort2.QuickSort[int32]},
	{name: "QuickSort 遍历", sortFn: sort2.QuickSortTraversal[int32]},
	{name: "slices.Sort (pdqsort)", sortFn: slices.Sort[[]int32]},
}

func buildCompareChart(title string, genData func(int) []int32) chartData {
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
			costs[j] = append(costs[j], measureSort(arr, algo.sortFn))
		}
	}
	ss := make([]chartSeries, len(compareAlgos))
	for i, algo := range compareAlgos {
		ss[i] = chartSeries{Name: algo.name, Data: costs[i]}
	}
	return chartData{
		Type: "line", Title: title,
		XAxis: xLabels, XAxisName: "元素数量", YAxisName: "耗时(ms)", Series: ss,
	}
}

func handleSortCompare(w http.ResponseWriter, _ *http.Request) {
	scenarios := []struct {
		title   string
		genData func(int) []int32
	}{
		{"随机数据（基准）", genRandom},
		{"近乎有序（1% 随机交换）", genNearlySorted},
		{"完全有序（升序）", genSorted},
		{"逆序", genReverse},
		{"大量重复（10 种值）", genManyDuplicates},
	}
	charts := make([]chartData, len(scenarios))
	for i, s := range scenarios {
		charts[i] = buildCompareChart(s.title, s.genData)
	}
	json.NewEncoder(w).Encode(pageData{Title: "快排 vs pdqsort", Charts: charts}) //nolint:errcheck
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "sort/", Title: "高效排序对比",
		Desc: "QuickSort / CountingSort / RadixSort / stdlib（10万~100万）",
		Path: "/api/sort/efficient", DataHandler: handleSortEfficient,
	})
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "sort/", Title: "简单排序对比",
		Desc: "SelectionSort / InsertionSort（2千~2万）",
		Path: "/api/sort/simple", DataHandler: handleSortSimple,
	})
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "sort/", Title: "快排 vs pdqsort",
		Desc: "5种输入场景（随机 / 近乎有序 / 有序 / 逆序 / 大量重复）",
		Path: "/api/sort/compare", DataHandler: handleSortCompare,
	})
}
