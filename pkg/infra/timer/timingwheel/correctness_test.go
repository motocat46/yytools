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
	var mu sync.Mutex
	rng := rand.New(rand.NewPCG(42, 0))

	var wg sync.WaitGroup
	wg.Add(total)

	for g := 0; g < goroutines; g++ {
		go func() {
			for i := 0; i < perG; i++ {
				mu.Lock()
				d := time.Duration(rng.IntN(50)+1) * time.Millisecond
				mu.Unlock()
				tw.AfterFunc(d, func() {
					count.Add(1)
					wg.Done()
				})
			}
		}()
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
func TestCorrectness_WaitSemantics(t *testing.T) {
	const n = 200
	tw := timingwheel.New()
	tw.Start()

	var count atomic.Int32
	for i := 0; i < n; i++ {
		tw.AfterFunc(1*time.Millisecond, func() {
			time.Sleep(2 * time.Millisecond)
			count.Add(1)
		})
	}
	time.Sleep(50 * time.Millisecond)

	tw.Stop() // 等待语义：返回后所有已投递回调应已完成

	// 50ms 睡眠 >> 1ms timer 间隔，所有 n 个 timer 均应在 Stop 前到期并入队 taskQueue。
	// Stop 的排水语义要求：Stop 返回时，taskQueue 中所有回调均已执行完毕。
	// 因此 count 必须等于 n，而非"至少 1"。
	got := int(count.Load())
	if got != n {
		t.Errorf("等待语义命题失败：Stop 返回后 count=%d，预期 %d（所有已投递回调应已完成）", got, n)
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

// TestCorrectness_ConcurrentSafety 命题：高压并发 Add/Cancel 无数据竞争。
// 依赖 -race 检测。总量 100万，满足压力测试 ≥ 100万要求。
func TestCorrectness_ConcurrentSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发压力测试")
	}
	const (
		goroutines = 50
		perG       = 20000 // 总量 1,000,000
	)
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(seed, 0))
			for i := 0; i < perG; i++ {
				d := time.Duration(rng.IntN(100)+1) * time.Millisecond
				timer, _ := tw.AfterFunc(d, func() {})
				if rng.IntN(2) == 0 {
					timer.Cancel()
				}
			}
		}(uint64(g))
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
}

// TestCorrectness_RepeatingInterval 命题：repeating timer 按 interval 持续触发（fixed-delay）。
func TestCorrectness_RepeatingInterval(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	const interval = 50 * time.Millisecond
	var count atomic.Int32
	timer, _ := tw.EveryFunc(interval, func() { count.Add(1) })

	time.Sleep(6 * interval)
	timer.Cancel()
	time.Sleep(2 * interval)

	got := int(count.Load())
	if got < 4 || got > 8 {
		t.Errorf("repeating 命题失败：%dms 内触发 %d 次，预期 4-8 次（interval=%v）",
			(6 * interval).Milliseconds(), got, interval)
	}
}
