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

	"github.com/motocat46/yytools/pkg/algorithms/mathx/overflow"
)

const overflowOpsPerMeasure = 1_000_000

func handleAlgoMathxOverflow(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：AddInt/SubInt/MulInt/DivInt 各操作耗时（混合：50% 不溢出 + 50% 溢出边界）
	// 目的：展示溢出检测开销，以及各操作的实现复杂度差异

	// 预生成操作数，避免循环内随机生成影响计时
	const n = 512
	type pair struct{ a, b int64 }
	// 不溢出对：普通数值范围内
	normalPairs := make([]pair, n)
	for i := range n {
		normalPairs[i] = pair{int64(i + 1), int64(i + 2)}
	}
	// 溢出对：MulInt 溢出（大数相乘）
	mulOverflowPairs := make([]pair, n)
	for i := range n {
		mulOverflowPairs[i] = pair{int64(1<<32 + i), int64(1<<32 + i + 1)}
	}
	// DivInt 溢出唯一情形：MinInt64 / -1
	const minInt64 = int64(-1 << 63)

	// overflowSink 防止编译器消除纯函数调用（AddInt/SubInt/MulInt/DivInt 无外部副作用）。
	var overflowSinkVal int64
	var overflowSinkOvf bool

	start := time.Now()
	for k := range overflowOpsPerMeasure {
		overflowSinkVal, overflowSinkOvf = overflow.AddInt(normalPairs[k%n].a, normalPairs[k%n].b)
	}
	addNs := time.Since(start).Nanoseconds() / overflowOpsPerMeasure

	start = time.Now()
	for k := range overflowOpsPerMeasure {
		overflowSinkVal, overflowSinkOvf = overflow.SubInt(normalPairs[k%n].a, normalPairs[k%n].b)
	}
	subNs := time.Since(start).Nanoseconds() / overflowOpsPerMeasure

	start = time.Now()
	for k := range overflowOpsPerMeasure {
		// 交替正常 / 溢出以测量混合场景（两个分支各执行一半）
		if k%2 == 0 {
			overflowSinkVal, overflowSinkOvf = overflow.MulInt(normalPairs[k%n].a, normalPairs[k%n].b)
		} else {
			overflowSinkVal, overflowSinkOvf = overflow.MulInt(mulOverflowPairs[k%n].a, mulOverflowPairs[k%n].b)
		}
	}
	mulNs := time.Since(start).Nanoseconds() / overflowOpsPerMeasure

	start = time.Now()
	for k := range overflowOpsPerMeasure {
		if k%200 == 0 {
			overflowSinkVal, overflowSinkOvf = overflow.DivInt(minInt64, int64(-1)) // 触发唯一溢出情形
		} else {
			overflowSinkVal, overflowSinkOvf = overflow.DivInt(normalPairs[k%n].a, normalPairs[k%n].b+1)
		}
	}
	divNs := time.Since(start).Nanoseconds() / overflowOpsPerMeasure
	_, _ = overflowSinkVal, overflowSinkOvf

	// 裸运算（无溢出检测）作为基准
	start = time.Now()
	sum := int64(0)
	for k := range overflowOpsPerMeasure {
		sum += normalPairs[k%n].a + normalPairs[k%n].b // 防止优化消除
	}
	_ = sum
	rawAddNs := time.Since(start).Nanoseconds() / overflowOpsPerMeasure

	start = time.Now()
	prod := int64(1)
	for k := range overflowOpsPerMeasure {
		prod *= normalPairs[k%n].a // 防止优化消除
	}
	_ = prod
	rawMulNs := time.Since(start).Nanoseconds() / overflowOpsPerMeasure

	xLabels := []string{"AddInt", "SubInt", "MulInt（混合）", "DivInt（混合）", "裸加法（对照）", "裸乘法（对照）"}
	data := []int64{addNs, subNs, mulNs, divNs, rawAddNs, rawMulNs}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Overflow 检测耗时",
		Charts: []chartData{{
			Type:      "bar",
			Title:     fmt.Sprintf("AddInt/SubInt/MulInt/DivInt 均摊耗时（%d 万次，混合溢出/正常场景）", overflowOpsPerMeasure/10000),
			XAxis:     xLabels,
			XAxisName: "操作",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "耗时（含溢出检测）", Data: data},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/overflow/", Title: "Overflow 检测耗时",
		Desc: "AddInt/SubInt/MulInt/DivInt 均摊 ns/op（100万次混合场景）vs 裸运算基准",
		Path: "/api/algo/mathx/overflow", DataHandler: handleAlgoMathxOverflow,
	})
}
