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
	"sort"
	"time"

	ss_pkg "github.com/motocat46/yytools/pkg/ds/sorted_set"
)

// ---- 公共辅助 ----

const ssOpsPerMeasure = 1000

// newFilledSortedSet 创建并预填充 n 个元素的有序集合，使用固定种子保证可复现。
// Key 为 [0, n)，Score 为随机整数。
func newFilledSortedSet(n int) *ss_pkg.SortedSet[int, struct{}] {
	rng := rand.New(rand.NewPCG(42, 0))
	ss := ss_pkg.NewSortedSet[int, struct{}]()
	for i := range n {
		node := ss_pkg.NewNodeData(i, float64(rng.IntN(10_000_000)), struct{}{})
		ss.Insert(node)
	}
	return ss
}

// ---- 各操作计时 ----

// measureSSInsert 向已有 n 个元素的集合插入 ssOpsPerMeasure 个元素，返回均摊 ns/op。
func measureSSInsert(n int) int64 {
	ss := newFilledSortedSet(n)
	rng := rand.New(rand.NewPCG(99, 0))
	start := time.Now()
	for j := range ssOpsPerMeasure {
		node := ss_pkg.NewNodeData(n+j, float64(rng.IntN(10_000_000)), struct{}{})
		ss.Insert(node)
	}
	return time.Since(start).Nanoseconds() / ssOpsPerMeasure
}

// measureSSGetByRank 对 n 个元素的集合做 ssOpsPerMeasure 次随机排名查询，返回均摊 ns/op。
func measureSSGetByRank(n int) int64 {
	ss := newFilledSortedSet(n)
	rng := rand.New(rand.NewPCG(99, 0))
	start := time.Now()
	for range ssOpsPerMeasure {
		rank := rng.IntN(n) + 1
		ss.GetByRank(rank)
	}
	return time.Since(start).Nanoseconds() / ssOpsPerMeasure
}

// measureSSUpdateScore 对 n 个元素的集合做 ssOpsPerMeasure 次随机分数更新，返回均摊 ns/op。
// 键在计时前预收集，避免 GetByRank 调用混入计时。
func measureSSUpdateScore(n int) int64 {
	ss := newFilledSortedSet(n)
	rng := rand.New(rand.NewPCG(77, 0))
	keys := make([]int, ssOpsPerMeasure)
	for j := range ssOpsPerMeasure {
		keys[j] = ss.GetByRank(rng.IntN(n) + 1).Key
	}
	rng2 := rand.New(rand.NewPCG(55, 0))
	start := time.Now()
	for _, k := range keys {
		ss.UpdateScore(k, float64(rng2.IntN(10_000_000)))
	}
	return time.Since(start).Nanoseconds() / ssOpsPerMeasure
}

// measureSSDelete 从 n+ssOpsPerMeasure 个元素的集合删除 ssOpsPerMeasure 个，返回均摊 ns/op。
// 键在计时前按排名 1..ssOpsPerMeasure 预收集，保证不重复。
func measureSSDelete(n int) int64 {
	ss := newFilledSortedSet(n + ssOpsPerMeasure)
	keys := make([]int, ssOpsPerMeasure)
	for j := range ssOpsPerMeasure {
		keys[j] = ss.GetByRank(j + 1).Key
	}
	start := time.Now()
	for _, k := range keys {
		ss.Delete(k)
	}
	return time.Since(start).Nanoseconds() / ssOpsPerMeasure
}

// ---- Handlers ----

func handleDsSortedSetOps(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 20000, 50000, 100000, 200000}
	xLabels := make([]string, len(sizes))
	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)
	}

	insertNs := make([]int64, len(sizes))
	getByRankNs := make([]int64, len(sizes))
	updateScoreNs := make([]int64, len(sizes))
	deleteNs := make([]int64, len(sizes))

	for i, n := range sizes {
		insertNs[i] = measureSSInsert(n)
		getByRankNs[i] = measureSSGetByRank(n)
		updateScoreNs[i] = measureSSUpdateScore(n)
		deleteNs[i] = measureSSDelete(n)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "SortedSet 各操作耗时",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("各操作均摊耗时 vs 规模（每规模 %d 次取均值）", ssOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "集合规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "Insert", Data: insertNs},
				{Name: "GetByRank", Data: getByRankNs},
				{Name: "UpdateScore", Data: updateScoreNs},
				{Name: "Delete", Data: deleteNs},
			},
		}},
	})
}

// insertIntoSortedSlice 将 score 插入有序 float64 切片的正确位置，维护升序，O(n)。
func insertIntoSortedSlice(s *[]float64, score float64) {
	i := sort.SearchFloat64s(*s, score)
	*s = append(*s, 0)
	copy((*s)[i+1:], (*s)[i:])
	(*s)[i] = score
}

func handleDsSortedSetVsSlice(w http.ResponseWriter, _ *http.Request) {
	// 规模需足够大才能体现 O(n) vs O(log n) 的差距；
	// 100k 时有序切片 copy 均摊约 2000 ns/op，跳表约 600 ns/op，差距约 3x；
	// 趋势更重要：切片 ns/op 线性增长，跳表近似平坦。
	sizes := []int{1000, 5000, 10000, 30000, 60000, 100000}
	xLabels := make([]string, len(sizes))
	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d", n)
	}

	ssNsPerOp := make([]int64, len(sizes))
	sliceNsPerOp := make([]int64, len(sizes))

	for i, n := range sizes {
		rng := rand.New(rand.NewPCG(42, 0))
		// 两个数据结构使用相同输入，保证比较公平
		scores := make([]float64, n)
		for j := range n {
			scores[j] = float64(rng.IntN(10_000_000))
		}

		// SortedSet：插入 n 个元素
		ss := ss_pkg.NewSortedSet[int, struct{}]()
		start := time.Now()
		for j, score := range scores {
			node := ss_pkg.NewNodeData(j, score, struct{}{})
			ss.Insert(node)
		}
		ssNsPerOp[i] = time.Since(start).Nanoseconds() / int64(n)

		// 有序切片：sort.Search 定位 + copy 移位，O(n)
		sl := make([]float64, 0, n)
		start = time.Now()
		for _, score := range scores {
			insertIntoSortedSlice(&sl, score)
		}
		sliceNsPerOp[i] = time.Since(start).Nanoseconds() / int64(n)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "跳表 vs 有序切片 — Insert 性能对比",
		Charts: []chartData{{
			Type:      "line",
			Title:     "批量插入均摊耗时（O(log n) vs O(n)）",
			XAxis:     xLabels,
			XAxisName: "集合规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "SortedSet（跳表，O(log n)）", Data: ssNsPerOp},
				{Name: "有序切片（sort.Search+copy，O(n)）", Data: sliceNsPerOp},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "sorted_set/", Title: "SortedSet 各操作耗时",
		Desc: "Insert / GetByRank / UpdateScore / Delete 均摊 ns/op（1万~20万）",
		Path: "/api/ds/sortedset/ops", DataHandler: handleDsSortedSetOps,
	})
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "sorted_set/", Title: "跳表 vs 有序切片",
		Desc: "批量 Insert 均摊 ns/op 对比：O(log n) vs O(n)（500~20000 元素）",
		Path: "/api/ds/sortedset/vs", DataHandler: handleDsSortedSetVsSlice,
	})
}
