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

package probability_distribution_test

import (
	"fmt"

	pd "github.com/motocat46/yytools/pkg/algorithms/mathx/probability_distribution"
)

// ExampleCalcIndexByWeight 展示一次性权重随机查找——候选集动态变化时使用。
func ExampleCalcIndexByWeight() {
	// 权重 30:50:20，总权重 100
	weights := []int32{30, 50, 20}
	idx := pd.CalcIndexByWeight(weights, 100)
	fmt.Println(idx >= 0 && idx < 3) // 结果一定在有效范围内
	// Output:
	// true
}

// ExampleCalcKeyByWeight 展示按 map 键的权重随机返回 key。
func ExampleCalcKeyByWeight() {
	weights := map[string]int32{"SSR": 5, "SR": 25, "R": 70}
	key := pd.CalcKeyByWeight(weights, 100)
	_, ok := weights[key]
	fmt.Println(ok) // 返回的 key 一定在 map 中
	// Output:
	// true
}

// ExampleNormalMethod 展示 NormalMethod：构建 O(n)，生成 O(log n)。
// 适合候选集固定、中等频次生成的场景。
func ExampleNormalMethod_Generate() {
	weights := []int32{30, 50, 20}
	gen := pd.NewNormalMethod(weights)
	idx := gen.Generate()
	fmt.Println(idx >= 0 && idx < 3)
	// Output:
	// true
}

// ExampleVoseAliasMethod 展示 VoseAliasMethod：构建 O(n)，生成 O(1)。
// 适合候选集固定、高频生成（如游戏帧循环）的场景。
func ExampleVoseAliasMethod_Generate() {
	weights := []int32{30, 50, 20}
	gen := pd.NewVoseAliasMethod(weights)
	idx := gen.Generate()
	fmt.Println(idx >= 0 && idx < 3)
	// Output:
	// true
}

// ExampleProbFactory 展示工厂函数统一入口。
func ExampleProbFactory() {
	weights := []int32{10, 60, 30}

	// 低频场景用 Normal（实现更简单）
	gen1 := pd.ProbFactory(pd.Normal, weights)
	fmt.Println(gen1.Generate() >= 0)

	// 高频场景用 VoseAlias（O(1) 生成）
	gen2 := pd.ProbFactory(pd.VoseAlias, weights)
	fmt.Println(gen2.Generate() >= 0)
	// Output:
	// true
	// true
}
