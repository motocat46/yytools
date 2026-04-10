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
	"sync"
	"testing"

	weight_cycle "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
)

// TestCorrectness_EngineShared_StateIsolated_NoDataRace 命题：多 goroutine 共享 Engine、
// 各持独立 State，并发调用 Next/ResetCycle，-race 下不触发 data race。
func TestCorrectness_EngineShared_StateIsolated_NoDataRace(t *testing.T) {
	items := []weight_cycle.Item{
		{Quota: 2, JoinAt: 0},
		{Quota: 1, JoinAt: 2},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 20},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: 1, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}

	const goroutines = 20
	const opsPerGoroutine = 500

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			state := eng.NewState(rand.New(rand.NewPCG(seed, 0)))
			eng.Init(state)
			for range opsPerGoroutine {
				eng.NextAutoReset(state) //nolint
			}
		}(uint64(i))
	}
	wg.Wait()
}

// TestCorrectness_SameSeed_SameSequence 命题：两个 goroutine 使用相同种子各自创建 State，
// 并发运行若干周期后，各自产生完全相同的 plan 序列（种子决定序列，互不干扰）。
func TestCorrectness_SameSeed_SameSequence(t *testing.T) {
	const seed uint64 = 42
	const cycles = 5

	items := []weight_cycle.Item{
		{Quota: 2, JoinAt: 0},
		{Quota: 1, JoinAt: 2},
	}
	eng, err := New(Config{
		ConfigBase:     ConfigBase{CycleLen: 10},
		ConfigStandard: ConfigStandard{Weight: stdWeight},
		ConfigSpecial:  ConfigSpecial{MinInterval: 1, Items: items},
	})
	if err != nil {
		t.Fatal(err)
	}

	// 采集一个 goroutine 运行 cycles 个周期的 plan 序列
	collectPlans := func(seed uint64) [][]int32 {
		state := eng.NewState(rand.New(rand.NewPCG(seed, 0)))
		eng.Init(state)
		var plans [][]int32
		for range cycles {
			plan := make([]int32, len(state.Plan()))
			copy(plan, state.Plan())
			plans = append(plans, plan)
			// 跑完一个完整周期
			for {
				res, _ := eng.Next(state)
				if res.CycleEnd {
					eng.ResetCycle(state)
					break
				}
			}
		}
		return plans
	}

	var plans1, plans2 [][]int32
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		plans1 = collectPlans(seed)
	}()
	go func() {
		defer wg.Done()
		plans2 = collectPlans(seed)
	}()
	wg.Wait()

	if len(plans1) != len(plans2) {
		t.Fatalf("plan 数量不一致: %d vs %d", len(plans1), len(plans2))
	}
	for i := range plans1 {
		if len(plans1[i]) != len(plans2[i]) {
			t.Errorf("周期 %d plan 长度不一致: %d vs %d", i, len(plans1[i]), len(plans2[i]))
			continue
		}
		for j, p := range plans1[i] {
			if p != plans2[i][j] {
				t.Errorf("周期 %d plan[%d] 不一致: %d vs %d", i, j, p, plans2[i][j])
			}
		}
	}
}
