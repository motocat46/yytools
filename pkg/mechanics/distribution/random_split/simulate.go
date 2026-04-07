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

package random_split

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// PositionStat 记录某个位置（0-indexed）在多轮模拟中的累积统计量。
// 使用 Welford 在线算法，O(1) 内存，适合大规模模拟。
type PositionStat struct {
	Position int   // 0-indexed
	count    int64 // 采样次数（== rounds）
	mean     float64
	m2       float64 // Welford 方差累积量
	Min, Max int64
}

// update 用 Welford 在线算法更新统计量（每轮调用一次）。
func (s *PositionStat) update(x int64) {
	s.count++
	delta := float64(x) - s.mean
	s.mean += delta / float64(s.count)
	delta2 := float64(x) - s.mean
	s.m2 += delta * delta2
	if s.count == 1 {
		s.Min, s.Max = x, x
	} else {
		if x < s.Min {
			s.Min = x
		}
		if x > s.Max {
			s.Max = x
		}
	}
}

// Mean 返回所有采样的均值。
func (s *PositionStat) Mean() float64 {
	return s.mean
}

// Variance 返回样本方差（除以 count-1）；count<=1 时返回 0。
func (s *PositionStat) Variance() float64 {
	if s.count <= 1 {
		return 0
	}
	return s.m2 / float64(s.count-1)
}

// StdDev 返回样本标准差。
func (s *PositionStat) StdDev() float64 {
	return math.Sqrt(s.Variance())
}

// SimResult 是一次完整模拟的汇总结果。
type SimResult struct {
	State     State          // 原始参数（初始值）
	Rounds    int            // 模拟轮数
	Positions []PositionStat // length == State.RemainCount，按位置 0-indexed 排列
	roundSums []int64        // 每轮实际总和，用于 CheckConservation（内部字段）
}

// CheckConservation 验证每轮所有份额之和 == 初始 RemainAmount，无一例外。
func (r *SimResult) CheckConservation() error {
	for i, s := range r.roundSums {
		if s != r.State.RemainAmount {
			return fmt.Errorf("第%d轮 sum=%d，期望 %d（守恒性违反）", i+1, s, r.State.RemainAmount)
		}
	}
	return nil
}

// CheckLegality 验证每个份额 >= MinPerPart，无一例外。
// 通过 PositionStat.Min 检查，无需遍历原始值。
func (r *SimResult) CheckLegality() error {
	for _, p := range r.Positions {
		if p.Min < r.State.MinPerPart {
			return fmt.Errorf("位置 %d 最小值 %d < MinPerPart=%d（合法性违反）",
				p.Position, p.Min, r.State.MinPerPart)
		}
	}
	return nil
}

// CheckMeans 验证各位置均值与全局均值（RemainAmount/RemainCount）的偏差均 < tol。
// 主要用于 Fixed 策略验证（期望 tol 极小）。
// 对 DoubleMean/Uniform，由于首位从 [min, 2×(S/N)] 采样，其期望值 > S/N；
// 后续位置剩余均值递减，形成系统性递减偏差，tol 应设宽松（建议 RemainAmount*0.1）。
func (r *SimResult) CheckMeans(tol float64) error {
	globalMean := float64(r.State.RemainAmount) / float64(r.State.RemainCount)
	for _, p := range r.Positions {
		diff := math.Abs(p.Mean() - globalMean)
		if diff >= tol {
			return fmt.Errorf("位置 %d 均值=%.2f，全局均值=%.2f，偏差=%.2f >= tol=%.2f",
				p.Position, p.Mean(), globalMean, diff, tol)
		}
	}
	return nil
}

// Summary 返回可打印的统计摘要（位置/均值/标准差/min/max）。
func (r *SimResult) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "State=%+v  Rounds=%d\n", r.State, r.Rounds)
	fmt.Fprintf(&sb, "%-8s %-10s %-10s %-8s %-8s\n", "pos", "mean", "stddev", "min", "max")
	for _, p := range r.Positions {
		fmt.Fprintf(&sb, "%-8d %-10.2f %-10.2f %-8d %-8d\n",
			p.Position, p.Mean(), p.StdDev(), p.Min, p.Max)
	}
	return sb.String()
}

// Simulate 用指定策略跑 rounds 轮完整分配，收集每个位置的统计数据。
// seed 固定随机源，保证可复现（内部使用 rand.NewPCG(seed, 0)）。
// 单个 rng 在所有轮次间共享串行使用：round N 的随机序列延续 round N-1 结束时的状态，
// 不是每轮重置，这是保证全局确定性的前提。
// rounds <= 0 返回 (nil, ErrInvalidParam)。
func Simulate(s State, fn SampleFunc, rounds int, seed uint64) (*SimResult, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if fn == nil {
		return nil, fmt.Errorf("SampleFunc must not be nil: %w", ErrInvalidParam)
	}
	if rounds <= 0 {
		return nil, fmt.Errorf("rounds=%d must be >= 1: %w", rounds, ErrInvalidParam)
	}

	n := int(s.RemainCount)
	positions := make([]PositionStat, n)
	for i := range positions {
		positions[i].Position = i
	}
	roundSums := make([]int64, rounds)

	rng := rand.New(rand.NewPCG(seed, 0))

	for round := 0; round < rounds; round++ {
		d, err := New(s, fn, rng)
		if err != nil {
			return nil, fmt.Errorf("round %d: New() 失败: %w", round, err)
		}
		alloc, err := d.Allocate()
		if err != nil {
			return nil, fmt.Errorf("round %d: Allocate() 失败: %w", round, err)
		}
		var roundSum int64
		for i, v := range alloc {
			positions[i].update(v)
			roundSum += v
		}
		roundSums[round] = roundSum
	}

	return &SimResult{
		State:     s,
		Rounds:    rounds,
		Positions: positions,
		roundSums: roundSums,
	}, nil
}
