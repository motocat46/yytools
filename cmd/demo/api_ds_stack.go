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
// 创建日期:2026/4/16
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	stack_pkg "github.com/motocat46/yytools/pkg/ds/stack"
)

// measureStackPush 在 n 元素稳态下（预分配 2n 容量）测量 heapOpsPerMeasure 次 Push，返回均摊 ns/op。
func measureStackPush(n int) int64 {
	s := stack_pkg.NewStackWithSize[int](n * 2)
	for i := range n {
		s.Push(i)
	}
	start := time.Now()
	for j := range heapOpsPerMeasure {
		s.Push(n + j)
	}
	return time.Since(start).Nanoseconds() / heapOpsPerMeasure
}

// measureStackPop 在 n 元素稳态下（预分配 2n 容量）测量 heapOpsPerMeasure 次 Pop，返回均摊 ns/op。
func measureStackPop(n int) int64 {
	s := stack_pkg.NewStackWithSize[int](n * 2)
	for i := range n {
		s.Push(i)
	}
	start := time.Now()
	for range heapOpsPerMeasure {
		s.Pop()
	}
	return time.Since(start).Nanoseconds() / heapOpsPerMeasure
}

const (
	stackSimN    = 1000 // 入栈元素数（随后等量出栈）
	stackSimStep = 10   // 每隔 stackSimStep 次操作采样一次
)

func handleDsStackOps(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 50000, 100000, 200000, 500000}
	xLabels := make([]string, len(sizes))
	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)
	}

	pushNs := make([]int64, len(sizes))
	popNs := make([]int64, len(sizes))

	for i, n := range sizes {
		pushNs[i] = measureStackPush(n)
		popNs[i] = measureStackPop(n)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Stack 稳态吞吐",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Push / Pop 均摊耗时 vs 规模（预分配容量，无扩缩容，每规模 %d 次取均值）", heapOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "栈规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "Push", Data: pushNs},
				{Name: "Pop", Data: popNs},
			},
		}},
	})
}

func handleDsStackResize(w http.ResponseWriter, _ *http.Request) {
	// 初始容量 DEFAULT_STACK_SIZE（16），入栈 stackSimN 个后等量出栈，全程记录 Len 与 Capacity
	s := stack_pkg.NewStack[int]()

	totalOps := stackSimN * 2
	numSamples := totalOps/stackSimStep + 1 // 含 op=0 的初始快照
	xLabels := make([]string, numSamples)
	lenData := make([]int64, numSamples)
	capData := make([]int64, numSamples)

	// op=0：初始快照
	idx := 0
	xLabels[idx] = "0"
	lenData[idx] = int64(s.Length())
	capData[idx] = int64(cap(s.Items))
	idx++

	// 入栈阶段
	for op := 1; op <= stackSimN; op++ {
		s.Push(op)
		if op%stackSimStep == 0 {
			xLabels[idx] = fmt.Sprintf("%d", op)
			lenData[idx] = int64(s.Length())
			capData[idx] = int64(cap(s.Items))
			idx++
		}
	}

	// 出栈阶段（操作序号从 stackSimN+1 开始，连续编号）
	for op := 1; op <= stackSimN; op++ {
		s.Pop()
		if op%stackSimStep == 0 {
			xLabels[idx] = fmt.Sprintf("%d", stackSimN+op)
			lenData[idx] = int64(s.Length())
			capData[idx] = int64(cap(s.Items))
			idx++
		}
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Stack 扩缩容行为",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Len vs Capacity（入栈 %d 个后出栈 %d 个，初始容量 %d）", stackSimN, stackSimN, stack_pkg.DEFAULT_STACK_SIZE),
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
		Pkg: "pkg/ds", SubPkg: "stack/", Title: "Stack 稳态吞吐",
		Desc: "Push / Pop 均摊 ns/op（预分配，无扩缩容，1万~50万）",
		Path: "/api/ds/stack/ops", DataHandler: handleDsStackOps,
	})
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "stack/", Title: "Stack 扩缩容行为",
		Desc: "入栈 1000 后出栈 1000，Len vs Capacity 阶梯变化（初始容量 16）",
		Path: "/api/ds/stack/resize", DataHandler: handleDsStackResize,
	})
}
