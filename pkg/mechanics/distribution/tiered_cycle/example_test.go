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

	weight_cycle "github.com/stormYuanYang/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
)

// Example_gachaSystem 演示游戏抽卡保底系统。
//
// 场景：100 抽一个大周期，4 种奖励分层解锁：
//
//	sum(Quota) = 4+3+2+1 = 10 个保底位置，随机分布在 100 抽中。
//
// JoinAt 梯度解锁效果：
//
//	特殊序号 0–3  → 仅 item[0] 可选（4星角色碎片，兜底选项）
//	特殊序号 4–6  → item[0], item[1] 可选（4星武器加入）
//	特殊序号 7–8  → item[0..2] 可选（5星限定武器加入）
//	特殊序号 9    → item[0..3] 全部可选（5星限定角色，终极保底）
func Example_gachaSystem() {
	items := []weight_cycle.Item{
		{Quota: 4, JoinAt: 0}, // 4星角色碎片：随时可出
		{Quota: 3, JoinAt: 4}, // 4星武器：从第5次保底起
		{Quota: 2, JoinAt: 7}, // 5星限定武器：从第8次保底起
		{Quota: 1, JoinAt: 9}, // 5星限定角色：仅最后1次保底可出
	}
	cfg := Config{
		ConfigBase: ConfigBase{
			R:        rand.New(rand.NewPCG(42, 0)),
			CycleLen: 100,
		},
		ConfigStandard: ConfigStandard{
			Weight: NewWeight(map[int]int32{0: 70, 1: 20, 2: 10}),
		},
		ConfigSpecial: ConfigSpecial{
			MinInterval: 5,
			Items:       items,
		},
	}

	eng, err := New(cfg)
	if err != nil {
		panic(err)
	}
	state := NewState()
	eng.Init(state)

	standardCount := 0
	specialCounts := make([]int, len(items))
	// 抽3个周期，engine自动重置，调用者只需要构造初始配置信息即可
	length := 3 * cfg.CycleLen
	for i := 0; i < int(length); i++ {
		res, callErr := eng.NextAutoReset(state)
		if callErr != nil {
			continue
		}
		switch res.Type {
		case Standard:
			standardCount++
		case Special:
			specialCounts[res.Index]++
		}
	}

	fmt.Printf("标准抽次数:   %d\n", standardCount)
	fmt.Printf("4星角色碎片: %d次\n", specialCounts[0])
	fmt.Printf("4星武器:     %d次\n", specialCounts[1])
	fmt.Printf("5星限定武器: %d次\n", specialCounts[2])
	fmt.Printf("5星限定角色: %d次\n", specialCounts[3])
}

// Example_multipleStates 演示同一 Engine 驱动多个玩家的独立 State。
//
// Engine 保存规则（不可变），State 保存每位玩家的进度。
// 多个 State 互不干扰，体现 Engine/State 职责分离的设计价值。
func Example_multipleStates() {
	items := []weight_cycle.Item{
		{Quota: 3, JoinAt: 0},
		{Quota: 2, JoinAt: 3},
	}
	cfg := Config{
		ConfigBase: ConfigBase{
			R:        rand.New(rand.NewPCG(1, 0)),
			CycleLen: 20,
		},
		ConfigStandard: ConfigStandard{
			Weight: NewWeight(map[int]int32{0: 1}),
		},
		ConfigSpecial: ConfigSpecial{
			MinInterval: 2,
			Items:       items,
		},
	}

	eng, err := New(cfg)
	if err != nil {
		panic(err)
	}

	// 3 个玩家，独立进度
	states := make([]*State, 3)
	for i := range states {
		states[i] = NewState()
		eng.Init(states[i])
	}

	// 每个玩家独立完整跑一个周期
	for playerID, st := range states {
		specialCount := 0
		for j := 0; j < int(cfg.CycleLen); j++ {
			res, _ := eng.NextAutoReset(st)
			if res.Type == Special {
				specialCount++
			}
		}
		fmt.Printf("玩家%d: 一个周期内特殊次数=%d\n", playerID, specialCount)
	}
}
