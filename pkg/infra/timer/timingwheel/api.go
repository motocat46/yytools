// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// 作者: yangyuan
// 创建日期: 2026-03-28
package timingwheel

import (
	"fmt"
	"time"
)

// AfterFunc 注册 one-shot 定时器，d 后执行 f。
// 返回 *Timer 可用于取消；若设置了 WithMaxTimeout 且 d 超出上限，返回 (nil, error)。
// d=0 时 expireAt==nowMs()，delta<=0，视为立即执行（与 time.AfterFunc 语义一致）。
func (tw *TimingWheel) AfterFunc(d time.Duration, f func()) (*Timer, error) {
	if err := tw.checkTimeout(d); err != nil {
		return nil, err
	}
	timer := &Timer{
		expireAt: tw.nowMs() + d.Milliseconds(),
		interval: 0,
		task:     f,
	}
	tw.add(timer)
	return timer, nil
}

// EveryFunc 注册 repeating 定时器，每隔 d 执行一次 f（fixed-delay）。
// 若设置了 WithMaxTimeout 且 d 超出上限，返回 (nil, error)。
func (tw *TimingWheel) EveryFunc(d time.Duration, f func()) (*Timer, error) {
	if err := tw.checkTimeout(d); err != nil {
		return nil, err
	}
	timer := &Timer{
		expireAt: tw.nowMs() + d.Milliseconds(),
		interval: d.Milliseconds(),
		task:     f,
	}
	tw.add(timer)
	return timer, nil
}

// GoAsync 将 f 包装为每次调用独立起一个 goroutine 的函数。
// 用于需要在定时器回调中执行阻塞操作的场景，避免阻塞 taskExecutor。
//
// 注意：safeexec 保护的是 taskExecutor 内"启动 goroutine"这一步；
// 新 goroutine 内的 f panic 不在保护范围——可在 f 内自行使用 safeexec.Safe 包裹：
//
//	tw.AfterFunc(d, GoAsync(func() { safeexec.Safe(heavyWork) }))
func GoAsync(f func()) func() {
	return func() { go f() }
}

// checkTimeout 验证 d 是否超出 maxTimeout（0=无上限）。
func (tw *TimingWheel) checkTimeout(d time.Duration) error {
	if tw.maxTimeout > 0 && d.Milliseconds() > tw.maxTimeout {
		return fmt.Errorf("duration %v exceeds max timeout %v",
			d, time.Duration(tw.maxTimeout)*time.Millisecond)
	}
	return nil
}
