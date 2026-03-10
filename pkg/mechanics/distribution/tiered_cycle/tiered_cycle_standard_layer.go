// Package tiered_cycle.

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
// 创建日期:2026/3/2
package tiered_cycle

import pd "github.com/motocat46/yytools/pkg/algorithms/mathx/probability_distribution"

type Weight struct {
	weightMap   map[int]int32
	totalWeight int32
}

// NewWeight 构造权重结构，totalWeight 自动由 weightMap 计算，避免调用方手动维护
func NewWeight(weightMap map[int]int32) Weight {
	var total int32
	for _, v := range weightMap {
		total += v
	}
	return Weight{weightMap: weightMap, totalWeight: total}
}

// StandardLayerState 普通层运行时状态（普通层在一个周期内无状态）
type StandardLayerState struct{}

// StandardLayer 普通层：封装普通权重随机分布的规则与算法
type StandardLayer struct {
	weight Weight
}

func newStandardLayer(w Weight) StandardLayer {
	return StandardLayer{weight: w}
}

func (l *StandardLayer) NewState() StandardLayerState {
	return StandardLayerState{}
}

// Generate 执行普通层抽取（纯粹的权重随机，算法在此可独立替换）
func (l *StandardLayer) Generate(_ *StandardLayerState) int {
	return pd.CalcKeyByWeight(l.weight.weightMap, l.weight.totalWeight)
}

// Reset 重置普通层状态（无状态，空操作，保持调用形式统一）
func (l *StandardLayer) Reset(_ *StandardLayerState) {}
