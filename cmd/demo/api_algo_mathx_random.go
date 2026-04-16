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

	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
)

const randTrials = 1_000_000

func handleAlgoMathxRandom(w http.ResponseWriter, _ *http.Request) {
	rng := random.NewRand(42)

	// Chart 1：RandInt[int64] 分布均匀性（[-50, 49] 共 100 个桶，100 万次）
	const lo64, hi64 = int64(-50), int64(49)
	const buckets = int(hi64 - lo64 + 1) // 100
	hitCount := make([]int64, buckets)
	for range randTrials {
		v := random.RandIntWith(rng, lo64, hi64)
		hitCount[v-lo64]++
	}
	xLabels1 := make([]string, buckets)
	for i := range buckets {
		xLabels1[i] = fmt.Sprintf("%d", lo64+int64(i))
	}
	expected := int64(randTrials / buckets)

	// Chart 2：RandInt 各宽度 vs 标准库 rand.Int64N 性能（每种 100 万次）
	const perfOps = 1_000_000
	rng2 := random.NewRand(99)
	stdRng := rand.New(rand.NewPCG(99, 0))

	start := time.Now()
	for range perfOps {
		random.RandIntWith(rng2, int8(-100), int8(100))
	}
	int8Ns := time.Since(start).Nanoseconds() / perfOps

	start = time.Now()
	for range perfOps {
		random.RandIntWith(rng2, int32(-1_000_000), int32(1_000_000))
	}
	int32Ns := time.Since(start).Nanoseconds() / perfOps

	start = time.Now()
	for range perfOps {
		random.RandIntWith(rng2, int64(-1_000_000_000), int64(1_000_000_000))
	}
	int64Ns := time.Since(start).Nanoseconds() / perfOps

	start = time.Now()
	for range perfOps {
		random.RandIntWith(rng2, uint64(0), uint64(1_000_000_000_000))
	}
	uint64Ns := time.Since(start).Nanoseconds() / perfOps

	// 标准库对照：rand.Int64N（只支持 int64，正数范围）
	start = time.Now()
	for range perfOps {
		stdRng.Int64N(2_000_000_001) // 等价 [0, 2G]，与 RandInt int64 范围同量级
	}
	stdInt64Ns := time.Since(start).Nanoseconds() / perfOps

	xLabels2 := []string{"RandInt[int8]", "RandInt[int32]", "RandInt[int64]", "RandInt[uint64]", "rand.Int64N（标准库）"}
	perfData := []int64{int8Ns, int32Ns, int64Ns, uint64Ns, stdInt64Ns}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "RandInt 均匀性与性能",
		Charts: []chartData{
			{
				Type:      "bar",
				Title:     fmt.Sprintf("RandInt[int64] 分布均匀性（范围 [%d,%d]，%d 万次，期望 %d 次/桶）", lo64, hi64, randTrials/10000, expected),
				XAxis:     xLabels1,
				XAxisName: "采样值",
				YAxisName: "命中次数",
				Series: []chartSeries{
					{Name: "实际命中次数", Data: hitCount},
				},
			},
			{
				Type:      "bar",
				Title:     fmt.Sprintf("RandInt 各宽度 vs 标准库 rand.Int64N 耗时（%d 万次均值）", perfOps/10000),
				XAxis:     xLabels2,
				XAxisName: "类型",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "均摊耗时", Data: perfData},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/random/", Title: "RandInt 均匀性与性能",
		Desc: "RandInt[int64] 分布均匀性（100万次）；各宽度类型耗时 vs 标准库 rand.Int64N",
		Path: "/api/algo/mathx/random", DataHandler: handleAlgoMathxRandom,
	})
}
