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
)

// Fixed 返回固定均值策略，每次返回当前 RemainAmount/RemainCount（整除）。
//
// 注意：返回值基于当前剩余状态计算，而非初始 S/N 的固定值。
//
// 示例（S=10, N=3, min=1）：
//
//	第1次调用：10/3 = 3
//	第2次调用：7/2  = 3
//	最后一份由 Next() 返回剩余 4（不经过 SampleFunc）
//
// variance≈0（确定性执行下），用途：作为统计对比基准，量化其他策略的分布偏差。
func Fixed() SampleFunc {
	return func(state State, _ *rand.Rand) (int64, error) {
		return state.RemainAmount / state.RemainCount, nil
	}
}

// Uniform 返回均匀可行策略，从 [MinPerPart, safeUpper] 均匀采样。
//
// safeUpper = RemainAmount - (RemainCount-1)*MinPerPart（保证剩余人都能拿到 min）
//
// 分布最均匀，但尾部集中效应比 DoubleMean 更强（最后几份方差更小）。
func Uniform() SampleFunc {
	return func(state State, rng *rand.Rand) (int64, error) {
		safeUpper := state.RemainAmount - (state.RemainCount-1)*state.MinPerPart
		return state.MinPerPart + rng.Int64N(safeUpper-state.MinPerPart+1), nil
	}
}

// MeanBounded 返回均值倍数策略，每次从 [min, min(safeUpper, floor(multiplier*avg))] 均匀采样。
//
//	safeUpper = RemainAmount - (RemainCount-1)*MinPerPart（保证剩余人都能拿到 min）
//	avg       = RemainAmount / RemainCount（当前状态，整除）
//
// multiplier 控制方差：越大分布越分散，越小越集中于均值。
// multiplier < 1.0 返回 (nil, ErrInvalidParam)（上界低于均值，分布无意义）。
func MeanBounded(multiplier float64) (SampleFunc, error) {
	if multiplier < 1.0 {
		return nil, fmt.Errorf("multiplier=%.2f must be >= 1.0: %w", multiplier, ErrInvalidParam)
	}
	return func(state State, rng *rand.Rand) (int64, error) {
		avg := state.RemainAmount / state.RemainCount
		safeUpper := state.RemainAmount - (state.RemainCount-1)*state.MinPerPart
		// multiplier 极大时 multiplier*float64(avg) 可能溢出为 +Inf；
		// int64(+Inf) 是未定义行为，显式检测后直接使用 safeUpper。
		floatUpper := multiplier * float64(avg)
		var upper int64
		if math.IsInf(floatUpper, 1) || floatUpper >= float64(math.MaxInt64) {
			upper = safeUpper
		} else {
			upper = min(safeUpper, int64(math.Floor(floatUpper)))
		}
		// 理论上 upper >= MinPerPart（合法 state + multiplier>=1.0 保证），
		// 此处防御极端浮点情形。
		if upper < state.MinPerPart {
			upper = state.MinPerPart
		}
		return state.MinPerPart + rng.Int64N(upper-state.MinPerPart+1), nil
	}, nil
}

// DoubleMean 是 MeanBounded(2.0) 的快捷方式，即传统二倍均值法。
// 不返回 error（2.0 是合法 multiplier）。
func DoubleMean() SampleFunc {
	fn, _ := MeanBounded(2.0) // 2.0 >= 1.0，永远成功
	return fn
}
