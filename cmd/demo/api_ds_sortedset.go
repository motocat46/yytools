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

// lbEntry 是有序切片实现的排行榜条目。
type lbEntry struct {
	key   int
	score float64
}

func handleDsSortedSetLeaderboard(w http.ResponseWriter, _ *http.Request) {
	// 混合负载：80% UpdateScore + 20% GetRangeByRankDesc（Top-10）
	// 衡量总吞吐（万 ops/秒）vs 排行榜规模
	const (
		lbOps    = 10_000 // 每规模操作次数
		lbTopK   = 10
		lbUpdate = 8 // 每 10 次操作中 8 次更新、2 次查询
	)

	sizes := []int{10000, 30000, 50000, 100000, 200000}
	xLabels := make([]string, len(sizes))
	ssThroughput := make([]int64, len(sizes))
	sliceThroughput := make([]int64, len(sizes))

	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)
		rng := rand.New(rand.NewPCG(42, 0))

		// 预填充 SortedSet
		ss := ss_pkg.NewSortedSet[int, struct{}]()
		scores := make([]float64, n)
		for j := range n {
			scores[j] = float64(rng.IntN(10_000_000))
			ss.Insert(ss_pkg.NewNodeData(j, scores[j], struct{}{}))
		}

		// SortedSet 混合负载
		rng2 := rand.New(rand.NewPCG(99, 0))
		start := time.Now()
		for op := range lbOps {
			if op%10 < lbUpdate {
				key := rng2.IntN(n)
				ss.UpdateScore(key, float64(rng2.IntN(10_000_000)))
			} else {
				ss.GetRangeByRankDesc(1, lbTopK)
			}
		}
		elapsed := time.Since(start)
		ssThroughput[i] = int64(lbOps) * 1_000_000 / elapsed.Microseconds() / 10_000

		// 有序切片混合负载（每次更新后 sort.Slice 重排）
		sl := make([]lbEntry, n)
		for j := range n {
			sl[j] = lbEntry{j, scores[j]}
		}
		sort.Slice(sl, func(a, b int) bool { return sl[a].score > sl[b].score })

		rng3 := rand.New(rand.NewPCG(99, 0))
		start = time.Now()
		for op := range lbOps {
			if op%10 < lbUpdate {
				key := rng3.IntN(n)
				newScore := float64(rng3.IntN(10_000_000))
				for k := range sl {
					if sl[k].key == key {
						sl[k].score = newScore
						break
					}
				}
				sort.Slice(sl, func(a, b int) bool { return sl[a].score > sl[b].score })
			} else {
				_ = sl[:lbTopK]
			}
		}
		elapsed = time.Since(start)
		sliceThroughput[i] = int64(lbOps) * 1_000_000 / elapsed.Microseconds() / 10_000
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "SortedSet 排行榜场景模拟",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("排行榜混合吞吐 vs 规模（%d%%更新分数 + %d%%查 Top-%d，共 %d 次操作）", lbUpdate*10, (10-lbUpdate)*10, lbTopK, lbOps),
			XAxis:     xLabels,
			XAxisName: "排行榜规模",
			YAxisName: "万 ops/秒",
			Series: []chartSeries{
				{Name: fmt.Sprintf("SortedSet（O(log n) 更新 + O(%d) 查询）", lbTopK), Data: ssThroughput},
				{Name: "有序切片（每次更新后 sort.Slice 重排，O(n log n)）", Data: sliceThroughput},
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
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "sorted_set/", Title: "排行榜场景模拟",
		Desc: "80% UpdateScore + 20% Top-10 查询混合吞吐（万ops/秒）vs 规模；SortedSet vs 有序切片重排",
		Path: "/api/ds/sortedset/leaderboard", DataHandler: handleDsSortedSetLeaderboard,
	})
}
