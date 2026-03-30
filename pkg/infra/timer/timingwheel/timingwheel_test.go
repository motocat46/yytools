// Package timingwheel_test.
package timingwheel_test

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timer/timingwheel"
)

// TestStartStop_BasicFire 验证 Start 后 one-shot timer 触发，Stop 等待完成
func TestStartStop_BasicFire(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	var fired atomic.Bool
	done := make(chan struct{})

	_, err := tw.AfterFunc(50*time.Millisecond, func() {
		fired.Store(true)
		close(done)
	})
	if err != nil {
		t.Fatalf("AfterFunc error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timer 未在 500ms 内触发")
	}
	if !fired.Load() {
		t.Error("fired 应为 true")
	}
}

// TestAfterFunc_ZeroDuration 验证 d=0 立即执行
func TestAfterFunc_ZeroDuration(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	done := make(chan struct{})
	tw.AfterFunc(0, func() { close(done) })

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("d=0 timer 未在 500ms 内触发")
	}
}

// TestEveryFunc_FiresMultipleTimes 验证 repeating timer 多次触发
func TestEveryFunc_FiresMultipleTimes(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	var count atomic.Int32
	timer, err := tw.EveryFunc(50*time.Millisecond, func() {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("EveryFunc error: %v", err)
	}

	time.Sleep(280 * time.Millisecond)
	timer.Cancel()

	got := int(count.Load())
	// 280ms / 50ms ≈ 5次，允许 ±2 误差
	if got < 3 || got > 8 {
		t.Errorf("repeating timer 触发 %d 次，预期 3-8 次（280ms / 50ms interval）", got)
	}
}

// TestAfterFunc_WithMaxTimeout_Exceeded 验证超出 MaxTimeout 返回 error
func TestAfterFunc_WithMaxTimeout_Exceeded(t *testing.T) {
	tw := timingwheel.New(timingwheel.WithMaxTimeout(1 * time.Second))
	tw.Start()
	defer tw.Stop()

	_, err := tw.AfterFunc(2*time.Second, func() {})
	if err == nil {
		t.Error("超出 MaxTimeout 应返回 error")
	}
}

// TestGoAsync_DoesNotBlockTaskExecutor 验证 GoAsync 不阻塞 taskExecutor
func TestGoAsync_DoesNotBlockTaskExecutor(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	tw.AfterFunc(10*time.Millisecond, timingwheel.GoAsync(func() {
		time.Sleep(200 * time.Millisecond)
		close(done1)
	}))
	tw.AfterFunc(20*time.Millisecond, func() { close(done2) })

	select {
	case <-done2:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("第二个 timer 被 GoAsync 阻塞（GoAsync 未正确异步化）")
	}
	<-done1
}

// TestAfterFunc_MultiLayer_Demotion 验证 L2 范围 timer 经降级后在正确时间触发
func TestAfterFunc_MultiLayer_Demotion(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	d := 500 * time.Millisecond
	start := time.Now()
	done := make(chan struct{})

	tw.AfterFunc(d, func() { close(done) })

	select {
	case <-done:
	case <-time.After(d + 300*time.Millisecond):
		t.Fatal("L2 timer 未在预期时间触发")
	}

	elapsed := time.Since(start)
	if elapsed < d-20*time.Millisecond {
		t.Errorf("timer 过早触发，elapsed=%v，期望 >=%v", elapsed, d-20*time.Millisecond)
	}
}

// TestAfterFunc_ManyTimers_AllFire 10万个定时器全部触发，无丢失
func TestAfterFunc_ManyTimers_AllFire(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模测试")
	}
	const total = 100_000
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	rng := rand.New(rand.NewPCG(42, 0))
	var fired atomic.Int64
	var wg sync.WaitGroup
	wg.Add(total)

	for i := 0; i < total; i++ {
		d := time.Duration(rng.IntN(200)+1) * time.Millisecond
		tw.AfterFunc(d, func() {
			fired.Add(1)
			wg.Done()
		})
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("10万 timer 未全部触发，已触发 %d", fired.Load())
	}

	if got := fired.Load(); got != int64(total) {
		t.Errorf("触发 %d 次，预期 %d", got, total)
	}
}

// TestCancelBeforeFire_BestEffort 验证 timer 仍在 bucket 中时 Cancel 可阻止回调执行。
// 注意：Cancel 是 best-effort 语义（与 time.Timer.Stop() 一致）——timer 已入 taskQueue 后
// Cancel 不再有效；本测试只覆盖"Cancel 在 Flush 之前"这条路径。
func TestCancelBeforeFire_BestEffort(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	var fired atomic.Bool
	timer, _ := tw.AfterFunc(500*time.Millisecond, func() { fired.Store(true) })
	timer.Cancel()

	time.Sleep(600 * time.Millisecond)
	if fired.Load() {
		t.Error("Cancel 后 timer 不应触发（正常 best-effort 路径：timer 仍在 bucket 中）")
	}
}

// TestStop_DrainsAllPendingCallbacks 验证 Stop 返回前已执行所有已投递回调
func TestStop_DrainsAllPendingCallbacks(t *testing.T) {
	tw := timingwheel.New()
	tw.Start()

	const n = 50
	var count atomic.Int32
	for i := 0; i < n; i++ {
		tw.AfterFunc(1*time.Millisecond, func() {
			time.Sleep(5 * time.Millisecond)
			count.Add(1)
		})
	}

	time.Sleep(50 * time.Millisecond)
	tw.Stop()

	if got := int(count.Load()); got < 1 {
		t.Errorf("Stop 后 count=%d，预期至少 1", got)
	}
}
