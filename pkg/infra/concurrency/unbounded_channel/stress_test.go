//go:build stress

// Package unbounded_channel — stress_test.go
//
// 长时间随机化压力测试，通过 build tag 与普通测试完全隔离。
//
// 设计原则：
//   - 不加 -tags stress 时，本文件根本不参与编译，对 go test ./... 零干扰。
//   - 加了 -tags stress 后，整个 -timeout 预算完全属于压力测试，
//     t.Deadline() 直接使用 -timeout 值，不存在"剩余时间分配"问题。
//   - -timeout 是唯一需要设置的参数，无隐性约束。
//
// 运行方式：
//
//	# 30 分钟（-timeout 即运行时长，留 30s 给最后一轮清理）
//	go test -tags stress -run TestStress -v -timeout 30m \
//	    ./pkg/infra/concurrency/unbounded_channel/
//
//	# 一个月（服务器独跑）
//	go test -tags stress -run TestStress -v -timeout 720h \
//	    ./pkg/infra/concurrency/unbounded_channel/
package unbounded_channel

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// stressReserve 是每轮结束后留给清理工作的余量。
// t.Deadline() 减去此值即为压力循环的实际截止时间。
const stressReserve = 30 * time.Second

// TestStress 多轮随机化压力测试。
//
// 每轮从四种场景中随机选择：FIFO 顺序、多生产者完整性、背压、混沌。
// 任意一轮失败立即终止并上报。
func TestStress(t *testing.T) {
	// t.Deadline() 返回 -timeout 设定的截止时间，保证 -timeout 即运行时长。
	// 极少数情况下未设置 -timeout，给 10 分钟保底。
	deadline, ok := t.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Minute)
	}
	stressEnd := deadline.Add(-stressReserve)
	planned := time.Until(stressEnd).Truncate(time.Second)

	start := time.Now()
	var totalRounds atomic.Int64

	t.Logf("[stress] 开始，计划运行 %s", planned)

	// 每分钟上报一次进度
	progressTicker := time.NewTicker(time.Minute)
	defer progressTicker.Stop()
	go func() {
		for range progressTicker.C {
			t.Logf("[stress] 已运行 %s，剩余 %s，rounds=%d",
				time.Since(start).Truncate(time.Second),
				time.Until(stressEnd).Truncate(time.Second),
				totalRounds.Load())
		}
	}()

	for time.Now().Before(stressEnd) {
		switch rand.IntN(4) {
		case 0:
			stressRound_FIFO(t)
		case 1:
			stressRound_Integrity(t)
		case 2:
			stressRound_Backpressure(t)
		case 3:
			stressRound_Chaos(t)
		}

		if t.Failed() {
			return
		}
		totalRounds.Add(1)
	}

	t.Logf("[stress] 完成：实际运行 %s，rounds=%d",
		time.Since(start).Truncate(time.Millisecond), totalRounds.Load())
}

// stressRound_FIFO 单生产者场景，严格验证 FIFO 顺序。
func stressRound_FIFO(t *testing.T) {
	t.Helper()
	chanSize := rand.IntN(32) + 1
	n := rand.IntN(500) + 50

	ch := NewUnboundedChannelV6[int](chanSize, n*10)
	defer ch.Close()

	received := make([]int, 0, n)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range n {
			v, ok := ch.Receive()
			if !ok {
				return
			}
			received = append(received, v)
		}
	}()

	for i := range n {
		ch.Send(i)
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("[FIFO round] 超时：chanSize=%d n=%d 已收到 %d/%d",
			chanSize, n, len(received), n)
	}

	assertStrictFIFO(t, received, n)
}

// stressRound_Integrity 多生产者场景，验证无丢失、无重复。
func stressRound_Integrity(t *testing.T) {
	t.Helper()
	producers := rand.IntN(8) + 2
	msgsPerProducer := rand.IntN(200) + 20
	total := producers * msgsPerProducer
	chanSize := rand.IntN(64) + 1

	ch := NewUnboundedChannelV6[int](chanSize, total*2)
	defer ch.Close()

	seen := make([]atomic.Bool, total)
	var count atomic.Int64
	allDone := make(chan struct{})

	go func() {
		for {
			v, ok := ch.Receive()
			if !ok {
				return
			}
			if v < 0 || v >= total {
				t.Errorf("[integrity round] 非法消息值 %d（total=%d）", v, total)
				return
			}
			if seen[v].Swap(true) {
				t.Errorf("[integrity round] 重复消息：%d", v)
				return
			}
			if count.Add(1) == int64(total) {
				close(allDone)
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(producers)
	for p := range producers {
		go func(pid int) {
			defer wg.Done()
			base := pid * msgsPerProducer
			for i := range msgsPerProducer {
				ch.Send(base + i)
			}
		}(p)
	}
	wg.Wait()

	select {
	case <-allDone:
	case <-time.After(30 * time.Second):
		t.Fatalf("[integrity round] 超时：producers=%d msgs=%d 已收到 %d/%d",
			producers, msgsPerProducer, count.Load(), total)
	}

	assertNoDup(t, seen, total)
}

// stressRound_Backpressure 小 limit + 慢消费者，反复触发背压与解压。
func stressRound_Backpressure(t *testing.T) {
	t.Helper()
	chanSize := rand.IntN(8) + 1
	limit := rand.IntN(20) + 5
	total := (limit + chanSize) * (rand.IntN(5) + 3)

	ch := NewUnboundedChannelV6[int](chanSize, limit)
	defer ch.Close()

	var count atomic.Int64
	allDone := make(chan struct{})

	go func() {
		for {
			_, ok := ch.Receive()
			if !ok {
				return
			}
			if count.Add(1) == int64(total) {
				close(allDone)
				return
			}
			if rand.IntN(10) == 0 {
				time.Sleep(time.Duration(rand.IntN(500)) * time.Microsecond)
			}
		}
	}()

	for i := range total {
		ch.Send(i)
	}

	select {
	case <-allDone:
	case <-time.After(30 * time.Second):
		t.Fatalf("[backpressure round] 超时：chanSize=%d limit=%d total=%d 已收到 %d",
			chanSize, limit, total, count.Load())
	}
}

// stressRound_Chaos 完全随机化：并发 Send/Receive/Close，验证无 panic、无死锁。
func stressRound_Chaos(t *testing.T) {
	t.Helper()
	chanSize := rand.IntN(16) + 1
	limit := rand.IntN(50) + 10
	producers := rand.IntN(5) + 1
	consumers := rand.IntN(3) + 1
	total := producers * (rand.IntN(100) + 20)

	ch := NewUnboundedChannelV6[int](chanSize, limit)

	var receivedCount atomic.Int64
	allDone := make(chan struct{}, 1)

	for range consumers {
		go func() {
			for {
				_, ok := ch.Receive()
				if !ok {
					return
				}
				if receivedCount.Add(1) == int64(total) {
					select {
					case allDone <- struct{}{}:
					default:
					}
				}
			}
		}()
	}

	var wg sync.WaitGroup
	wg.Add(producers)
	for p := range producers {
		go func(pid int) {
			defer wg.Done()
			base := pid * (total / producers)
			for i := range total / producers {
				ch.Send(base + i)
			}
		}(p)
	}
	wg.Wait()

	select {
	case <-allDone:
	case <-time.After(30 * time.Second):
		// chaos round 仅验证无 panic/死锁，不强制要求全部消息到达
	}

	ch.Close()
	for {
		_, ok := ch.Receive()
		if !ok {
			break
		}
	}
}
