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

	lru_pkg "github.com/motocat46/yytools/pkg/ds/lru"
)

const (
	lruKeySpace = 1000  // 总 key 空间
	lruHotKeys  = 200   // 热点 key 数量（Pareto 20%）
	lruHotProb  = 80    // 热点访问比例（80/20 规则）
	lruAccessN  = 100_000
	lruWarmupN  = 10_000 // 预热次数，不计入命中统计
)

// lruNextKey 按 80/20 分布生成下一个访问 key。
func lruNextKey(rng *rand.Rand) int {
	if rng.IntN(100) < lruHotProb {
		return rng.IntN(lruHotKeys) // 热点：[0, 200)
	}
	return lruHotKeys + rng.IntN(lruKeySpace-lruHotKeys) // 冷点：[200, 1000)
}

// simulateHitRate 模拟 lruWarmupN+lruAccessN 次访问，返回稳态命中率（整数百分比）。
func simulateHitRate(cacheSize int) int64 {
	cache := lru_pkg.New[int, int](cacheSize, 0)
	rng := rand.New(rand.NewPCG(42, 0))

	// 预热：让缓存达到稳态，不计入命中统计
	for range lruWarmupN {
		key := lruNextKey(rng)
		if _, ok := cache.Get(key); !ok {
			cache.Put(key, key)
		}
	}

	// 统计阶段
	hits := 0
	for range lruAccessN {
		key := lruNextKey(rng)
		if _, ok := cache.Get(key); ok {
			hits++
		} else {
			cache.Put(key, key)
		}
	}
	return int64(hits * 100 / lruAccessN)
}

func handleDsLRUHitRate(w http.ResponseWriter, _ *http.Request) {
	cacheSizes := []int{10, 25, 50, 100, 150, 200, 250, 350, 500, 750, 1000}
	xLabels := make([]string, len(cacheSizes))
	hitRates := make([]int64, len(cacheSizes))

	for i, size := range cacheSizes {
		xLabels[i] = fmt.Sprintf("%d", size)
		hitRates[i] = simulateHitRate(size)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "LRU 命中率 vs 缓存容量",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("命中率 vs 缓存容量（key 空间=%d，热点=%d个占%d%%访问，%d 万次模拟）", lruKeySpace, lruHotKeys, lruHotProb, lruAccessN/10000),
			XAxis:     xLabels,
			XAxisName: "缓存容量",
			YAxisName: "命中率(%)",
			Series: []chartSeries{
				{Name: "命中率", Data: hitRates},
			},
		}},
	})
}

// ---- Get/Put 均摊耗时 ----

func handleDsLRUOps(w http.ResponseWriter, _ *http.Request) {
	sizes := []int{10000, 50000, 100000, 200000, 500000}
	xLabels := make([]string, len(sizes))
	getNs := make([]int64, len(sizes))
	putNs := make([]int64, len(sizes))

	for i, n := range sizes {
		xLabels[i] = fmt.Sprintf("%d万", n/10000)

		// 预填充到容量 n
		cache := lru_pkg.New[int, int](n, 0)
		for j := range n {
			cache.Put(j, j)
		}

		// 测量 Get（全命中，不触发 Put 路径）
		rng := rand.New(rand.NewPCG(99, 0))
		start := time.Now()
		for range heapOpsPerMeasure {
			key := rng.IntN(n)
			cache.Get(key)
		}
		getNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure

		// 测量 Put（全部为已有 key 的更新，容量不变）
		start = time.Now()
		for range heapOpsPerMeasure {
			key := rng.IntN(n)
			cache.Put(key, key*2)
		}
		putNs[i] = time.Since(start).Nanoseconds() / heapOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "LRU Get/Put 耗时",
		Charts: []chartData{{
			Type:      "line",
			Title:     fmt.Sprintf("Get / Put 均摊耗时 vs 缓存规模（每规模 %d 次取均值，含锁开销）", heapOpsPerMeasure),
			XAxis:     xLabels,
			XAxisName: "缓存规模",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "Get（全命中）", Data: getNs},
				{Name: "Put（更新已有 key）", Data: putNs},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "lru/", Title: "LRU 命中率 vs 缓存容量",
		Desc: "80/20 访问模型，10万次模拟，展示肘部曲线与热点容量选型依据",
		Path: "/api/ds/lru/hitrate", DataHandler: handleDsLRUHitRate,
	})
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "lru/", Title: "LRU Get/Put 耗时",
		Desc: "Get / Put 均摊 ns/op（含 RWMutex 开销，1万~50万）",
		Path: "/api/ds/lru/ops", DataHandler: handleDsLRUOps,
	})
}
