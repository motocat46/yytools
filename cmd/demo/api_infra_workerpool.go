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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	wp "github.com/motocat46/yytools/pkg/infra/concurrency/workerpool"
)

const wpTotalTasks = 100_000

// measureWPThroughput 用 submitters 个 goroutine 并发向 pool 提交 wpTotalTasks 个无操作任务，
// 等待全部完成后返回万tasks/秒。
func measureWPThroughput(pool *wp.WorkerPool, submitters int) int64 {
	perSubmitter := wpTotalTasks / submitters
	ctx := context.Background()

	var wg sync.WaitGroup
	start := time.Now()
	for range submitters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range perSubmitter {
				pool.Submit(ctx, func() {}) //nolint:errcheck
			}
		}()
	}
	wg.Wait()
	pool.Wait()
	elapsed := time.Since(start)

	// 万 tasks/秒
	return int64(wpTotalTasks) * 1_000_000 / elapsed.Microseconds() / 10_000
}

func handleInfraWorkerPool(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：吞吐 vs worker 数（单提交方，无操作任务）
	workerCounts := []int{1, 2, 4, 8, 16, 32, 64}
	xLabels1 := make([]string, len(workerCounts))
	throughput1 := make([]int64, len(workerCounts))

	for i, n := range workerCounts {
		xLabels1[i] = fmt.Sprintf("%d", n)
		pool := wp.NewWorkerPool(n, wpTotalTasks)
		throughput1[i] = measureWPThroughput(pool, 1)
		pool.Close()
	}

	// Chart 2：RWMutex vs Mutex — Submit 吞吐 vs 并发提交方数（固定 16 workers）
	const fixedWorkers = 16
	submitterCounts := []int{1, 2, 4, 8, 16, 32, 64}
	xLabels2 := make([]string, len(submitterCounts))
	rwThroughput := make([]int64, len(submitterCounts))
	mutexThroughput := make([]int64, len(submitterCounts))

	for i, s := range submitterCounts {
		xLabels2[i] = fmt.Sprintf("%d", s)

		rwPool := wp.NewWorkerPool(fixedWorkers, wpTotalTasks)
		rwThroughput[i] = measureWPThroughput(rwPool, s)
		rwPool.Close()

		mutexPool := wp.NewWorkerPoolMutex(fixedWorkers, wpTotalTasks)
		mutexThroughput[i] = measureWPThroughput(mutexPool, s)
		mutexPool.Close()
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "WorkerPool 并发吞吐",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("吞吐 vs worker 数（单提交方，%d 个无操作任务）", wpTotalTasks),
				XAxis:     xLabels1,
				XAxisName: "worker 数",
				YAxisName: "万 tasks/秒",
				Series: []chartSeries{
					{Name: "RWMutex Pool 吞吐", Data: throughput1},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("RWMutex vs Mutex — Submit 吞吐 vs 并发提交方数（固定 %d workers，%d 个任务）", fixedWorkers, wpTotalTasks),
				XAxis:     xLabels2,
				XAxisName: "并发提交方数",
				YAxisName: "万 tasks/秒",
				Series: []chartSeries{
					{Name: "RWMutex（Submit 并发友好）", Data: rwThroughput},
					{Name: "Mutex（Submit 串行）", Data: mutexThroughput},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/infra", SubPkg: "concurrency/workerpool/", Title: "WorkerPool 并发吞吐",
		Desc: "吞吐 vs worker 数（10万无操作任务）；RWMutex vs Mutex Submit 吞吐 vs 并发提交方数",
		Path: "/api/infra/workerpool", DataHandler: handleInfraWorkerPool,
	})
}
