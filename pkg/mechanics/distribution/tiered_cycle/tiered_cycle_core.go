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

type Weight struct {
	weightMap   map[int]int32
	totalWeight int32
}

// NewWeight 构造权重结构，totalWeight 自动由 weightMap 计算，避免调用方手动维护
func NewWeight(weightMap map[int]int32) Weight {
	var total int32
	for _, v := range weightMap {
		total += v
	}
	return Weight{weightMap: weightMap, totalWeight: total}
}

// standardCycleCore 标准分布(纯粹的权重随机)
func standardCycleCore(w Weight) int {
	return pd.CalcKeyByWeight(w.weightMap, w.totalWeight)
}

// buildSpecialPlan 初始化特殊抽取的周期分布计划
func buildSpecialPlan(r *rand.Rand, pondCycle int32, pondMinInterval int32, count int32) []int32 {
	assert.Assert(count <= pondCycle)
	rangeMax := pondCycle - pondMinInterval*(count-1)
	assert.Assert(rangeMax > 0 && rangeMax <= pondCycle, rangeMax)

	// 从 [0, pondCycle-1] 选 count 个位置，最小间隔 pondMinInterval
	return sampling.SampleWithMinGap[int32](0, pondCycle-1, int(count), int(pondMinInterval), r)
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

// specialCycleCore 执行特殊抽取选择。
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
