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
	"math/rand/v2"
	"net/http"

	probdist "github.com/motocat46/yytools/pkg/algorithms/mathx/probability_distribution"
	pwc "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
	tiered "github.com/motocat46/yytools/pkg/mechanics/distribution/tiered_cycle"
)

// ---- TieredCycle ----

const (
	tieredCycleLen  = 100
	tieredSimCycles = 500
)

type tieredSimResult struct {
	standardCounts   [3]int
	specialCounts    [4]int
	specialPosCounts [tieredCycleLen]int
}

func runTieredCycleSim() tieredSimResult {
	items := []pwc.Item{
		{Quota: 4, JoinAt: 0},
		{Quota: 3, JoinAt: 4},
		{Quota: 2, JoinAt: 7},
		{Quota: 1, JoinAt: 9},
	}
	cfg := tiered.Config{
		ConfigBase:     tiered.ConfigBase{CycleLen: tieredCycleLen},
		ConfigStandard: tiered.ConfigStandard{Weight: tiered.NewWeight(map[int]int32{0: 70, 1: 20, 2: 10})},
		ConfigSpecial:  tiered.ConfigSpecial{MinInterval: 5, Items: items},
	}
	eng, _ := tiered.New(cfg)
	state := eng.NewState(rand.New(rand.NewPCG(42, 0)))
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

func handleDistTiered(w http.ResponseWriter, _ *http.Request) {
	sim := runTieredCycleSim()

	// 图表1：奖励分布
	xLabels1 := []string{"普A(70%)", "普B(20%)", "普C(10%)", "4★碎片(×4)", "4★武器(×3)", "5★武器(×2)", "5★角色(×1)"}
	actual := []int64{
		int64(sim.standardCounts[0]) / tieredSimCycles,
		int64(sim.standardCounts[1]) / tieredSimCycles,
		int64(sim.standardCounts[2]) / tieredSimCycles,
		int64(sim.specialCounts[0]) / tieredSimCycles,
		int64(sim.specialCounts[1]) / tieredSimCycles,
		int64(sim.specialCounts[2]) / tieredSimCycles,
		int64(sim.specialCounts[3]) / tieredSimCycles,
	}
	expected := []int64{63, 18, 9, 4, 3, 2, 1}

	// 图表2：特殊位置散布
	xLabels2 := make([]string, tieredCycleLen)
	posData := make([]int64, tieredCycleLen)
	for i := range tieredCycleLen {
		xLabels2[i] = fmt.Sprintf("%d", i)
		posData[i] = int64(sim.specialPosCounts[i])
	}

	json.NewEncoder(w).Encode(pageData{
		Title: "分层周期引擎",
		Charts: []chartData{
			{
				Type:  "bar",
				Title: fmt.Sprintf("奖励分布（每周期平均，模拟 %d 个周期）", tieredSimCycles),
				XAxis: xLabels1,
				Series: []chartSeries{
					{Name: "实际", Data: actual},
					{Name: "期望", Data: expected},
				},
			},
			{
				Type:      "bar",
				Title:     fmt.Sprintf("特殊位置散布（MinInterval=5，%d 个周期）", tieredSimCycles),
				XAxis:     xLabels2,
				XAxisName: "周期内位置",
				Series:    []chartSeries{{Name: "命中次数", Data: posData}},
			},
		},
	})
}

// ---- ProgressiveWeightCycle ----

const pwcSimCycles = 1000

func handleDistPWC(w http.ResponseWriter, _ *http.Request) {
	items := []pwc.Item{
		{Quota: 4, JoinAt: 0},
		{Quota: 2, JoinAt: 2},
		{Quota: 1, JoinAt: 5},
		{Quota: 1, JoinAt: 7},
		{Quota: 1, JoinAt: 8},
	}
	names := []string{"普通材料(×4)", "稀有材料(×2)", "史诗装备(×1)", "传说碎片(×1)", "神话装备(×1)"}

	layer := pwc.NewWeightCycleLayer(items)
	total := pwc.TotalQuota(items)
	counts := make([]int64, len(items))
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
	expected := make([]int64, len(items))
	for i, item := range items {
		expected[i] = int64(item.Quota) * pwcSimCycles
	}

	json.NewEncoder(w).Encode(pageData{
		Title: "渐进权重周期",
		Charts: []chartData{{
			Type:  "bar",
			Title: fmt.Sprintf("奖励分布实际 vs 期望（%d 个周期，每周期 %d 次）", pwcSimCycles, total),
			XAxis: names,
			Series: []chartSeries{
				{Name: "实际", Data: counts},
				{Name: "期望", Data: expected},
			},
		}},
	})
}

// ---- 概率分布 ----

func handleProb(w http.ResponseWriter, _ *http.Request) {
	weights := []int32{10, 20, 30, 40}
	total := int32(0)
	for _, wt := range weights {
		total += wt
	}
	const iterations = 1_000_000
	n := len(weights)
	normalCounts := make([]int64, n)
	voseCounts := make([]int64, n)
	nm := probdist.NewNormalMethod(weights)
	vm := probdist.NewVoseAliasMethod(weights)
	for range iterations {
		normalCounts[nm.Generate()]++
		voseCounts[vm.Generate()]++
	}
	xLabels := make([]string, n)
	for i, wt := range weights {
		xLabels[i] = fmt.Sprintf("idx%d (%.1f%%)", i, float64(wt)/float64(total)*100)
	}
	json.NewEncoder(w).Encode(pageData{
		Title: "概率分布对比",
		Charts: []chartData{{
			Type:      "bar",
			Title:     fmt.Sprintf("概率分布对比（%d 万次采样，权重 %v）", iterations/10000, weights),
			XAxis:     xLabels,
			XAxisName: "元素（期望概率）",
			YAxisName: "命中次数",
			Series: []chartSeries{
				{Name: "NormalMethod (二分)", Data: normalCounts},
				{Name: "VoseAliasMethod", Data: voseCounts},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/mechanics", SubPkg: "distribution/tiered_cycle/", Title: "分层周期引擎",
		Desc: "奖励分布 + 特殊位置散布（500 周期）",
		Path: "/api/dist/tiered", DataHandler: handleDistTiered,
	})
	Register(VisEntry{
		Pkg: "pkg/mechanics", SubPkg: "distribution/progressive_weight_cycle/", Title: "渐进权重周期",
		Desc: "奖励分布实际 vs 期望（1000 周期）",
		Path: "/api/dist/pwc", DataHandler: handleDistPWC,
	})
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/probability_distribution/", Title: "概率分布对比",
		Desc: "NormalMethod vs VoseAliasMethod（100万次采样）",
		Path: "/api/prob", DataHandler: handleProb,
	})
}
