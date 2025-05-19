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
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

func CountingSortTest(cnt int) {
	fmt.Printf("计数排序测试开始..\n")
	arr := make([]int32, 1e6)
	totalDuration := int64(0)
	for j := 0; j < cnt; j++ {
		for i := 0; i < len(arr); i++ {
			arr[i] = random.RandInt32(1, 1e5)
			// sign := random.RandInt32(1, 2)
			// if sign == 1 {
			// 	arr[i] = -arr[i]
			// }
		}
		before := map[int32]int32{}
		for _, v := range arr {
			before[v]++
		}
		start := time.Now().UnixNano()
		CountingSort(arr)
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
			// 判断排序结束后是否升序
			assert.Assert(arr[z-1] <= arr[z])
		}
		//for _, v := range arr {
		//	fmt.Printf("%d\t", v)
		//}
		//println()
	}
	if cnt > 1 {
		fmt.Printf("平均耗时:%d\nms", totalDuration/int64(cnt))
	}
	fmt.Printf("计数排序测试完毕..\n")
}