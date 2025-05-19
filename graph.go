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
// 创建日期:2024/2/20
package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
	sort2 "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/sort"
	"net/http"
	syssort "sort"
	"time"
)

// generate random data for line chart
func generateLineItems() []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 12; i++ {
		items = append(items, opts.LineData{Value: random.RandInt32(1, 30)})
	}
	return items
}

func generateLineData(data []int64) []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < len(data); i++ {
		items = append(items, opts.LineData{Value: data[i]})
	}
	return items
}

type SortInfo struct {
	Name   string
	CostMs []int64
}

type CompareArr struct {
	Arr []int32
}

func (this *CompareArr) Len() int {
	return len(this.Arr)
}

func (this *CompareArr) Less(i, j int) bool {
	return this.Arr[i] < this.Arr[j]
}

func (this *CompareArr) Swap(i, j int) {
	this.Arr[i], this.Arr[j] = this.Arr[j], this.Arr[i]
}

func createLine() *charts.Line {
	insertionSortInfo := &SortInfo{
		Name:   "insertion sort",
		CostMs: []int64{},
	}
	quickSortInfo := &SortInfo{
		Name:   "quick sort",
		CostMs: []int64{},
	}
	// quickSortTraversalInfo := &SortInfo{
	// 	Name:   "quick sort traversal",
	// 	CostMs: []int64{},
	// }
	systemQuickSortInfo := &SortInfo{
		Name:   "golang quick sort",
		CostMs: []int64{},
	}
	coutingSortInfo := &SortInfo{
		Name:   "Counting sort",
		CostMs: make([]int64, 0),
	}
	for i := 0; i < 10; i++ {
		limit := 1e5 * (i + 1)
		arr := make([]int32, limit, limit)
		for j := 0; j < int(limit); j++ {
			r := random.RandInt32(1, 100000)
			arr[j] = r
		}

		auxArr := make([]int32, limit, limit)
		// copy(auxArr, arr)
		start := time.Now().UnixNano()
		// sort.InsertionSort(auxArr)
		end := time.Now().UnixNano()
		duration := (end - start) / 1e6
		insertionSortInfo.CostMs = append(insertionSortInfo.CostMs, duration)

		copy(auxArr, arr)
		start = time.Now().UnixNano()
		sort2.QuickSort(auxArr)
		end = time.Now().UnixNano()
		duration = (end - start) / 1e6
		quickSortInfo.CostMs = append(quickSortInfo.CostMs, duration)

		copy(auxArr, arr)
		start = time.Now().UnixNano()
		sort2.CountingSort(auxArr)
		end = time.Now().UnixNano()
		duration = (end - start) / 1e6
		coutingSortInfo.CostMs = append(coutingSortInfo.CostMs, duration)

		// copy(auxArr, arr)
		// start = timeutils.Now().UnixNano()
		// sort.QuickSortTraversal(auxArr)
		// end = timeutils.Now().UnixNano()
		// duration = (end - start) / 1e6
		// quickSortTraversalInfo.CostMs = append(quickSortTraversalInfo.CostMs, duration)

		copy(auxArr, arr)
		start = time.Now().UnixNano()
		syssort.Sort(&CompareArr{Arr: auxArr})
		end = time.Now().UnixNano()
		duration = (end - start) / 1e6
		systemQuickSortInfo.CostMs = append(systemQuickSortInfo.CostMs, duration)
	}

	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "排序",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Cost timeutils(ms)",
			SplitLine: &opts.SplitLine{
				Show: opts.Bool(true),
			},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Elements",
		}),
	)

	// Put data into instance
	line.SetXAxis([]string{"10万", "20万", "30万", "40万", "50万", "60万", "70万", "80万", "90万", "100万"})

	// line.AddSeries(insertionSortInfo.Name, generateLineData(insertionSortInfo.CostMs),
	// 	charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "bottom"}))
	line.AddSeries(quickSortInfo.Name, generateLineData(quickSortInfo.CostMs),
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top"}))
	line.AddSeries(coutingSortInfo.Name, generateLineData(coutingSortInfo.CostMs),
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top"}))
	// line.AddSeries(quickSortTraversalInfo.Name, generateLineData(quickSortTraversalInfo.CostMs),
	// 	charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top"}))
	line.AddSeries(systemQuickSortInfo.Name, generateLineData(systemQuickSortInfo.CostMs),
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top"}))

	line.SetSeriesOptions(
		// charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{
		//     Name: "Average",
		//     Type: "average",
		// }),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
			Label: &opts.Label{
				Show:      opts.Bool(true),
				Formatter: "{a}: {b}",
			},
		}),
	)
	return line
}

func graphHttpServer(w http.ResponseWriter, _ *http.Request) {
	page := components.NewPage()

	page.AddCharts(
		createLine(),
	)
	err := page.Render(w)
	if err != nil {
		panic(err)
	}
}