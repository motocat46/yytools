// Package heap.

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
// 创建日期:2023/7/11
package heap

import (
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

// 用单调递减的变量来表示元素的顺序
var muniq = 1000000
var mmax = muniq

func MaxHeapOp_PushItem(heap InterfaceHeap, num int) interface{} {
	for i := 0; i < num; i++ {
		one := &Item{
			Data:   nil,
			Weight: muniq,
		}
		heap.PushItem(one)
		muniq--

		assert.Assert(heap.Length() > 0)
	}
	return nil
}

func MaxHeapOp_PopItem(heap InterfaceHeap, num int) interface{} {
	res := make([]*Item, 0, num)
	for i := 0; i < num; i++ {
		if heap.Length() > 0 {
			tmp := heap.PeekItem()

			item := heap.PopItem()
			assert.Assert(item.Weight == mmax)
			assert.Assert(item == tmp)

			mmax--
			res = append(res, item)
		}
	}
	// 必须是从大到小的
	for i := 0; i < len(res)-1; i++ {
		assert.Assert(res[i].Weight > res[i+1].Weight)
	}
	return res
}

func MaxHeapOp_PeekItem(heap InterfaceHeap, num int) interface{} {
	for i := 0; i < num; i++ {
		oldLength := heap.Length()
		if oldLength > 0 {
			item := heap.PeekItem()
			assert.Assert(item.Weight == mmax)
			assert.Assert(oldLength == heap.Length())
		}
	}
	return nil
}

var MaxHeap_handlers = []func(heap InterfaceHeap, num int) interface{}{
	MaxHeapOp_PushItem,
	MaxHeapOp_PopItem,
	MaxHeapOp_PeekItem,
}

func MaxHeapMustBeLegal(heap InterfaceHeap, deleted []*Item) {
	items := MaxHeapOp_PopItem(heap, heap.Length()).([]*Item)
	assert.Assert(heap.Length() == 0)
	deleted = append(deleted, items...)
	// 必须是从大到小的
	for i := 0; i < len(deleted)-1; i++ {
		assert.Assert(deleted[i].Weight > deleted[i+1].Weight)
	}
}

func MaxHeapTest(num int) {
	println("最大堆测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	// 起始规模
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000, 100000, 1000000}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			var heap InterfaceHeap = NewMaxHeap()
			deleted := []*Item{}
			// 需要重置数据起始值
			muniq = 1000000
			mmax = muniq
			MaxHeapOp_PushItem(heap, s)

			// 十万次
			opCnt := 100000
			handlerLength := len(MaxHeap_handlers)
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, handlerLength-1)
				handler := MaxHeap_handlers[r]
				res := handler(heap, 1)
				if res != nil {
					deleted = append(deleted, res.([]*Item)...)
				}
			}
			MaxHeapMustBeLegal(heap, deleted)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, heap.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("最大堆测试完毕...")
}