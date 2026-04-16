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

	"github.com/motocat46/yytools/pkg/infra/timer/timingwheel"
)

const twOpsPerMeasure = 1000

// measureTWAfterFunc 预填充 n 个 timer，稳态下测量 twOpsPerMeasure 次 (Cancel+AfterFunc) 对，返回均摊 ns/op。
// 维持集合规模不变：取消一个，新增一个。
func measureTWAfterFunc(tw *timingwheel.TimingWheel, n int) int64 {
	timers := make([]*timingwheel.Timer, n)
	for i := range n {
		t, _ := tw.AfterFunc(time.Duration(i+1)*time.Hour, func() {})
		timers[i] = t
	}
	start := time.Now()
	for k := range twOpsPerMeasure {
		timers[k%n].Cancel()
		newT, _ := tw.AfterFunc(time.Duration(k%n+1)*time.Hour, func() {})
		timers[k%n] = newT
	}
	return time.Since(start).Nanoseconds() / twOpsPerMeasure
}

// measureTWConcurrent 用 parallelism 个 goroutine 并发执行 Cancel+AfterFunc，
// 预填充 n=100K timers，返回均摊 ns/op。
func measureTWConcurrent(tw *timingwheel.TimingWheel, parallelism int) int64 {
	const fixedN = 100_000
	const concOps = 10_000

	timers := make([]*timingwheel.Timer, fixedN)
	for i := range fixedN {
		t, _ := tw.AfterFunc(time.Duration(i+1)*time.Hour, func() {})
		timers[i] = t
	}

	done := make(chan int64, parallelism)
	start := time.Now()
	for p := range parallelism {
		go func(pIdx int) {
			perP := concOps / parallelism
			for k := range perP {
				slot := (pIdx*perP + k) % fixedN
				timers[slot].Cancel()
				newT, _ := tw.AfterFunc(time.Duration(slot+1)*time.Hour, func() {})
				timers[slot] = newT
			}
			done <- 0
		}(p)
	}
	for range parallelism {
		<-done
	}
	elapsed := time.Since(start).Nanoseconds()
	return elapsed / concOps
}

func handleInfraTimingWheel(w http.ResponseWriter, _ *http.Request) {
	tw := timingwheel.New()
	tw.Start()
	defer tw.Stop()

	// Chart 1：AfterFunc+Cancel 稳态耗时 vs 规模（单调用方）
	sizes := []int{1000, 10000, 100000, 1000000}
	xLabels1 := make([]string, len(sizes))
	opsNs := make([]int64, len(sizes))
	for i, n := range sizes {
		if n >= 10000 {
			xLabels1[i] = fmt.Sprintf("%d万", n/10000)
		} else {
			xLabels1[i] = fmt.Sprintf("%d", n)
		}
		opsNs[i] = measureTWAfterFunc(tw, n)
	}

	// Chart 2：并发 AfterFunc+Cancel 吞吐 vs 并发度（固定 100K timers）
	parallelisms := []int{1, 4, 16, 64}
	xLabels2 := make([]string, len(parallelisms))
	concNs := make([]int64, len(parallelisms))
	for i, p := range parallelisms {
		xLabels2[i] = fmt.Sprintf("p=%d", p)
		concNs[i] = measureTWConcurrent(tw, p)
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "TimingWheel 吞吐",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Cancel+AfterFunc 均摊耗时 vs 已注册 timer 规模（O(1)，每规模 %d 次均值）", twOpsPerMeasure),
				XAxis:     xLabels1,
				XAxisName: "timer 规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Cancel+AfterFunc（稳态维持规模 n）", Data: opsNs},
				},
			},
			{
				Type:      "bar",
				Title:     "并发 AfterFunc+Cancel 均摊耗时 vs 并发度（固定 10万 timers，总 1万次操作）",
				XAxis:     xLabels2,
				XAxisName: "并发度",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Cancel+AfterFunc（多调用方并发）", Data: concNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/infra", SubPkg: "timer/timingwheel/", Title: "TimingWheel 吞吐",
		Desc: "Cancel+AfterFunc O(1) 稳态耗时 vs 规模（1K~100万）；并发度 1/4/16/64 吞吐对比",
		Path: "/api/infra/timingwheel", DataHandler: handleInfraTimingWheel,
	})
}
