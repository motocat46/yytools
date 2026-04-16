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

	"github.com/motocat46/yytools/pkg/infra/timer/delayqueue"
)

// dqTestItem 是用于 demo 的简单 DelayQueue 元素。
type dqTestItem struct {
	expireAtMs int64
}

func (d *dqTestItem) ExpireAt() int64 { return d.expireAtMs }

const dqOpsPerMeasure = 1000

// nowMs 返回当前 Unix 毫秒时间戳。
func nowMs() int64 {
	return time.Now().UnixMilli()
}

// fillDQ 向 DelayQueue 填充 n 个未来到期（1小时后）的元素，返回各元素引用。
func fillDQ(dq *delayqueue.DelayQueue[*dqTestItem], n int) []*dqTestItem {
	items := make([]*dqTestItem, n)
	future := nowMs() + 3_600_000 // 1小时后
	for i := range n {
		item := &dqTestItem{expireAtMs: future + int64(i)}
		items[i] = item
		dq.Offer(item)
	}
	return items
}

func handleInfraDelayQueue(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：Offer 吞吐 vs 队列规模（维持规模稳定：先 TryPoll 出一个过期元素，再 Offer 一个）
	sizes := []int{1000, 10000, 100000, 500000, 1000000}
	xLabels1 := make([]string, len(sizes))
	offerNs := make([]int64, len(sizes))

	for i, n := range sizes {
		if n >= 10000 {
			xLabels1[i] = fmt.Sprintf("%d万", n/10000)
		} else {
			xLabels1[i] = fmt.Sprintf("%d", n)
		}
		dq := delayqueue.New[*dqTestItem](nowMs)
		fillDQ(dq, n)
		future := nowMs() + 3_600_000

		// 测量：稳态 Offer（队列大小保持 n，每次 Offer 一个新未来元素）
		start := time.Now()
		for k := range dqOpsPerMeasure {
			dq.Offer(&dqTestItem{expireAtMs: future + int64(k)})
		}
		offerNs[i] = time.Since(start).Nanoseconds() / dqOpsPerMeasure
	}

	// Chart 2：TryPoll 吞吐 — 命中（已到期）vs 未命中（未到期）
	const tryPollN = 100000
	const tryPollOps = 100_000

	// 未命中：队列中全是未来元素
	dqMiss := delayqueue.New[*dqTestItem](nowMs)
	fillDQ(dqMiss, tryPollN)
	start := time.Now()
	for range tryPollOps {
		dqMiss.TryPoll()
	}
	missNs := time.Since(start).Nanoseconds() / tryPollOps

	// 命中：队列中全是已到期元素（expireAt = 0 远小于当前时间）
	dqHit := delayqueue.New[*dqTestItem](nowMs)
	hitItems := make([]*dqTestItem, tryPollN)
	for i := range tryPollN {
		hitItems[i] = &dqTestItem{expireAtMs: int64(i)} // 远过去
		dqHit.Offer(hitItems[i])
	}
	// 取出后重新入队，维持规模（TryPoll 取出后立即 Offer 回去）
	start = time.Now()
	for range tryPollOps {
		if item, ok := dqHit.TryPoll(); ok {
			item.expireAtMs = 0 // 保持已到期
			dqHit.Offer(item)
		}
	}
	hitNs := time.Since(start).Nanoseconds() / tryPollOps

	// 空队列：边界情况
	dqEmpty := delayqueue.New[*dqTestItem](nowMs)
	start = time.Now()
	for range tryPollOps {
		dqEmpty.TryPoll()
	}
	emptyNs := time.Since(start).Nanoseconds() / tryPollOps

	xLabels2 := []string{"未命中（未到期）", "命中（已到期，O(log n)）", "空队列（O(1) 直接返回）"}
	tryPollData := []int64{missNs, hitNs, emptyNs}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "DelayQueue 吞吐",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Offer 均摊耗时 vs 队列规模（O(log n)，每规模 %d 次均值）", dqOpsPerMeasure),
				XAxis:     xLabels1,
				XAxisName: "队列规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Offer（堆插入，O(log n)）", Data: offerNs},
				},
			},
			{
				Type:      "bar",
				Title:     fmt.Sprintf("TryPoll 三种场景耗时对比（队列规模 %d，%d 次均值）", tryPollN, tryPollOps),
				XAxis:     xLabels2,
				XAxisName: "场景",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "TryPoll 耗时", Data: tryPollData},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/infra", SubPkg: "timer/delayqueue/", Title: "DelayQueue 吞吐",
		Desc: "Offer O(log n) 耗时 vs 规模（1K~100万）；TryPoll 命中/未命中/空队列三场景对比",
		Path: "/api/infra/delayqueue", DataHandler: handleInfraDelayQueue,
	})
}
