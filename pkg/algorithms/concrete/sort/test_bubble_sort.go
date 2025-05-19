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
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
)

func BubbleSortTest(cnt int) {
	fmt.Printf("冒泡排序测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt32(1, 1e3)
		}
		BubbleSort(arr)
		for z := 1; z < len(arr); z++ {
			// 判断排序结束后是否升序
			assert.Assert(arr[z-1] <= arr[z])
		}
		//for _, v := range arr {
		//	fmt.Printf("%d\t", v)
		//}
		//println()
	}
	fmt.Printf("冒泡排序测试完毕..\n")
}

func BubbleSortDescTest(cnt int) {
	fmt.Printf("冒泡排序(降序)测试开始..\n")
	arr := make([]int32, 1e3)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt32(1, 1e3)
		}
		BubbleSortDesc(arr)
		for z := 1; z < len(arr); z++ {
			// 判断排序结束后是否降序序
			assert.Assert(arr[z-1] >= arr[z])
		}
		//for _, v := range arr {
		//	fmt.Printf("%d\t", v)
		//}
		//println()
	}
	fmt.Printf("冒泡排序(降序)测试完毕..\n")
}