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
	
	weight_cycle "github.com/motocat46/yytools/pkg/mechanics/distribution/progressive_weight_cycle"
)

type Config struct {
	ConfigBase
	ConfigStandard
	ConfigSpecial
}

type ConfigBase struct {
	// R 随机源，仅用于每轮周期的特殊分布计划生成；标准分布使用全局随机源
	R        *rand.Rand
	CycleLen int32 // 一个循环大周期分布数量
}

type ConfigStandard struct {
	Weight Weight // 普通分布的权重结构
}

type ConfigSpecial struct {
	MinInterval int32               // 两个特殊位置之间的最小间隔（0 表示不限）
	Items       []weight_cycle.Item // 各特殊结果配置，长度即特殊分布数量
}

// State 每个玩家/对象的进度，持有各层的运行时状态
type State struct {
	posInCycle int32              // 当前分布位置，由 Engine 推进（周期内位置）
	standard   StandardLayerState // 普通层状态
	special    SpecialLayerState  // 特殊层状态
}

// NewState 创建一个新的空 State，调用 Engine.Init 后方可使用
func NewState() *State {
	return &State{}
}

// PosInCycle 返回当前周期内位置（只读）
func (s *State) PosInCycle() int32 {
	return s.posInCycle
}

// Plan 返回本周期特殊分布计划（只读切片）
func (s *State) Plan() []int32 {
	return s.special.plan
}

type DistributionType int

const (
	Invalid  DistributionType = -1
	Standard DistributionType = 0
	Special  DistributionType = 1
)

// Engine 持有不可变规则，纯调度；非线程安全（因 rand），多 goroutine 使用需加锁
//	v1（当前）：固定两层 + 固定 state 形状，追求好用、稳定、易测
//	v2（未来）：当出现需要不同 state 的 layer 时，引入可插拔 state 的 EngineV2（可以同仓库并存）
type Engine struct {
	cycleLen int32
	r        *rand.Rand
	standard StandardLayer
	special  SpecialLayer
}

// New 根据配置创建 Engine。Config 固定两层结构，覆盖多数使用场景。
// 如未来需要可插拔 Layer，参见 tiered_cycle_design.md §六"未来演进策略"。
func New(cfg Config) (*Engine, error) {
	if cfg.CycleLen <= 0 {
		return nil, fmt.Errorf("CycleLen must be > 0, got %d", cfg.CycleLen)
	}
	if weight_cycle.TotalQuota(cfg.Items) > 0 && cfg.R == nil {
		return nil, fmt.Errorf("R must not be nil when Items is non-empty")
	}
	n := weight_cycle.TotalQuota(cfg.Items)
	if n > cfg.CycleLen {
		return nil, fmt.Errorf("sum(Quota)=%d > CycleLen=%d", n, cfg.CycleLen)
	}
	if n > 0 && cfg.CycleLen-cfg.MinInterval*(n-1) < n {
		return nil, fmt.Errorf("infeasible: CycleLen=%d cannot accommodate %d special occurrences with MinInterval=%d",
			cfg.CycleLen, n, cfg.MinInterval)
	}
	return &Engine{
		cycleLen: cfg.CycleLen,
		r:        cfg.R,
		standard: newStandardLayer(cfg.Weight),
		special:  newSpecialLayer(cfg.Items, cfg.MinInterval),
	}, nil
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
	res := Result{Index: -1, Type: Invalid}
	var err error
	
	occIdx := e.special.GetOccIdx(&state.special, state.posInCycle)
	if occIdx == -1 {
		res.Index = e.standard.Generate(&state.standard)
		res.Type = Standard
	} else {
		var index int
		index, err = e.special.Generate(&state.special, int32(occIdx))
		if err == nil {
			res.Index = index
			res.Type = Special
		}
	}
	
	// 状态推进：无论是否出错位置都前进，防止调用方陷入同一特殊位置的无限重试
	state.posInCycle++
	if state.posInCycle >= e.cycleLen {
		res.CycleEnd = true
	}
	return res, err
}

func (e *Engine) ResetCycle(state *State) {
	state.posInCycle = 0
	e.standard.Reset(&state.standard)
	e.special.Reset(&state.special, e.r, e.cycleLen)
}