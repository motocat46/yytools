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
	"fmt"
	"math/rand/v2"
	
	pd "github.com/stormYuanYang/yytools/pkg/algorithms/mathx/probability_distribution"
	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/sampling"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
)

// SpecialItem 每个特殊结果的配置，下标即 Result.Index
type SpecialItem struct {
	Quota  int32 // 此特殊结果在一个周期内可出现的最大次数
	JoinAt int32 // 第几次特殊抽（0-based 特殊出现序号）时该奖励才开始进入候选池
}

// SpecialLayerState 特殊层的运行时状态（仅属于特殊层）
type SpecialLayerState struct {
	plan     []int32                   // 本周期特殊位置计划（有序）
	dw       *pd.DynamicWeights[int32] // 动态权重机
	unlocked map[int]bool              // 已加入动态权重机的权重 key
}

// SpecialLayer 特殊层：封装特殊保底分布的规则与算法
type SpecialLayer struct {
	items       []SpecialItem
	minInterval int32
}

func newSpecialLayer(items []SpecialItem, minInterval int32) SpecialLayer {
	return SpecialLayer{items: items, minInterval: minInterval}
}

func (l *SpecialLayer) NewState() SpecialLayerState {
	return SpecialLayerState{
		plan:     make([]int32, 0),
		dw:       newEmptyDW(),
		unlocked: make(map[int]bool),
	}
}

// GetOccIdx 返回 posInCycle 在特殊计划中的序号（0-based），-1 表示不是特殊位置
func (l *SpecialLayer) GetOccIdx(state *SpecialLayerState, posInCycle int32) int {
	return getSpecialCycleIndex(state.plan, posInCycle)
}

// Generate 执行特殊层抽取（算法在此，可独立替换）
func (l *SpecialLayer) Generate(state *SpecialLayerState, occIdx int32) (int, error) {
	return specialCycleCoreV2(state.unlocked, occIdx, l.items, state.dw)
}

// Reset 重置特殊层状态，并生成新的周期计划
func (l *SpecialLayer) Reset(state *SpecialLayerState, r *rand.Rand, cycleLen int32) {
	n := totalQuota(l.items)
	if n == 0 {
		if state.plan == nil {
			state.plan = make([]int32, 0)
		} else {
			state.plan = state.plan[:0]
		}
	} else {
		state.plan = buildSpecialPlan(r, cycleLen, l.minInterval, n)
	}
	state.dw = newEmptyDW()
	state.unlocked = make(map[int]bool)
}

func newEmptyDW() *pd.DynamicWeights[int32] {
	return pd.NewDynamicWeightsProgressive[int32]()
}

// buildSpecialPlan 初始化特殊抽取的周期分布计划
func buildSpecialPlan(r *rand.Rand, pondCycle int32, pondMinInterval int32, count int32) []int32 {
	assert.Assert(count <= pondCycle)
	rangeMax := pondCycle - pondMinInterval*(count-1)
	assert.Assert(rangeMax > 0 && rangeMax <= pondCycle, rangeMax)
	
	// 从 [0, pondCycle-1] 选 count 个位置，最小间隔 pondMinInterval
	return sampling.SampleWithMinGap(0, pondCycle-1, int(count), int(pondMinInterval), r)
}

// totalQuota 返回所有 SpecialItem 的 Quota 之和，即一个周期内特殊位置总数
func totalQuota(items []SpecialItem) int32 {
	var n int32
	for _, item := range items {
		n += item.Quota
	}
	return n
}

func checkSpecialCycleCoreParams(used map[int32]int32, specialOccIdx int32, items []SpecialItem) error {
	if used == nil {
		return fmt.Errorf("invalid used map(nil)")
	}
	if specialOccIdx < 0 {
		return fmt.Errorf("invalid specialOccIdx:%d", specialOccIdx)
	}
	if len(items) == 0 {
		return fmt.Errorf("invalid items: empty")
	}
	return nil
}

// specialCycleCore 执行特殊抽取选择（V1 实现，保留用于对比测试）。
// specialOccIdx 为当前特殊出现序号（0-based），即本次特殊抽是周期内第几次特殊抽。
// items[i].JoinAt 表示第几次特殊抽（0-based）时该奖励才开始进入候选池。
// 例如：JoinAt=0 从第0次特殊抽起即可出现；JoinAt=3 从第4次特殊抽起才能出现。
func specialCycleCore(used map[int32]int32, specialOccIdx int32, items []SpecialItem) (selectedIndex int, err error) {
	err = checkSpecialCycleCoreParams(used, specialOccIdx, items)
	if err != nil {
		return -1, err
	}
	weightMap := make(map[int]int32)
	var totalWeight int32
	
	for i, item := range items {
		// 检查：
		// 1. 当前特殊序号已达到加入门槛，且 2. 该奖励还有出现次数配额
		diff := item.Quota - used[int32(i)]
		if specialOccIdx >= item.JoinAt && diff > 0 {
			weightMap[i] = diff
			totalWeight += diff
		}
	}
	
	if len(weightMap) == 0 || totalWeight == 0 {
		return -1, fmt.Errorf("no candidate")
	}
	selectedIndex = pd.CalcKeyByWeight(weightMap, totalWeight)
	return selectedIndex, nil
}

// specialCycleCoreV2 执行特殊抽取选择（V2 实现，使用动态权重机）。
// specialOccIdx 为当前特殊出现序号（0-based），即本次特殊抽是周期内第几次特殊抽。
// items[i].JoinAt 表示第几次特殊抽（0-based）时该奖励才开始进入候选池。
func specialCycleCoreV2(unlocked map[int]bool, specialOccIdx int32, items []SpecialItem, dw *pd.DynamicWeights[int32]) (selectedIndex int, err error) {
	for i, item := range items {
		if specialOccIdx >= item.JoinAt && !unlocked[i] {
			unlocked[i] = true
			dw.Weights[i] = item.Quota
			dw.TtlWght += item.Quota
		}
	}
	if !dw.CanGenerate() {
		return -1, fmt.Errorf("no candidate: specialOccIdx=%d", specialOccIdx)
	}
	return dw.Generate().(int), nil
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