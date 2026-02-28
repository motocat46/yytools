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
)

// SpecialItem 每个特殊结果的配置，下标即 Result.Index
type SpecialItem struct {
	Quota  int32 // 此特殊结果在一个周期内可出现的最大次数
	JoinAt int32 // 第几次特殊抽（0-based 特殊出现序号）时该奖励才开始进入候选池
}

type Config struct {
	ConfigBase
	ConfigSpecial
}

type ConfigBase struct {
	// R 随机源，仅用于每轮周期的特殊分布计划生成；标准分布使用全局随机源
	R        *rand.Rand
	CycleLen int32  // 一个循环大周期分布数量
	Weight   Weight // 普通分布的权重结构
}

type ConfigSpecial struct {
	MinInterval int32         // 两个特殊位置之间的最小间隔（0 表示不限）
	Items       []SpecialItem // 各特殊结果配置，长度即特殊分布数量
}

type State struct {
	posInCycle  int32           // 当前分布位置，由 Engine 推进（周期内位置）
	specialPlan []int32         // 特殊分布计划（有序）
	specialUsed map[int32]int32 // 特殊分布已出现次数
}

func NewState() *State {
	return &State{
		posInCycle:  0,
		specialPlan: make([]int32, 0),
		specialUsed: make(map[int32]int32),
	}
}

// PosInCycle 返回当前周期内位置（只读）
func (s *State) PosInCycle() int32 {
	return s.posInCycle
}

// Plan 返回本周期特殊分布计划（只读切片）
func (s *State) Plan() []int32 {
	return s.specialPlan
}

type DistributionType int

const (
	Invalid  DistributionType = -1
	Standard DistributionType = 0
	Special  DistributionType = 1
)

// 非线程安全(主要是因为rand)，多个goroutine使用要加锁
type Engine struct{ cfg Config }

func New(cfg Config) (*Engine, error) {
	if cfg.CycleLen <= 0 {
		return nil, fmt.Errorf("CycleLen must be > 0, got %d", cfg.CycleLen)
	}
	if totalQuota(cfg.Items) > 0 && cfg.R == nil {
		return nil, fmt.Errorf("R must not be nil when Items is non-empty")
	}
	n := totalQuota(cfg.Items)
	if n > cfg.CycleLen {
		return nil, fmt.Errorf("sum(Quota)=%d > CycleLen=%d", n, cfg.CycleLen)
	}
	if n > 0 && cfg.CycleLen-cfg.MinInterval*(n-1) < n {
		return nil, fmt.Errorf("infeasible: CycleLen=%d cannot accommodate %d special occurrences with MinInterval=%d",
			cfg.CycleLen, n, cfg.MinInterval)
	}
	return &Engine{cfg: cfg}, nil
}

func (e *Engine) Init(state *State) {
	e.ResetCycle(state)
}

type Result struct {
	Index    int
	Type     DistributionType
	CycleEnd bool
}

func (e *Engine) NextAutoReset(state *State) (Result, error) {
	res, err := e.Next(state)
	if res.CycleEnd {
		e.ResetCycle(state)
	}
	return res, err
}

func (e *Engine) Next(state *State) (Result, error) {
	// 1.调用者先构造好基础数据，再使用Next
	// 2.Next()执行；
	// 3.调用者更新相关数据，方便下次再调用Next
	res := Result{Index: -1, Type: Invalid}
	var err error
	
	specialOccIdx := getSpecialCycleIndex(state.specialPlan, state.posInCycle)
	if specialOccIdx == -1 {
		res.Index = standardCycleCore(e.cfg.Weight)
		res.Type = Standard
	} else {
		var index int
		index, err = specialCycleCore(state.specialUsed, int32(specialOccIdx), e.cfg.Items)
		if err == nil {
			res.Index = index
			res.Type = Special
			state.specialUsed[int32(index)]++
		}
	}
	
	// 状态推进：无论是否出错位置都前进，防止调用方陷入同一特殊位置的无限重试
	state.posInCycle++
	if state.posInCycle >= e.cfg.CycleLen {
		// 不直接重置周期，由上层决定是否重置
		res.CycleEnd = true
	}
	return res, err
}

func (e *Engine) ResetCycle(state *State) {
	// 重置周期
	state.posInCycle = 0
	// 重置特殊分布计划数据
	n := totalQuota(e.cfg.Items)
	if n == 0 {
		if state.specialPlan == nil {
			state.specialPlan = make([]int32, 0)
		} else {
			state.specialPlan = state.specialPlan[:0]
		}
	} else {
		state.specialPlan = buildSpecialPlan(e.cfg.R, e.cfg.CycleLen, e.cfg.MinInterval, n)
	}
	// 清理不同特殊分布命中数据
	if state.specialUsed == nil {
		state.specialUsed = make(map[int32]int32)
	} else {
		clear(state.specialUsed)
	}
}