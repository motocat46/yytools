// Package benchsort.

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
package benchsort

import (
	"fmt"
	"time"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
	"github.com/motocat46/yytools/pkg/algorithms/sort"
	"github.com/motocat46/yytools/pkg/common/assert"
)

func BubbleSortTest(cnt int) {
	fmt.Printf("冒泡排序测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](1, 1e3)
		}
		sort.BubbleSort(arr)
		for z := 1; z < len(arr); z++ {
			assert.Assert(arr[z-1] <= arr[z])
		}
	}
	fmt.Printf("冒泡排序测试完毕..\n")
}

func BubbleSortDescTest(cnt int) {
	fmt.Printf("冒泡排序(降序)测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](1, 1e3)
		}
		sort.BubbleSortDesc(arr)
		for z := 1; z < len(arr); z++ {
			assert.Assert(arr[z-1] >= arr[z])
		}
	}
	fmt.Printf("冒泡排序(降序)测试完毕..\n")
}

func InsertionSortTest(cnt int) {
	fmt.Printf("插入排序测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](1, 1e3)
		}
		sort.InsertionSort(arr)
		for z := 1; z < len(arr); z++ {
			assert.Assert(arr[z-1] <= arr[z])
		}
	}
	fmt.Printf("插入排序测试完毕..\n")
}

func InsertionSortDescTest(cnt int) {
	fmt.Printf("插入排序(降序)测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](1, 1e3)
		}
		sort.InsertionSortDesc(arr)
		for z := 1; z < len(arr); z++ {
			assert.Assert(arr[z-1] >= arr[z])
		}
	}
	fmt.Printf("插入排序(降序)测试完毕..\n")
}

func runSortBenchmark(sortFunc func(arr []int32), cnt int, desc bool) {
	arr := make([]int32, 1e6)
	totalDuration := int64(0)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](1, 1e5)
		}
		start := time.Now().UnixNano()
		sortFunc(arr)
		end := time.Now().UnixNano()
		duration := (end - start) / 1e6
		fmt.Printf("测试%d耗时:%dms\n", j+1, duration)
		totalDuration += duration
		for z := 1; z < len(arr); z++ {
			if desc {
				assert.Assert(arr[z-1] >= arr[z])
			} else {
				assert.Assert(arr[z-1] <= arr[z])
			}
		}
	}
	if cnt > 1 {
		fmt.Printf("平均耗时:%dms\n", totalDuration/int64(cnt))
	}
}

func QuickSortTest(cnt int) {
	fmt.Printf("快速排序测试开始\n")
	runSortBenchmark(sort.QuickSort[int32], cnt, false)
	fmt.Printf("快速排序测试完毕..\n")
}

func QuickSortTraversalTest(cnt int) {
	fmt.Printf("快速排序(遍历)测试开始..\n")
	runSortBenchmark(sort.QuickSortTraversal[int32], cnt, false)
	fmt.Printf("快速排序(遍历)测试完毕..\n")
}

func QuickSortDescTest(cnt int) {
	fmt.Printf("快速排序(降序)测试开始..\n")
	runSortBenchmark(sort.QuickSortDesc[int32], cnt, true)
	fmt.Printf("快速排序(降序)测试完毕..\n")
}

func QuickSortDescTraversalTest(cnt int) {
	fmt.Printf("快速排序(遍历)(降序)测试开始..\n")
	runSortBenchmark(sort.QuickSortDescTraversal[int32], cnt, true)
	fmt.Printf("快速排序(遍历)(降序)测试完毕..\n")
}

func CountingSortTest(cnt int) {
	fmt.Printf("计数排序测试开始..\n")
	arr := make([]int32, 1e6)
	totalDuration := int64(0)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](1, 1e5)
		}
		before := map[int32]int32{}
		for _, v := range arr {
			before[v]++
		}
		start := time.Now().UnixNano()
		sort.CountingSort(arr)
		end := time.Now().UnixNano()
		duration := (end - start) / 1e6
		fmt.Printf("测试%d耗时:%dms\n", j+1, duration)
		totalDuration += duration
		after := map[int32]int32{}
		for _, v := range arr {
			after[v]++
		}
		assert.Assert(len(before) == len(after))
		for k, v := range before {
			assert.Assert(after[k] == v)
		}
		for z := 1; z < len(arr); z++ {
			assert.Assert(arr[z-1] <= arr[z])
		}
	}
	if cnt > 1 {
		fmt.Printf("平均耗时:%dms\n", totalDuration/int64(cnt))
	}
	fmt.Printf("计数排序测试完毕..\n")
}

func RadixSortTest(cnt int) {
	fmt.Printf("基数排序测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt[int32](100, 999)
		}
		before := map[int32]int32{}
		for _, v := range arr {
			before[v]++
		}
		sort.RadixSort(arr)
		after := map[int32]int32{}
		for _, v := range arr {
			after[v]++
		}
		assert.Assert(len(before) == len(after))
		for k, v := range before {
			assert.Assert(after[k] == v)
		}
		for z := 1; z < len(arr); z++ {
			assert.Assert(arr[z-1] <= arr[z])
		}
	}
	fmt.Printf("基数排序测试完毕..\n")
}

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
