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
	"time"

	rb_pkg "github.com/motocat46/yytools/pkg/ds/ring_buffer"
)

// ---- 吞吐计时 ----

// measureRBEnqueue 向已满的 RingBuffer（容量 n）执行 heapOpsPerMeasure 次覆盖写，返回均摊 ns/op。
func measureRBEnqueue(n int) int64 {
	rb := rb_pkg.NewRingBuffer[int](n)
	for i := range n {
		rb.Enqueue(i)
	}
	start := time.Now()
	for j := range heapOpsPerMeasure {
		rb.Enqueue(n + j) // 覆盖最旧元素
	}
	return time.Since(start).Nanoseconds() / heapOpsPerMeasure
}

// measureRBDequeue 从满载 RingBuffer（容量 n）执行 heapOpsPerMeasure 次出队，返回均摊 ns/op。
// n >> heapOpsPerMeasure（n≥10000），出队后不会触发 Empty panic。
func measureRBDequeue(n int) int64 {
	rb := rb_pkg.NewRingBuffer[int](n)
	for i := range n {
		rb.Enqueue(i)
	}
	start := time.Now()
	for range heapOpsPerMeasure {
		rb.Dequeue()
	}
	return time.Since(start).Nanoseconds() / heapOpsPerMeasure
}

func handleDsRingBufOps(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 50000, 100000, 200000, 500000}
	xLabels := make([]string, len(sizes))
	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)
	}

	rbEnqNs := make([]int64, len(sizes))
	rbDeqNs := make([]int64, len(sizes))
	qEnqNs := make([]int64, len(sizes))
	qDeqNs := make([]int64, len(sizes))

	for i, n := range sizes {
		rbEnqNs[i] = measureRBEnqueue(n)
		rbDeqNs[i] = measureRBDequeue(n)
		// 复用 Queue 稳态计时（预分配 2n 容量，无扩缩容）
		qEnqNs[i] = measureQueueEnqueue(n)
		qDeqNs[i] = measureQueueDequeue(n)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "RingBuffer vs Queue 吞吐对比",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Enqueue / Dequeue 均摊耗时 vs 规模（每规模 %d 次取均值）", heapOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "RingBuffer Enqueue（覆盖写）", Data: rbEnqNs},
				{Name: "RingBuffer Dequeue", Data: rbDeqNs},
				{Name: "Queue Enqueue（预分配，无扩容）", Data: qEnqNs},
				{Name: "Queue Dequeue（预分配，无缩容）", Data: qDeqNs},
			},
		}},
	})
}

// ---- 覆盖写语义演示 ----

const (
	rbOverwriteCap = 100 // 演示用容量
	rbOverwriteN   = 200 // 总入队次数（= 2× 容量）
)

func handleDsRingBufOverwrite(w http.ResponseWriter, _ *http.Request) {
	rb := rb_pkg.NewRingBuffer[int](rbOverwriteCap)

	xLabels := make([]string, rbOverwriteN)
	newestData := make([]int64, rbOverwriteN) // 最新入队元素（= 入队序号）
	oldestData := make([]int64, rbOverwriteN)  // 最旧可见元素（Peek）
	lenData := make([]int64, rbOverwriteN)     // 当前缓冲区元素数

	for k := 1; k <= rbOverwriteN; k++ {
		rb.Enqueue(k)
		xLabels[k-1] = fmt.Sprintf("%d", k)
		newestData[k-1] = int64(k)
		oldestData[k-1] = int64(rb.Peek())
		lenData[k-1] = int64(rb.Len())
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "RingBuffer 覆盖写语义",
		Charts: []chartData{{
			Type:  "line",
			Title: fmt.Sprintf("满载后覆盖最旧元素（容量 %d，共入队 %d 次）", rbOverwriteCap, rbOverwriteN),
			XAxis: xLabels,
			XAxisName: "已入队总数",
			YAxisName: "元素值 / 数量",
			Series: []chartSeries{
				{Name: fmt.Sprintf("Len（最大 %d）", rbOverwriteCap), Data: lenData},
				{Name: "最新元素（Newest）", Data: newestData},
				{Name: "最旧元素 Peek（Oldest）", Data: oldestData},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "ring_buffer/", Title: "RingBuffer vs Queue 吞吐对比",
		Desc: "Enqueue / Dequeue 均摊 ns/op，固定容量 vs 预分配动态队列（1万~50万）",
		Path: "/api/ds/ringbuf/ops", DataHandler: handleDsRingBufOps,
	})
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "ring_buffer/", Title: "RingBuffer 覆盖写语义",
		Desc: "容量 100，入队 200 次：Len 封顶、Oldest 追踪、Newest 持续增长",
		Path: "/api/ds/ringbuf/overwrite", DataHandler: handleDsRingBufOverwrite,
	})
}
