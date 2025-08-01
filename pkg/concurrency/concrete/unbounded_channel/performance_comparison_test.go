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
	uc := NewUnboundedChannel()
	defer uc.Close()

	// 模拟不同的负载场景
	testCases := []struct {
		name     string
		msgCount int
		delay    time.Duration
	}{
		{"低负载", 100, 0},
		{"中等负载", 500, 0},
		{"高负载", 2000, 0},
		{"突发负载", 1000, 10 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var received []int64
			var mu sync.Mutex

			// 启动消费者
			go func() {
				for msg := range uc.Channel {
					mu.Lock()
					received = append(received, msg.ID)
					mu.Unlock()
					// 模拟处理延迟
					if tc.delay > 0 {
						time.Sleep(tc.delay)
					}
				}
			}()

			startTime := time.Now()

			// 发送消息
			for i := int64(0); i < int64(tc.msgCount); i++ {
				msg := &Msg{
					ID:   i,
					Data: []byte(fmt.Sprintf("test_%s_%d", tc.name, i)),
				}
				if err := uc.SendMsg(msg); err != nil {
					t.Fatalf("SendMsg failed: %v", err)
				}
			}

			time.Sleep(100 * time.Millisecond)

			duration := time.Since(startTime)
			throughput := float64(tc.msgCount) / duration.Seconds()

			// 验证消息数量
			if len(received) != tc.msgCount {
				t.Fatalf("Expected %d messages, got %d", tc.msgCount, len(received))
			}

			// 验证FIFO顺序
			for i := int64(0); i < int64(tc.msgCount); i++ {
				if received[i] != i {
					t.Fatalf("FIFO order violated at index %d: expected %d, got %d", i, i, received[i])
				}
			}

			// 获取统计信息
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

	// 启动消费者
	go func() {
		for range uc.Channel {
			// 模拟处理
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

	// 模拟活跃状态
	for i := int64(0); i < 1000; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("efficiency_test_%d", i)),
		}
		uc.SendMsg(msg)
	}

	time.Sleep(100 * time.Millisecond)
	stats2 := uc.GetStats()
	t.Logf("活跃状态统计: transfer_count=%v, avg_interval=%vμs",
		stats2["transfer_count"], stats2["avg_transfer_interval_us"])

	// 验证transfer次数增加
	if stats2["transfer_count"].(int64) <= stats1["transfer_count"].(int64) {
		t.Fatal("Transfer count should increase during active state")
	}
}
