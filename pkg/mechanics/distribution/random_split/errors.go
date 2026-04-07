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

import "errors"

var (
	// ErrDone 在 Done()==true 后调用 Next() 时返回。
	ErrDone = errors.New("random_split: allocation already done")

	// ErrInvalidState 由 New() 返回，当 State.Validate() 失败时。
	// 始终以 fmt.Errorf("...: %w", ErrInvalidState) 包装，调用方可用 errors.Is 检查。
	ErrInvalidState = errors.New("random_split: invalid state")

	// ErrInvalidParam 参数非法，以下情形触发：
	//   - MeanBounded(multiplier<1.0)
	//   - New(fn=nil)
	//   - Simulate(rounds<=0)
	// 同样以 %w 包装。
	ErrInvalidParam = errors.New("random_split: invalid parameter")

	// ErrAlreadyStarted 由 Allocate() 在 Distributor 非全新状态时返回，包括：
	//   - 已调用过 Next()（started==true）
	//   - 已成功完成过一次 Allocate()（内部调用过 Next，started==true）
	ErrAlreadyStarted = errors.New("random_split: allocate requires a fresh distributor")
)
