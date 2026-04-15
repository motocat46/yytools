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

// 作者:  yangyuan
// 创建日期:2026/4/15
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	uc "github.com/motocat46/yytools/pkg/infra/concurrency/unbounded_channel"
)

// slowConsumerDelay 模拟每次消费的处理耗时（1ms → 1000 ops/sec）
const slowConsumerDelay = 1 * time.Millisecond

const ucTotalSends = 100_000

// measureUCThroughput 用 producers 个 goroutine 向 UnboundedChannel 发送 ucTotalSends 条消息，
// consumers 个 goroutine 用 Receive() 接收，测量全程耗时，返回万msg/秒。
func measureUCThroughput(producers, consumers, chanSize int) int64 {
	ch := uc.NewUnboundedChannel[int](chanSize, ucTotalSends*2)
	defer ch.Close()

	perProducer := ucTotalSends / producers
	var sendWg, recvWg sync.WaitGroup

	start := time.Now()
	for range consumers {
		recvWg.Add(1)
		go func() {
			defer recvWg.Done()
			for range ucTotalSends / consumers {
				ch.Receive() //nolint:errcheck
			}
		}()
	}
	for range producers {
		sendWg.Add(1)
		go func() {
			defer sendWg.Done()
			for i := range perProducer {
				ch.Send(i)
			}
		}()
	}
	sendWg.Wait()
	recvWg.Wait()
	elapsed := time.Since(start)
	return int64(ucTotalSends) * 1_000_000 / elapsed.Microseconds() / 10_000
}

// measureBufChThroughput 用 buffered channel 做相同全程测量。
func measureBufChThroughput(producers, consumers, chanSize int) int64 {
	ch := make(chan int, chanSize)
	perProducer := ucTotalSends / producers
	var sendWg, recvWg sync.WaitGroup

	start := time.Now()
	for range consumers {
		recvWg.Add(1)
		go func() {
			defer recvWg.Done()
			count := 0
			for range ch {
				count++
				if count == ucTotalSends/consumers {
					return
				}
			}
		}()
	}
	for range producers {
		sendWg.Add(1)
		go func() {
			defer sendWg.Done()
			for i := range perProducer {
				ch <- i
			}
		}()
	}
	sendWg.Wait()
	recvWg.Wait()
	elapsed := time.Since(start)
	return int64(ucTotalSends) * 1_000_000 / elapsed.Microseconds() / 10_000
}

// measureDecouplingMs 在慢消费者（每条消息 sleep slowConsumerDelay）场景下，
// 只计量生产者侧完成时间（ms）。
// UC：Send 进 buffer 不阻塞 → 生产者侧极快完成；
// Buffered：chanSize 满后 Send 阻塞 → 生产者等待慢消费者拖动，完成时间接近消费者完成时间。
const burstSize = 100 // 发送条数；消费者 100×1ms=100ms 完成

func measureDecouplingMs(chanSize int, isUC bool) int64 {
	var sendWg sync.WaitGroup

	if isUC {
		ch := uc.NewUnboundedChannel[int](chanSize, burstSize*10)
		defer ch.Close()
		go func() { // 慢消费者
			for range burstSize {
				time.Sleep(slowConsumerDelay)
				ch.Receive() //nolint:errcheck
			}
		}()
		start := time.Now()
		sendWg.Add(1)
		go func() {
			defer sendWg.Done()
			for i := range burstSize {
				ch.Send(i)
			}
		}()
		sendWg.Wait()
		return time.Since(start).Microseconds()
	}

	ch := make(chan int, chanSize)
	go func() { // 慢消费者
		count := 0
		for range ch {
			time.Sleep(slowConsumerDelay)
			count++
			if count >= burstSize {
				return
			}
		}
	}()
	start := time.Now()
	sendWg.Add(1)
	go func() {
		defer sendWg.Done()
		for i := range burstSize {
			ch <- i
		}
	}()
	sendWg.Wait()
	return time.Since(start).Microseconds()
}

func handleInfraUnboundedChannel(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：全程吞吐 vs 生产者数（4 消费者，chanSize=1024，快速路径为主）
	const fastChanSize = 1024
	const fastConsumers = 4
	producerCounts := []int{1, 2, 4, 8, 16, 32}
	xLabels1 := make([]string, len(producerCounts))
	ucFastNs := make([]int64, len(producerCounts))
	bufFastNs := make([]int64, len(producerCounts))

	for i, p := range producerCounts {
		xLabels1[i] = fmt.Sprintf("%d", p)
		ucFastNs[i] = measureUCThroughput(p, fastConsumers, fastChanSize)
		bufFastNs[i] = measureBufChThroughput(p, fastConsumers, fastChanSize)
	}

	// Chart 2：慢消费者（每条 sleep 1ms）下生产者侧完成时间 vs chanSize
	// 单生产者发 burstSize 条，消费者共需 burstSize×1ms 完成
	// UC：进 buffer 不阻塞 → 生产者毫秒级完成
	// Buffered：chanSize 满后阻塞 → 完成时间趋近消费者时间（burstSize ms）
	chanSizes := []int{10, 25, 50, 75, 100}
	xLabels2 := make([]string, len(chanSizes))
	ucDecoupMs := make([]int64, len(chanSizes))
	bufDecoupMs := make([]int64, len(chanSizes))

	for i, cs := range chanSizes {
		xLabels2[i] = fmt.Sprintf("%d", cs)
		ucDecoupMs[i] = measureDecouplingMs(cs, true)
		bufDecoupMs[i] = measureDecouplingMs(cs, false)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "UnboundedChannel 吞吐对比",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("全程吞吐 vs 生产者数（%d 消费者，chanSize=%d，%d 万消息）", fastConsumers, fastChanSize, ucTotalSends/10000),
				XAxis:     xLabels1,
				XAxisName: "生产者数",
				YAxisName: "万 msg/秒",
				Series: []chartSeries{
					{Name: "UnboundedChannel", Data: ucFastNs},
					{Name: "buffered channel", Data: bufFastNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("慢消费者（每条 %v）下生产者侧完成时间 vs chanSize（发送 %d 条，理论消费完成≈%dms）", slowConsumerDelay, burstSize, burstSize),
				XAxis:     xLabels2,
				XAxisName: "chanSize",
				YAxisName: "生产者完成时间（µs）",
				Series: []chartSeries{
					{Name: "UnboundedChannel（进 buffer 不阻塞）", Data: ucDecoupMs},
					{Name: "buffered channel（满后阻塞）", Data: bufDecoupMs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/infra", SubPkg: "concurrency/unbounded_channel/", Title: "UnboundedChannel 吞吐对比",
		Desc: "全程吞吐 vs 生产者数（chanSize=1024）；慢消费者场景生产者侧完成时间 vs chanSize（Send 不阻塞的解耦效果）",
		Path: "/api/infra/unbounded_channel", DataHandler: handleInfraUnboundedChannel,
	})
}
