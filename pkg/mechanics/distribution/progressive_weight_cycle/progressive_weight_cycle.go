// Package progressive_weight_cycle.

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
// 创建日期:2026/3/3
package progressive_weight_cycle

import (
	"fmt"
	
	pd "github.com/motocat46/yytools/pkg/algorithms/mathx/probability_distribution"
)

// Item 每个特殊结果的配置，下标即 Result.Index
type Item struct {
	Quota  int32 // 此特殊结果在一个周期内可出现的最大次数
	JoinAt int32 // 第几次特殊抽（0-based 特殊出现序号）时该奖励才开始进入候选池
}

// State 特殊层的运行时状态（仅属于特殊层）
type State struct {
	Dw       *pd.DynamicWeights[int, int32] // 动态权重机（key=items下标，weight=剩余配额）
	Unlocked map[int]bool                   // 已加入动态权重机的权重 key
}

// Layer 特殊层：封装特殊保底分布的规则与算法
type Layer struct {
	Items      []Item
	TotalQuota int32 // 所有项目的总配额
}

// NewWeightCycleLayer 创建特殊层，计算并缓存 items 的总配额。
func NewWeightCycleLayer(items []Item) *Layer {
	layer := &Layer{Items: items}
	layer.TotalQuota = TotalQuota(items)
	return layer
}

// NewState 创建空的特殊层运行时状态（动态权重机和已解锁集合均为空）。
func NewState() *State {
	return &State{
		Dw:       newEmptyDW(),
		Unlocked: make(map[int]bool),
	}
}

// Generate 执行一次特殊层抽取，返回命中的 items 下标；occIdx 为当前周期内第几次特殊抽（0-based）。
// 所有候选项权重耗尽时返回 (-1, error)。
func (l *Layer) Generate(state *State, occIdx int32) (int, error) {
	return specialCycleCoreV2(state.Unlocked, occIdx, l.Items, state.Dw)
}

// Reset 重置特殊层状态以开始新的一轮周期，清空动态权重机和已解锁记录。
func (s *State) Reset() {
	s.Dw = newEmptyDW()
	s.Unlocked = make(map[int]bool)
}

// TotalQuota 返回所有 SpecialItem 的 Quota 之和，即一个周期内特殊位置总数
func TotalQuota(items []Item) int32 {
	var n int32
	for _, item := range items {
		n += item.Quota
	}
	return n
}

// newEmptyDW 创建初始为空的动态权重机，供 NewState 和 Reset 使用。
func newEmptyDW() *pd.DynamicWeights[int, int32] {
	return pd.NewDynamicWeightsProgressive[int, int32]()
}

// checkSpecialCycleCoreParams 校验 specialCycleCore V1 的入参合法性。
func checkSpecialCycleCoreParams(used map[int32]int32, specialOccIdx int32, items []Item) error {
	if used == nil {
		return fmt.Errorf("invalid used map(nil)")
	}
	if specialOccIdx < 0 {
		return fmt.Errorf("invalid specialOccIdx:%d", specialOccIdx)
	}
	if len(items) == 0 {
		return fmt.Errorf("invalid Items: empty")
	}
	return nil
}

// specialCycleCore 执行特殊抽取选择（V1 实现，保留用于对比测试）。
// specialOccIdx 为当前特殊出现序号（0-based），即本次特殊抽是周期内第几次特殊抽。
// Items[i].JoinAt 表示第几次特殊抽（0-based）时该奖励才开始进入候选池。
// 例如：JoinAt=0 从第0次特殊抽起即可出现；JoinAt=3 从第4次特殊抽起才能出现。
func specialCycleCore(used map[int32]int32, specialOccIdx int32, items []Item) (selectedIndex int, err error) {
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
// Items[i].JoinAt 表示第几次特殊抽（0-based）时该奖励才开始进入候选池。
func specialCycleCoreV2(unlocked map[int]bool, specialOccIdx int32, items []Item, dw *pd.DynamicWeights[int, int32]) (selectedIndex int, err error) {
	for i, item := range items {
		if specialOccIdx >= item.JoinAt && !unlocked[i] {
			unlocked[i] = true
			dw.AddWeight(i, item.Quota)
		}
	}
	if !dw.CanGenerate() {
		return -1, fmt.Errorf("no candidate: specialOccIdx=%d", specialOccIdx)
	}
	return dw.Generate(), nil
}