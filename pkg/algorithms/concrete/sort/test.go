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

func SortTest(cnt int) {
	// 冒泡排序
	BubbleSortTest(cnt)
	BubbleSortDescTest(cnt)
	// 插入排序
	InsertionSortTest(cnt)
	InsertionSortDescTest(cnt)
	// 快速排序
	QuickSortTest(cnt)
	QuickSortTraversalTest(cnt)
	QuickSortDescTest(cnt)
	QuickSortDescTraversalTest(cnt)
	// 计数排序
	CountingSortTest(cnt)
	// 基数排序
	// RadixSortTest(cnt)
}