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
	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/common/base"
)

/*
*	提供动态的和概率相关的工具方法
*   动态指：元素的权重值会随着算法的进行改变(减少)
*   区别于 probability_distribution.go里的方法
 */

type IDynamicProbDistr[T base.Signed] interface {
	CanGenerate() bool
	Generate() interface{}
	SetReduce(reduce T)
}

// 可以得到一个周期的完整分布
type DynamicWeights[T base.Signed] struct {
	Weights map[interface{}]T // 权重map
	TtlWght T                 // 总权重
	Reduce  T                 // 权重减少的值
}

// 一般而言reduce为1,表示减去一个单位的权重
func NewDynamicWeights[T base.Signed](weights map[interface{}]T) *DynamicWeights[T] {
	return NewDynamicWeightsWithReduce[T](weights, 1)
}

func NewDynamicWeightsWithReduce[T base.Signed](weights map[interface{}]T, reduce T) *DynamicWeights[T] {
	assert.Assert(len(weights) > 0)
	assert.Assert(reduce > 0)
	total := T(0)
	for _, w := range weights {
		total += w
		assert.Assert(w > 0)
	}
	assert.Assert(total > 0, "总权重需要大于0：", total)
	return &DynamicWeights[T]{
		Weights: weights,
		TtlWght: total,
		Reduce:  reduce,
	}
}

// NewDynamicWeightsProgressive 创建初始为空、支持后续动态增加权重的 DW。
// 与 NewDynamicWeights 不同，此构造器不要求初始权重非空。
func NewDynamicWeightsProgressive[T base.Signed]() *DynamicWeights[T] {
	return &DynamicWeights[T]{
		Weights: make(map[interface{}]T),
		TtlWght: 0,
		Reduce:  1,
	}
}

// 判断是否可以继续获得
func (this *DynamicWeights[T]) CanGenerate() bool {
	if this.TtlWght > 0 {
		return true
	}
	return false
}

func (this *DynamicWeights[T]) SetReduce(reduce T) {
	this.Reduce = reduce
}

// 遍历查找
// 调用者去判断是否可以继续获得(CanGenerate判断)
// 时间复杂度O(n)
func (this *DynamicWeights[T]) Generate() interface{} {
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
	return nil
}