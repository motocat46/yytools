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
    "github.com/motocat46/yytools/pkg/common/base"
)

// CountingSort 对整数切片原地升序排序，时间复杂度 O(n+k)，k = max-min+1。
// 适用于元素数量多、但数值集中在较小范围内的场景；支持负数、混合正负数和纯正数数组。
// 当 max-min 超过 1e7 时 panic，请改用 QuickSort。
//
// 示例：
//
//	CountingSort([]int{-3, 1, 0, -1, 2}) → [-3, -1, 0, 1, 2]
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
	// 防止 max-min 过大导致 OOM（建议差值 ≤ 1e7，超出此范围请改用 QuickSort）
	const maxRange = 1e7
	if uint64(max-min) > maxRange {
		panic("CountingSort: max-min 差值超出限制，请改用 QuickSort")
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