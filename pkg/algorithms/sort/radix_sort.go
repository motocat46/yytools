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
// 创建日期:2023/10/26
package sort

import (
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/common/base"
	"math"
)

// 基数排序（Radix Sort）：基数排序适用于排序非负整数序列，其时间复杂度为O(d * (n + k))，
// 其中d是最大数字的位数，n是元素个数，k是每个位的范围大小。
// 基数排序按照数字的个位、十位、百位等依次进行排序，直到最高位排序完成，得到有序序列。
func RadixSort[T base.Integer](array []T) {
	if len(array) < 2 {
		return
	}
	assert.Assert(array[0] >= 0)
	min := array[0]
	max := array[0]
	for i := 1; i < len(array); i++ {
		assert.Assert(array[i] >= 0)
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

	digits := int(math.Log10(float64(max))) + 1
	aux := make([]int, 10)
	tmp := make([]T, len(array))
	radix := T(1)
	for i := 0; i < digits; i++ {
		// 辅助数组全部重置为0
		for j := 0; j < len(aux); j++ {
			aux[j] = 0
		}
		// 遍历数组得到当前位上的数字数量
		for _, v := range array {
			single := (v / radix) % 10
			aux[single]++
		}
		// 计算得到数字对应的位置和数量
		for j := 1; j < len(aux); j++ {
			aux[j] += aux[j-1]
		}
		// 从右往左遍历，将元素按当前位的大小放入临时数组（保证稳定性）
		for k := len(array) - 1; k >= 0; k-- {
			single := (array[k] / radix) % 10
			aux[single]--
			tmp[aux[single]] = array[k]
		}
		// 将排序后的数组拷贝回原数组
		copy(array, tmp)
		// 增大基数，进行下一位判断
		radix *= 10
	}
}