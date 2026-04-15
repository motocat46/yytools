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

	queue_pkg "github.com/motocat46/yytools/pkg/ds/queue"
)

// ---- 稳态吞吐计时 ----

// measureQueueEnqueue 在 n 元素稳态下测量 heapOpsPerMeasure 次 Enqueue，返回均摊 ns/op。
// 使用 2n 预分配容量，避免扩容干扰计时。
func measureQueueEnqueue(n int) int64 {
	q := queue_pkg.NewQueueWithSize[int](n * 2)
	for i := range n {
		q.Enqueue(i)
	}
	start := time.Now()
	for j := range heapOpsPerMeasure {
		q.Enqueue(n + j)
	}
	return time.Since(start).Nanoseconds() / heapOpsPerMeasure
}

// measureQueueDequeue 在 n 元素稳态下测量 heapOpsPerMeasure 次 Dequeue，返回均摊 ns/op。
// 使用 2n 预分配容量，避免缩容干扰计时（n-1000 远大于 capacity/4 = n/2，n≥10000 时成立）。
func measureQueueDequeue(n int) int64 {
	q := queue_pkg.NewQueueWithSize[int](n * 2)
	for i := range n {
		q.Enqueue(i)
	}
	start := time.Now()
	for range heapOpsPerMeasure {
		q.Dequeue()
	}
	return time.Since(start).Nanoseconds() / heapOpsPerMeasure
}

// ---- 扩缩容行为模拟 ----

const (
	queueSimN    = 2000 // 入队元素数（随后等量出队）
	queueSimStep = 20   // 每隔 queueSimStep 次操作采样一次
)

func handleDsQueueOps(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 50000, 100000, 200000, 500000}
	xLabels := make([]string, len(sizes))
	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)
	}

	enqueueNs := make([]int64, len(sizes))
	dequeueNs := make([]int64, len(sizes))

	for i, n := range sizes {
		enqueueNs[i] = measureQueueEnqueue(n)
		dequeueNs[i] = measureQueueDequeue(n)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Queue 稳态吞吐",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Enqueue / Dequeue 均摊耗时 vs 规模（预分配容量，无扩缩容，每规模 %d 次取均值）", heapOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "队列规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "Enqueue", Data: enqueueNs},
				{Name: "Dequeue", Data: dequeueNs},
			},
		}},
	})
}

func handleDsQueueResize(w http.ResponseWriter, _ *http.Request) {
	// 初始容量 DEFAULT_QUEUE_SIZE（16），入队 2000 个后等量出队，全程记录 Len 与 Capacity
	q := queue_pkg.NewQueue[int]()

	totalOps := queueSimN * 2
	numSamples := totalOps/queueSimStep + 1 // 含 op=0 的初始快照
	xLabels := make([]string, numSamples)
	lenData := make([]int64, numSamples)
	capData := make([]int64, numSamples)

	// op=0：初始快照
	idx := 0
	xLabels[idx] = "0"
	lenData[idx] = int64(q.Len())
	capData[idx] = int64(q.Capacity())
	idx++

	// 入队阶段
	for op := 1; op <= queueSimN; op++ {
		q.Enqueue(op)
		if op%queueSimStep == 0 {
			xLabels[idx] = fmt.Sprintf("%d", op)
			lenData[idx] = int64(q.Len())
			capData[idx] = int64(q.Capacity())
			idx++
		}
	}

	// 出队阶段（操作序号从 queueSimN+1 开始，连续编号）
	for op := 1; op <= queueSimN; op++ {
		q.Dequeue()
		if op%queueSimStep == 0 {
			xLabels[idx] = fmt.Sprintf("%d", queueSimN+op)
			lenData[idx] = int64(q.Len())
			capData[idx] = int64(q.Capacity())
			idx++
		}
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Queue 扩缩容行为",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Len vs Capacity（入队 %d 个后出队 %d 个，初始容量 %d）", queueSimN, queueSimN, queue_pkg.DEFAULT_QUEUE_SIZE),
			XAxis:     xLabels,
			XAxisName: "操作序号",
			YAxisName: "数量",
			Series: []chartSeries{
				{Name: "Len（元素数量）", Data: lenData},
				{Name: "Capacity（容量）", Data: capData},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "queue/", Title: "Queue 稳态吞吐",
		Desc: "Enqueue / Dequeue 均摊 ns/op（预分配，无扩缩容，1万~50万）",
		Path: "/api/ds/queue/ops", DataHandler: handleDsQueueOps,
	})
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "queue/", Title: "Queue 扩缩容行为",
		Desc: "入队 2000 后出队 2000，Len vs Capacity 阶梯变化（初始容量 16）",
		Path: "/api/ds/queue/resize", DataHandler: handleDsQueueResize,
	})
}
