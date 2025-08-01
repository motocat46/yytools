package unbounded_channel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestFIFOOrder 测试FIFO顺序
func TestFIFOOrder(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	// 发送1000条消息，超过通道容量
	const msgCount = 1000
	var received []int64
	var mu sync.Mutex

	// 启动消费者
	go func() {
		for msg := range uc.Channel {
			mu.Lock()
			received = append(received, msg.ID)
			mu.Unlock()
		}
	}()

	// 发送消息
	for i := int64(0); i < msgCount; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("data_%d", i)),
		}
		if err := uc.SendMsg(msg); err != nil {
			t.Fatalf("SendMsg failed: %v", err)
		}
	}

	// 等待所有消息被消费
	time.Sleep(100 * time.Millisecond)

	// 验证FIFO顺序
	if len(received) != msgCount {
		t.Fatalf("Expected %d messages, got %d", msgCount, len(received))
	}

	for i := int64(0); i < msgCount; i++ {
		if received[i] != i {
			t.Fatalf("FIFO order violated at index %d: expected %d, got %d", i, i, received[i])
		}
	}

	t.Logf("FIFO order test passed: %d messages processed in correct order", msgCount)
}

// TestConcurrentSend 测试并发发送
func TestConcurrentSend(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	const goroutineCount = 10
	const msgPerGoroutine = 100
	const totalMsgCount = goroutineCount * msgPerGoroutine

	var received []int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 启动消费者
	go func() {
		for msg := range uc.Channel {
			mu.Lock()
			received = append(received, msg.ID)
			mu.Unlock()
		}
	}()

	// 启动多个发送goroutine
	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			startID := int64(goroutineID * msgPerGoroutine)
			for j := int64(0); j < msgPerGoroutine; j++ {
				msg := &Msg{
					ID:   startID + j,
					Data: []byte(fmt.Sprintf("goroutine_%d_msg_%d", goroutineID, j)),
				}
				if err := uc.SendMsg(msg); err != nil {
					t.Errorf("SendMsg failed: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// 验证所有消息都被接收
	if len(received) != totalMsgCount {
		t.Fatalf("Expected %d messages, got %d", totalMsgCount, len(received))
	}

	// 验证消息ID的唯一性
	seen := make(map[int64]bool)
	for _, id := range received {
		if seen[id] {
			t.Fatalf("Duplicate message ID: %d", id)
		}
		seen[id] = true
	}

	t.Logf("Concurrent send test passed: %d messages processed", totalMsgCount)
}

// TestChannelCapacity 测试通道容量边界
func TestChannelCapacity(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	// 通道容量是1000，发送1001条消息
	const msgCount = 1001
	var received []int64
	var mu sync.Mutex

	// 启动消费者
	go func() {
		for msg := range uc.Channel {
			mu.Lock()
			received = append(received, msg.ID)
			mu.Unlock()
		}
	}()

	// 发送消息
	for i := int64(0); i < msgCount; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("data_%d", i)),
		}
		if err := uc.SendMsg(msg); err != nil {
			t.Fatalf("SendMsg failed: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// 验证所有消息都被接收
	if len(received) != msgCount {
		t.Fatalf("Expected %d messages, got %d", msgCount, len(received))
	}

	t.Logf("Channel capacity test passed: %d messages processed", msgCount)
}

// TestNilMessage 测试空消息处理
func TestNilMessage(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	err := uc.SendMsg(nil)
	if err == nil {
		t.Fatal("Expected error when sending nil message")
	}

	t.Logf("Nil message test passed: %v", err)
}

// TestClosedChannel 测试关闭的通道
func TestClosedChannel(t *testing.T) {
	uc := NewUnboundedChannel()
	uc.Close()

	msg := &Msg{
		ID:   1,
		Data: []byte("test"),
	}

	err := uc.SendMsg(msg)
	if err == nil {
		t.Fatal("Expected error when sending to closed channel")
	}

	if !uc.IsClosed() {
		t.Fatal("Channel should be marked as closed")
	}

	t.Logf("Closed channel test passed: %v", err)
}

// TestStressTest 压力测试
func TestStressTest(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	const msgCount = 10000
	const goroutineCount = 20
	const msgPerGoroutine = msgCount / goroutineCount

	var received []int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 启动消费者
	go func() {
		for msg := range uc.Channel {
			mu.Lock()
			received = append(received, msg.ID)
			mu.Unlock()
		}
	}()

	// 启动多个发送goroutine
	wg.Add(goroutineCount)
	startTime := time.Now()

	for i := 0; i < goroutineCount; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			startID := int64(goroutineID * msgPerGoroutine)
			for j := int64(0); j < msgPerGoroutine; j++ {
				msg := &Msg{
					ID:   startID + j,
					Data: []byte(fmt.Sprintf("stress_goroutine_%d_msg_%d", goroutineID, j)),
				}
				if err := uc.SendMsg(msg); err != nil {
					t.Errorf("SendMsg failed: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	duration := time.Since(startTime)
	throughput := float64(msgCount) / duration.Seconds()

	// 验证所有消息都被接收
	if len(received) != msgCount {
		t.Fatalf("Expected %d messages, got %d", msgCount, len(received))
	}

	t.Logf("Stress test passed: %d messages in %v (%.2f msg/sec)", msgCount, duration, throughput)
}

// TestMixedLoad 测试混合负载（快速路径和慢路径）
func TestMixedLoad(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	const msgCount = 2000
	var received []int64
	var mu sync.Mutex

	// 启动消费者
	go func() {
		for msg := range uc.Channel {
			mu.Lock()
			received = append(received, msg.ID)
			mu.Unlock()
		}
	}()

	// 先快速发送一些消息（快速路径）
	for i := int64(0); i < 500; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("fast_%d", i)),
		}
		if err := uc.SendMsg(msg); err != nil {
			t.Fatalf("SendMsg failed: %v", err)
		}
	}

	// 暂停一下，让消费者处理一些消息
	time.Sleep(10 * time.Millisecond)

	// 继续发送更多消息（可能触发慢路径）
	for i := int64(500); i < msgCount; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("slow_%d", i)),
		}
		if err := uc.SendMsg(msg); err != nil {
			t.Fatalf("SendMsg failed: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// 验证所有消息都被接收
	if len(received) != msgCount {
		t.Fatalf("Expected %d messages, got %d", msgCount, len(received))
	}

	// 验证FIFO顺序
	for i := int64(0); i < msgCount; i++ {
		if received[i] != i {
			t.Fatalf("FIFO order violated at index %d: expected %d, got %d", i, i, received[i])
		}
	}

	t.Logf("Mixed load test passed: %d messages processed in correct order", msgCount)
}

// TestChannelLen 测试通道长度统计
func TestChannelLen(t *testing.T) {
	uc := NewUnboundedChannel()
	defer uc.Close()

	// 发送一些消息
	for i := int64(0); i < 100; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("data_%d", i)),
		}
		if err := uc.SendMsg(msg); err != nil {
			t.Fatalf("SendMsg failed: %v", err)
		}
	}

	// 检查通道长度
	channelLen := uc.ChannelLen()
	if channelLen == 0 {
		t.Log("Channel is empty (all messages transferred to slow path)")
	} else {
		t.Logf("Channel has %d messages", channelLen)
	}

	// 检查List长度
	listLen := uc.Len()
	t.Logf("List has %d messages", listLen)

	t.Logf("Channel length test passed: channel=%d, list=%d", channelLen, listLen)
}

// BenchmarkSendMsg 性能基准测试
func BenchmarkSendMsg(b *testing.B) {
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
				Data: []byte(fmt.Sprintf("benchmark_%d", i)),
			}
			if err := uc.SendMsg(msg); err != nil {
				b.Fatalf("SendMsg failed: %v", err)
			}
			i++
		}
	})
}
