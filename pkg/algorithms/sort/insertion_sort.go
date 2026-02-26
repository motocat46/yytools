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
    "github.com/stormYuanYang/yytools/pkg/common/base"
)

func InsertionSort[T base.Integer](arr []T) {
	insertionSort(arr, 0, len(arr))
}

// 插入排序 默认按升序排列
func insertionSort[T base.Integer](arr []T, start, end int) {
	for i := start + 1; i < end; i++ {
		// 理论上这里的遍历查找可以优化成用二分搜索
		// 但移动元素的时间复杂度是O(n)不变的
		// 并且插入排序大部分的使用场景都是数组数据量比较小的时候
		// 所以就没必要用二分搜索
		for j := start; j < i; j++ {
			if arr[i] < arr[j] {
				tmp := arr[i]
				// 将arr[j,i)拷贝到arr[j+1,i+1)
				// 即将元素向右都移动一位
				copy(arr[j+1:i+1], arr[j:i])
				// 然后插入指定位置
				arr[j] = tmp
				break
			}
		}
	}
}

func InsertionSortDesc[T base.Integer](arr []T) {
	insertionSortDesc(arr, 0, len(arr))
}

// 插入排序 默认按降序排序
func insertionSortDesc[T base.Integer](arr []T, start, end int) {
	for i := start + 1; i < end; i++ {
		for j := start; j < i; j++ {
			// 理论上这里的遍历查找可以优化成用二分搜索
			// 但移动元素的时间复杂度是O(n)不变的
			// 并且插入排序大部分的使用场景都是数组数据量比较小的时候
			// 所以就没必要用二分搜索
			if arr[i] > arr[j] {
				tmp := arr[i]
				// 将arr[j,i)拷贝到arr[j+1,i+1)
				// 即将元素向右都移动一位
				copy(arr[j+1:i+1], arr[j:i])
				// 然后插入指定位置
				arr[j] = tmp
				break
			}
		}
	}
}