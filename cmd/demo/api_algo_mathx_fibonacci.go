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

	"github.com/motocat46/yytools/pkg/algorithms/mathx"
)

// fibSink 防止编译器优化掉 FibN / Calculate 调用
var fibSink int64

func handleAlgoMathxFibonacci(w http.ResponseWriter, _ *http.Request) {
	// 测试的 FibN 下标（0-indexed，有效范围 0..92 for int64）
	// 对应 Calculate 参数为 n+1（1-indexed）
	nValues := []int{1, 5, 10, 20, 30, 50, 70, 92}

	// FibN 和暖缓存：O(1) 路径，百万次取均值
	const fibFastOps = 1_000_000
	// 冷启路径：每次新建实例，成本包含内存分配 + O(n) 计算
	const fibColdOps = 50_000

	xLabels := make([]string, len(nValues))
	fibNNs := make([]int64, len(nValues))
	coldNs := make([]int64, len(nValues))
	warmNs := make([]int64, len(nValues))

	// 预热一个 Fibonacci 实例到最大下标，用于暖缓存测量
	warmed := mathx.NewFibMem[int64]()
	warmed.Calculate(int64(mathx.FibNMax[int64]() + 1)) // 计算到 F(92)，全量缓存

	for i, n := range nValues {
		xLabels[i] = fmt.Sprintf("n=%d", n)

		// FibN：O(1) 静态查找表（数组下标访问）
		start := time.Now()
		for range fibFastOps {
			fibSink += mathx.FibN[int64](n)
		}
		fibNNs[i] = time.Since(start).Nanoseconds() / fibFastOps

		// Calculate 冷启：每次新建实例，首次计算 O(n)（分配 + 计算 n-1 个新值）
		start = time.Now()
		for range fibColdOps {
			f := mathx.NewFibMem[int64]()
			fibSink += f.Calculate(int64(n + 1))
		}
		coldNs[i] = time.Since(start).Nanoseconds() / fibColdOps

		// Calculate 暖缓存：已预热，重复查询 O(1)（切片下标访问）
		start = time.Now()
		for range fibFastOps {
			fibSink += warmed.Calculate(int64(n + 1))
		}
		warmNs[i] = time.Since(start).Nanoseconds() / fibFastOps
	}

	// Chart 2：各整数类型支持的最大 FibN 下标
	typeLabels := []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64"}
	typeMaxIdx := []int64{
		int64(mathx.FibNMax[int8]()),
		int64(mathx.FibNMax[int16]()),
		int64(mathx.FibNMax[int32]()),
		int64(mathx.FibNMax[int64]()),
		int64(mathx.FibNMax[uint8]()),
		int64(mathx.FibNMax[uint16]()),
		int64(mathx.FibNMax[uint32]()),
		int64(mathx.FibNMax[uint64]()),
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Fibonacci 三路实现耗时",
		Charts: []chartData{
			{
				Type: "line",
				Title: fmt.Sprintf(
					"FibN O(1) vs Calculate 冷启 O(n) vs Calculate 暖缓存 O(1)（快速路径 %d 万次均值；冷启 %d 次均值）",
					fibFastOps/10000, fibColdOps,
				),
				XAxis:     xLabels,
				XAxisName: "斐波那契下标 n（FibN 0-indexed）",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "FibN（静态查找表，O(1)）", Data: fibNNs},
					{Name: "Calculate 暖缓存（已预热，O(1)）", Data: warmNs},
					{Name: "Calculate 冷启（新实例，O(n)）", Data: coldNs},
				},
			},
			{
				Type:      "bar",
				Title:     "各整数类型支持的最大 FibN 下标（unsigned 比同宽 signed 多容纳 1 个）",
				XAxis:     typeLabels,
				YAxisName: "FibNMax（最大有效下标）",
				Series: []chartSeries{
					{Name: "FibNMax", Data: typeMaxIdx},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/", Title: "Fibonacci 三路耗时",
		Desc: "FibN O(1)静态查找表 vs Calculate 冷启O(n) vs Calculate 暖缓存O(1)；各整数类型最大下标对比",
		Path: "/api/algo/mathx/fibonacci", DataHandler: handleAlgoMathxFibonacci,
	})
}
