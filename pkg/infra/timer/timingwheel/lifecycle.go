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
	"github.com/motocat46/yytools/pkg/infra/safeexec"
)

// Start 启动 reaper goroutine（驱动时钟推进）和 taskExecutor goroutine（执行回调）。
// 必须在 AfterFunc/EveryFunc 前调用。
func (tw *TimingWheel) Start() {
	tw.wg.Add(2)
	go tw.reaper()
	go tw.taskExecutor()
}

// Stop 停止时间轮，等待所有已投递回调执行完毕后返回。
// 调用 Stop 后不得再调用 AfterFunc/EveryFunc。
func (tw *TimingWheel) Stop() {
	tw.cancel()  // 通知 reaper 退出
	tw.wg.Wait() // 等待 reaper 和 taskExecutor 均退出
}

// reaper 是时钟推进 goroutine：从 DelayQueue 取到期 bucket，持 writeLock 推进时钟并 Flush。
func (tw *TimingWheel) reaper() {
	defer func() {
		tw.taskQueue.Close() // reaper 退出后关闭 taskQueue，通知 taskExecutor 排水后退出
		tw.wg.Done()
	}()

	for {
		b, ok := tw.delayQueue.Poll(tw.ctx)
		if !ok {
			return // ctx 被 Stop() 取消
		}

		tw.mu.Lock()
		for b != nil {
			tw.advanceClock(b.ExpireAt())
			b.Flush(tw.addOrRun)
			b, _ = tw.delayQueue.TryPoll()
		}
		tw.mu.Unlock()
	}
}

// taskExecutor 是回调执行 goroutine：串行执行到期回调，处理 repeating 重注册。
// 单 goroutine 保证回调不并发（业务层无需额外同步）。
func (tw *TimingWheel) taskExecutor() {
	defer tw.wg.Done()

	for timer := range tw.taskQueue.Out() {
		// safeexec 隔离每个回调的 panic：一个回调崩溃不影响后续定时器执行
		safeexec.SafeExec("timingwheel", timer.task)

		// repeating：从实际执行时刻重新计时（fixed-delay，防止 GC pause 后连锁补发）
		if timer.interval > 0 && !timer.cancelled.Load() {
			timer.expireAt = tw.nowMs() + timer.interval
			tw.add(timer) // readLock，与 advanceClock（writeLock）互斥
		}
	}
}
