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

package binary_search_test

import (
	"fmt"

	bs "github.com/motocat46/yytools/pkg/algorithms/binary_search"
)

// ExampleBinarySearch 展示精确查找——有重复元素时返回任意匹配下标。
func ExampleBinarySearch() {
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println(bs.BinarySearch(nums, 3))  // 2
	fmt.Println(bs.BinarySearch(nums, 99)) // -1
	// Output:
	// 2
	// -1
}

// ExampleLeftBound 展示查找重复元素的最左下标。
func ExampleLeftBound() {
	nums := []int{1, 2, 2, 3, 3, 3, 4}
	fmt.Println(bs.LeftBound(nums, 3))  // 3（第一个 3 的下标）
	fmt.Println(bs.LeftBound(nums, 99)) // -1
	// Output:
	// 3
	// -1
}

// ExampleRightBound 展示查找重复元素的最右下标。
func ExampleRightBound() {
	nums := []int{1, 2, 2, 3, 3, 3, 4}
	fmt.Println(bs.RightBound(nums, 3))  // 5（最后一个 3 的下标）
	fmt.Println(bs.RightBound(nums, 99)) // -1
	// Output:
	// 5
	// -1
}

// ExampleSearchBound 展示同时获取重复元素的左右边界。
func ExampleSearchBound() {
	nums := []int{1, 2, 2, 3, 3, 3, 4}
	l, r := bs.SearchBound(nums, 3)
	fmt.Println(l, r) // 3 5

	l, r = bs.SearchBound(nums, 99)
	fmt.Println(l, r) // -1 -1
	// Output:
	// 3 5
	// -1 -1
}
