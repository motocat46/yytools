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

// Config 是 Engine 的完整配置，由基础参数、普通层配置和特殊层配置组合而成。
type Config struct {
	ConfigBase
	ConfigStandard
	ConfigSpecial
}

// ConfigBase 是所有层共用的基础参数。
type ConfigBase struct {
	CycleLen int32 // 一个循环大周期分布数量
}

// ConfigStandard 是普通层配置。
type ConfigStandard struct {
	Weight Weight // 普通分布的权重结构
}

// ConfigSpecial 是特殊层配置。
type ConfigSpecial struct {
	MinInterval int32               // 两个特殊位置之间的最小间隔（0 表示不限）
	Items       []weight_cycle.Item // 各特殊结果配置，长度即特殊分布数量
}

// State 每个玩家/对象的进度，持有各层的运行时状态
type State struct {
	posInCycle int32              // 当前分布位置，由 Engine 推进（周期内位置）
	r          *rand.Rand         // 随机源，仅用于每轮周期的特殊分布计划生成
	standard   StandardLayerState // 普通层状态
	special    SpecialLayerState  // 特殊层状态
}

// PosInCycle 返回当前周期内位置（只读）
func (s *State) PosInCycle() int32 {
	return s.posInCycle
}

// Plan 返回本周期特殊分布计划（只读切片）
func (s *State) Plan() []int32 {
	return s.special.plan
}

// DistributionType 标识每次抽取结果属于哪个层。
type DistributionType int

const (
	Invalid  DistributionType = -1 // 无效（通常伴随 error）
	Standard DistributionType = 0  // 普通层结果
	Special  DistributionType = 1  // 特殊层结果
)

// Engine 持有不可变规则，纯调度，goroutine-safe。
//
//	v1（当前）：固定两层 + 固定 state 形状，追求好用、稳定、易测
//	v2（未来）：当出现需要不同 state 的 layer 时，引入可插拔 state 的 EngineV2（可以同仓库并存）
type Engine struct {
	cycleLen int32
	standard StandardLayer
	special  SpecialLayer
}

// New 根据配置创建 Engine。Config 固定两层结构，覆盖多数使用场景。
// 如未来需要可插拔 Layer，参见 tiered_cycle_design.md §六"未来演进策略"。
func New(cfg Config) (*Engine, error) {
	if cfg.CycleLen <= 0 {
		return nil, fmt.Errorf("CycleLen must be > 0, got %d", cfg.CycleLen)
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
		standard: newStandardLayer(cfg.Weight),
		special:  newSpecialLayer(cfg.Items, cfg.MinInterval),
	}, nil
}

// NewState 创建与本 Engine 关联的 State，调用 Engine.Init 后方可使用。
// Engine 含特殊层（Items 非空）时 r 不可为 nil；无特殊层时 r 传 nil 即可。
func (e *Engine) NewState(r *rand.Rand) *State {
	if e.special.TotalQuota > 0 && r == nil {
		panic("tiered_cycle: r must not be nil when Engine has special items")
	}
	return &State{r: r}
}

// Init 初始化 state 以开始第一个周期，等价于 ResetCycle。
func (e *Engine) Init(state *State) {
	e.ResetCycle(state)
}

// Result 是单次抽取的结果。
// Index 为命中的奖励下标；Type 标识来自哪一层；CycleEnd 为 true 时表示当前周期已结束。
type Result struct {
	Index    int
	Type     DistributionType
	CycleEnd bool
}

// NextAutoReset 执行一次抽取，语义见 Next；周期结束时自动调用 ResetCycle 开始新周期。
func (e *Engine) NextAutoReset(state *State) (Result, error) {
	res, err := e.Next(state)
	if res.CycleEnd {
		e.ResetCycle(state)
	}
	return res, err
}

// Next 执行一次抽取，返回命中结果和 error（特殊层无候选时 error 非 nil）。
// 无论是否出错，位置都会前进，防止调用方陷入同一特殊位置的无限重试。
// 周期结束时 Result.CycleEnd 为 true，调用方可选择手动调用 ResetCycle 或改用 NextAutoReset。
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

// ResetCycle 重置 state 以开始新的一轮周期，重新生成特殊分布计划。
func (e *Engine) ResetCycle(state *State) {
	state.posInCycle = 0
	e.standard.Reset(&state.standard)
	e.special.Reset(&state.special, state.r, e.cycleLen)
}
