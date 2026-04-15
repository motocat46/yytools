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
	"sort"
	"sync"
	"sync/atomic"
	"time"

	sf "github.com/motocat46/yytools/pkg/algorithms/idgen/snowflake"
)

func handleAlgoSnowflake(w http.ResponseWriter, _ *http.Request) {
	g, _ := sf.NewGenerator(0)

	// Chart 1：吞吐 vs goroutine 数
	// 每个并发级别生成 totalIDs 个，测量耗时计算万IDs/秒
	goroutineCounts := []int{1, 2, 4, 8, 16, 32, 64}
	xLabels1 := make([]string, len(goroutineCounts))
	throughputs := make([]int64, len(goroutineCounts))

	const totalIDs = 100_000
	for i, n := range goroutineCounts {
		xLabels1[i] = fmt.Sprintf("%d", n)
		collected := make([]int64, totalIDs)
		var idx atomic.Int64
		var wg sync.WaitGroup
		start := time.Now()
		for range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					pos := idx.Add(1) - 1
					if pos >= totalIDs {
						return
					}
					collected[pos] = g.NewID()
				}
			}()
		}
		wg.Wait()
		_ = collected
		elapsed := time.Since(start)
		// 万IDs/秒
		throughputs[i] = int64(totalIDs) * 1_000_000 / elapsed.Microseconds() / 10000
	}

	// Chart 2：每毫秒实际生成量（64 goroutine 突发 40960 个 ID）
	// 展示序号上限（默认 4096/ms）与实际利用率
	const (
		chart2Goroutines = 64
		chart2Total      = 40_960 // = 10 × 4096，预期约跨 10ms
	)
	ids2 := make([]int64, chart2Total)
	var idx2 atomic.Int64
	var wg2 sync.WaitGroup
	for range chart2Goroutines {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			for {
				pos := idx2.Add(1) - 1
				if pos >= chart2Total {
					return
				}
				ids2[pos] = g.NewID()
			}
		}()
	}
	wg2.Wait()

	// 按毫秒时间戳分组计数
	msBuckets := make(map[int64]int64)
	for _, id := range ids2 {
		parts := g.ParseID(id)
		msBuckets[parts.Timestamp]++
	}
	msKeys := make([]int64, 0, len(msBuckets))
	for k := range msBuckets {
		msKeys = append(msKeys, k)
	}
	sort.Slice(msKeys, func(a, b int) bool { return msKeys[a] < msKeys[b] })

	xLabels2 := make([]string, len(msKeys))
	msCountData := make([]int64, len(msKeys))
	maxSeqLine := make([]int64, len(msKeys)) // 水平参考线：4096
	for i, ms := range msKeys {
		xLabels2[i] = fmt.Sprintf("+%dms", ms-msKeys[0])
		msCountData[i] = msBuckets[ms]
		maxSeqLine[i] = sf.MaxSequence + 1 // 4096
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Snowflake ID 并发吞吐",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("单节点吞吐 vs goroutine 数（每级别生成 %d 个 ID）", totalIDs),
				XAxis:     xLabels1,
				XAxisName: "goroutine 数",
				YAxisName: "万 IDs/秒",
				Series: []chartSeries{
					{Name: "吞吐（万 IDs/秒）", Data: throughputs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("每毫秒实际生成量（%d goroutines，共 %d 个 ID，理论上限 %d/ms）", chart2Goroutines, chart2Total, sf.MaxSequence+1),
				XAxis:     xLabels2,
				XAxisName: "时间偏移",
				YAxisName: "IDs/ms",
				Series: []chartSeries{
					{Name: "实际生成量", Data: msCountData},
					{Name: fmt.Sprintf("序号上限（%d）", sf.MaxSequence+1), Data: maxSeqLine},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "idgen/snowflake/", Title: "Snowflake ID 并发吞吐",
		Desc: "CAS 无锁吞吐 vs goroutine 数（10万IDs）；64 goroutine 突发时每毫秒实际生成量 vs 序号上限（4096）",
		Path: "/api/algo/snowflake", DataHandler: handleAlgoSnowflake,
	})
}
