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
	"math/rand/v2"
)

// State 描述当前剩余分配状态，值语义，不可变传递。
type State struct {
	RemainAmount int64 // 剩余总量
	RemainCount  int64 // 剩余份数
	MinPerPart   int64 // 每份最小值（>= 1）
}

// Validate 验证 State 是否合法，失败时返回包装了 ErrInvalidState 的错误。
//
// 三条规则：
//   - RemainCount > 0
//   - MinPerPart >= 1
//   - RemainAmount >= RemainCount * MinPerPart
func (s State) Validate() error {
	if s.RemainCount <= 0 {
		return fmt.Errorf("RemainCount=%d must be > 0: %w", s.RemainCount, ErrInvalidState)
	}
	if s.MinPerPart < 1 {
		return fmt.Errorf("MinPerPart=%d must be >= 1: %w", s.MinPerPart, ErrInvalidState)
	}
	if s.RemainAmount < s.RemainCount*s.MinPerPart {
		return fmt.Errorf("RemainAmount=%d < RemainCount*MinPerPart=%d: %w",
			s.RemainAmount, s.RemainCount*s.MinPerPart, ErrInvalidState)
	}
	return nil
}

// SampleFunc 是策略函数类型。
// 给定当前状态（副本）和随机源，返回本次分配量。
//
// 实现约定：
//   - 返回值必须在 [state.MinPerPart, state.RemainAmount-(state.RemainCount-1)*state.MinPerPart] 内
//   - 调用方（Distributor.Next）保证 state 已通过 Validate，且 RemainCount >= 2
//   - 内置策略在合法 state 下永远返回 nil error
//   - 自定义策略仅在无法生成合法值时返回 error（如依赖外部资源失败）
//     Next() 遇到 SampleFunc 返回 error 时立即透传，不更新内部状态
type SampleFunc func(state State, rng *rand.Rand) (int64, error)

// Distributor 按序管理分配过程，非线程安全。
// 并发场景：调用 Allocate() 预生成完整 slice，slice 只读可安全并发读取。
// Allocate() 要求 Distributor 处于全新状态；Next() 和 Allocate() 不混用。
type Distributor struct {
	state   State
	sample  SampleFunc
	rng     *rand.Rand
	started bool // 是否已调用过 Next()，用于 Allocate() 的前置检查
}

// New 创建 Distributor，验证 state 合法性。
// rng 为 nil 时使用随机种子（生产场景）；传固定种子的 rng 用于测试复现。
// fn 为 nil 时返回 (nil, ErrInvalidParam)。
func New(s State, fn SampleFunc, rng *rand.Rand) (*Distributor, error) {
	if fn == nil {
		return nil, fmt.Errorf("SampleFunc must not be nil: %w", ErrInvalidParam)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if rng == nil {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	return &Distributor{state: s, sample: fn, rng: rng}, nil
}

// Done 报告分配是否已完成（RemainCount == 0）。
func (d *Distributor) Done() bool {
	return d.state.RemainCount == 0
}

// Remaining 返回当前剩余状态的副本（快照，不持有引用）。
func (d *Distributor) Remaining() State {
	return d.state
}

// Next 生成下一份分配量，更新内部状态。
// Done() 为 true 时调用返回 (0, ErrDone)，不 panic。
// SampleFunc 返回 error 时透传，不更新内部状态（可重试或放弃）。
func (d *Distributor) Next() (int64, error) {
	if d.Done() {
		return 0, ErrDone
	}
	d.started = true

	// 最后一份：直接返回全部剩余，不调用 SampleFunc
	if d.state.RemainCount == 1 {
		amount := d.state.RemainAmount
		d.state.RemainAmount = 0
		d.state.RemainCount = 0
		return amount, nil
	}

	// 调用策略生成本次份额
	amount, err := d.sample(d.state, d.rng)
	if err != nil {
		return 0, err // 状态不变，可重试
	}

	d.state.RemainAmount -= amount
	d.state.RemainCount--
	return amount, nil
}

// Allocate 一次性生成所有份额，返回完整 slice（长度 == 初始 RemainCount）。
// 要求 Distributor 处于全新状态（未调用过 Next() 或 Allocate()），否则返回 (nil, ErrAlreadyStarted)。
// 返回的 slice 只读，可安全并发读取。
// 推荐并发场景：预生成后按原子 index 分发。
func (d *Distributor) Allocate() ([]int64, error) {
	if d.started {
		return nil, ErrAlreadyStarted
	}
	n := int(d.state.RemainCount)
	result := make([]int64, 0, n)
	for !d.Done() {
		v, err := d.Next()
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
