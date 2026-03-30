// Package delayqueue_test.
package delayqueue_test

import (
	"context"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timer/delayqueue"
)

type testItem struct{ exp int64 }

func (t *testItem) ExpireAt() int64 { return t.exp }

// TestOffer_ThenTryPoll 验证 Offer 后 TryPoll 按到期时间顺序取出元素
func TestOffer_ThenTryPoll(t *testing.T) {
	now := int64(0)
	clock := func() int64 { return now }
	q := delayqueue.New[*testItem](clock)

	q.Offer(&testItem{exp: 10})
	q.Offer(&testItem{exp: 5})
	q.Offer(&testItem{exp: 20})

	// 还没到期，TryPoll 返回 false
	item, ok := q.TryPoll()
	if ok {
		t.Fatalf("now=0, 最小 exp=5，预期 TryPoll 返回 false，got item=%v", item)
	}

	// 推进时钟到 5，取出 exp=5
	now = 5
	item, ok = q.TryPoll()
	if !ok || item.exp != 5 {
		t.Fatalf("now=5, 预期取出 exp=5，got ok=%v item=%v", ok, item)
	}

	// 推进到 10，取出 exp=10
	now = 10
	item, ok = q.TryPoll()
	if !ok || item.exp != 10 {
		t.Fatalf("now=10, 预期取出 exp=10，got ok=%v item=%v", ok, item)
	}
}

// TestOffer_HeapOrder 验证 Offer 维持最小堆顺序
func TestOffer_HeapOrder(t *testing.T) {
	now := int64(100)
	q := delayqueue.New[*testItem](func() int64 { return now })

	exps := []int64{50, 30, 80, 10, 60}
	for _, e := range exps {
		q.Offer(&testItem{exp: e})
	}

	// 全部到期后，应按升序弹出
	want := []int64{10, 30, 50, 60, 80}
	for _, w := range want {
		item, ok := q.TryPoll()
		if !ok || item.exp != w {
			t.Fatalf("预期 exp=%d，got ok=%v item=%v", w, ok, item)
		}
	}
}

// TestPoll_BlocksUntilExpiry 验证 Poll 在元素到期前阻塞，到期后立即返回
func TestPoll_BlocksUntilExpiry(t *testing.T) {
	start := time.Now()
	q := delayqueue.New[*testItem](func() int64 {
		return time.Since(start).Milliseconds()
	})
	q.Offer(&testItem{exp: 50})

	ctx := context.Background()
	item, ok := q.Poll(ctx)
	elapsed := time.Since(start)

	if !ok {
		t.Fatal("Poll 返回 false，预期返回元素")
	}
	if item.exp != 50 {
		t.Fatalf("预期 exp=50，got %d", item.exp)
	}
	if elapsed < 45*time.Millisecond {
		t.Fatalf("Poll 过早返回，elapsed=%v，预期 ≥45ms", elapsed)
	}
}

// TestPoll_CtxCancel 验证 ctx 取消后 Poll 立即返回 false
func TestPoll_CtxCancel(t *testing.T) {
	q := delayqueue.New[*testItem](func() int64 { return 0 })
	q.Offer(&testItem{exp: 10000}) // 10秒后到期

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, ok := q.Poll(ctx)
	if ok {
		t.Fatal("ctx 取消后预期 Poll 返回 false")
	}
}

// TestPoll_EarlyOfferWakesBlockedPoll 验证 Offer 更早元素会唤醒阻塞的 Poll
func TestPoll_EarlyOfferWakesBlockedPoll(t *testing.T) {
	start := time.Now()
	q := delayqueue.New[*testItem](func() int64 {
		return time.Since(start).Milliseconds()
	})
	q.Offer(&testItem{exp: 10000}) // 初始元素很晚

	// Poll 阻塞中，20ms 后 Offer 一个更早的元素
	go func() {
		time.Sleep(20 * time.Millisecond)
		q.Offer(&testItem{exp: 30}) // 30ms 后到期
	}()

	item, ok := q.Poll(context.Background())
	elapsed := time.Since(start)

	if !ok || item.exp != 30 {
		t.Fatalf("预期取出 exp=30，got ok=%v", ok)
	}
	if elapsed > 100*time.Millisecond {
		t.Fatalf("Poll 等待过久，elapsed=%v，预期 <100ms", elapsed)
	}
}

// TestDelayQueue_Concurrent_NoLostNoDuplicate 多生产者并发 Offer，消费者 Poll，验证无丢失无重复。
// 总量 100,000，满足集成测试 ≥ 10万要求。
func TestDelayQueue_Concurrent_NoLostNoDuplicate(t *testing.T) {
	const (
		producers = 20
		perProd   = 5000 // 总量 100,000
		total     = producers * perProd
	)
	start := time.Now()
	q := delayqueue.New[*testItem](func() int64 {
		return time.Since(start).Milliseconds()
	})

	rng := rand.New(rand.NewPCG(42, 0)) // rand.New 返回的 rng 非并发安全，需 mu 保护
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perProd; j++ {
				mu.Lock()
				exp := time.Since(start).Milliseconds() + int64(rng.IntN(50)+1)
				mu.Unlock()
				q.Offer(&testItem{exp: exp})
			}
		}()
	}
	wg.Wait()

	// 消费所有元素（最多等待 2s，适配 10 万元素 + 最大 50ms 到期）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got := 0
	for {
		_, ok := q.Poll(ctx)
		if !ok {
			break
		}
		got++
	}
	if got != total {
		t.Errorf("消费 %d 个元素，预期 %d（producers=%d perProd=%d）", got, total, producers, perProd)
	}
}
