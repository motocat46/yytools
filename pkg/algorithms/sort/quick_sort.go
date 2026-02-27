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
// 创建日期:2023/10/19
package sort

import (
	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/base"
	"github.com/stormYuanYang/yytools/pkg/ds/stack"
)

const maxInsertion = 12

func QuickSort[T base.Integer](arr []T) {
	quickSort(arr, 0, len(arr))
}

func QuickSortTraversal[T base.Integer](arr []T) {
	length := len(arr)
	if length < 2 {
		return
	}
	quickSortTraversal(arr, 0, length)
}

type StackData struct {
	Start int
	End   int
}

// 利用栈的辅助,遍历实现快速排序
// 避免函数的递归调用
func quickSortTraversal[T base.Integer](arr []T, start, end int) {
	s := stack.NewStack[*StackData]()
	s.Push(&StackData{Start: start, End: end})
	for !s.Empty() {
		tmp := s.Pop()
		if tmp.End-tmp.Start <= maxInsertion {
			// 元素较少时，插入排序的效率是很高的
			insertionSort(arr, tmp.Start, tmp.End)
			continue
		}

		pivot := partition(arr, tmp.Start, tmp.End)
		s.Push(&StackData{Start: pivot + 1, End: tmp.End})
		s.Push(&StackData{Start: tmp.Start, End: pivot})
	}
}

func quickSort[T base.Integer](arr []T, start, end int) {
	if end <= start+1 {
		return
	}
	if end-start <= maxInsertion {
		// 元素较少时，插入排序的效率是很高的
		// 元素较少时采用插入排序,减小快排的递归深度
		insertionSort(arr, start, end)
		return
	}
	pivot := partition(arr, start, end)
	quickSort(arr, start, pivot)
	quickSort(arr, pivot+1, end)
}

func partition[T base.Integer](arr []T, start, end int) int {
	r := random.RandInt(start, end-1)
	arr[r], arr[end-1] = arr[end-1], arr[r]

	i := start
	j := end - 1
	for ; i < j; i++ {
		// 从左往右找到第一个大于等于val的元素
		if arr[i] >= arr[end-1] {
			// 然后从右往左，找到一个比指定值小的数值
			for j > i && arr[j] >= arr[end-1] {
				j--
			}
			if j == i {
				break
			}
			// swap
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	// 循环结束后，arr[i]左边的值都小于等于arr[i],arr[i]右边的值都大于等于arr[i]
	arr[i], arr[end-1] = arr[end-1], arr[i]
	return i
}

func QuickSortDesc[T base.Integer](arr []T) {
	quickSortDesc(arr, 0, len(arr))
}

func QuickSortDescTraversal[T base.Integer](arr []T) {
	length := len(arr)
	if length < 2 {
		return
	}
	quickSortDescTraversal(arr, 0, length)
}

// 利用栈的辅助,遍历实现快速排序
// 避免函数的递归调用
func quickSortDescTraversal[T base.Integer](arr []T, start, end int) {
	s := stack.NewStack[*StackData]()
	s.Push(&StackData{Start: start, End: end})
	for !s.Empty() {
		tmp := s.Pop()
		if tmp.End-tmp.Start <= maxInsertion {
			// 元素较少时，插入排序的效率是很高的
			insertionSortDesc(arr, tmp.Start, tmp.End)
			continue
		}

		pivot := partitionDesc(arr, tmp.Start, tmp.End)
		s.Push(&StackData{Start: pivot + 1, End: tmp.End})
		s.Push(&StackData{Start: tmp.Start, End: pivot})
	}
}

func quickSortDesc[T base.Integer](arr []T, start, end int) {
	if end-start <= maxInsertion {
		// 元素较少时，插入排序的效率是很高的
		// 元素较少时采用插入排序,减小快排的递归深度
		insertionSortDesc(arr, start, end)
		return
	}
	pivot := partitionDesc(arr, start, end)
	quickSortDesc(arr, start, pivot)
	quickSortDesc(arr, pivot+1, end)
}

func partitionDesc[T base.Integer](arr []T, start, end int) int {
	r := random.RandInt(start, end-1)
	arr[r], arr[end-1] = arr[end-1], arr[r]

	i := start
	j := end - 1
	for ; i < j; i++ {
		// 从左往右找到第一个小于等于val的元素
		if arr[i] <= arr[end-1] {
			// 然后从右往左，找到一个比指定值大的数值
			for j > i && arr[j] <= arr[end-1] {
				j--
			}
			if j == i {
				break
			}
			// swap
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	// 循环结束后，arr[i]左边的值都小于等于arr[i],arr[i]右边的值都大于等于arr[i]
	arr[i], arr[end-1] = arr[end-1], arr[i]
	return i
}
