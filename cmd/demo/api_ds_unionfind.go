// Copyright [yangyuan]
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
//
// 作者:  yangyuan
// 创建日期:2026/4/20
package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	uf "github.com/motocat46/yytools/pkg/ds/unionfind"
)

const ufOpsPerMeasure = 10_000

func handleDsUnionFind(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{100, 1_000, 10_000, 100_000, 1_000_000}
	xLabels := make([]string, len(sizes))
	unionNs := make([]int64, len(sizes))
	findNs := make([]int64, len(sizes))
	connNs := make([]int64, len(sizes))

	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d", n)
		rng := rand.New(rand.NewPCG(42, 0))

		// 预填充：注册所有节点，建链式结构（路径压缩有工作可做）
		u := uf.New[int]()
		for j := range n {
			u.Find(j)
		}
		// 链式 Union：形成一条链，让路径压缩产生效果
		for j := range n - 1 {
			u.Union(j, j+1)
		}

		// 测量 Find
		start := time.Now()
		for range ufOpsPerMeasure {
			u.Find(rng.IntN(n))
		}
		findNs[i] = time.Since(start).Nanoseconds() / ufOpsPerMeasure

		// 测量 Connected
		start = time.Now()
		for range ufOpsPerMeasure {
			u.Connected(rng.IntN(n), rng.IntN(n))
		}
		connNs[i] = time.Since(start).Nanoseconds() / ufOpsPerMeasure

		// 重建独立节点，测量 Union（随机合并）
		u2 := uf.New[int]()
		for j := range n {
			u2.Find(j)
		}
		start = time.Now()
		for range ufOpsPerMeasure {
			a, b := rng.IntN(n), rng.IntN(n)
			u2.Union(a, b)
		}
		unionNs[i] = time.Since(start).Nanoseconds() / ufOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "UnionFind 性能",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Union/Find/Connected 耗时 vs 元素规模（每规模 %d 次）", ufOpsPerMeasure),
				XAxis:     xLabels,
				XAxisName: "元素规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Union（O(α) 均摊）", Data: unionNs},
					{Name: "Find（O(α) 均摊，含路径压缩）", Data: findNs},
					{Name: "Connected（O(α) 均摊）", Data: connNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "unionfind/", Title: "UnionFind 性能",
		Desc:        "Union/Find/Connected 耗时 vs 元素规模；验证路径压缩 + 按大小合并的 O(α) 均摊性质",
		Path:        "/api/ds/unionfind",
		DataHandler: handleDsUnionFind,
	})
}
