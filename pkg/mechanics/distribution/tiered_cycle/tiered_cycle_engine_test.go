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
	"testing"
)

// newTestRand 用固定种子创建随机源，保证计划生成可重放
func newTestRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 0))
}

// stdWeight 用于标准分布的简单权重（只有一种结果）
var stdWeight = NewWeight(map[int]int32{0: 1, 1: 1, 2: 1})

// TestEngine_Config_Validation 验证各种非法 Config 返回 error
func TestEngine_Config_Validation(t *testing.T) {
	r := newTestRand(1)

	cases := []struct {
		name string
		cfg  Config
	}{
		{
			name: "CycleLen=0",
			cfg: Config{
				ConfigBase:     ConfigBase{CycleLen: 0, R: r},
				ConfigStandard: ConfigStandard{Weight: stdWeight},
			},
		},
		{
			name: "CycleLen<0",
			cfg: Config{
				ConfigBase:     ConfigBase{CycleLen: -1, R: r},
				ConfigStandard: ConfigStandard{Weight: stdWeight},
			},
		},
		{
			name: "Items非空但R为nil",
			cfg: Config{
				ConfigBase:     ConfigBase{CycleLen: 10},
				ConfigStandard: ConfigStandard{Weight: stdWeight},
				ConfigSpecial:  ConfigSpecial{Items: []SpecialItem{{Quota: 1, JoinAt: 0}}},
			},
		},
		{
			// sum(Quota)=3 > CycleLen=2
			name: "sum(Quota)>CycleLen",
			cfg: Config{
				ConfigBase:     ConfigBase{CycleLen: 2, R: r},
				ConfigStandard: ConfigStandard{Weight: stdWeight},
				ConfigSpecial:  ConfigSpecial{Items: []SpecialItem{{1, 0}, {1, 0}, {1, 0}}},
			},
		},
		{
			name: "MinInterval过大导致不可行",
			// CycleLen=5, MinInterval=3, n=3: 5-3*(3-1)=5-6=-1 < 3
			cfg: Config{
				ConfigBase:     ConfigBase{CycleLen: 5, R: r},
				ConfigStandard: ConfigStandard{Weight: stdWeight},
				ConfigSpecial:  ConfigSpecial{MinInterval: 3, Items: []SpecialItem{{1, 0}, {1, 0}, {1, 0}}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := New(c.cfg)
			if err == nil {
				t.Errorf("期望返回 error，但 New() 成功了")
			}
		})
	}

	// 合法配置应该成功
	t.Run("合法配置", func(t *testing.T) {
		_, err := New(Config{
			ConfigBase:     ConfigBase{CycleLen: 10, R: newTestRand(1)},
			ConfigStandard: ConfigStandard{Weight: stdWeight},
			ConfigSpecial:  ConfigSpecial{MinInterval: 1, Items: []SpecialItem{{1, 0}, {1, 0}}},
		})
		if err != nil {
			t.Errorf("合法配置不应返回 error: %v", err)
		}
	})

	// 无特殊分布时 R 可以为 nil
	t.Run("无Items时R可nil", func(t *testing.T) {
		_, err := New(Config{
			ConfigBase:     ConfigBase{CycleLen: 10},
			ConfigStandard: ConfigStandard{Weight: stdWeight},
		})
		if err != nil {
			t.Errorf("无 Items 时 R=nil 应合法: %v", err)
		}
	})
}

// runOneCycle 运行一整个周期，返回每次 Next 前的 posInCycle 及对应结果
func runOneCycle(t *testing.T, eng *Engine, state *State) (positions []int32, results []Result, errs []error) {
	t.Helper()
	for {
		pos := state.PosInCycle()
		res, err := eng.Next(state)
		positions = append(positions, pos)
		results = append(results, res)
		errs = append(errs, err)
		if res.CycleEnd {
			break
		}
	}
	return
}

// TestEngine_FullCycle_PositionTypeMapping 全周期内：plan 位置 → Special，其余 → Standard
func TestEngine_FullCycle_PositionTypeMapping(t *testing.T) {
	items := []SpecialItem{
		{Quota: 1, JoinAt: 0},
		{Quota: 1, JoinAt: 0},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 10, R: newTestRand(42)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: 1, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	// 记录特殊计划位置集合（Init 之后立刻获取，此时 posInCycle=0）
	planSet := make(map[int32]bool)
	for _, p := range state.Plan() {
		planSet[p] = true
	}

	positions, results, errs := runOneCycle(t, eng, state)

	for i, pos := range positions {
		err := errs[i]
		res := results[i]
		if err != nil {
			// 出错时类型为 Invalid，跳过类型断言（由 SpecialNoCandidate 专项覆盖）
			continue
		}
		if planSet[pos] {
			if res.Type != Special {
				t.Errorf("pos=%d 在 plan 中，期望 Special，实际 %v", pos, res.Type)
			}
		} else {
			if res.Type != Standard {
				t.Errorf("pos=%d 不在 plan 中，期望 Standard，实际 %v", pos, res.Type)
			}
		}
	}
}

// TestEngine_FullCycle_JoinAtEnforced JoinAt 是特殊出现序号（0-based），非绝对位置
//
// items: quota=[2,2,2], joinAt=[0,2,4]，sum=6
//   - occIdx 0,1: 只有 item0 可选（item1.JoinAt=2>0,1；item2.JoinAt=4>0,1）
//   - occIdx 2,3: item0,item1 可选
//   - occIdx 4,5: 三者都可选
func TestEngine_FullCycle_JoinAtEnforced(t *testing.T) {
	items := []SpecialItem{
		{Quota: 2, JoinAt: 0},
		{Quota: 2, JoinAt: 2},
		{Quota: 2, JoinAt: 4},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 20, R: newTestRand(7)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: 2, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	for cycle := 0; cycle < 3; cycle++ {
		// occIdx 跟踪当前周期内已遇到的特殊位置数（等于传给 specialCycleCore 的 specialOccIdx）
		occIdx := int32(0)
		for {
			res, err2 := eng.Next(state)
			if err2 == nil && res.Type == Special {
				item := items[res.Index]
				if item.JoinAt > occIdx {
					t.Errorf("cycle=%d occIdx=%d: item[%d].JoinAt=%d 未满足（应 <= %d）",
						cycle, occIdx, res.Index, item.JoinAt, occIdx)
				}
				occIdx++
			}
			if res.CycleEnd {
				break
			}
		}
		eng.ResetCycle(state)
	}
}

// TestEngine_FullCycle_QuotaEnforced 每个周期内 Special 命中次数不超过 Quota
func TestEngine_FullCycle_QuotaEnforced(t *testing.T) {
	items := []SpecialItem{
		{Quota: 1, JoinAt: 0},
		{Quota: 2, JoinAt: 0},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 20, R: newTestRand(99)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: 1, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	counts := make([]int32, len(items))
	for cycle := 0; cycle < 5; cycle++ {
		for i := range counts {
			counts[i] = 0
		}
		_, results, errs := runOneCycle(t, eng, state)
		for i, res := range results {
			if errs[i] == nil && res.Type == Special {
				counts[res.Index]++
			}
		}
		for i, item := range items {
			if counts[i] > item.Quota {
				t.Errorf("cycle=%d item[%d] 命中 %d 次，超出 Quota=%d",
					cycle, i, counts[i], item.Quota)
			}
		}
		eng.ResetCycle(state)
	}
}

// TestEngine_FullCycle_CycleEnd 只有最后一次 Next 返回 CycleEnd=true
func TestEngine_FullCycle_CycleEnd(t *testing.T) {
	const cycleLen = 8
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: cycleLen, R: newTestRand(1)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: 2, Items: []SpecialItem{{1, 0}, {1, 0}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	_, results, _ := runOneCycle(t, eng, state)

	if len(results) != cycleLen {
		t.Fatalf("期望 %d 次 Next，实际 %d 次", cycleLen, len(results))
	}
	for i, res := range results {
		isLast := i == cycleLen-1
		if res.CycleEnd != isLast {
			t.Errorf("results[%d].CycleEnd=%v，期望 %v", i, res.CycleEnd, isLast)
		}
	}
}

// TestEngine_ResetCycle_StateCleared Reset 后状态正确归零、新 Plan 生成
func TestEngine_ResetCycle_StateCleared(t *testing.T) {
	items := []SpecialItem{{Quota: 1, JoinAt: 0}}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 5, R: newTestRand(3)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	// 跑几步推进 posInCycle 和特殊层状态
	eng.Next(state) //nolint
	eng.Next(state) //nolint

	if state.PosInCycle() != 2 {
		t.Fatalf("期望 posInCycle=2，实际 %d", state.PosInCycle())
	}

	eng.ResetCycle(state)

	if state.PosInCycle() != 0 {
		t.Errorf("Reset 后 posInCycle 应为 0，实际 %d", state.PosInCycle())
	}
	if state.special.dw.TtlWght != 0 {
		t.Errorf("Reset 后 dw.TtlWght 应为 0，实际 %d", state.special.dw.TtlWght)
	}
	if len(state.special.unlocked) != 0 {
		t.Errorf("Reset 后 unlocked 应为空，实际 %v", state.special.unlocked)
	}
	// Plan 长度应等于 sum(Quota)，此处 totalQuota=1=len(items)，值相同但语义为特殊位置总数
	if len(state.Plan()) != int(totalQuota(items)) {
		t.Errorf("Reset 后 Plan 长度应为 %d（sum of Quota），实际 %d", totalQuota(items), len(state.Plan()))
	}
}

// TestEngine_NextAutoReset 到达 CycleEnd 后自动 Reset，下轮继续
func TestEngine_NextAutoReset(t *testing.T) {
	const cycleLen = 6
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: cycleLen, R: newTestRand(5)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{Items: []SpecialItem{{1, 0}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	// 跑 2 个完整周期
	const totalCycles = 2
	cycleEndCount := 0
	for i := 0; i < cycleLen*totalCycles; i++ {
		res, _ := eng.NextAutoReset(state)
		if res.CycleEnd {
			cycleEndCount++
			// 自动 Reset 后 posInCycle 应立即为 0
			if state.PosInCycle() != 0 {
				t.Errorf("第 %d 次 CycleEnd 后 posInCycle 应为 0，实际 %d",
					cycleEndCount, state.PosInCycle())
			}
		}
	}
	if cycleEndCount != totalCycles {
		t.Errorf("期望 %d 次 CycleEnd，实际 %d 次", totalCycles, cycleEndCount)
	}
}

// TestEngine_NoSpecials Items 为空时全部走 Standard
func TestEngine_NoSpecials(t *testing.T) {
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 10},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		// R=nil, Items=nil —— 合法：无特殊分布
	})
	if err != nil {
		t.Fatal(err)
	}
	state := NewState()
	eng.Init(state)

	if len(state.Plan()) != 0 {
		t.Errorf("无 Items 时 Plan 应为空，实际 %v", state.Plan())
	}

	_, results, errs := runOneCycle(t, eng, state)
	for i, res := range results {
		if errs[i] != nil {
			t.Errorf("无 Items 时不应返回 error: %v", errs[i])
		}
		if res.Type != Standard {
			t.Errorf("无 Items 时 results[%d] 应为 Standard，实际 %v", i, res.Type)
		}
	}
}

// TestSpecialCycleCore_V1 保留旧实现的直接单元测试，用于对比 V1/V2 决策
func TestSpecialCycleCore_V1(t *testing.T) {
	items := []SpecialItem{{Quota: 1, JoinAt: 5}}
	_, err := specialCycleCore(map[int32]int32{}, 3, items)
	if err == nil {
		t.Fatal("期望 error: JoinAt=5 > specialOccIdx=3")
	}
	_, err = specialCycleCore(map[int32]int32{}, 5, items)
	if err != nil {
		t.Fatalf("specialOccIdx=5 >= JoinAt=5 时不应报错: %v", err)
	}
}

// TestEngine_SpecialNoCandidate JoinAt（特殊序号门槛）不满足时返回 error，但状态仍推进
func TestEngine_SpecialNoCandidate(t *testing.T) {
	// sum(quota)=1，plan 只有 1 个位置，该位置 occIdx 恒为 0，JoinAt=1 > 0 → 永远 no candidate
	items2 := []SpecialItem{{Quota: 1, JoinAt: 1}}
	e2, err2 := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 5, R: newTestRand(42)},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{Items: items2},
	})
	if err2 != nil {
		t.Fatal(err2)
	}
	s2 := NewState()
	e2.Init(s2)

	errCount := 0
	for {
		prevPos := s2.PosInCycle()
		res, callErr := e2.Next(s2)
		if callErr != nil {
			errCount++
			// 即使出错，posInCycle 也应前进
			if s2.PosInCycle() != prevPos+1 {
				t.Errorf("出错后 posInCycle 应前进，%d → 期望 %d，实际 %d",
					prevPos, prevPos+1, s2.PosInCycle())
			}
		}
		if res.CycleEnd {
			break
		}
	}
	if errCount == 0 {
		t.Error("期望至少一次 no candidate error（JoinAt=1 > occIdx=0 恒不满足）")
	}
}

// TestEngine_Replay 相同种子产生完全相同的周期计划（plan 位置）
func TestEngine_Replay(t *testing.T) {
	items := []SpecialItem{
		{Quota: 1, JoinAt: 0},
		{Quota: 1, JoinAt: 0},
	}
	const seed uint64 = 12345
	const cycleLen = 10

	getPlan := func() []int32 {
		eng, err := New(Config{
			ConfigBase:     ConfigBase{CycleLen: cycleLen, R: newTestRand(seed)},
			ConfigStandard: ConfigStandard{Weight: stdWeight},
			ConfigSpecial:  ConfigSpecial{MinInterval: 1, Items: items},
		})
		if err != nil {
			t.Fatal(err)
		}
		state := NewState()
		eng.Init(state)
		plan := make([]int32, len(state.Plan()))
		copy(plan, state.Plan())
		return plan
	}

	plan1 := getPlan()
	plan2 := getPlan()

	if len(plan1) != len(plan2) {
		t.Fatalf("plan 长度不一致: %d vs %d", len(plan1), len(plan2))
	}
	for i, p := range plan1 {
		if p != plan2[i] {
			t.Errorf("plan[%d] 不一致: %d vs %d", i, p, plan2[i])
		}
	}
}

// --- 各层独立单元测试 ---

// TestStandardLayer_Generate_Distribution 验证普通层生成结果在权重范围内
func TestStandardLayer_Generate_Distribution(t *testing.T) {
	sl := newStandardLayer(NewWeight(map[int]int32{0: 1, 1: 2, 2: 3}))
	state := sl.NewState()
	for range 1000 {
		idx := sl.Generate(&state)
		if idx < 0 || idx > 2 {
			t.Errorf("Generate 返回了非法下标 %d（应在 [0,2]）", idx)
		}
	}
}

// TestStandardLayer_Reset_NoOp 验证普通层 Reset 是无状态空操作
func TestStandardLayer_Reset_NoOp(t *testing.T) {
	sl := newStandardLayer(NewWeight(map[int]int32{0: 1}))
	state := sl.NewState()
	// Reset 不应 panic，也不改变任何外部可观测状态
	sl.Reset(&state)
	idx := sl.Generate(&state)
	if idx != 0 {
		t.Errorf("Reset 后 Generate 应返回 0，实际 %d", idx)
	}
}

// TestSpecialLayer_GetOccIdx 验证 GetOccIdx 正确返回特殊位置序号
func TestSpecialLayer_GetOccIdx(t *testing.T) {
	items := []SpecialItem{{Quota: 1, JoinAt: 0}, {Quota: 1, JoinAt: 0}}
	sl := newSpecialLayer(items, 1)
	state := sl.NewState()
	// 手动注入计划，隔离随机性
	state.plan = []int32{3, 7}

	cases := []struct {
		pos    int32
		expect int
	}{
		{0, -1},
		{3, 0},
		{7, 1},
		{5, -1},
		{2, -1},
	}
	for _, c := range cases {
		got := sl.GetOccIdx(&state, c.pos)
		if got != c.expect {
			t.Errorf("GetOccIdx(pos=%d)=%d，期望 %d", c.pos, got, c.expect)
		}
	}
}

// TestSpecialLayer_Generate_JoinAt 直接测试特殊层的 JoinAt 约束（各阶段候选池正确）
//
// items: quota=[2,2], joinAt=[0,2]
//   - occIdx=0,1: 只有 item[0] 在候选池（item[1].JoinAt=2 未满足）→ 单一候选，结果确定
//   - occIdx=2: item[0] 配额耗尽，item[1] 进入候选池 → 结果确定为 item[1]
func TestSpecialLayer_Generate_JoinAt(t *testing.T) {
	items := []SpecialItem{
		{Quota: 2, JoinAt: 0},
		{Quota: 2, JoinAt: 2},
	}
	sl := newSpecialLayer(items, 0)
	state := sl.NewState()

	// occIdx=0：只有 item[0]，单候选必命中
	idx, err := sl.Generate(&state, 0)
	if err != nil {
		t.Fatalf("occIdx=0 不应报错: %v", err)
	}
	if idx != 0 {
		t.Errorf("occIdx=0 应命中 item[0]，实际 idx=%d", idx)
	}

	// occIdx=1：item[0] 还剩 1 次配额，item[1].JoinAt=2 未满足，单候选
	idx, err = sl.Generate(&state, 1)
	if err != nil {
		t.Fatalf("occIdx=1 不应报错: %v", err)
	}
	if idx != 0 {
		t.Errorf("occIdx=1 应命中 item[0]（配额剩 1），实际 idx=%d", idx)
	}

	// occIdx=2：item[0] 配额耗尽，item[1] 满足 JoinAt=2，单候选
	idx, err = sl.Generate(&state, 2)
	if err != nil {
		t.Fatalf("occIdx=2 不应报错: %v", err)
	}
	if idx != 1 {
		t.Errorf("occIdx=2 应命中 item[1]（item[0] 已耗尽），实际 idx=%d", idx)
	}
}

// TestSpecialLayer_Generate_NoCandidate JoinAt 门槛未满足时返回 error
func TestSpecialLayer_Generate_NoCandidate(t *testing.T) {
	items := []SpecialItem{{Quota: 1, JoinAt: 3}}
	sl := newSpecialLayer(items, 0)
	state := sl.NewState()

	_, err := sl.Generate(&state, 2) // occIdx=2 < JoinAt=3
	if err == nil {
		t.Error("occIdx=2 < JoinAt=3，应返回 no candidate error")
	}
}
