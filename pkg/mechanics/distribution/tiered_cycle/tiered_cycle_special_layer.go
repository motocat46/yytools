// Package tiered_cycle.

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
// 创建日期:2026/2/28
package tiered_cycle

import (
	"math/rand/v2"
	
	"github.com/motocat46/yytools/pkg/algorithms/mathx/sampling"
	"github.com/motocat46/yytools/pkg/common/assert"
	weight_cycle "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
)

// SpecialLayerState 特殊层的运行时状态（仅属于特殊层）
type SpecialLayerState struct {
	plan []int32 // 本周期特殊位置计划（有序）
	*weight_cycle.State
}

// SpecialLayer 特殊层：封装特殊保底分布的规则与算法
type SpecialLayer struct {
	minInterval int32
	*weight_cycle.Layer
}

// newSpecialLayer 以 items 和最小间隔创建特殊层。
func newSpecialLayer(items []weight_cycle.Item, minInterval int32) SpecialLayer {
	layer := SpecialLayer{minInterval: minInterval}
	layer.Layer = weight_cycle.NewWeightCycleLayer(items)
	return layer
}

// NewState 创建空的特殊层运行时状态（plan 为空切片，weight_cycle.State 为初始状态）。
func (l *SpecialLayer) NewState() SpecialLayerState {
	state := SpecialLayerState{
		plan: make([]int32, 0),
	}
	state.State = weight_cycle.NewState()
	return state
}

// GetOccIdx 返回 posInCycle 在特殊计划中的序号（0-based），-1 表示不是特殊位置
func (l *SpecialLayer) GetOccIdx(state *SpecialLayerState, posInCycle int32) int {
	return getSpecialCycleIndex(state.plan, posInCycle)
}

// Generate 执行特殊层抽取（算法在此，可独立替换）
func (l *SpecialLayer) Generate(state *SpecialLayerState, occIdx int32) (int, error) {
	return l.Layer.Generate(state.State, occIdx)
}

// Reset 重置特殊层状态，并生成新的周期计划
func (l *SpecialLayer) Reset(state *SpecialLayerState, r *rand.Rand, cycleLen int32) {
	n := l.TotalQuota
	if n == 0 {
		if state.plan == nil {
			state.plan = make([]int32, 0)
		} else {
			state.plan = state.plan[:0]
		}
	} else {
		state.plan = buildSpecialPlan(r, cycleLen, l.minInterval, n)
	}
	state.State = weight_cycle.NewState()
}

// buildSpecialPlan 初始化特殊抽取的周期分布计划
func buildSpecialPlan(r *rand.Rand, pondCycle int32, pondMinInterval int32, count int32) []int32 {
	assert.Assert(count <= pondCycle)
	rangeMax := pondCycle - pondMinInterval*(count-1)
	assert.Assert(rangeMax > 0 && rangeMax <= pondCycle, rangeMax)
	
	// 从 [0, pondCycle-1] 选 count 个位置，最小间隔 pondMinInterval
	return sampling.SampleWithMinGap(0, pondCycle-1, int(count), int(pondMinInterval), r)
}

// getSpecialCycleIndex 判断当前位置是否是特殊循环位置
// specialPlan 是有序切片，使用二分搜索 O(log n)
// 返回：-1 表示不是特殊循环位置；>= 0 表示是第几个特殊循环位置（从0开始）
func getSpecialCycleIndex(period []int32, currentIndex int32) int {
	lo, hi := 0, len(period)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if period[mid] == currentIndex {
			return mid
		} else if period[mid] < currentIndex {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return -1
}