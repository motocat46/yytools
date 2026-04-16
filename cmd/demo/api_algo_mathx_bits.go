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
	mathbits "math/bits"
	"net/http"
	"time"

	ybits "github.com/motocat46/yytools/pkg/algorithms/mathx/bits"
)

const bitsOpsPerMeasure = 1_000_000

// bitsSink 防止编译器消除纯函数调用（CountingBits/OnesCount64 返回值无副作用）。
var bitsSink int

// makeBitPatterns 生成各种位密度的测试序列：0-bit、稀疏、中等、稠密、全 1。
// 每种密度返回 n 个 uint64 样本。
func makeBitPatterns(density, n int) []uint64 {
	out := make([]uint64, n)
	for i := range n {
		switch density {
		case 0: // 0-bit：全零
			out[i] = 0
		case 1: // 1-bit：2 的幂次（只有 1 个 1）
			out[i] = 1 << uint(i%64)
		case 8: // 8-bit：低 8 位全 1
			out[i] = uint64((i % 255) + 1)
		case 32: // 32-bit：低 32 位随机模式
			out[i] = uint64(i*0x9e3779b9) & 0xFFFFFFFF
		case 48: // 48-bit：低 48 位随机模式
			out[i] = uint64(i*0x9e3779b9) & 0xFFFFFFFFFFFF
		case 64: // 64-bit：全 1
			out[i] = ^uint64(0)
		}
	}
	return out
}

func handleAlgoMathxBits(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：CountingBits vs math/bits.OnesCount64 — 不同位密度下的性能对比
	// Kernighan 算法 O(k)（k = 1-bit 数量），OnesCount64 用 POPCNT 指令 O(1)
	densities := []struct {
		label   string
		density int
	}{
		{"0-bit（全零）", 0},
		{"1-bit（2的幂）", 1},
		{"8-bit", 8},
		{"32-bit", 32},
		{"48-bit", 48},
		{"64-bit（全1）", 64},
	}

	xLabels := make([]string, len(densities))
	countingBitsNs := make([]int64, len(densities))
	onesCountNs := make([]int64, len(densities))

	const sampleN = 256 // 不同样本循环使用，避免测同一个值
	for i, d := range densities {
		xLabels[i] = d.label
		patterns := makeBitPatterns(d.density, sampleN)

		start := time.Now()
		for k := range bitsOpsPerMeasure {
			bitsSink += ybits.CountingBits(patterns[k%sampleN])
		}
		countingBitsNs[i] = time.Since(start).Nanoseconds() / bitsOpsPerMeasure

		start = time.Now()
		for k := range bitsOpsPerMeasure {
			bitsSink += mathbits.OnesCount64(patterns[k%sampleN])
		}
		onesCountNs[i] = time.Since(start).Nanoseconds() / bitsOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Bits 操作性能",
		Charts: []chartData{{
			Type:      "bar",
			Title:     fmt.Sprintf("CountingBits（Kernighan O(k)）vs math/bits.OnesCount64（POPCNT O(1)）（%d 万次均值）", bitsOpsPerMeasure/10000),
			XAxis:     xLabels,
			XAxisName: "位密度",
			YAxisName: "ns/op",
			Series: []chartSeries{
				{Name: "CountingBits（O(k)，稀疏时更快）", Data: countingBitsNs},
				{Name: "math/bits.OnesCount64（POPCNT，恒定）", Data: onesCountNs},
			},
		}},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "mathx/bits/", Title: "Bits 操作性能",
		Desc: "CountingBits（Kernighan O(k)）vs math/bits.OnesCount64（POPCNT）— 不同位密度下的自适应性",
		Path: "/api/algo/mathx/bits", DataHandler: handleAlgoMathxBits,
	})
}
