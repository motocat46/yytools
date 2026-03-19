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
// 创建日期:2023/10/24
package sort

import (
    "github.com/motocat46/yytools/pkg/common/base"
)

func SelectionSort[T base.Integer](arr []T) {
	selectionSort(arr, 0, len(arr))
}

// 选择排序 默认按升序排列
func selectionSort[T base.Integer](arr []T, start, end int) {
	for i := start + 1; i < end; i++ {
		head := i - 1
		min := arr[head]
		minIndex := head
		// 找到最小元素
		for j := i; j < end; j++ {
			if min > arr[j] {
				min = arr[j]
				minIndex = j
			}
		}
		// 把最小的元素换到最前面
		if minIndex != head {
			arr[head], arr[minIndex] = arr[minIndex], arr[head]
		}
	}
}

func SelectionSortDesc[T base.Integer](arr []T) {
	selectionSortDesc(arr, 0, len(arr))
}

// 选择排序 默认按降序排序
func selectionSortDesc[T base.Integer](arr []T, start, end int) {
	for i := start + 1; i < end; i++ {
		head := i - 1
		max := arr[head]
		maxIndex := head
		// 找到最大元素
		for j := i; j < end; j++ {
			if max < arr[j] {
				max = arr[j]
				maxIndex = j
			}
		}
		// 把最大的元素换到最前面
		if maxIndex != head {
			arr[head], arr[maxIndex] = arr[maxIndex], arr[head]
		}
	}
}