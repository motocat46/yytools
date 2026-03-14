// Package unbounded_channel.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unbounded_channel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// 公共定义
// =============================================================================

// iUnboundedChan 统一接口，让V5和V6可以跑相同的测试逻辑
type iUnboundedChan interface {
	Send(msg any) bool
	Receive() (any, bool)
	Close()
	Out() <-chan any
}

// testImpls 所有待测实现，新增版本只需在此追加
var testImpls = []struct {
	name    string
	factory func(chanSize, limit int) iUnboundedChan
}{
	{"V6", func(cs, lim int) iUnboundedChan { return NewUnboundedChannelV6[any](cs, lim) }},
}

// =============================================================================
// 正确性测试
// =============================================================================

// TestCompare_FIFO 单生产者发送N条消息，验证消费顺序严格递增
func TestCompare_FIFO(t *testing.T) {
	const msgCount = 50000
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(100, msgCount*2)
			defer uc.Close()
			
			received := make([]int, 0, msgCount)
			allReceived := make(chan struct{})
			
			go func() {
				for len(received) < msgCount {
					v, ok := uc.Receive()
					if !ok {
						break
					}
					received = append(received, v.(int))
				}
				close(allReceived)
			}()
			
			for i := 0; i < msgCount; i++ {
				uc.Send(i)
			}
			
			select {
			case <-allReceived:
			case <-time.After(15 * time.Second):
				t.Fatalf("超时：仅收到 %d/%d 条消息", len(received), msgCount)
			}
			
			if len(received) != msgCount {
				t.Fatalf("消息数量不符：期望 %d，实际 %d", msgCount, len(received))
			}
			for i, v := range received {
				if v != i {
					t.Fatalf("FIFO乱序：index=%d，期望=%d，实际=%d", i, i, v)
				}
			}
		})
	}
}

// TestCompare_ConcurrentProducers 多生产者并发写入，验证消息不丢失、不重复
func TestCompare_ConcurrentProducers(t *testing.T) {
	const (
		producers   = 20
		msgsPerProd = 5000
		total       = producers * msgsPerProd
	)
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(100, total*2)
			defer uc.Close()
			
			seen := make([]atomic.Bool, total)
			var receivedCount atomic.Int64
			allReceived := make(chan struct{})
			
			// 消费者
			go func() {
				for {
					v, ok := uc.Receive()
					if !ok {
						return
					}
					id := v.(int)
					seen[id].Store(true)
					if receivedCount.Add(1) == int64(total) {
						close(allReceived)
						return
					}
				}
			}()
			
			// 多生产者并发发送
			var wg sync.WaitGroup
			wg.Add(producers)
			for p := 0; p < producers; p++ {
				go func(id int) {
					defer wg.Done()
					base := id * msgsPerProd
					for i := 0; i < msgsPerProd; i++ {
						uc.Send(base + i)
					}
				}(p)
			}
			wg.Wait()
			
			select {
			case <-allReceived:
			case <-time.After(20 * time.Second):
				t.Fatalf("超时：仅收到 %d/%d 条消息", receivedCount.Load(), total)
			}
			
			// 验证无丢失、无重复
			for i := 0; i < total; i++ {
				if !seen[i].Load() {
					t.Fatalf("消息丢失：id=%d 未被消费", i)
				}
			}
		})
	}
}

// TestCompare_SlowConsumer 快速生产+慢速消费，验证buffer在背压下不丢数据
func TestCompare_SlowConsumer(t *testing.T) {
	const msgCount = 500
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(10, msgCount*2)
			defer uc.Close()
			
			var count atomic.Int64
			allReceived := make(chan struct{})
			
			// 慢速消费者：每条消息处理 2ms
			go func() {
				for {
					_, ok := uc.Receive()
					if !ok {
						return
					}
					time.Sleep(2 * time.Millisecond)
					if count.Add(1) == int64(msgCount) {
						close(allReceived)
						return
					}
				}
			}()
			
			// 瞬间全部发送
			for i := 0; i < msgCount; i++ {
				uc.Send(i)
			}
			
			select {
			case <-allReceived:
			case <-time.After(30 * time.Second):
				t.Fatalf("超时：仅收到 %d/%d 条消息", count.Load(), msgCount)
			}
		})
	}
}

// TestCompare_BurstThenDrain 先无消费者堆满buffer，再启动消费者全部取出，验证FIFO
func TestCompare_BurstThenDrain(t *testing.T) {
	const (
		msgCount = 5000
		chanSize = 10
	)
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(chanSize, msgCount*2)
			
			// 阶段1：无消费者，全部发送（channel满后溢出到buffer）
			var wgSend sync.WaitGroup
			wgSend.Add(1)
			go func() {
				defer wgSend.Done()
				for i := 0; i < msgCount; i++ {
					uc.Send(i)
				}
			}()
			wgSend.Wait()
			
			// 阶段2：启动消费者，全部取出
			received := make([]int, 0, msgCount)
			allReceived := make(chan struct{})
			go func() {
				for len(received) < msgCount {
					v, ok := uc.Receive()
					if !ok {
						break
					}
					received = append(received, v.(int))
				}
				close(allReceived)
			}()
			
			select {
			case <-allReceived:
			case <-time.After(15 * time.Second):
				t.Fatalf("超时：仅收到 %d/%d 条消息", len(received), msgCount)
			}
			uc.Close()
			
			if len(received) != msgCount {
				t.Fatalf("消息数量不符：期望 %d，实际 %d", msgCount, len(received))
			}
			// 单生产者场景验证FIFO
			for i, v := range received {
				if v != i {
					t.Fatalf("FIFO乱序：index=%d，期望=%d，实际=%d", i, i, v)
				}
			}
		})
	}
}

// =============================================================================
// 时延对比测试（核心差异场景）
// =============================================================================

// TestCompare_BurstDrainLatency
// 重点场景：buffer积压 + 无新生产者 + 消费者开始排空。
// V5依赖listCheck的退避唤醒；V6依赖Receive()后的信号触发。
// 该场景最能体现两个版本的时延差异。
func TestCompare_BurstDrainLatency(t *testing.T) {
	const (
		msgCount = 10000
		chanSize = 32
		rounds   = 5 // 多轮取平均，减少调度抖动
	)
	
	type result struct {
		total time.Duration
		count int
	}
	results := make(map[string]*result)
	for _, impl := range testImpls {
		results[impl.name] = &result{}
	}
	
	for round := 0; round < rounds; round++ {
		for _, impl := range testImpls {
			uc := impl.factory(chanSize, msgCount*2)
			
			// 阶段1：无消费者，先堆满buffer
			for i := 0; i < msgCount; i++ {
				uc.Send(i)
			}
			
			// 阶段2：计时消费者排空全部消息
			start := time.Now()
			count := 0
			for count < msgCount {
				_, ok := uc.Receive()
				if !ok {
					break
				}
				count++
			}
			elapsed := time.Since(start)
			uc.Close()
			
			if count != msgCount {
				t.Errorf("[%s] round=%d 消息丢失：期望%d，实际%d", impl.name, round, msgCount, count)
				continue
			}
			results[impl.name].total += elapsed
			results[impl.name].count++
		}
	}
	
	t.Logf("BurstDrainLatency（%d条消息，chanSize=%d，%d轮平均）：", msgCount, chanSize, rounds)
	for _, impl := range testImpls {
		r := results[impl.name]
		if r.count == 0 {
			continue
		}
		avg := r.total / time.Duration(r.count)
		throughput := float64(msgCount) / avg.Seconds()
		t.Logf("  %-4s  平均耗时=%-12v  吞吐=%.0f msg/s", impl.name, avg, throughput)
	}
}

// =============================================================================
// 性能基准测试
// =============================================================================

// benchRun 通用基准逻辑：RunParallel并发写入，单消费者持续消费
func benchRun(b *testing.B, uc iUnboundedChan) {
	b.Helper()
	defer uc.Close()
	
	go func() {
		for {
			if _, ok := uc.Receive(); !ok {
				return
			}
		}
	}()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			uc.Send(i)
			i++
		}
	})
	b.StopTimer()
}

// --- 低负载：channel容量大，buffer几乎不触发，主要测快速路径 ---

func BenchmarkV6_LowLoad(b *testing.B) {
	benchRun(b, NewUnboundedChannelV6[any](100000, 1000000))
}

// --- 高负载：channel容量小，buffer频繁触发，测慢路径+worker效率 ---

func BenchmarkV6_HighLoad(b *testing.B) {
	benchRun(b, NewUnboundedChannelV6[any](16, 1000000))
}

// --- 多生产者：测并发竞争下的锁争用 ---

func benchMultiProducer(b *testing.B, uc iUnboundedChan, producers int) {
	b.Helper()
	defer uc.Close()
	
	go func() {
		for {
			if _, ok := uc.Receive(); !ok {
				return
			}
		}
	}()
	
	var wg sync.WaitGroup
	msgPerProd := b.N / producers
	if msgPerProd == 0 {
		msgPerProd = 1
	}
	
	b.ResetTimer()
	wg.Add(producers)
	for p := 0; p < producers; p++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < msgPerProd; i++ {
				uc.Send(id*msgPerProd + i)
			}
		}(p)
	}
	wg.Wait()
	b.StopTimer()
}

func BenchmarkV6_MultiProducer8(b *testing.B) {
	benchMultiProducer(b, NewUnboundedChannelV6[any](100, 1000000), 8)
}

func BenchmarkV6_MultiProducer32(b *testing.B) {
	benchMultiProducer(b, NewUnboundedChannelV6[any](100, 1000000), 32)
}

// --- 吞吐量对比（固定消息量，测总耗时）---

func benchThroughput(b *testing.B, uc iUnboundedChan, msgCount int) {
	b.Helper()
	defer uc.Close()
	
	var received atomic.Int64
	done := make(chan struct{})
	go func() {
		for {
			_, ok := uc.Receive()
			if !ok {
				return
			}
			if received.Add(1) == int64(msgCount) {
				close(done)
				return
			}
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < msgCount; i++ {
		uc.Send(i)
	}
	<-done
	b.StopTimer()
	
	b.ReportMetric(float64(msgCount)/b.Elapsed().Seconds(), "msg/s")
}

func BenchmarkV6_Throughput(b *testing.B) {
	benchThroughput(b, NewUnboundedChannelV6[any](1000, 1000000), 100000)
}

// =============================================================================
// 严格 FIFO 语义验证
// =============================================================================
//
// FIFO 保证的范围：
//   - 单生产者：严格全序（消费顺序与发送顺序完全一致）
//   - 多生产者：每个生产者的消息在全局接收序列中保持相对顺序；
//               不同生产者之间的顺序由 OS 调度决定，与原生 channel 语义一致
//
// 多消费者的验证注意：
//   两个消费者并发读取时，recording 顺序 ≠ dequeue 顺序。
//   因此跨消费者的全局 FIFO 验证必须通过单消费者完成；
//   多消费者测试只能验证：无丢失、无重复、以及单个消费者视角的局部有序。
// =============================================================================

// fifoMsg 携带生产者标识和序列号，用于 FIFO 验证
type fifoMsg struct {
	producerID int
	seq        int
}

// TestFIFO_SingleProd_HeavyBuffer
// 最严格的 FIFO 测试：单生产者 + 极小 chanSize，强制绝大多数消息经过
// buffer→channel 转移路径，验证转移过程是否保持严格有序。
func TestFIFO_SingleProd_HeavyBuffer(t *testing.T) {
	const (
		msgCount = 100000
		chanSize = 16 // 极小，强制走 buffer
	)
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(chanSize, msgCount*2)
			defer uc.Close()
			
			received := make([]int, 0, msgCount)
			done := make(chan struct{})
			
			go func() {
				for len(received) < msgCount {
					v, ok := uc.Receive()
					if !ok {
						break
					}
					received = append(received, v.(fifoMsg).seq)
				}
				close(done)
			}()
			
			for i := 0; i < msgCount; i++ {
				uc.Send(fifoMsg{0, i})
			}
			
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatalf("超时：已收 %d/%d", len(received), msgCount)
			}
			
			if len(received) != msgCount {
				t.Fatalf("消息丢失：期望 %d，实际 %d", msgCount, len(received))
			}
			for i, v := range received {
				if v != i {
					t.Fatalf("FIFO乱序：index=%d 期望=%d 实际=%d", i, i, v)
				}
			}
		})
	}
}

// TestFIFO_MultiProd_SingleCons
// 多生产者 + 单消费者。单消费者顺序读取，可严格验证：
// 每个生产者的消息在全局 dequeue 序列中保持相对顺序。
// chanSize 设为小值，强制走 buffer，压测转移路径的 FIFO 保证。
func TestFIFO_MultiProd_SingleCons(t *testing.T) {
	const (
		producers   = 20
		msgsPerProd = 5000
		chanSize    = 64
	)
	total := producers * msgsPerProd
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(chanSize, total*2)
			defer uc.Close()
			
			// 每个生产者最后一次收到的 seq（初始 -1）
			lastSeq := make([]int, producers)
			for i := range lastSeq {
				lastSeq[i] = -1
			}
			receivedCount := 0
			done := make(chan struct{})
			
			// 单消费者：记录每个生产者的 seq 递增性
			go func() {
				for receivedCount < total {
					v, ok := uc.Receive()
					if !ok {
						break
					}
					m := v.(fifoMsg)
					if m.seq <= lastSeq[m.producerID] {
						t.Errorf("生产者%d FIFO乱序：收到seq=%d，上次已收到seq=%d",
							m.producerID, m.seq, lastSeq[m.producerID])
					}
					lastSeq[m.producerID] = m.seq
					receivedCount++
				}
				close(done)
			}()
			
			var wg sync.WaitGroup
			wg.Add(producers)
			for p := 0; p < producers; p++ {
				go func(id int) {
					defer wg.Done()
					for i := 0; i < msgsPerProd; i++ {
						uc.Send(fifoMsg{id, i})
					}
				}(p)
			}
			wg.Wait()
			
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatalf("超时：已收 %d/%d", receivedCount, total)
			}
		})
	}
}

// TestFIFO_MultiProd_MultiCons
// 多生产者 + 多消费者。
// 可验证：无丢失、无重复、每个消费者局部视角内的 per-producer 有序。
// 注意：由于并发消费，无法通过此测试验证跨消费者的全局 dequeue 顺序，
// 全局顺序需依赖 TestFIFO_MultiProd_SingleCons 保证。
func TestFIFO_MultiProd_MultiCons(t *testing.T) {
	const (
		producers   = 20
		msgsPerProd = 5000
		chanSize    = 64
		consumers   = 5
	)
	total := producers * msgsPerProd
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(chanSize, total*2)
			defer uc.Close()
			
			// 用于验证无重复：记录每条消息是否被消费过
			received := make([]atomic.Bool, total) // [producerID*msgsPerProd + seq]
			var totalReceived atomic.Int64
			allDone := make(chan struct{})
			var once sync.Once
			
			for c := 0; c < consumers; c++ {
				go func(consumerID int) {
					// 每个消费者独立追踪 per-producer 的最后 seq
					lastSeq := make([]int, producers)
					for i := range lastSeq {
						lastSeq[i] = -1
					}
					for {
						v, ok := uc.Receive()
						if !ok {
							return
						}
						m := v.(fifoMsg)
						idx := m.producerID*msgsPerProd + m.seq
						
						// 验证无重复
						if !received[idx].CompareAndSwap(false, true) {
							t.Errorf("消费者%d：消息重复消费 producerID=%d seq=%d",
								consumerID, m.producerID, m.seq)
						}
						
						// 验证局部 per-producer 有序（必要非充分条件）
						if m.seq <= lastSeq[m.producerID] {
							t.Errorf("消费者%d：生产者%d FIFO乱序 seq=%d <= lastSeq=%d",
								consumerID, m.producerID, m.seq, lastSeq[m.producerID])
						}
						lastSeq[m.producerID] = m.seq
						
						if totalReceived.Add(1) == int64(total) {
							once.Do(func() { close(allDone) })
						}
					}
				}(c)
			}
			
			var wg sync.WaitGroup
			wg.Add(producers)
			for p := 0; p < producers; p++ {
				go func(id int) {
					defer wg.Done()
					for i := 0; i < msgsPerProd; i++ {
						uc.Send(fifoMsg{id, i})
					}
				}(p)
			}
			wg.Wait()
			
			select {
			case <-allDone:
			case <-time.After(30 * time.Second):
				t.Fatalf("超时：已收 %d/%d", totalReceived.Load(), total)
			}
			
			// 验证无丢失
			for p := 0; p < producers; p++ {
				for s := 0; s < msgsPerProd; s++ {
					if !received[p*msgsPerProd+s].Load() {
						t.Errorf("消息丢失：producerID=%d seq=%d", p, s)
					}
				}
			}
		})
	}
}

// TestFIFO_SingleProd_MultiCons
// 单生产者 + 多消费者，混合使用 Receive() 和 Out()。
//
// 多消费者场景下全局 dequeue 顺序无法通过并发 recording 验证
// （receipt 号分配与 dequeue 是两个非原子步骤，调度会使两者错位）。
// 因此本测试只验证可靠的必要条件：
//   1. 无消息丢失（seen 位图全部置位）
//   2. 无消息重复（CompareAndSwap 保证）
//   3. 每个消费者局部视角内消息严格递增（单生产者，每个消费者的子序列必须有序）
//
// 全局严格 FIFO 已由 TestFIFO_SingleProd_HeavyBuffer（单消费者）覆盖。
func TestFIFO_SingleProd_MultiCons(t *testing.T) {
	const (
		msgCount  = 50000
		chanSize  = 64
		consumers = 4 // 前2个用 Receive()，后2个用 Out()
	)
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(chanSize, msgCount*2)
			defer uc.Close()
			
			seen := make([]atomic.Bool, msgCount)
			var totalReceived atomic.Int64
			allDone := make(chan struct{})
			var once sync.Once
			
			consume := func(val int) {
				if !seen[val].CompareAndSwap(false, true) {
					t.Errorf("消息重复消费：val=%d", val)
				}
				if totalReceived.Add(1) == int64(msgCount) {
					once.Do(func() { close(allDone) })
				}
			}
			
			// 消费者 1、2：使用 Receive()，验证局部有序
			for c := 0; c < consumers/2; c++ {
				go func() {
					lastVal := -1
					for {
						v, ok := uc.Receive()
						if !ok {
							return
						}
						val := v.(fifoMsg).seq
						if val <= lastVal {
							t.Errorf("Receive()消费者局部乱序：val=%d <= lastVal=%d", val, lastVal)
						}
						lastVal = val
						consume(val)
					}
				}()
			}
			
			// 消费者 3、4：使用 Out()，验证局部有序
			for c := 0; c < consumers-consumers/2; c++ {
				go func() {
					lastVal := -1
					for v := range uc.Out() {
						val := v.(fifoMsg).seq
						if val <= lastVal {
							t.Errorf("Out()消费者局部乱序：val=%d <= lastVal=%d", val, lastVal)
						}
						lastVal = val
						consume(val)
					}
				}()
			}
			
			// 单生产者按序发送
			for i := 0; i < msgCount; i++ {
				uc.Send(fifoMsg{0, i})
			}
			
			select {
			case <-allDone:
			case <-time.After(30 * time.Second):
				t.Fatalf("超时：已收 %d/%d", totalReceived.Load(), msgCount)
			}
			
			// 验证无丢失
			for i := range seen {
				if !seen[i].Load() {
					t.Errorf("消息丢失：val=%d", i)
				}
			}
		})
	}
}

// TestFIFO_Stress_HighConcurrency
// 压力场景：大量生产者 + 大量消费者 + 小 chanSize，
// 在极端并发下验证无丢失、无重复，以及局部 per-producer 有序。
func TestFIFO_Stress_HighConcurrency(t *testing.T) {
	const (
		producers   = 100
		msgsPerProd = 1000
		chanSize    = 32
		consumers   = 10
	)
	total := producers * msgsPerProd
	
	for _, impl := range testImpls {
		t.Run(impl.name, func(t *testing.T) {
			uc := impl.factory(chanSize, total*2)
			defer uc.Close()
			
			received := make([]atomic.Bool, total)
			var totalReceived atomic.Int64
			allDone := make(chan struct{})
			var once sync.Once
			
			for c := 0; c < consumers; c++ {
				go func(consumerID int) {
					lastSeq := make([]int, producers)
					for i := range lastSeq {
						lastSeq[i] = -1
					}
					for {
						v, ok := uc.Receive()
						if !ok {
							return
						}
						m := v.(fifoMsg)
						idx := m.producerID*msgsPerProd + m.seq
						
						if !received[idx].CompareAndSwap(false, true) {
							t.Errorf("消息重复：producerID=%d seq=%d", m.producerID, m.seq)
						}
						if m.seq <= lastSeq[m.producerID] {
							t.Errorf("消费者%d：生产者%d 局部乱序 seq=%d <= lastSeq=%d",
								consumerID, m.producerID, m.seq, lastSeq[m.producerID])
						}
						lastSeq[m.producerID] = m.seq
						
						if totalReceived.Add(1) == int64(total) {
							once.Do(func() { close(allDone) })
						}
					}
				}(c)
			}
			
			var wg sync.WaitGroup
			wg.Add(producers)
			for p := 0; p < producers; p++ {
				go func(id int) {
					defer wg.Done()
					for i := 0; i < msgsPerProd; i++ {
						uc.Send(fifoMsg{id, i})
					}
				}(p)
			}
			wg.Wait()
			
			select {
			case <-allDone:
			case <-time.After(60 * time.Second):
				t.Fatalf("超时：已收 %d/%d", totalReceived.Load(), total)
			}
			
			// 验证无丢失
			lost := 0
			for i := range received {
				if !received[i].Load() {
					lost++
				}
			}
			if lost > 0 {
				t.Errorf("消息丢失：%d 条", lost)
			}
		})
	}
}

// =============================================================================
// 接近真实项目参数的测试（chanSize=1万）
// =============================================================================
//
// 场景说明：
//   - chanSize=10000 模拟实际项目配置
//   - 快速路径命中率高，buffer 仅在突发时触发
//   - 重点关注：突发排空延迟、多生产者吞吐、慢消费者下的buffer行为

const realChanSize = 10000

// BenchmarkReal_LowLoad_V6
// 生产速率平稳，channel 基本不满，几乎全走快速路径
func BenchmarkReal_LowLoad_V6(b *testing.B) {
	benchRun(b, NewUnboundedChannelV6[any](realChanSize, 1000000))
}

// BenchmarkReal_BurstOverflow_V6
// 突发流量超过 chanSize，buffer 被触发，测搬运效率
func BenchmarkReal_BurstOverflow_V6(b *testing.B) {
	benchMultiProducer(b, NewUnboundedChannelV6[any](realChanSize, 1000000), 32)
}

// BenchmarkReal_Throughput_V6
// 固定发送50万条消息，测从开始到全部消费完的端到端吞吐
func BenchmarkReal_Throughput_V6(b *testing.B) {
	benchThroughput(b, NewUnboundedChannelV6[any](realChanSize, 1000000), 500000)
}

// TestReal_BurstDrainLatency 突发堆积后排空的延迟对比（chanSize=1万）
// 与 TestCompare_BurstDrainLatency 对比，观察大 chanSize 对延迟的影响
func TestReal_BurstDrainLatency(t *testing.T) {
	const (
		msgCount = 100000 // 10倍于 chanSize，确保大量消息进入 buffer
		rounds   = 5
	)
	
	type result struct {
		total time.Duration
		count int
	}
	results := make(map[string]*result)
	for _, impl := range testImpls {
		results[impl.name] = &result{}
	}
	
	for round := 0; round < rounds; round++ {
		for _, impl := range testImpls {
			uc := impl.factory(realChanSize, msgCount*2)
			
			// 阶段1：无消费者，全部发送
			for i := 0; i < msgCount; i++ {
				uc.Send(i)
			}
			
			// 阶段2：计时排空
			start := time.Now()
			count := 0
			for count < msgCount {
				_, ok := uc.Receive()
				if !ok {
					break
				}
				count++
			}
			elapsed := time.Since(start)
			uc.Close()
			
			if count != msgCount {
				t.Errorf("[%s] round=%d 消息丢失：期望%d 实际%d", impl.name, round, msgCount, count)
				continue
			}
			results[impl.name].total += elapsed
			results[impl.name].count++
		}
	}
	
	t.Logf("Real_BurstDrainLatency（%d条，chanSize=%d，%d轮平均）：", msgCount, realChanSize, rounds)
	for _, impl := range testImpls {
		r := results[impl.name]
		if r.count == 0 {
			continue
		}
		avg := r.total / time.Duration(r.count)
		throughput := float64(msgCount) / avg.Seconds()
		t.Logf("  %-4s  平均耗时=%-12v  吞吐=%.0f msg/s", impl.name, avg, throughput)
	}
}

// =============================================================================
// 并发扩展性测试：生产者数量从8递增到1000，观察 V6 的效率变化趋势
// 使用小 chanSize=64，强制大量消息走慢路径，让 signal() 被频繁调用
// =============================================================================

func benchScale(b *testing.B, uc iUnboundedChan, producers int) {
	b.Helper()
	benchMultiProducer(b, uc, producers)
}

// chanSize=64（小），强制慢路径，最能暴露 signal() 开销
func BenchmarkScale_V6_Chan64_P8(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](64, 1000000), 8)
}
func BenchmarkScale_V6_Chan64_P32(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](64, 1000000), 32)
}
func BenchmarkScale_V6_Chan64_P64(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](64, 1000000), 64)
}
func BenchmarkScale_V6_Chan64_P128(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](64, 1000000), 128)
}
func BenchmarkScale_V6_Chan64_P256(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](64, 1000000), 256)
}
func BenchmarkScale_V6_Chan64_P1000(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](64, 1000000), 1000)
}

// chanSize=10000（大），对照组：快速路径为主，signal() 几乎不触发
func BenchmarkScale_V6_Chan10k_P8(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](10000, 1000000), 8)
}
func BenchmarkScale_V6_Chan10k_P64(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](10000, 1000000), 64)
}
func BenchmarkScale_V6_Chan10k_P256(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](10000, 1000000), 256)
}
func BenchmarkScale_V6_Chan10k_P1000(b *testing.B) {
	benchScale(b, NewUnboundedChannelV6[any](10000, 1000000), 1000)
}

// =============================================================================
// signal() 开销隔离测试
// 直接测量 N 个 goroutine 并发调用 signal() 的代价，排除其他干扰
// =============================================================================

func BenchmarkSignalContention_P1(b *testing.B) {
	uc := NewUnboundedChannelV6[any](100, 1000000)
	defer uc.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			uc.signal()
		}
	})
}

func BenchmarkSignalContention_P8(b *testing.B) {
	uc := NewUnboundedChannelV6[any](100, 1000000)
	defer uc.Close()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			uc.signal()
		}
	})
}

func BenchmarkSignalContention_P64(b *testing.B) {
	uc := NewUnboundedChannelV6[any](100, 1000000)
	defer uc.Close()
	b.SetParallelism(64)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			uc.signal()
		}
	})
}

func BenchmarkSignalContention_P256(b *testing.B) {
	uc := NewUnboundedChannelV6[any](100, 1000000)
	defer uc.Close()
	b.SetParallelism(256)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			uc.signal()
		}
	})
}

// 对照：测 atomic.Int32.Add 的并发开销（V5 慢路径也用到）
func BenchmarkAtomicAdd_P256(b *testing.B) {
	var v atomic.Int32
	b.SetParallelism(256)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Add(1)
		}
	})
}

// =============================================================================
// 与原生 channel 的吞吐量对比
// 量化"无限容量"特性的额外开销
// =============================================================================

// benchNativeChan 原生 buffered channel 基准，作为性能基线
// 注意：原生 channel 满时生产者会阻塞；V5/V6 满时溢出到 buffer
// 因此只有在 channel 不满的场景下（chanSize 足够大）才是公平对比
func benchNativeChan(b *testing.B, chanSize int) {
	b.Helper()
	ch := make(chan any, chanSize)
	defer close(ch)
	
	go func() {
		for range ch {
		}
	}()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ch <- i
			i++
		}
	})
	b.StopTimer()
}

// --- 场景1：chanSize 足够大，buffer 几乎不触发 ---
// 公平对比：两者都走快速路径，体现 V6 包装层的固定开销

func BenchmarkQPS_NativeChan_Large(b *testing.B) { benchNativeChan(b, 100000) }
func BenchmarkQPS_V6_Large(b *testing.B)         { benchRun(b, NewUnboundedChannelV6[any](100000, 1000000)) }

// --- 场景2：chanSize 较小，buffer 频繁触发 ---
// 非对等对比：原生 channel 满时阻塞生产者；V6 溢出到 buffer 继续接收
// 体现"无限容量"在高负载下的额外代价（以及避免阻塞的价值）

func BenchmarkQPS_NativeChan_Small(b *testing.B) { benchNativeChan(b, 64) }
func BenchmarkQPS_V6_Small(b *testing.B)         { benchRun(b, NewUnboundedChannelV6[any](64, 1000000)) }

// --- 场景3：接近实际项目参数（chanSize=1万）---

func BenchmarkQPS_NativeChan_10k(b *testing.B) { benchNativeChan(b, 10000) }
func BenchmarkQPS_V6_10k(b *testing.B)         { benchRun(b, NewUnboundedChannelV6[any](10000, 1000000)) }

// --- 场景4：固定消息量，精确测量端到端吞吐（msg/s）---

func BenchmarkQPS_Throughput_NativeChan(b *testing.B) {
	ch := make(chan any, 10000)
	var done atomic.Int64
	allDone := make(chan struct{})
	const msgCount = 200000
	go func() {
		for range ch {
			if done.Add(1) == int64(msgCount)*int64(b.N) {
				close(allDone)
				return
			}
		}
	}()
	b.ResetTimer()
	for range b.N {
		for i := range msgCount {
			ch <- i
		}
	}
	<-allDone
	b.StopTimer()
	b.ReportMetric(float64(msgCount*b.N)/b.Elapsed().Seconds(), "msg/s")
}

func BenchmarkQPS_Throughput_V6(b *testing.B) {
	uc := NewUnboundedChannelV6[any](10000, 1000000)
	defer uc.Close()
	var done atomic.Int64
	allDone := make(chan struct{})
	const msgCount = 200000
	go func() {
		for {
			if _, ok := uc.Receive(); !ok {
				return
			}
			if done.Add(1) == int64(msgCount)*int64(b.N) {
				close(allDone)
				return
			}
		}
	}()
	b.ResetTimer()
	for range b.N {
		for i := range msgCount {
			uc.Send(i)
		}
	}
	<-allDone
	b.StopTimer()
	b.ReportMetric(float64(msgCount*b.N)/b.Elapsed().Seconds(), "msg/s")
}

// init 打印分隔，方便阅读测试输出
func init() {
	fmt.Println("=== NativeChan vs V6 对比测试 ===")
}