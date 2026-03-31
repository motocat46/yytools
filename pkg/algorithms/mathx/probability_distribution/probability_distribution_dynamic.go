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

// 作者:  yangyuan
// 创建日期:2023/7/26
package probability_distribution

import (
	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/base"
)

/*
*	提供动态的和概率相关的工具方法
*   动态指：元素的权重值会随着算法的进行改变(减少)
*   区别于 probability_distribution.go里的方法
 */

// IDynamicProbDistr 是动态权重概率分布生成器的接口。
// 与 IProbDist 的区别：每次 Generate 后被命中 key 的权重会减少，直到耗尽。
type IDynamicProbDistr[K comparable, T base.Signed] interface {
	CanGenerate() bool
	Generate() K
	SetReduce(reduce T)
}

// DynamicWeights 可以得到一个周期的完整分布。
// K 为权重 key 类型（comparable），T 为权重值类型（Signed）。
type DynamicWeights[K comparable, T base.Signed] struct {
	Weights map[K]T // 权重map
	TtlWght T        // 总权重
	Reduce  T        // 权重减少的值
}

// NewDynamicWeights 创建初始权重固定（每次命中减少 1）的动态分布器，语义见 NewDynamicWeightsWithReduce。
func NewDynamicWeights[K comparable, T base.Signed](weights map[K]T) *DynamicWeights[K, T] {
	return NewDynamicWeightsWithReduce[K, T](weights, 1)
}

// NewDynamicWeightsWithReduce 创建动态分布器，每次命中后将对应 key 的权重减少 reduce。
// weights 所有权重必须 > 0，reduce 必须 > 0；某 key 权重降至 <= 0 时自动从候选集中移除。
func NewDynamicWeightsWithReduce[K comparable, T base.Signed](weights map[K]T, reduce T) *DynamicWeights[K, T] {
	assert.Assert(len(weights) > 0)
	assert.Assert(reduce > 0)
	total := T(0)
	for _, w := range weights {
		total += w
		assert.Assert(w > 0)
	}
	assert.Assert(total > 0, "总权重需要大于0：", total)
	return &DynamicWeights[K, T]{
		Weights: weights,
		TtlWght: total,
		Reduce:  reduce,
	}
}

// NewDynamicWeightsProgressive 创建初始为空、支持后续动态增加权重的 DW。
// 与 NewDynamicWeights 不同，此构造器不要求初始权重非空。
func NewDynamicWeightsProgressive[K comparable, T base.Signed]() *DynamicWeights[K, T] {
	return &DynamicWeights[K, T]{
		Weights: make(map[K]T),
		TtlWght: 0,
		Reduce:  1,
	}
}

// CanGenerate 返回当前总权重是否大于 0，即是否还能继续生成。
// Generate 前应先调用此方法确认，总权重为 0 时调用 Generate 会触发 assert。
func (this *DynamicWeights[K, T]) CanGenerate() bool {
	return this.TtlWght > 0
}

// SetReduce 设置每次命中后权重的减少量。
func (this *DynamicWeights[K, T]) SetReduce(reduce T) {
	this.Reduce = reduce
}

// AddWeight 向 DW 中新增（或累加）一个 key 的权重。
// 用于 Progressive 模式：在运行时动态解锁新的候选项。
func (this *DynamicWeights[K, T]) AddWeight(key K, weight T) {
	this.Weights[key] += weight
	this.TtlWght += weight
}

// Generate 遍历查找，返回命中的 key。
// 调用者应先通过 CanGenerate() 确认可以继续生成。
// 时间复杂度O(n)
func (this *DynamicWeights[K, T]) Generate() K {
	assert.Assert(this.TtlWght > 0, "总权重需要大于0：", this.TtlWght)
	traverse := T(0)
	// 先根据总权重计算一个随机值，范围在[1,totalWeight]
	r := random.RandInt[T](1, this.TtlWght)
	for key, weight := range this.Weights {
		// 最后一次循环后，traverse会等于totalWeight,此时必然有r <= totalWeight
		traverse += weight
		if r <= traverse {
			// 命中区间
			// 当前key对应的权重减少
			this.Weights[key] -= this.Reduce
			if this.Weights[key] <= 0 {
				// 如果对应key的权重小于等于0了，则从权重集合中移除
				delete(this.Weights, key)
			}
			// 总权重减少
			this.TtlWght -= this.Reduce
			// 返回命中的key
			return key
		}
	}
	// 直接断言 逻辑不应该执行到这里
	assert.Assert(false, "未命中任何区间,r:", r, "totalWeight:", this.TtlWght)
	var zero K
	return zero
}
