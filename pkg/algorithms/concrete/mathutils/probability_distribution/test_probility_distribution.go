// Package test.

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

// 作者:  yangyuan
// 创建日期:2023/6/14
package probability_distribution

import (
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "time"
)

func CalcIndexWithWeightTest(x []int) {
	println("遍历概率生成方法:")
	// 一百万次
	N := 1000000
	totalWeight := 0
	for _, one := range x {
		totalWeight += one
	}
	
	tmp := make(map[int]int)
	for i := 0; i < N; i++ {
		index := CalcIndexByWeight(x, totalWeight)
		tmp[index]++
	}
	for i, _ := range x {
		fmt.Printf("%.3f\t", float64(tmp[i])/float64(N))
	}
	println()
}

func CalcKeyWithWeightTest(x []int) {
	println("在map中遍历概率的生成方法:")
	// 一百万次
	N := 1000000
	totalWeight := 0
	for _, one := range x {
		totalWeight += one
	}
	
	xMap := make(map[int]int, len(x))
	for i, v := range x {
		xMap[i] = v
	}
	
	tmp := make(map[int]int)
	for i := 0; i < N; i++ {
		index := CalcKeyByWeight(xMap, totalWeight)
		tmp[index]++
	}
	for i, _ := range x {
		fmt.Printf("%.3f\t", float64(tmp[i])/float64(N))
	}
	println()
}


func NormalMethodTest(x []int) {
	// 创建一个包含map字段的数据对象
	println("普通概率生成方法:")
	// 一百万次
	N := 1000000
	
	method := ProbFactory(Normal, x)
	tmp := make(map[int]int)
	for i := 0; i < N; i++ {
		index := method.Generate()
		tmp[index]++
	}
	for i, _ := range x {
		fmt.Printf("%.3f\t", float64(tmp[i])/float64(N))
	}
	println()
}

func VoseAliasMethodTest(x []int) {
	println("vose的别名方法:")
	// 一百万次
	N := 1000000
	method := ProbFactory(VoseAlias, x)
	tmp := make(map[int]int)
	for i := 0; i < N; i++ {
		index := method.Generate()
		tmp[index]++
	}
	for i, _ := range x {
		fmt.Printf("%.3f\t", float64(tmp[i])/float64(N))
	}
	println()
}

func DynamicWeightsTest(x []int) {
	println("动态计算权重:")
	weights := make(map[interface{}]int, len(x))
	for k, v := range x {
		weights[k] = v
	}
	method := NewDynamicWeights(weights)
	totalWeights := method.TtlWght
	tmp := make(map[int]int)
	for method.CanGenerate() {
		index := method.Generate().(int)
		tmp[index]++
	}
	for i, _ := range x {
		fmt.Printf("%.3f\t", float64(tmp[i])/float64(totalWeights))
	}
	println()
}

var ProbabilityDistribution_handlers = []func(num []int){
    CalcIndexWithWeightTest,
    CalcKeyWithWeightTest,
    NormalMethodTest,
    VoseAliasMethodTest,
    DynamicWeightsTest,
}

func ProbabilityDistributionTest(num int) {
	println("概率分布测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		handlerLength := len(ProbabilityDistribution_handlers)
		var x []int
		count := 10
		for a := 0; a < count; a++ {
			r := random.RandInt(1, 100)
			x = append(x, r)
		}
		println("理论概率分布：")
		total := 0
		for _, v := range x {
			total += v
		}
		for _, v := range x {
			fmt.Printf("%.3f\t", float64(v)/float64(total))
		}
		println()
		for k := 0; k < handlerLength; k++ {
			handler := ProbabilityDistribution_handlers[k]
			handler(x)
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("概率分布测试完毕...")
}