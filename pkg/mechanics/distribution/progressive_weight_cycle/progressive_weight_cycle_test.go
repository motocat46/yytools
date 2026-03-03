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
	"testing"
)

// TestTotalQuota 验证 TotalQuota 正确累加所有 Item 的 Quota
func TestTotalQuota(t *testing.T) {
	cases := []struct {
		items  []Item
		expect int32
	}{
		{nil, 0},
		{[]Item{}, 0},
		{[]Item{{Quota: 1, JoinAt: 0}}, 1},
		{[]Item{{Quota: 2, JoinAt: 0}, {Quota: 3, JoinAt: 1}}, 5},
	}
	for _, c := range cases {
		got := TotalQuota(c.items)
		if got != c.expect {
			t.Errorf("TotalQuota(%v)=%d，期望 %d", c.items, got, c.expect)
		}
	}
}

// TestState_Reset 验证 Reset 后状态归零
func TestState_Reset(t *testing.T) {
	s := NewState()
	// 模拟消耗后重置
	s.Dw.Weights[0] = 3
	s.Dw.TtlWght = 3
	s.Unlocked[0] = true

	s.Reset()

	if s.Dw.TtlWght != 0 {
		t.Errorf("Reset 后 Dw.TtlWght 应为 0，实际 %d", s.Dw.TtlWght)
	}
	if len(s.Dw.Weights) != 0 {
		t.Errorf("Reset 后 Dw.Weights 应为空，实际 %v", s.Dw.Weights)
	}
	if len(s.Unlocked) != 0 {
		t.Errorf("Reset 后 Unlocked 应为空，实际 %v", s.Unlocked)
	}
}

// TestLayer_Generate_JoinAt 验证 V2 实现的 JoinAt 约束（各阶段候选池正确）
//
// items: quota=[2,2], joinAt=[0,2]
//   - occIdx=0,1: 只有 item[0] 在候选池（item[1].JoinAt=2 未满足）→ 单一候选，结果确定
//   - occIdx=2: item[0] 配额耗尽，item[1] 进入候选池 → 结果确定为 item[1]
func TestLayer_Generate_JoinAt(t *testing.T) {
	items := []Item{
		{Quota: 2, JoinAt: 0},
		{Quota: 2, JoinAt: 2},
	}
	layer := NewWeightCycleLayer(items)
	state := NewState()

	// occIdx=0：只有 item[0]，单候选必命中
	idx, err := layer.Generate(state, 0)
	if err != nil {
		t.Fatalf("occIdx=0 不应报错: %v", err)
	}
	if idx != 0 {
		t.Errorf("occIdx=0 应命中 item[0]，实际 idx=%d", idx)
	}

	// occIdx=1：item[0] 还剩 1 次配额，item[1].JoinAt=2 未满足，单候选
	idx, err = layer.Generate(state, 1)
	if err != nil {
		t.Fatalf("occIdx=1 不应报错: %v", err)
	}
	if idx != 0 {
		t.Errorf("occIdx=1 应命中 item[0]（配额剩 1），实际 idx=%d", idx)
	}

	// occIdx=2：item[0] 配额耗尽，item[1] 满足 JoinAt=2，单候选
	idx, err = layer.Generate(state, 2)
	if err != nil {
		t.Fatalf("occIdx=2 不应报错: %v", err)
	}
	if idx != 1 {
		t.Errorf("occIdx=2 应命中 item[1]（item[0] 已耗尽），实际 idx=%d", idx)
	}
}

// TestLayer_Generate_NoCandidate JoinAt 门槛未满足时返回 error
func TestLayer_Generate_NoCandidate(t *testing.T) {
	items := []Item{{Quota: 1, JoinAt: 3}}
	layer := NewWeightCycleLayer(items)
	state := NewState()

	_, err := layer.Generate(state, 2) // occIdx=2 < JoinAt=3
	if err == nil {
		t.Error("occIdx=2 < JoinAt=3，应返回 no candidate error")
	}
}

// TestLayer_Generate_QuotaEnforced 验证 Quota 耗尽后不再输出
func TestLayer_Generate_QuotaEnforced(t *testing.T) {
	items := []Item{{Quota: 2, JoinAt: 0}}
	layer := NewWeightCycleLayer(items)
	state := NewState()

	// 连续抽取，第3次应无候选
	for i := range int32(2) {
		_, err := layer.Generate(state, i)
		if err != nil {
			t.Fatalf("第 %d 次抽取不应报错: %v", i, err)
		}
	}
	_, err := layer.Generate(state, 2)
	if err == nil {
		t.Error("Quota=2 耗尽后第3次抽取应返回 no candidate error")
	}
}

// TestSpecialCycleCore_V1 V1 实现的直接单元测试，保留用于与 V2 行为对比
func TestSpecialCycleCore_V1(t *testing.T) {
	items := []Item{{Quota: 1, JoinAt: 5}}

	// specialOccIdx=3 < JoinAt=5，应报 no candidate
	_, err := specialCycleCore(map[int32]int32{}, 3, items)
	if err == nil {
		t.Fatal("期望 error: JoinAt=5 > specialOccIdx=3")
	}

	// specialOccIdx=5 >= JoinAt=5，不应报错
	_, err = specialCycleCore(map[int32]int32{}, 5, items)
	if err != nil {
		t.Fatalf("specialOccIdx=5 >= JoinAt=5 时不应报错: %v", err)
	}
}
