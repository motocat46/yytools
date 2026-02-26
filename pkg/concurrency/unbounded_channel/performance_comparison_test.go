package unbounded_channel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 性能对比测试：固定间隔 vs 自适应间隔

// TestAdaptiveIntervalBehavior 测试自适应间隔的行为
func TestAdaptiveIntervalBehavior(t *testing.T) {
	testCases := []struct {
		name     string
		msgCount int
		delay    time.Duration
	}{
		{"低负载", 100, 0},
		{"中等负载", 500, 0},
		{"高负载", 2000, 0},
		// 突发负载：原参数 1000条*10ms=10s，缩减为可接受的测试时长
		{"突发负载", 20, 5 * time.Millisecond},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// 每个子测试独立创建 uc，避免多个消费者 goroutine 跨用例竞争同一 Channel
			uc := NewUnboundedChannel()
			defer uc.Close()

			var mu sync.Mutex
			var received []int64
			allReceived := make(chan struct{})

			go func() {
				for msg := range uc.Channel {
					mu.Lock()
					received = append(received, msg.ID)
					count := len(received)
					mu.Unlock()

					// 收齐即通知，不等延迟处理完再通知
					if count == tc.msgCount {
						close(allReceived)
					}

					if tc.delay > 0 {
						time.Sleep(tc.delay)
					}
				}
			}()

			startTime := time.Now()
			for i := int64(0); i < int64(tc.msgCount); i++ {
				msg := &Msg{
					ID:   i,
					Data: []byte(fmt.Sprintf("test_%s_%d", tc.name, i)),
				}
				if err := uc.SendMsg(msg); err != nil {
					t.Fatalf("SendMsg failed: %v", err)
				}
			}

			// 等待所有消息收齐，超时时间按负载动态计算
			timeout := time.Duration(tc.msgCount)*tc.delay + 2*time.Second
			select {
			case <-allReceived:
			case <-time.After(timeout):
				mu.Lock()
				got := len(received)
				mu.Unlock()
				t.Fatalf("超时: 期望 %d 条消息，实际收到 %d 条", tc.msgCount, got)
			}

			duration := time.Since(startTime)
			throughput := float64(tc.msgCount) / duration.Seconds()

			mu.Lock()
			receivedCopy := make([]int64, len(received))
			copy(receivedCopy, received)
			mu.Unlock()

			if len(receivedCopy) != tc.msgCount {
				t.Fatalf("期望 %d 条消息，实际收到 %d 条", tc.msgCount, len(receivedCopy))
			}

			// 验证FIFO顺序
			for i := int64(0); i < int64(tc.msgCount); i++ {
				if receivedCopy[i] != i {
					t.Fatalf("FIFO顺序错误: 索引 %d 期望 %d，实际 %d", i, i, receivedCopy[i])
				}
			}

			stats := uc.GetStats()
			t.Logf("[%s] 处理完成: %d 消息, 耗时: %v, 吞吐量: %.2f msg/sec",
				tc.name, tc.msgCount, duration, throughput)
			t.Logf("[%s] 统计信息: transfer_count=%v, avg_interval=%vμs",
				tc.name, stats["transfer_count"], stats["avg_transfer_interval_us"])
		})
	}
}

// BenchmarkAdaptiveInterval 自适应间隔版本的基准测试
func BenchmarkAdaptiveInterval(b *testing.B) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	go func() {
		for range uc.Channel {
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			msg := &Msg{
				ID:   i,
				Data: []byte(fmt.Sprintf("benchmark_adaptive_%d", i)),
			}
			if err := uc.SendMsg(msg); err != nil {
				b.Fatalf("SendMsg failed: %v", err)
			}
			i++
		}
	})
}

// TestAdaptiveIntervalEfficiency 测试自适应间隔的效率
func TestAdaptiveIntervalEfficiency(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	// 模拟空闲状态
	time.Sleep(100 * time.Millisecond)
	stats1 := uc.GetStats()
	t.Logf("空闲状态统计: transfer_count=%v", stats1["transfer_count"])

	// 启动慢速消费者：消费速度远低于发送速度，确保 Channel 填满并触发慢路径
	go func() {
		for range uc.Channel {
			time.Sleep(200 * time.Microsecond)
		}
	}()

	// 发送超过 Channel 容量(1000)的消息，迫使多余消息进入慢路径，触发 transfer
	for i := int64(0); i < 3000; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("efficiency_test_%d", i)),
		}
		uc.SendMsg(msg)
	}

	time.Sleep(300 * time.Millisecond)
	stats2 := uc.GetStats()
	t.Logf("活跃状态统计: transfer_count=%v, avg_interval=%vμs",
		stats2["transfer_count"], stats2["avg_transfer_interval_us"])

	// 验证 transfer 次数增加
	if stats2["transfer_count"].(int64) <= stats1["transfer_count"].(int64) {
		t.Fatal("活跃状态下 transfer_count 应该增加")
	}
}
