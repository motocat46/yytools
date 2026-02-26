// Package sort.

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
// 创建日期:2023/10/25
package sort

import (
    "github.com/stormYuanYang/yytools/pkg/common/base"
)

// 计数排序 时间复杂度O(n+k)
// 排序的效率非常高，但使用的场景受限:
// 最小值和最大值的差值不能过大(差值和额外申请的内存大小成正比);
// 适用于元素数量很多，但是其数值集中在一定范围内的情况
// 适用于:
// 1.负数的数组;
// 2.有负数有正数的数组;
// 3.正数的数组;
func CountingSort[T base.Integer](array []T) {
	if len(array) < 2 {
		return
	}
	min := array[0]
	max := array[0]
	for i := 1; i < len(array); i++ {
		if min > array[i] {
			min = array[i]
		} else if max < array[i] {
			max = array[i]
		}
	}
	if min == max {
		// 数组中的数字都相等，无需再排序
		return
	}
	
	aux := make([]T, max-min+1)
	for _, v := range array {
		// 根据偏移量计算元素对应的数量
		aux[v-min]++
	}
	
	j := 0
	for i := 0; i < len(aux); i++ {
		for aux[i] > 0 {
			array[j] = T(i) + min
			aux[i]--
			j++
		}
	}
}