// Package probability_distribution.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package probability_distribution

import (
	"math"
	"testing"
)

const sampleCount = 10000     // 采样次数
const toleranceRate = 0.10    // 允许误差率 ±10%

// checkDistribution 验证采样计数与期望比例的偏差在容差范围内
func checkDistribution(t *testing.T, name string, counts []int, weights []int, samples int) {
	t.Helper()
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	for i, count := range counts {
		expected := float64(samples) * float64(weights[i]) / float64(totalWeight)
		deviation := math.Abs(float64(count)-expected) / expected
		if deviation > toleranceRate {
			t.Errorf("%s: 下标 %d 命中 %d 次，期望约 %.0f 次（偏差 %.1f%%，允许 %.0f%%）",
				name, i, count, expected, deviation*100, toleranceRate*100)
		}
	}
}

// ---------- CalcIndexByWeight ----------

func TestCalcIndexByWeight(t *testing.T) {
	t.Run("单个权重永远返回0", func(t *testing.T) {
		weights := []int{5}
		for i := 0; i < 100; i++ {
			idx := CalcIndexByWeight(weights, 5)
			if idx != 0 {
				t.Errorf("单权重应该永远返回 0，实际 %d", idx)
			}
		}
	})

	t.Run("等权重均匀分布", func(t *testing.T) {
		weights := []int{1, 1, 1, 1}
		total := 4
		counts := make([]int, len(weights))
		for i := 0; i < sampleCount; i++ {
			idx := CalcIndexByWeight(weights, total)
			if idx < 0 || idx >= len(weights) {
				t.Fatalf("返回下标越界: %d", idx)
			}
			counts[idx]++
		}
		checkDistribution(t, "等权重", counts, weights, sampleCount)
	})

	t.Run("不等权重分布符合比例", func(t *testing.T) {
		weights := []int{1, 2, 7}
		total := 10
		counts := make([]int, len(weights))
		for i := 0; i < sampleCount; i++ {
			idx := CalcIndexByWeight(weights, total)
			if idx < 0 || idx >= len(weights) {
				t.Fatalf("返回下标越界: %d", idx)
			}
			counts[idx]++
		}
		checkDistribution(t, "不等权重[1,2,7]", counts, weights, sampleCount)
	})

	t.Run("返回下标在合法范围内", func(t *testing.T) {
		weights := []int{3, 5, 2}
		total := 10
		for i := 0; i < 1000; i++ {
			idx := CalcIndexByWeight(weights, total)
			if idx < 0 || idx >= len(weights) {
				t.Errorf("下标越界: %d，期望 [0, %d)", idx, len(weights))
			}
		}
	})
}

// ---------- CalcKeyByWeight ----------

func TestCalcKeyByWeight(t *testing.T) {
	t.Run("返回key在合法集合内", func(t *testing.T) {
		weightMap := map[string]int{
			"A": 1,
			"B": 3,
			"C": 6,
		}
		total := 10
		for i := 0; i < 1000; i++ {
			key := CalcKeyByWeight(weightMap, total)
			if _, ok := weightMap[key]; !ok {
				t.Errorf("返回了不存在的 key: %s", key)
			}
		}
	})

	t.Run("单key永远返回该key", func(t *testing.T) {
		weightMap := map[string]int{"only": 5}
		for i := 0; i < 100; i++ {
			key := CalcKeyByWeight(weightMap, 5)
			if key != "only" {
				t.Errorf("单 key 应该永远返回 'only'，实际 '%s'", key)
			}
		}
	})
}

// ---------- NormalMethod ----------

func TestNormalMethod(t *testing.T) {
	t.Run("Generate返回合法下标", func(t *testing.T) {
		weights := []int{3, 5, 2}
		nm := NewNormalMethod(weights)
		for i := 0; i < 1000; i++ {
			idx := nm.Generate()
			if idx < 0 || idx >= len(weights) {
				t.Errorf("Generate 返回越界下标: %d", idx)
			}
		}
	})

	t.Run("分布符合权重比例", func(t *testing.T) {
		weights := []int{1, 2, 7}
		nm := NewNormalMethod(weights)
		counts := make([]int, len(weights))
		for i := 0; i < sampleCount; i++ {
			idx := nm.Generate()
			counts[idx]++
		}
		checkDistribution(t, "NormalMethod[1,2,7]", counts, weights, sampleCount)
	})

	t.Run("全零权重等概率随机", func(t *testing.T) {
		weights := []int{0, 0, 0, 0}
		nm := NewNormalMethod(weights)
		counts := make([]int, len(weights))
		for i := 0; i < sampleCount; i++ {
			idx := nm.Generate()
			if idx < 0 || idx >= len(weights) {
				t.Fatalf("下标越界: %d", idx)
			}
			counts[idx]++
		}
		// 等概率，期望每个下标各出现约 sampleCount/4 次
		equalWeights := []int{1, 1, 1, 1}
		checkDistribution(t, "NormalMethod全零权重", counts, equalWeights, sampleCount)
	})
}

// ---------- VoseAliasMethod ----------

func TestVoseAliasMethod(t *testing.T) {
	t.Run("Generate返回合法下标", func(t *testing.T) {
		weights := []int{3, 5, 2}
		vam := NewVoseAliasMethod(weights)
		for i := 0; i < 1000; i++ {
			idx := vam.Generate()
			if idx < 0 || idx >= len(weights) {
				t.Errorf("Generate 返回越界下标: %d", idx)
			}
		}
	})

	t.Run("分布符合权重比例", func(t *testing.T) {
		weights := []int{1, 2, 7}
		vam := NewVoseAliasMethod(weights)
		counts := make([]int, len(weights))
		for i := 0; i < sampleCount; i++ {
			idx := vam.Generate()
			counts[idx]++
		}
		checkDistribution(t, "VoseAliasMethod[1,2,7]", counts, weights, sampleCount)
	})

	t.Run("单个权重永远返回0", func(t *testing.T) {
		weights := []int{5}
		vam := NewVoseAliasMethod(weights)
		for i := 0; i < 100; i++ {
			idx := vam.Generate()
			if idx != 0 {
				t.Errorf("单权重应永远返回 0，实际 %d", idx)
			}
		}
	})
}

// ---------- NormalMethod 与 VoseAliasMethod 趋同验证 ----------

func TestNormalAndVoseConsistency(t *testing.T) {
	weights := []int{1, 3, 6}
	nm := NewNormalMethod(weights)
	vam := NewVoseAliasMethod(weights)

	nmCounts := make([]int, len(weights))
	vamCounts := make([]int, len(weights))

	for i := 0; i < sampleCount; i++ {
		nmCounts[nm.Generate()]++
		vamCounts[vam.Generate()]++
	}

	// 两种算法都应与理论值接近
	checkDistribution(t, "Normal一致性", nmCounts, weights, sampleCount)
	checkDistribution(t, "VoseAlias一致性", vamCounts, weights, sampleCount)
}

// ---------- ProbFactory ----------

func TestProbFactory(t *testing.T) {
	weights := []int{2, 3, 5}

	t.Run("Normal类型可正常构造和生成", func(t *testing.T) {
		dist := ProbFactory(Normal, weights)
		if dist == nil {
			t.Fatal("ProbFactory(Normal) 返回了 nil")
		}
		idx := dist.Generate()
		if idx < 0 || idx >= len(weights) {
			t.Errorf("Generate 返回越界下标: %d", idx)
		}
	})

	t.Run("VoseAlias类型可正常构造和生成", func(t *testing.T) {
		dist := ProbFactory(VoseAlias, weights)
		if dist == nil {
			t.Fatal("ProbFactory(VoseAlias) 返回了 nil")
		}
		idx := dist.Generate()
		if idx < 0 || idx >= len(weights) {
			t.Errorf("Generate 返回越界下标: %d", idx)
		}
	})

	t.Run("未知类型触发panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("未知类型应该触发 panic")
			}
		}()
		ProbFactory(MethodType(99), weights)
	})
}

// ---------- DynamicWeights ----------

func TestDynamicWeights_CanGenerate(t *testing.T) {
	t.Run("初始状态可以生成", func(t *testing.T) {
		dw := NewDynamicWeights(map[string]int{"a": 2, "b": 3})
		if !dw.CanGenerate() {
			t.Error("初始总权重 > 0，CanGenerate 应返回 true")
		}
	})

	t.Run("权重耗尽后不可生成", func(t *testing.T) {
		dw := NewDynamicWeights(map[string]int{"a": 1})
		dw.Generate() // 消耗唯一权重
		if dw.CanGenerate() {
			t.Error("权重耗尽后 CanGenerate 应返回 false")
		}
	})
}

func TestDynamicWeights_Generate(t *testing.T) {
	t.Run("生成结果是合法key", func(t *testing.T) {
		weights := map[string]int{"x": 3, "y": 5, "z": 2}
		dw := NewDynamicWeights(weights)
		validKeys := map[string]struct{}{"x": {}, "y": {}, "z": {}}
		for dw.CanGenerate() {
			got := dw.Generate()
			if _, ok := validKeys[got]; !ok {
				t.Errorf("Generate 返回了非法 key: %v", got)
			}
		}
	})

	t.Run("总采样次数等于总权重", func(t *testing.T) {
		// 每次 Generate 权重减 1，所以可以生成 totalWeight 次
		dw := NewDynamicWeights(map[string]int{"a": 3, "b": 2})
		count := 0
		for dw.CanGenerate() {
			dw.Generate()
			count++
		}
		if count != 5 {
			t.Errorf("总权重为 5，期望生成 5 次，实际 %d 次", count)
		}
	})

	t.Run("每个key最多被选中其初始权重次", func(t *testing.T) {
		initial := map[string]int{"a": 2, "b": 3}
		dw := NewDynamicWeights(initial)
		hitCount := map[string]int{}
		for dw.CanGenerate() {
			k := dw.Generate()
			hitCount[k]++
		}
		if hitCount["a"] != 2 {
			t.Errorf("'a' 初始权重为 2，期望被选中 2 次，实际 %d 次", hitCount["a"])
		}
		if hitCount["b"] != 3 {
			t.Errorf("'b' 初始权重为 3，期望被选中 3 次，实际 %d 次", hitCount["b"])
		}
	})
}

func TestDynamicWeights_SetReduce(t *testing.T) {
	t.Run("SetReduce修改每次减少量", func(t *testing.T) {
		// 总权重 6，reduce=2，则可生成 3 次
		dw := NewDynamicWeightsWithReduce(map[string]int{"a": 6}, 2)
		count := 0
		for dw.CanGenerate() {
			dw.Generate()
			count++
		}
		if count != 3 {
			t.Errorf("总权重 6，reduce=2，期望生成 3 次，实际 %d 次", count)
		}
	})

	t.Run("运行中修改reduce", func(t *testing.T) {
		dw := NewDynamicWeights(map[string]int{"a": 4})
		dw.Generate() // reduce=1，总权重变为 3
		dw.SetReduce(3)
		dw.Generate() // reduce=3，总权重变为 0
		if dw.CanGenerate() {
			t.Error("修改 reduce 后权重应耗尽，CanGenerate 应返回 false")
		}
	})
}

func TestNewDynamicWeightsWithReduce(t *testing.T) {
	t.Run("自定义reduce构造", func(t *testing.T) {
		dw := NewDynamicWeightsWithReduce(map[string]int{"k": 10}, 5)
		if dw.Reduce != 5 {
			t.Errorf("期望 Reduce=5，实际 %d", dw.Reduce)
		}
		if dw.TtlWght != 10 {
			t.Errorf("期望 TtlWght=10，实际 %d", dw.TtlWght)
		}
	})
}
