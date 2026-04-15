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
	"math/rand/v2"
	"net/http"
	"time"

	heap_pkg "github.com/motocat46/yytools/pkg/ds/heap"
)

const heapOpsPerMeasure = 1000

// fillMinHeap 预填充 n 个元素到最小堆，使用固定种子。
func fillMinHeap(n int) *heap_pkg.Heap[struct{}] {
	rng := rand.New(rand.NewPCG(42, 0))
	h := heap_pkg.NewHeap[struct{}]()
	for range n {
		h.PushItem(&heap_pkg.Item[struct{}]{Weight: rng.IntN(10_000_000)})
	}
	return h
}

// fillPriorityQueue 预填充 n 个元素到优先级队列，返回队列及所有元素引用（供 UpdatePriority 使用）。
func fillPriorityQueue(n int) (*heap_pkg.PriorityQueue[struct{}], []*heap_pkg.PriorityItem[struct{}]) {
	rng := rand.New(rand.NewPCG(42, 0))
	pq := heap_pkg.NewPriorityQueue[struct{}]()
	items := make([]*heap_pkg.PriorityItem[struct{}], n)
	for i := range n {
		item := &heap_pkg.PriorityItem[struct{}]{Priority: rng.IntN(10_000_000)}
		pq.PushItem(item)
		items[i] = item
	}
	return pq, items
}

func handleDsHeap(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 50000, 100000, 200000, 500000}
	xLabels := make([]string, len(sizes))
	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)
	}

	minPushNs := make([]int64, len(sizes))
	minPopNs := make([]int64, len(sizes))
	pqPushNs := make([]int64, len(sizes))
	pqPopNs := make([]int64, len(sizes))
	pqUpdateNs := make([]int64, len(sizes))

	rng := rand.New(rand.NewPCG(99, 0))

	for i, n := range sizes {
		// MinHeap Push：向 n 元素堆压入 heapOpsPerMeasure 个元素
		h := fillMinHeap(n)
		start := time.Now()
		for range heapOpsPerMeasure {
			h.PushItem(&heap_pkg.Item[struct{}]{Weight: rng.IntN(10_000_000)})
		}
		minPushNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure

		// MinHeap Pop：从 n 元素堆弹出 heapOpsPerMeasure 个元素
		h = fillMinHeap(n)
		start = time.Now()
		for range heapOpsPerMeasure {
			h.PopItem()
		}
		minPopNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure

		// PriorityQueue Push
		pq, _ := fillPriorityQueue(n)
		start = time.Now()
		for range heapOpsPerMeasure {
			pq.PushItem(&heap_pkg.PriorityItem[struct{}]{Priority: rng.IntN(10_000_000)})
		}
		pqPushNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure

		// PriorityQueue Pop
		pq, _ = fillPriorityQueue(n)
		start = time.Now()
		for range heapOpsPerMeasure {
			pq.PopItem()
		}
		pqPopNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure

		// PriorityQueue UpdatePriority：预选 heapOpsPerMeasure 个元素引用，计时更新优先级
		pq, items := fillPriorityQueue(n)
		selected := make([]*heap_pkg.PriorityItem[struct{}], heapOpsPerMeasure)
		for j := range heapOpsPerMeasure {
			selected[j] = items[rng.IntN(n)]
		}
		start = time.Now()
		for _, item := range selected {
			pq.UpdatePriority(item, rng.IntN(10_000_000))
		}
		pqUpdateNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Heap 各操作耗时",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Push / Pop / UpdatePriority 均摊耗时 vs 规模（每规模 %d 次取均值）", heapOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "堆规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "MinHeap Push", Data: minPushNs},
				{Name: "MinHeap Pop", Data: minPopNs},
				{Name: "PriorityQueue Push", Data: pqPushNs},
				{Name: "PriorityQueue Pop", Data: pqPopNs},
				{Name: "PriorityQueue UpdatePriority", Data: pqUpdateNs},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "heap/", Title: "Heap 各操作耗时",
		Desc: "MinHeap / PriorityQueue Push、Pop、UpdatePriority 均摊 ns/op（1万~50万）",
		Path: "/api/ds/heap", DataHandler: handleDsHeap,
	})
}
