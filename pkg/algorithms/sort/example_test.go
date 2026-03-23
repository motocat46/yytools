// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sort_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/algorithms/sort"
)

// ExampleQuickSort 展示通用快速排序（升序）。
func ExampleQuickSort() {
	nums := []int{5, 3, 8, 1, 9, 2}
	sort.QuickSort(nums)
	fmt.Println(nums)
	// Output:
	// [1 2 3 5 8 9]
}

// ExampleCountingSort 展示计数排序——值域集中时比快排更快。
// 适合元素数量多、值域范围小（max-min ≤ 1e7）的整数数组。
func ExampleCountingSort() {
	scores := []int{85, 92, 78, 95, 88, 78, 92}
	sort.CountingSort(scores)
	fmt.Println(scores)
	// Output:
	// [78 78 85 88 92 92 95]
}

// ExampleQuickSortDesc 展示降序快速排序。
func ExampleQuickSortDesc() {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
	sort.QuickSortDesc(nums)
	fmt.Println(nums)
	// Output:
	// [9 6 5 4 3 2 1 1]
}
