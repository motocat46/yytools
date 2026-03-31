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
	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
	"github.com/motocat46/yytools/pkg/common/base"
	"github.com/motocat46/yytools/pkg/ds/stack"
)

// maxInsertion 是触发插入排序优化的分区规模阈值；元素数 ≤ maxInsertion 时改用插入排序。
const maxInsertion = 12

// QuickSort 对整数切片原地升序排序，平均时间复杂度 O(n log n)。
// 采用随机三路划分（荷兰国旗算法），大量重复元素时性能从 O(n²) 提升到 O(n)。
// 小分区（≤12 个元素）自动切换为插入排序以减少递归开销。
func QuickSort[T base.Integer](arr []T) {
	quickSort(arr, 0, len(arr))
}

// QuickSortTraversal 对整数切片原地升序排序，语义见 QuickSort。
// 用显式栈替代递归，避免极端输入下的调用栈溢出。
func QuickSortTraversal[T base.Integer](arr []T) {
	length := len(arr)
	if length < 2 {
		return
	}
	quickSortTraversal(arr, 0, length)
}

// StackData 保存待排序分区的左右边界，供 QuickSortTraversal / QuickSortDescTraversal 的显式栈使用。
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
		lt, gt := partition3way(arr, tmp.Start, tmp.End)
		s.Push(&StackData{Start: gt + 1, End: tmp.End})
		s.Push(&StackData{Start: tmp.Start, End: lt})
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
	lt, gt := partition3way(arr, start, end)
	quickSort(arr, start, lt)
	quickSort(arr, gt+1, end)
}

// partition3way 三路划分（荷兰国旗算法）
// 将 arr[start, end) 划分为三段：
//   < pivot  | == pivot | > pivot
//
// 返回 (lt, gt)，其中 arr[lt, gt] 全部等于 pivot，无需再参与递归。
// 相比二路划分，大量重复元素时性能从 O(n²) 提升到 O(n)。
func partition3way[T base.Integer](arr []T, start, end int) (lt, gt int) {
	r := random.RandInt(start, end-1)
	pivot := arr[r]

	lt = start  // [start, lt) < pivot
	gt = end - 1 // (gt, end) > pivot
	i := start

	for i <= gt {
		if arr[i] < pivot {
			arr[lt], arr[i] = arr[i], arr[lt]
			lt++
			i++
		} else if arr[i] > pivot {
			arr[i], arr[gt] = arr[gt], arr[i]
			gt--
			// arr[i] 换来的新值尚未检查，i 不递增
		} else {
			i++
		}
	}
	return lt, gt + 1 // 返回 gt+1 使调用方直接用作切片右边界
}

// QuickSortDesc 对整数切片原地降序排序，语义见 QuickSort。
func QuickSortDesc[T base.Integer](arr []T) {
	quickSortDesc(arr, 0, len(arr))
}

// QuickSortDescTraversal 对整数切片原地降序排序，语义见 QuickSortTraversal。
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
		lt, gt := partition3wayDesc(arr, tmp.Start, tmp.End)
		s.Push(&StackData{Start: gt, End: tmp.End})
		s.Push(&StackData{Start: tmp.Start, End: lt})
	}
}

func quickSortDesc[T base.Integer](arr []T, start, end int) {
	if end <= start+1 {
		return
	}
	if end-start <= maxInsertion {
		// 元素较少时，插入排序的效率是很高的
		// 元素较少时采用插入排序,减小快排的递归深度
		insertionSortDesc(arr, start, end)
		return
	}
	lt, gt := partition3wayDesc(arr, start, end)
	quickSortDesc(arr, start, lt)
	quickSortDesc(arr, gt, end)
}

// partition3wayDesc 降序三路划分
// 将 arr[start, end) 划分为三段：
//
//	> pivot | == pivot | < pivot
func partition3wayDesc[T base.Integer](arr []T, start, end int) (lt, gt int) {
	r := random.RandInt(start, end-1)
	pivot := arr[r]

	lt = start
	gt = end - 1
	i := start

	for i <= gt {
		if arr[i] > pivot {
			arr[lt], arr[i] = arr[i], arr[lt]
			lt++
			i++
		} else if arr[i] < pivot {
			arr[i], arr[gt] = arr[gt], arr[i]
			gt--
		} else {
			i++
		}
	}
	return lt, gt + 1
}