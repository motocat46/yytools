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
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/motocat46/yytools/pkg/algorithms/mathx"
)

const gcdOpsPerMeasure = 100_000

// gcdPair 是一对非负整数，用于 GCD 计算输入。
type gcdPair struct{ x, y int64 }

// makeGCDPairs 在 [1, maxVal] 范围内生成 n 对随机正整数，使用固定种子保证可复现。
func makeGCDPairs(n int, maxVal int64) []gcdPair {
	rng := rand.New(rand.NewPCG(42, 0))
	pairs := make([]gcdPair, n)
	for i := range n {
		pairs[i] = gcdPair{
			x: rng.Int64N(maxVal) + 1,
			y: rng.Int64N(maxVal) + 1,
		}
	}
	return pairs
}

func handleAlgoMathxGcd(w http.ResponseWriter, _ *http.Request) {
	// 不同输入量级下，GcdR（递归）vs GcdI（迭代）耗时对比
	magnitudes := []struct {
		label  string
		maxVal int64
	}{
		{"≤100", 100},
		{"≤10^4", 10_000},
		{"≤10^6", 1_000_000},
		{"≤10^8", 100_000_000},
		{"≤10^12", 1_000_000_000_000},
	}

	xLabels := make([]string, len(magnitudes))
	gcdRNs := make([]int64, len(magnitudes))
	gcdINs := make([]int64, len(magnitudes))

	for i, mag := range magnitudes {
		xLabels[i] = mag.label
		pairs := makeGCDPairs(gcdOpsPerMeasure, mag.maxVal)

		start := time.Now()
		for _, p := range pairs {
			mathx.GcdR(p.x, p.y)
		}
		gcdRNs[i] = time.Since(start).Nanoseconds() / gcdOpsPerMeasure

		start = time.Now()
		for _, p := range pairs {
			mathx.GcdI(p.x, p.y)
		}
		gcdINs[i] = time.Since(start).Nanoseconds() / gcdOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "GCD 递归 vs 迭代",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("GcdR（递归）vs GcdI（迭代）耗时 vs 输入量级（每组 %d 对，固定种子）", gcdOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "输入值范围",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "GcdR（欧几里得递归，O(log n) 栈深）", Data: gcdRNs},
				{Name: "GcdI（欧几里得迭代，O(1) 空间）", Data: gcdINs},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/", Title: "GCD 递归 vs 迭代",
		Desc: "GcdR（递归）vs GcdI（迭代）耗时 vs 输入量级（5档，每档10万对）",
		Path: "/api/algo/mathx/gcd", DataHandler: handleAlgoMathxGcd,
	})
}
