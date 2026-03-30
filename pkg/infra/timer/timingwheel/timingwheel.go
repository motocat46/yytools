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
	"context"
	"sync"
	"time"

	uc "github.com/motocat46/yytools/pkg/infra/concurrency/unbounded_channel"
	"github.com/motocat46/yytools/pkg/infra/timer/delayqueue"
)

// TimingWheel 是分层时间轮定时器，O(1) add/cancel，毫秒精度，覆盖约 50 天。
// 使用单调时钟（time.Since(startTime)），避免 NTP 调整导致时间倒退。
// 必须调用 Start() 后才能触发定时器；使用完毕调用 Stop()。
type TimingWheel struct {
	startTime time.Time

	mu          sync.RWMutex
	currentTime int64 // 单调时间（ms），相对于 startTime

	wheels     [5]*wheel
	delayQueue *delayqueue.DelayQueue[*bucket]

	overflow timerHeap // overflow min-heap（超出 ~50天的定时器）
	// overflowMu 串行化 add()（持 mu.RLock）路径中多 goroutine 对 overflow 的并发写入。
	// advanceClock()（持 mu.Lock write）独占时不加 overflowMu——writeLock 已排他。
	overflowMu sync.Mutex

	// taskQueue 携带 *Timer 而非 func()，taskExecutor 据此重新注册 repeating timer
	taskQueue *uc.UnboundedChannel[*Timer]

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	maxTimeout int64 // 0=无上限；>0=最大允许时长（ms）
}

// Option 是 TimingWheel 构造选项。
type Option func(*TimingWheel)

// WithMaxTimeout 设置允许的最大定时时长。超出时 AfterFunc/EveryFunc 返回 error。
// d=0 表示不限制（默认）。
func WithMaxTimeout(d time.Duration) Option {
	return func(tw *TimingWheel) {
		if d > 0 {
			tw.maxTimeout = d.Milliseconds()
		}
	}
}

// New 创建 TimingWheel，tick=1ms，5 层固定结构，覆盖约 50 天。
// 调用 Start() 后开始推进时钟。
func New(opts ...Option) *TimingWheel {
	ctx, cancel := context.WithCancel(context.Background())
	tw := &TimingWheel{
		startTime: time.Now(),
		taskQueue: uc.NewUnboundedChannel[*Timer](1024, 1_000_000),
		ctx:       ctx,
		cancel:    cancel,
	}
	for i := range tw.wheels {
		size := l2Size
		if i == 0 {
			size = l1Size
		}
		tw.wheels[i] = newWheel(size)
	}
	// 注入单调时钟（startTime 在上方已固定）
	tw.delayQueue = delayqueue.New[*bucket](tw.nowMs)
	for _, opt := range opts {
		opt(tw)
	}
	return tw
}

// nowMs 返回相对于 startTime 的单调毫秒偏移。
func (tw *TimingWheel) nowMs() int64 {
	return time.Since(tw.startTime).Milliseconds()
}

// addInternal 将 timer 路由到对应层 bucket 或 overflow heap（已在 readLock 或 writeLock 保护下）。
func (tw *TimingWheel) addInternal(timer *Timer) {
	delta := timer.expireAt - tw.currentTime
	switch {
	case delta <= 0:
		tw.addOrRun(timer)
		return

	case delta < l1Interval:
		slot := int(timer.expireAt & l1Mask)
		bucketExp := timer.expireAt // L1 levelShift=0，无需对齐
		tw.placeInBucket(tw.wheels[0].buckets[slot], timer, bucketExp)

	case delta < l2Interval:
		slot := int((timer.expireAt >> l2Shift) & l2Mask)
		bucketExp := (timer.expireAt >> l2Shift) << l2Shift
		tw.placeInBucket(tw.wheels[1].buckets[slot], timer, bucketExp)

	case delta < l3Interval:
		slot := int((timer.expireAt >> l3Shift) & l2Mask)
		bucketExp := (timer.expireAt >> l3Shift) << l3Shift
		tw.placeInBucket(tw.wheels[2].buckets[slot], timer, bucketExp)

	case delta < l4Interval:
		slot := int((timer.expireAt >> l4Shift) & l2Mask)
		bucketExp := (timer.expireAt >> l4Shift) << l4Shift
		tw.placeInBucket(tw.wheels[3].buckets[slot], timer, bucketExp)

	case delta < l5Interval:
		slot := int((timer.expireAt >> l5Shift) & l2Mask)
		bucketExp := (timer.expireAt >> l5Shift) << l5Shift
		tw.placeInBucket(tw.wheels[4].buckets[slot], timer, bucketExp)

	default:
		// 超出 5 层覆盖范围，入 overflow heap（overflowMu 串行化并发 add）
		tw.overflowMu.Lock()
		heapPush(&tw.overflow, timer)
		tw.overflowMu.Unlock()
	}
}

// placeInBucket 将 timer 加入 bucket，若是首个 timer 则 CAS 设置 expireAt 并入 DelayQueue。
func (tw *TimingWheel) placeInBucket(b *bucket, timer *Timer, bucketExp int64) {
	isFirst := b.Add(timer)
	if isFirst {
		// CAS 确保只有第一个 timer 触发 DelayQueue.Offer（Flush 后 expireAt 重置为 -1）
		if b.expireAt.CompareAndSwap(-1, bucketExp) {
			tw.delayQueue.Offer(b)
		}
	}
}

// addOrRun 决定 timer 立即执行（已到期）或重新插入时间轮（未到期）。
func (tw *TimingWheel) addOrRun(timer *Timer) {
	if timer.cancelled.Load() {
		return // Flush 清除 bucket → addOrRun 窗口期内被 Cancel
	}
	if timer.expireAt <= tw.currentTime {
		tw.taskQueue.Send(timer)
	} else {
		tw.addInternal(timer)
	}
}

// advanceClock 推进时钟至 timeMs，批量提升进入调度窗口的 overflow timer。
// 调用方必须持有 writeLock（由 reaper goroutine 保证）。
func (tw *TimingWheel) advanceClock(timeMs int64) {
	if timeMs < tw.currentTime+tickMs {
		return
	}
	tw.currentTime = timeMs

	// 将进入调度窗口的 overflow timer 批量提升到时间轮。
	// 条件严格 <：若用 <=，恰好 delta==l5Interval 的 timer 提升后路由回 overflow，死循环。
	for heapPeek(&tw.overflow) != nil &&
		heapPeek(&tw.overflow).expireAt < tw.currentTime+l5Interval {
		timer := heapPop(&tw.overflow)
		if timer.cancelled.Load() {
			continue
		}
		tw.addInternal(timer)
	}
}

// add 将 timer 加入时间轮，持 readLock（与 advanceClock 的 writeLock 互斥）。
func (tw *TimingWheel) add(timer *Timer) {
	tw.mu.RLock()
	tw.addInternal(timer)
	tw.mu.RUnlock()
}

// Cancel 取消定时器，O(1)，幂等。
func (t *Timer) Cancel() {
	b := t.bucket.Load()
	if b == nil {
		t.cancelled.Store(true)
		return
	}
	b.mu.Lock()
	// double-check：Flush 可能在我们拿锁前已清零 bucket
	if t.bucket.Load() != b {
		b.mu.Unlock()
		t.cancelled.Store(true)
		return
	}
	t.bucket.Store(nil)
	b.detach(t)
	b.mu.Unlock()
}
