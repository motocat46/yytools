// Package timingwheel_test 正确性命题测试：验证 TimingWheel 的并发不变量。
//
// 全部在 -race 下运行：
//
//	go test -race -run TestCorrectness -v ./pkg/infra/timer/timingwheel/
package timingwheel_test

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timer/timingwheel"
)

// TestCorrectness_PreciselyOnce 命题：one-shot timer 恰好触发一次，不多不少。
// 多 goroutine 并发 add，验证无丢失无重复。
func TestCorrectness_PreciselyOnce(t *testing.T) {
	const (
		goroutines = 20
		perG       = 5000 // 总量 100,000，满足集成测试 ≥ 10万要求
		total      = goroutines * perG
	)
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	var count atomic.Int64

	var wg sync.WaitGroup
	wg.Add(total)

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			// 每个 goroutine 独立 rng，消除锁竞争，保持确定性
			rng := rand.New(rand.NewPCG(uint64(g), 0))
			for i := 0; i < perG; i++ {
				d := time.Duration(rng.IntN(50)+1) * time.Millisecond
				tw.AfterFunc(d, func() {
					count.Add(1)
					wg.Done()
				})
			}
		}(g)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("精确一次命题失败：%d/%d 触发（超时）", count.Load(), total)
	}

	if got := count.Load(); got != int64(total) {
		t.Errorf("精确一次命题失败：触发 %d 次，预期 %d（goroutines=%d perG=%d）",
			got, total, goroutines, perG)
	}
}

// TestCorrectness_WaitSemantics 命题：Stop 返回后，所有已投递回调均已完成（等待语义）。
//
// 设计：用 allStarted WaitGroup 精确等待所有回调进入 taskExecutor（而非靠 time.Sleep 猜测），
// 保证在 Stop() 之前所有回调都已被 taskExecutor 认领，此后 Stop 必须排水至完成。
// taskExecutor 是单 goroutine 串行执行，allStarted.Wait() 返回时：
//   - 回调 1..n-1 已全部完成；
//   - 回调 n 恰好已启动（调用了 allStarted.Done()）但尚未完成（仍在 sleep 中）；
//
// Stop() 的排水语义要求：Stop 返回时，回调 n 的 sleep 已结束，count == n。
func TestCorrectness_WaitSemantics(t *testing.T) {
	const n = 200
	tw := timingwheel.New()
	tw.Start()

	var count atomic.Int32
	var allStarted sync.WaitGroup
	allStarted.Add(n)

	for i := 0; i < n; i++ {
		tw.AfterFunc(1*time.Millisecond, func() {
			allStarted.Done()           // 通知：已进入 taskExecutor
			time.Sleep(2 * time.Millisecond)
			count.Add(1)
		})
	}

	// 等待所有回调均已被 taskExecutor 认领（此时 taskQueue 中无积压，排水场景由最后一个回调体现）
	waitDone := make(chan struct{})
	go func() { allStarted.Wait(); close(waitDone) }()
	select {
	case <-waitDone:
	case <-time.After(10 * time.Second):
		t.Fatalf("等待语义命题：%d 个 1ms timer 未在 10s 内全部分发到 taskExecutor", n)
	}

	tw.Stop() // 等待语义：返回后回调 n 的 sleep 已完成，count 必须 == n

	if got := int(count.Load()); got != n {
		t.Errorf("等待语义命题失败：Stop 返回后 count=%d，预期 %d", got, n)
	}
}

// TestCorrectness_PanicIsolation 命题：单个回调 panic 不影响后续定时器（异常隔离）。
func TestCorrectness_PanicIsolation(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	tw.AfterFunc(10*time.Millisecond, func() {
		panic("预期的测试 panic，safeexec 应隔离")
	})

	done := make(chan struct{})
	tw.AfterFunc(60*time.Millisecond, func() { close(done) })

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("异常隔离命题失败：回调 panic 后 taskExecutor 停止响应")
	}
}

// TestCorrectness_ConcurrentSafety 命题：高压并发 Add/Cancel 无数据竞争，且无重复触发。
// 依赖 -race 检测。总量 100万，满足压力测试 ≥ 100万要求。
// 不变量：fired <= total（无 timer 重复触发）。
func TestCorrectness_ConcurrentSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发压力测试")
	}
	const (
		goroutines = 50
		perG       = 20000 // 总量 1,000,000
		total      = goroutines * perG
	)
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	var fired atomic.Int64
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(seed, 0))
			for i := 0; i < perG; i++ {
				d := time.Duration(rng.IntN(100)+1) * time.Millisecond
				timer, _ := tw.AfterFunc(d, func() { fired.Add(1) })
				if rng.IntN(2) == 0 {
					timer.Cancel()
				}
			}
		}(uint64(g))
	}
	wg.Wait()

	// 等待所有未取消 timer 触发（最长 100ms + buffer）
	time.Sleep(500 * time.Millisecond)

	// 不变量：触发次数不超过添加次数（重复触发说明存在 bug）
	if got := fired.Load(); got > total {
		t.Errorf("触发次数 %d 超出添加次数 %d（存在重复触发 bug）", got, total)
	}
	if fired.Load() == 0 {
		t.Error("触发次数为 0，所有 timer 均未触发")
	}
}

// TestCorrectness_RepeatingInterval 命题：repeating timer 按 interval 持续触发，Cancel 后精确停止。
// 分两段验证：
//  1. Cancel 前：6 个间隔内触发次数在合理范围（验证"持续触发"）；
//  2. Cancel 后：count 不再增加，最多允许多 1 次（best-effort：Cancel 时回调可能已在 taskQueue 中）。
func TestCorrectness_RepeatingInterval(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	const interval = 50 * time.Millisecond
	var count atomic.Int32
	timer, _ := tw.EveryFunc(interval, func() { count.Add(1) })

	time.Sleep(6 * interval) // ~300ms
	countBeforeCancel := count.Load()
	timer.Cancel()
	time.Sleep(2 * interval) // 100ms，等待已入 taskQueue 的最后一次回调完成
	countAfterCancel := count.Load()

	// 第一段：6 个间隔，预期 4-8 次（含调度抖动）
	if countBeforeCancel < 4 || countBeforeCancel > 8 {
		t.Errorf("Cancel 前触发 %d 次，预期 4-8 次（interval=%v，等待 %v）",
			countBeforeCancel, interval, 6*interval)
	}
	// 第二段：Cancel 后最多多触发 1 次（best-effort 语义：已入 taskQueue 的回调执行后不再重注册）
	if extra := countAfterCancel - countBeforeCancel; extra > 1 {
		t.Errorf("Cancel 后又触发了 %d 次（Cancel前=%d，Cancel后=%d），repeating 应停止",
			extra, countBeforeCancel, countAfterCancel)
	}
}

// TestCorrectness_CancelBestEffort 命题：Cancel 是 best-effort——
// timer 已入 taskQueue 时，Cancel 无法阻止本次回调执行（与 time.Timer.Stop() 语义一致）。
//
// 验证路径：让 taskExecutor 忙碌 → 目标 timer 到期入队 → 此时 Cancel → 解除阻塞 → 回调仍执行一次。
func TestCorrectness_CancelBestEffort(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	// 先加一个阻塞回调，让 taskExecutor 忙碌
	blocked := make(chan struct{})
	unblock := make(chan struct{})
	tw.AfterFunc(1*time.Millisecond, func() {
		close(blocked)
		<-unblock
	})

	// 等待 taskExecutor 确实阻塞住
	select {
	case <-blocked:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("第一个 timer 未在 500ms 内触发")
	}

	// 添加目标 timer，等其到期并入 taskQueue（taskExecutor 此时忙碌，timer 只能排队）
	var fired atomic.Bool
	target, _ := tw.AfterFunc(1*time.Millisecond, func() { fired.Store(true) })
	time.Sleep(50 * time.Millisecond) // 50ms >> 1ms，目标 timer 必已入 taskQueue

	// Cancel：此时 timer 在 taskQueue 中（bucket==nil），Cancel 只能设 cancelled=true，
	// 但 taskExecutor 在执行回调前不检查 cancelled，回调仍会执行一次
	target.Cancel()

	// 解除阻塞，让 taskExecutor 继续处理 taskQueue
	close(unblock)
	time.Sleep(100 * time.Millisecond) // 等待目标回调执行完毕

	if !fired.Load() {
		t.Error("best-effort 命题失败：timer 已入 taskQueue，回调应执行一次（Cancel 无法撤回）")
	}
}
