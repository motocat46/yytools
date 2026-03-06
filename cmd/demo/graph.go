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
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	probdist "github.com/stormYuanYang/yytools/pkg/algorithms/mathx/probability_distribution"
	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	sort2 "github.com/stormYuanYang/yytools/pkg/algorithms/sort"
)

// ---- 通用辅助 ----

func generateLineData(data []int64) []opts.LineData {
	items := make([]opts.LineData, 0, len(data))
	for _, v := range data {
		items = append(items, opts.LineData{Value: v})
	}
	return items
}

// measureSort 复制数组后执行 sortFn，返回耗时毫秒数
func measureSort(arr []int32, sortFn func([]int32)) int64 {
	aux := make([]int32, len(arr))
	copy(aux, arr)
	start := time.Now().UnixNano()
	sortFn(aux)
	return (time.Now().UnixNano() - start) / 1e6
}

// measureGoSort 使用 slices.Sort（泛型，无接口开销）排序，返回耗时毫秒数
func measureGoSort(arr []int32) int64 {
	aux := make([]int32, len(arr))
	copy(aux, arr)
	start := time.Now().UnixNano()
	slices.Sort(aux)
	return (time.Now().UnixNano() - start) / 1e6
}

func newSortLine(title string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithYAxisOpts(opts.YAxis{Name: "耗时(ms)", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}),
		charts.WithXAxisOpts(opts.XAxis{Name: "元素数量"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
	)
	line.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))
	return line
}

// ---- 排序图表 ----

// createEfficientSortLine 高效排序对比（10万~100万）
// 包含：QuickSort、QuickSortTraversal、CountingSort、RadixSort、Go slices.Sort
func createEfficientSortLine() *charts.Line {
	type entry struct {
		name   string
		costs  []int64
		sortFn func([]int32) // nil 表示使用 Go stdlib
	}
	series := []*entry{
		{name: "Quick Sort", sortFn: sort2.QuickSort[int32]},
		{name: "Quick Sort (Traversal)", sortFn: sort2.QuickSortTraversal[int32]},
		{name: "Counting Sort", sortFn: sort2.CountingSort[int32]},
		{name: "Radix Sort", sortFn: sort2.RadixSort[int32]},
		{name: "Go slices.Sort", sortFn: nil},
	}

	xLabels := make([]string, 10)
	for i := range xLabels {
		xLabels[i] = fmt.Sprintf("%d万", (i+1)*10)
		n := int(1e5) * (i + 1)
		arr := make([]int32, n)
		for j := range arr {
			arr[j] = random.RandInt[int32](1, 100000)
		}
		for _, s := range series {
			var ms int64
			if s.sortFn == nil {
				ms = measureGoSort(arr)
			} else {
				ms = measureSort(arr, s.sortFn)
			}
			s.costs = append(s.costs, ms)
		}
	}

	line := newSortLine("高效排序对比（10万~100万元素）")
	line.SetXAxis(xLabels)
	for _, s := range series {
		line.AddSeries(s.name, generateLineData(s.costs))
	}
	return line
}

// createSimpleSortLine 简单排序对比（2千~2万）
// 包含：Bubble Sort、Insertion Sort，以及 Quick Sort 作为性能参照
func createSimpleSortLine() *charts.Line {
	type entry struct {
		name   string
		costs  []int64
		sortFn func([]int32)
	}
	series := []*entry{
		{name: "Bubble Sort", sortFn: sort2.BubbleSort[int32]},
		{name: "Insertion Sort", sortFn: sort2.InsertionSort[int32]},
		{name: "Quick Sort (参照)", sortFn: sort2.QuickSort[int32]},
	}

	xLabels := make([]string, 10)
	for i := range xLabels {
		n := 2000 * (i + 1)
		xLabels[i] = fmt.Sprintf("%d", n)
		arr := make([]int32, n)
		for j := range arr {
			arr[j] = random.RandInt[int32](1, 100000)
		}
		for _, s := range series {
			s.costs = append(s.costs, measureSort(arr, s.sortFn))
		}
	}

	line := newSortLine("简单排序对比（2千~2万元素）")
	line.SetXAxis(xLabels)
	for _, s := range series {
		line.AddSeries(s.name, generateLineData(s.costs))
	}
	return line
}

// ---- 概率分布图表 ----

// createProbDistBar 概率分布直方图
// 固定权重 [10, 20, 30, 40]，采样 100 万次，对比 NormalMethod 和 VoseAliasMethod 的实际命中分布
func createProbDistBar() *charts.Bar {
	weights := []int32{10, 20, 30, 40}
	total := int32(0)
	for _, w := range weights {
		total += w
	}

	const iterations = 1_000_000
	n := len(weights)
	normalCounts := make([]int, n)
	voseCounts := make([]int, n)

	nm := probdist.NewNormalMethod(weights)
	vm := probdist.NewVoseAliasMethod(weights)
	for range iterations {
		normalCounts[nm.Generate()]++
		voseCounts[vm.Generate()]++
	}

	xLabels := make([]string, n)
	for i, w := range weights {
		xLabels[i] = fmt.Sprintf("idx%d (%.1f%%)", i, float64(w)/float64(total)*100)
	}

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title:    "概率分布对比（100万次采样）",
			Subtitle: fmt.Sprintf("权重: %v  总权重: %d", weights, total),
		}),
		charts.WithYAxisOpts(opts.YAxis{Name: "命中次数", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}),
		charts.WithXAxisOpts(opts.XAxis{Name: "元素（期望概率）"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
	)
	bar.SetXAxis(xLabels)
	bar.AddSeries("NormalMethod (二分)", intsToBarData(normalCounts))
	bar.AddSeries("VoseAliasMethod", intsToBarData(voseCounts))
	return bar
}

// ---- HTTP 入口 ----

const indexHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>yytools 可视化</title>
  <style>
    body { font-family: sans-serif; max-width: 700px; margin: 60px auto; }
    h1   { font-size: 1.4em; margin-bottom: 0.4em; }
    h2   { font-size: 1em; color: #555; margin: 1.6em 0 0.4em; }
    ul   { list-style: none; padding: 0; }
    li   { margin: 10px 0; }
    a    { font-size: 1.05em; text-decoration: none; color: #1a73e8; }
    a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <h1>yytools 可视化示例</h1>

  <h2>排序算法</h2>
  <ul>
    <li><a href="/sort/efficient">高效排序对比（QuickSort / CountingSort / RadixSort / Go stdlib）</a></li>
    <li><a href="/sort/simple">简单排序对比（BubbleSort / InsertionSort）</a></li>
    <li><a href="/sort/compare">快排 vs pdqsort — 不同输入场景对比（随机 / 近乎有序 / 有序 / 逆序 / 大量重复）</a></li>
  </ul>

  <h2>概率分布</h2>
  <ul>
    <li><a href="/prob">概率分布对比（NormalMethod / VoseAliasMethod）</a></li>
  </ul>

  <h2>分布机制</h2>
  <ul>
    <li><a href="/dist/tiered">分层周期引擎（奖励分布 + 特殊位置散布）</a></li>
    <li><a href="/dist/pwc">渐进权重周期（奖励分布实际 vs 期望）</a></li>
  </ul>
</body>
</html>`

func renderCharts(w http.ResponseWriter, chartFns ...func() components.Charter) {
	page := components.NewPage()
	for _, fn := range chartFns {
		page.AddCharts(fn())
	}
	if err := page.Render(w); err != nil {
		panic(err)
	}
}

func graphHttpServer(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/sort/efficient":
		renderCharts(w, func() components.Charter { return createEfficientSortLine() })
	case "/sort/simple":
		renderCharts(w, func() components.Charter { return createSimpleSortLine() })
	case "/prob":
		renderCharts(w, func() components.Charter { return createProbDistBar() })
	case "/dist/tiered":
		sim := runTieredCycleSim()
		renderCharts(w,
			func() components.Charter { return createTieredRewardBar(sim) },
			func() components.Charter { return createTieredSpecialPosBar(sim) },
		)
	case "/dist/pwc":
		renderCharts(w, func() components.Charter { return createPWCRewardBar() })
	case "/sort/compare":
		page := components.NewPage()
		for _, c := range createSortCompareCharts() {
			page.AddCharts(c)
		}
		if err := page.Render(w); err != nil {
			panic(err)
		}
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	}
}