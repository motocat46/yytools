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
	"math"
	"net/http"

	rs "github.com/motocat46/yytools/pkg/mechanics/distribution/random_split"
)

const (
	rsTotal  = 1000 // 总量
	rsCount  = 10   // 份数
	rsMin    = 1    // 每份最小值
	rsRounds = 100_000
)

// runSim 使用指定策略跑 rsRounds 轮，返回各位置均值和标准差（均四舍五入为 int64）。
func runSim(fn rs.SampleFunc, seed uint64) (means, stdDevs []int64) {
	state := rs.State{RemainAmount: rsTotal, RemainCount: rsCount, MinPerPart: rsMin}
	result, _ := rs.Simulate(state, fn, rsRounds, seed)
	means = make([]int64, rsCount)
	stdDevs = make([]int64, rsCount)
	for i, p := range result.Positions {
		means[i] = int64(math.Round(p.Mean()))
		stdDevs[i] = int64(math.Round(p.StdDev()))
	}
	return
}

func handleMechRandomSplit(w http.ResponseWriter, _ *http.Request) {
	xLabels := make([]string, rsCount)
	for i := range rsCount {
		xLabels[i] = fmt.Sprintf("位置%d", i)
	}

	fixedMeans, fixedStdDevs := runSim(rs.Fixed(), 42)
	dmMeans, dmStdDevs := runSim(rs.DoubleMean(), 42)
	unifMeans, unifStdDevs := runSim(rs.Uniform(), 42)

	// 全局均值参考线
	idealMean := make([]int64, rsCount)
	for i := range rsCount {
		idealMean[i] = rsTotal / rsCount
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "RandomSplit 策略对比",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("各策略每位置均值（总量=%d，%d份，%d万轮，理想均值=%d）", rsTotal, rsCount, rsRounds/10000, rsTotal/rsCount),
				XAxis:     xLabels,
				XAxisName: "分配位置",
				YAxisName: "均值",
				Series: []chartSeries{
					{Name: fmt.Sprintf("理想均值（%d）", rsTotal/rsCount), Data: idealMean},
					{Name: "Fixed（确定性）", Data: fixedMeans},
					{Name: "DoubleMean（二倍均值）", Data: dmMeans},
					{Name: "Uniform（均匀可行）", Data: unifMeans},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("各策略每位置标准差（总量=%d，%d份，%d万轮）", rsTotal, rsCount, rsRounds/10000),
				XAxis:     xLabels,
				XAxisName: "分配位置",
				YAxisName: "标准差",
				Series: []chartSeries{
					{Name: "Fixed（确定性）", Data: fixedStdDevs},
					{Name: "DoubleMean（二倍均值）", Data: dmStdDevs},
					{Name: "Uniform（均匀可行）", Data: unifStdDevs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/mechanics", SubPkg: "distribution/random_split/", Title: "RandomSplit 策略对比",
		Desc: "Fixed/DoubleMean/Uniform 三策略 10 万轮模拟：每位置均值与标准差，展示位置偏差与方差分布",
		Path: "/api/mech/random_split", DataHandler: handleMechRandomSplit,
	})
}
