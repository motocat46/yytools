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
package progressive_weight_cycle_test

import (
	"fmt"

	pwc "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
)

// ExampleLayer_rewardBox 演示副本通关宝箱的渐进式奖励分配。
//
// 场景：玩家每通关一次副本可获得一个特殊奖励，奖励共 5 种，随通关次数逐步解锁：
//
//	item[0] 普通材料  Quota=4, JoinAt=0 → 任意时刻均可出现，基础兜底
//	item[1] 稀有材料  Quota=2, JoinAt=2 → 第 3 次特殊抽起解锁（确保前两次有材料积累）
//	item[2] 史诗装备  Quota=1, JoinAt=5 → 第 6 次特殊抽起解锁
//	item[3] 传说碎片  Quota=1, JoinAt=7 → 第 8 次特殊抽起解锁
//	item[4] 神话装备  Quota=1, JoinAt=8 → 仅最后一次特殊抽可出现
//
// 一个大周期共 9 次特殊抽（sum(Quota) = 4+2+1+1+1 = 9）。
func ExampleLayer_rewardBox() {
	items := []pwc.Item{
		{Quota: 4, JoinAt: 0}, // 普通材料：随时可出
		{Quota: 2, JoinAt: 2}, // 稀有材料：第3次起解锁
		{Quota: 1, JoinAt: 5}, // 史诗装备：第6次起解锁
		{Quota: 1, JoinAt: 7}, // 传说碎片：第8次起解锁
		{Quota: 1, JoinAt: 8}, // 神话装备：仅最后一次
	}
	names := []string{"普通材料", "稀有材料", "史诗装备", "传说碎片", "神话装备"}

	layer := pwc.NewWeightCycleLayer(items)
	total := pwc.TotalQuota(items)

	// 模拟一个周期：连续通关 total 次副本
	state := pwc.NewState()
	counts := make([]int, len(items))
	for occIdx := range total {
		idx, err := layer.Generate(state, occIdx)
		if err != nil {
			// 正常业务中不应出现，此处仅作防御性处理
			fmt.Printf("occIdx=%d 抽取异常: %v\n", occIdx, err)
			continue
		}
		counts[idx]++
	}

	fmt.Printf("一个周期（%d 次特殊抽）奖励统计:\n", total)
	for i, name := range names {
		fmt.Printf("  %s(Quota=%d): 出现 %d 次\n", name, items[i].Quota, counts[i])
	}
}

// ExampleLayer_multipleStates 演示同一 Layer 规则驱动多个独立 State（多玩家场景）。
//
// Layer 持有不可变规则，State 持有每个玩家的独立运行时进度，二者严格分离。
// 多个玩家共用同一套奖励规则，但各自的抽取进度互不影响。
func ExampleLayer_multipleStates() {
	items := []pwc.Item{
		{Quota: 3, JoinAt: 0}, // 普通奖励：随时可出
		{Quota: 1, JoinAt: 2}, // 稀有奖励：第3次起解锁
		{Quota: 1, JoinAt: 3}, // 终极奖励：第4次起解锁
	}
	names := []string{"普通奖励", "稀有奖励", "终极奖励"}

	layer := pwc.NewWeightCycleLayer(items)
	total := pwc.TotalQuota(items) // = 5

	const playerCount = 3
	states := make([]*pwc.State, playerCount)
	for i := range states {
		states[i] = pwc.NewState()
	}

	// 每位玩家独立完成一个周期
	for playerID, state := range states {
		counts := make([]int, len(items))
		for occIdx := range total {
			idx, err := layer.Generate(state, occIdx)
			if err != nil {
				continue
			}
			counts[idx]++
		}
		fmt.Printf("玩家%d:", playerID)
		for i, name := range names {
			fmt.Printf(" %s×%d", name, counts[i])
		}
		fmt.Println()
	}
}
