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
// 创建日期:2026/4/10
package tiered_cycle

import (
	"math/rand/v2"
	"testing"

	weight_cycle "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
)

// verifyCyclePlan 验证单个周期的 plan 满足三个约束：
//   - 长度等于 sum(Quota)
//   - 严格升序
//   - 相邻位置差值 >= minInterval（语义来自 sampling.SampleWithMinGap：gap=N 表示相邻差值 >= N）
func verifyCyclePlan(t *testing.T, plan []int32, minInterval int32, expectedLen int, cycleIdx int) {
	t.Helper()
	if len(plan) != expectedLen {
		t.Errorf("cycle %d: plan 长度=%d，期望 %d", cycleIdx, len(plan), expectedLen)
		return
	}
	for i := 1; i < len(plan); i++ {
		if plan[i] <= plan[i-1] {
			t.Errorf("cycle %d: plan 未升序：plan[%d]=%d >= plan[%d]=%d",
				cycleIdx, i-1, plan[i-1], i, plan[i])
		}
		if gap := plan[i] - plan[i-1]; gap < minInterval {
			t.Errorf("cycle %d: plan[%d]=%d plan[%d]=%d gap=%d，期望 >= %d（MinInterval）",
				cycleIdx, i-1, plan[i-1], i, plan[i], gap, minInterval)
		}
	}
}

// verifyCycleQuota 验证一个周期内各 item 命中次数不超过 Quota。
func verifyCycleQuota(t *testing.T, itemCounts []int32, items []weight_cycle.Item, cycleIdx int) {
	t.Helper()
	for i, item := range items {
		if itemCounts[i] > item.Quota {
			t.Errorf("cycle %d: item[%d] 命中 %d 次，超出 Quota=%d",
				cycleIdx, i, itemCounts[i], item.Quota)
		}
	}
}

// runCyclesWithInvariantCheck 运行 numCycles 个完整周期，每个周期后验证不变量。
// 每个周期使用 eng.Next + eng.ResetCycle（而非 NextAutoReset），使两条代码路径都得到覆盖。
func runCyclesWithInvariantCheck(
	t *testing.T,
	eng *Engine,
	state *State,
	items []weight_cycle.Item,
	minInterval int32,
	numCycles int,
) {
	t.Helper()
	expectedSpecials := int(weight_cycle.TotalQuota(items))
	itemCounts := make([]int32, len(items))

	for c := range numCycles {
		// 记录周期开始时的 plan（Init 或 ResetCycle 后立即有效）
		plan := make([]int32, len(state.Plan()))
		copy(plan, state.Plan())
		verifyCyclePlan(t, plan, minInterval, expectedSpecials, c)

		// 重置计数器
		for i := range itemCounts {
			itemCounts[i] = 0
		}

		// 推进一整个周期
		for {
			res, callErr := eng.Next(state)
			if callErr == nil && res.Type == Special {
				itemCounts[res.Index]++
			}
			if res.CycleEnd {
				break
			}
		}

		// 验证 Quota 约束
		verifyCycleQuota(t, itemCounts, items, c)

		// 手动重置（下一周期 state.Plan() 更新）
		eng.ResetCycle(state)
	}
}

// TestEngine_LargeScale_CycleInvariants 大规模不变量验证。
// 1000 个周期 × 100 次抽取 = 100,000 次抽取，验证每个周期的 plan 约束和 Quota 约束。
func TestEngine_LargeScale_CycleInvariants(t *testing.T) {
	const (
		cycleLen    = 100
		minInterval = 5
		numCycles   = 1_000 // 1000 × 100 = 100,000 次抽取
	)
	items := []weight_cycle.Item{
		{Quota: 4, JoinAt: 0},
		{Quota: 3, JoinAt: 4},
		{Quota: 2, JoinAt: 7},
		{Quota: 1, JoinAt: 9},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: cycleLen},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: minInterval, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := eng.NewState(rand.New(rand.NewPCG(42, 0)))
	eng.Init(state)
	runCyclesWithInvariantCheck(t, eng, state, items, minInterval, numCycles)
}

// TestEngine_Stress_CycleInvariants 压力测试。
// 10,000 个周期 × 100 次抽取 = 1,000,000 次抽取，-short 跳过。
func TestEngine_Stress_CycleInvariants(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模压力测试")
	}
	const (
		cycleLen    = 100
		minInterval = 5
		numCycles   = 10_000 // 10,000 × 100 = 1,000,000 次抽取
	)
	items := []weight_cycle.Item{
		{Quota: 4, JoinAt: 0},
		{Quota: 3, JoinAt: 4},
		{Quota: 2, JoinAt: 7},
		{Quota: 1, JoinAt: 9},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: cycleLen},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: minInterval, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := eng.NewState(rand.New(rand.NewPCG(99, 0)))
	eng.Init(state)
	runCyclesWithInvariantCheck(t, eng, state, items, minInterval, numCycles)
}
