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
	"math/rand/v2"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	pwc "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
	tiered "github.com/motocat46/yytools/pkg/mechanics/distribution/tiered_cycle"
)

// ---- 通用辅助 ----

func intsToBarData(vals []int) []opts.BarData {
	data := make([]opts.BarData, len(vals))
	for i, v := range vals {
		data[i] = opts.BarData{Value: v}
	}
	return data
}

func floatsToBarData(vals []float64) []opts.BarData {
	data := make([]opts.BarData, len(vals))
	for i, v := range vals {
		data[i] = opts.BarData{Value: v}
	}
	return data
}

func newDistBar(title, subtitle string) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{Title: title, Subtitle: subtitle}),
		charts.WithYAxisOpts(opts.YAxis{SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
	)
	return bar
}

// ---- TieredCycle 可视化 ----

const (
	tieredCycleLen  = 100
	tieredSimCycles = 500
)

type tieredSimResult struct {
	standardCounts  [3]int
	specialCounts   [4]int
	specialPosCounts [tieredCycleLen]int
}

// runTieredCycleSim 模拟抽卡保底场景：
//   - 周期长度 100，MinInterval=5
//   - 标准层权重：普A=70, 普B=20, 普C=10
//   - 特殊层：4★碎片(×4)、4★武器(×3)、5★武器(×2)、5★角色(×1)，共 10 个特殊位
func runTieredCycleSim() tieredSimResult {
	items := []pwc.Item{
		{Quota: 4, JoinAt: 0},
		{Quota: 3, JoinAt: 4},
		{Quota: 2, JoinAt: 7},
		{Quota: 1, JoinAt: 9},
	}
	cfg := tiered.Config{
		ConfigBase: tiered.ConfigBase{
			R:        rand.New(rand.NewPCG(42, 0)),
			CycleLen: tieredCycleLen,
		},
		ConfigStandard: tiered.ConfigStandard{
			Weight: tiered.NewWeight(map[int]int32{0: 70, 1: 20, 2: 10}),
		},
		ConfigSpecial: tiered.ConfigSpecial{
			MinInterval: 5,
			Items:       items,
		},
	}
	eng, _ := tiered.New(cfg)
	state := tiered.NewState()
	eng.Init(state)

	var res tieredSimResult
	for range tieredSimCycles * tieredCycleLen {
		pos := state.PosInCycle()
		r, err := eng.NextAutoReset(state)
		if err != nil {
			continue
		}
		switch r.Type {
		case tiered.Standard:
			if r.Index >= 0 && r.Index < len(res.standardCounts) {
				res.standardCounts[r.Index]++
			}
		case tiered.Special:
			if r.Index >= 0 && r.Index < len(res.specialCounts) {
				res.specialCounts[r.Index]++
			}
			if pos >= 0 && int(pos) < len(res.specialPosCounts) {
				res.specialPosCounts[pos]++
			}
		}
	}
	return res
}

// createTieredRewardBar 每周期平均奖励次数：实际 vs 期望
//
// 标准层期望 = 每周期标准位数(90) × 权重占比
// 特殊层期望 = Quota（每周期恰好出现 Quota 次，由动态权重机保证）
func createTieredRewardBar(sim tieredSimResult) *charts.Bar {
	xLabels := []string{
		"普A(70%)", "普B(20%)", "普C(10%)",
		"4★碎片(×4)", "4★武器(×3)", "5★武器(×2)", "5★角色(×1)",
	}
	actual := []float64{
		float64(sim.standardCounts[0]) / tieredSimCycles,
		float64(sim.standardCounts[1]) / tieredSimCycles,
		float64(sim.standardCounts[2]) / tieredSimCycles,
		float64(sim.specialCounts[0]) / tieredSimCycles,
		float64(sim.specialCounts[1]) / tieredSimCycles,
		float64(sim.specialCounts[2]) / tieredSimCycles,
		float64(sim.specialCounts[3]) / tieredSimCycles,
	}
	// 标准层期望：每周期 90 次普通抽 × 权重比例
	// 特殊层期望：每周期恰好 Quota 次
	expected := []float64{63, 18, 9, 4, 3, 2, 1}

	bar := newDistBar(
		"分层周期引擎 — 奖励分布（每周期平均）",
		fmt.Sprintf("模拟 %d 个周期，周期长度 %d", tieredSimCycles, tieredCycleLen),
	)
	bar.SetXAxis(xLabels)
	bar.AddSeries("实际", floatsToBarData(actual))
	bar.AddSeries("期望", floatsToBarData(expected))
	return bar
}

// createTieredSpecialPosBar 特殊位置散布图
//
// 展示 MinInterval=5 约束下，特殊位置在周期内 0~99 各处的实际落点次数。
// 理想情况下分布均匀；相邻特殊位之间始终保持 ≥5 的间距。
func createTieredSpecialPosBar(sim tieredSimResult) *charts.Bar {
	xLabels := make([]string, tieredCycleLen)
	data := make([]int, tieredCycleLen)
	for i := range tieredCycleLen {
		xLabels[i] = fmt.Sprintf("%d", i)
		data[i] = sim.specialPosCounts[i]
	}

	bar := newDistBar(
		"特殊位置散布（MinInterval=5 效果验证）",
		fmt.Sprintf("各位置（0~%d）被选为特殊抽的累计次数，%d 个周期", tieredCycleLen-1, tieredSimCycles),
	)
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine, Width: "1200px"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "周期内位置"}),
	)
	bar.SetXAxis(xLabels)
	bar.AddSeries("命中次数", intsToBarData(data))
	return bar
}

// ---- ProgressiveWeightCycle 可视化 ----

const pwcSimCycles = 1000

// createPWCRewardBar 渐进权重周期 — 奖励分布（实际 vs 期望）
//
// 场景：副本宝箱，5 种奖励随通关次数逐步解锁。
// 每周期恰好消耗完所有 Quota，因此期望 = Quota × 模拟周期数。
func createPWCRewardBar() *charts.Bar {
	items := []pwc.Item{
		{Quota: 4, JoinAt: 0}, // 普通材料：随时可出
		{Quota: 2, JoinAt: 2}, // 稀有材料：第3次起解锁
		{Quota: 1, JoinAt: 5}, // 史诗装备：第6次起解锁
		{Quota: 1, JoinAt: 7}, // 传说碎片：第8次起解锁
		{Quota: 1, JoinAt: 8}, // 神话装备：仅最后一次
	}
	names := []string{"普通材料(×4)", "稀有材料(×2)", "史诗装备(×1)", "传说碎片(×1)", "神话装备(×1)"}

	layer := pwc.NewWeightCycleLayer(items)
	total := pwc.TotalQuota(items)
	counts := make([]int, len(items))

	for range pwcSimCycles {
		state := pwc.NewState()
		for occIdx := range total {
			idx, err := layer.Generate(state, occIdx)
			if err != nil {
				continue
			}
			if idx >= 0 && idx < len(counts) {
				counts[idx]++
			}
		}
	}

	expected := make([]int, len(items))
	for i, item := range items {
		expected[i] = int(item.Quota) * pwcSimCycles
	}

	bar := newDistBar(
		"渐进权重周期 — 奖励分布（实际 vs 期望）",
		fmt.Sprintf("模拟 %d 个周期，每周期 %d 次特殊抽", pwcSimCycles, total),
	)
	bar.SetXAxis(names)
	bar.AddSeries("实际", intsToBarData(counts))
	bar.AddSeries("期望", intsToBarData(expected))
	return bar
}
