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
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/motocat46/yytools/pkg/infra/safeexec"
)

func handleInfraSafeExec(w http.ResponseWriter, _ *http.Request) {
	// 无 panic 路径：大量重复，ns 级精度
	const noPanicOps = 1_000_000

	// 裸函数调用基准：最小可能开销（循环 + 间接调用）
	noop := func() {}
	start := time.Now()
	for range noPanicOps {
		noop()
	}
	rawNs := time.Since(start).Nanoseconds() / noPanicOps

	// SafeExec 无 panic：defer+recover 设置/拆除 + nil 检查
	start = time.Now()
	for range noPanicOps {
		safeexec.SafeExec("bench", noop)
	}
	safeExecNoPanicNs := time.Since(start).Nanoseconds() / noPanicOps

	// SafeExecErr 无 panic（返回 error 变体）
	noopErr := func() error { return nil }
	start = time.Now()
	for range noPanicOps {
		safeexec.SafeExecErr("bench", noopErr) //nolint:errcheck
	}
	safeExecErrNoPanicNs := time.Since(start).Nanoseconds() / noPanicOps

	// panic 路径：少量重复，μs 级；压制 slog 输出避免 I/O 干扰测量
	// （SafeExec 的 panic 路径写 slog.Error + debug.Stack()，I/O 会主导结果）
	const panicOps = 1_000
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	panicFunc := func() { panic("bench") }
	start = time.Now()
	for range panicOps {
		safeexec.SafeExec("bench", panicFunc)
	}
	safeExecPanicNs := time.Since(start).Nanoseconds() / panicOps

	slog.SetDefault(orig)

	xLabels := []string{
		"裸调用（基准）",
		"SafeExec 无 panic",
		"SafeExecErr 无 panic",
		fmt.Sprintf("SafeExec 触发 panic\n（含 debug.Stack，%d次均值）", panicOps),
	}
	data := []int64{rawNs, safeExecNoPanicNs, safeExecErrNoPanicNs, safeExecPanicNs}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "SafeExec 开销",
		Charts: []chartData{{
			Type:  "bar",
			Title: fmt.Sprintf("三条路径耗时对比（无 panic 路径 %d 万次均值；panic 路径 %d 次均值）", noPanicOps/10000, panicOps),
			XAxis: xLabels,
			Series: []chartSeries{
				{Name: "ns/op（注意：panic 路径为 ns，实际为 μs 级）", Data: data},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/infra", SubPkg: "safeexec/", Title: "SafeExec 开销",
		Desc: "无 panic 路径：裸调用 vs SafeExec vs SafeExecErr（100万次）；panic 路径含 debug.Stack()（1000次）",
		Path: "/api/infra/safeexec", DataHandler: handleInfraSafeExec,
	})
}
