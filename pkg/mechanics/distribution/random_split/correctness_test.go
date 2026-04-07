// Copyright 2026 yangyuan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package random_split correctness_test.go
// 正确性命题测试：用 Simulate 驱动 10 万轮，固定种子 42，三种策略分别验证 5 条命题。
// 运行方式（必须带 -race）：
//
//	go test -race -run TestCorrectness -v ./pkg/mechanics/distribution/random_split/
package random_split

import (
	"math"
	"math/rand/v2"
	"testing"
)

const (
	propRounds = 100_000
	propSeed   = 42
)

var propState = State{RemainAmount: 100, RemainCount: 10, MinPerPart: 1}

func allStrategies() map[string]SampleFunc {
	return map[string]SampleFunc{
		"DoubleMean": DoubleMean(),
		"Uniform":    Uniform(),
		"Fixed":      Fixed(),
	}
}

// TestCorrectness_Conservation 命题1：每轮 sum(allocations) == RemainAmount，无一例外。
func TestCorrectness_Conservation(t *testing.T) {
	for name, fn := range allStrategies() {
		t.Run(name, func(t *testing.T) {
			result, err := Simulate(propState, fn, propRounds, propSeed)
			if err != nil {
				t.Fatalf("Simulate() 失败: %v", err)
			}
			if err := result.CheckConservation(); err != nil {
				t.Errorf("命题1 守恒性违反（%s）: %v", name, err)
			}
		})
	}
}

// TestCorrectness_Legality 命题2：每个值 >= MinPerPart，无一例外。
func TestCorrectness_Legality(t *testing.T) {
	for name, fn := range allStrategies() {
		t.Run(name, func(t *testing.T) {
			result, err := Simulate(propState, fn, propRounds, propSeed)
			if err != nil {
				t.Fatalf("Simulate() 失败: %v", err)
			}
			if err := result.CheckLegality(); err != nil {
				t.Errorf("命题2 合法性违反（%s）: %v", name, err)
			}
		})
	}
}

// TestCorrectness_Completeness 命题3：每轮恰好产出 RemainCount 个值，不多不少。
func TestCorrectness_Completeness(t *testing.T) {
	for name, fn := range allStrategies() {
		t.Run(name, func(t *testing.T) {
			result, err := Simulate(propState, fn, propRounds, propSeed)
			if err != nil {
				t.Fatalf("Simulate() 失败: %v", err)
			}
			if len(result.Positions) != int(propState.RemainCount) {
				t.Errorf("命题3 完整性违反（%s）: len(Positions)=%d，期望 %d",
					name, len(result.Positions), propState.RemainCount)
			}
			for _, p := range result.Positions {
				if p.count != int64(propRounds) {
					t.Errorf("命题3 完整性违反（%s）：位置 %d count=%d，期望 %d",
						name, p.Position, p.count, propRounds)
				}
			}
		})
	}
}

// TestCorrectness_Boundedness 命题4：MeanBounded(m) 每次采样 x 满足
// x <= floor(m * currentAvg)，currentAvg = currentRemain/currentCount（整除）。
// 最后一份由 Next() 直接返回剩余全部，不受此约束。
// 此命题直接在单轮分配中验证，以便观察当前状态。
func TestCorrectness_Boundedness(t *testing.T) {
	const multiplier = 2.0
	fn := DoubleMean()
	state := propState

	for round := 0; round < 10000; round++ {
		rng := rand.New(rand.NewPCG(uint64(round), 0))
		d, _ := New(state, fn, rng)
		cur := d.Remaining()
		for !d.Done() {
			if cur.RemainCount == 1 {
				d.Next() //nolint：最后一份不受 SampleFunc 约束
				break
			}
			avg := cur.RemainAmount / cur.RemainCount
			safeUpper := cur.RemainAmount - (cur.RemainCount-1)*cur.MinPerPart
			upper := min(safeUpper, int64(math.Floor(multiplier*float64(avg))))

			v, err := d.Next()
			if err != nil {
				t.Fatalf("round=%d Next() 错误: %v", round, err)
			}
			if v > upper {
				t.Fatalf("命题4 有界性违反 round=%d: v=%d > upper=%d（avg=%d safeUpper=%d）",
					round, v, upper, avg, safeUpper)
			}
			if v < cur.MinPerPart {
				t.Fatalf("命题4 合法性违反 round=%d: v=%d < min=%d", round, v, cur.MinPerPart)
			}
			cur = d.Remaining()
		}
	}
}

// TestCorrectness_FixedBaseline 命题5：Fixed 策略每个位置均值偏差 < 1.0。
func TestCorrectness_FixedBaseline(t *testing.T) {
	result, err := Simulate(propState, Fixed(), propRounds, propSeed)
	if err != nil {
		t.Fatalf("Simulate() 失败: %v", err)
	}
	if err := result.CheckMeans(1.0); err != nil {
		t.Errorf("命题5 基准性违反: %v", err)
	}
}
