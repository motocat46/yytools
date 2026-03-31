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
	"github.com/motocat46/yytools/pkg/common/base"
	"math"
)

// RadixSort 对非负整数切片原地升序排序，时间复杂度 O(d*(n+10))，d 为最大值的十进制位数。
// 按个位、十位、百位…依次稳定排序，每轮使用计数排序作为子程序。
// 传入负数时 panic。
//
// 示例：
//
//	RadixSort([]int{170, 45, 75, 90, 802}) → [45, 75, 90, 170, 802]
func RadixSort[T base.Integer](array []T) {
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
	if min < 0 {
		panic("RadixSort: 仅支持非负整数，传入了负数")
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