// Package tools.

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
// 创建日期:2022/6/15
package probability_distribution

import (
	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	. "github.com/stormYuanYang/yytools/pkg/common/base"
	"github.com/stormYuanYang/yytools/pkg/ds/stack"
)

/**
*	提供静态的和概率相关的工具方法
*   静态指：元素的概率和权重值不会改变(不会因为算法的进行而改变)
 */

// 遍历查找
// 时间复杂度O(n)
func CalcIndexByWeight[T Signed](weightList []T, totalWeight T) int {
	assert.Assert(totalWeight > 0, "总权重需要大于0：", totalWeight)
	traverse := T(0)
	// 先根据总权重计算一个随机值，范围在[1,totalWeight]
	r := random.RandInt[T](1, totalWeight)
	for i, weight := range weightList {
		// 最后一次循环后，traverse会等于totalWeight,此时必然有r <= totalWeight
		traverse += weight
		if r <= traverse {
			// 命中区间
			return i
		}
	}
	// 直接断言 逻辑不应该执行到这里
	assert.Assert(false, "未命中任何区间,r:", r, "totalWeight:", totalWeight)
	return -1
}

// 遍历查找(在map中根据权重查找)
// 时间复杂度O(n)
func CalcKeyByWeight[K comparable, V Signed](weightMap map[K]V, totalWeight V) K {
	assert.Assert(totalWeight > 0, "总权重需要大于0：", totalWeight)
	traverse := V(0)
	// 先根据总权重计算一个随机值，范围在[1,totalWeight]
	r := random.RandInt[V](1, totalWeight)
	for key, weight := range weightMap {
		// 最后一次循环后，traverse会等于totalWeight,此时必然有r <= totalWeight
		traverse += weight
		if r <= traverse {
			// 命中区间
			return key
		}
	}
	// 直接断言 逻辑不应该执行到这里
	assert.Assert(false, "未命中任何区间,r:", r, "totalWeight:", totalWeight)
	var k K
	return k
}

// 概率分布生成接口
// 返回权重数组对应权重的下标
type IProbDist interface {
	Generate() int
}

/*
		普通实现 离散分布(利用二分搜索对查找进行优化)
	 	构建时间复杂度:O(n),空间复杂度O(n)
		生成时间复杂度:O(logn)
		比起vose's alias method效率要低一些（但实现要简单很多，也更容易理解）
*/
type NormalMethod[T Signed] struct {
	WeightsSum []T // 权重和数组
}

func NewNormalMethod[T Signed](weights []T) *NormalMethod[T] {
	weightsSum := make([]T, 0, len(weights))
	var totalWeight T
	for _, weight := range weights {
		assert.Assert(weight >= 0, "元素的权重不能小于0:", weight)
		// 计算权重
		totalWeight += weight
		weightsSum = append(weightsSum, totalWeight)
	}
	assert.Assert(len(weightsSum) > 0, "权重数组长度要大于0")
	return &NormalMethod[T]{
		WeightsSum: weightsSum,
	}
}

// 时间复杂度:O(logn)
// 返回权重数组的下标
func (this *NormalMethod[T]) Generate() int {
	// 考虑一种特殊情况：即数组中所有元素的权重都为0
	// 此时可以认为就是等概率计算各个元素的概率
	length := len(this.WeightsSum)
	totalWeight := this.WeightsSum[length-1]
	if totalWeight == 0 {
		// 等概率随机数组下标
		return random.RandInt(0, length-1)
	}
	// 接下来，总权重至少为1
	// 利用二分搜索提高查找的效率(比起遍历，时间复杂度从O(n)改善到O(logn))
	randNum := random.RandInt[T](1, totalWeight)
	index := searchLeftBound(this.WeightsSum, randNum)
	assert.Assert(index != -1)
	return index
}

// 二分搜索
// 查找左边界(区别普通二分查找)
func searchLeftBound[T Integer](weightsSum []T, target T) int {
	left := 0
	right := len(weightsSum) - 1
	// 目标值在(weightsSum[i-1], weightsSum[i]] 就意味着命中
	// 实际上就是查找左边界
	for left <= right {
		mid := left + (right-left)/2
		if target <= weightsSum[mid] {
			// 缩小右边范围
			right = mid - 1
		} else {
			// 缩小左边范围
			left = mid + 1
		}
	}
	// 当循环结束时, left == right + 1
	if left < 0 || left >= len(weightsSum) {
		// 找不到
		return -1
	}
	// 目标值在(weightsSum[i-1], weightsSum[i]] 就意味着命中
	// 判断是否在指定的区间中
	// 这里只需要判断右边
	if target <= weightsSum[left] {
		return left
	}
	return -1
}

// 在区间中进行二分查找（区别于普通的二分查找）
// (l[0], l[1]]
// (l[1], l[2]]
// (l[2], l[3]]
// ...
// (l[n-1], l[n]]
func binarySearchInRange[T Integer](tmpList []T, n T) int {
	length := len(tmpList)
	if length == 0 {
		return -1
	}

	for low, high := 1, length-1; low <= high; {
		mid := low + (high-low)/2

		if n > tmpList[mid-1] && n <= tmpList[mid] {
			// 命中
			return mid
		} else if n > tmpList[mid] {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}

// Vose's Alias Method(Vose的别名方法)
// 十分高效而优雅的实现方式
// 初始化阶段:
// 时间复杂度O(n),空间复杂度O(n)
// 生成阶段:
// 时间复杂度O(1) [初始化阶段决定了:生成器本身就具有O(n)的空间复杂度]
// 参考: https://www.keithschwarz.com/darts-dice-coins/
type VoseAliasMethod struct {
	Prob  []float // 概率数组
	Alias []int   // 别名数组(记录的是下标)
}

type float float64

func NewVoseAliasMethod[T Integer](weights []T) *VoseAliasMethod {
	totalWeight := T(0)
	for _, w := range weights {
		totalWeight += w
	}
	assert.Assert(totalWeight > 0, "总权重肯定要大于0:", totalWeight)
	n := len(weights)
	prob := make([]float, n)
	alias := make([]int, n)

	small := stack.NewStackWithSize[int](n / 2)
	large := stack.NewStackWithSize[int](n / 2)

	// 初始化概率数组(每个原概率值都乘以n,等比例放大;总体的概率分布是不变的)
	// 概率值等比例放大后，概率平均值为1，概率总和为n
	for i := 0; i < n; i++ {
		prob[i] = float(weights[i]) / float(totalWeight) * float(n)
		// 设置小概率和大概率
		if prob[i] < 1 {
			small.Push(i)
		} else {
			large.Push(i)
		}
	}

	// 当小概率集合和大概率集合都不为空时，循环处理
	// 直到有一个集合为空，则结束循环
	for !small.Empty() && !large.Empty() {
		l := small.Pop() // 小概率的下标
		g := large.Pop() // 大概率的下标
		alias[l] = g     // 别名数组记录另一部分概率的下标
		// This is a more numerically stable option
		// 比起 prob[g] - (1 - prob[l])数值更稳定
		prob[g] = prob[g] + prob[l] - 1
		if prob[g] < 1 {
			small.Push(g) // 下标压入小概率下标集合
		} else {
			large.Push(g) // 下标压入大概率下标集合
		}
	}

	// 判断大概率下标集合
	for !large.Empty() {
		g := large.Pop() // 得到下标
		prob[g] = 1      // 经过前面的处理，这里的概率肯定是1
		alias[g] = -1    // 对应别名下标(该情况下，没有对应别名所以其下标设置为一个无效的-1)
	}
	// 判断小概率下标集合
	for !small.Empty() {
		// 能进入这里是因为数值精度的不稳定(This is only possible due to numerical instability)
		l := small.Pop() // 得到下标
		prob[l] = 1      // 经过前面的处理，这里的概率肯定是1
		alias[l] = -1    // 对应别名下标(该情况下，没有对应别名所以其下标设置为一个无效的-1)
	}
	// 至此，概率数组和别名数组都已构建完成

	ret := &VoseAliasMethod{
		Prob:  prob,
		Alias: alias,
	}
	return ret
}

// 返回概率(权重)数组的某一概率(权重)的下标
// 效率非常高的：时间复杂度O(1)
func (this *VoseAliasMethod) Generate() int {
	n := len(this.Prob)
	i := random.RandInt(0, n-1) // 随机得到一个概率数组的下标
	p := random.RandFloat64()   // 范围：[0.0,1.0)
	// 在[0.0, 1.0)范围内随机得到一个值,用这个值去判断得到该概率还是其别名
	if float(p) < this.Prob[i] {
		return i
	} else {
		index := this.Alias[i] // 别名数组保存的是下标
		assert.Assert(index >= 0 && index < n, "out of range:", index)
		return index
	}
}

/*
工厂模式 可以更方便使用概率分布方法
*/
type MethodType int32

const (
	Normal    MethodType = iota // 0 普通方法
	VoseAlias                   // 1 vose的别名方法
)

func ProbFactory[T Signed](typ MethodType, weights []T) IProbDist {
	switch typ {
	case Normal:
		return NewNormalMethod(weights)
	case VoseAlias:
		return NewVoseAliasMethod(weights)
	default:
		panic("unsupported type")
	}
}
