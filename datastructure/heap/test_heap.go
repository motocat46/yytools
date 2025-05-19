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
// 创建日期:2023/7/10
package heap

import (
    "fmt"
	"github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

// 用单调递增的变量来表示元素的顺序
var uniq = 1
var min = uniq

func HeapOp_PushItem(heap InterfaceHeap, num int) interface{} {
	for i := 0; i < num; i++ {
		one := &Item{
			Data:   nil,
			Weight: uniq,
		}
		heap.PushItem(one)
		uniq++

		assert.Assert(heap.Length() > 0)
	}
	return nil
}

func HeapOp_PopItem(heap InterfaceHeap, num int) interface{} {
	res := make([]*Item, 0, num)
	for i := 0; i < num; i++ {
		if heap.Length() > 0 {
			tmp := heap.PeekItem()

			item := heap.PopItem()
			assert.Assert(item.Weight == min)
			assert.Assert(item == tmp)

			min++
			res = append(res, item)
		}
	}
	// 必须是从小到大的
	for i := 0; i < len(res)-1; i++ {
		assert.Assert(res[i].Weight < res[i+1].Weight)
	}
	return res
}

func HeapOp_PeekItem(heap InterfaceHeap, num int) interface{} {
	for i := 0; i < num; i++ {
		oldLength := heap.Length()
		if oldLength > 0 {
			item := heap.PeekItem()
			assert.Assert(item.Weight == min)
			assert.Assert(oldLength == heap.Length())
		}
	}
	return nil
}

var Heap_handlers = []func(heap InterfaceHeap, num int) interface{}{
	HeapOp_PushItem,
	HeapOp_PopItem,
	HeapOp_PeekItem,
}

func HeapMustBeLegal(heap InterfaceHeap, deleted []*Item) {
	items := HeapOp_PopItem(heap, heap.Length()).([]*Item)
	assert.Assert(heap.Length() == 0)
	deleted = append(deleted, items...)
	// 必须是从小到大的
	for i := 0; i < len(deleted)-1; i++ {
		assert.Assert(deleted[i].Weight < deleted[i+1].Weight)
	}
}

func HeapTest(num int) {
	println("最小堆测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	// 起始规模
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000, 100000, 1000000}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			var heap InterfaceHeap = NewHeap()
			deleted := []*Item{}
			// 需要重置数据起始值
			uniq = 1
			min = uniq
			HeapOp_PushItem(heap, s)

			// 十万次
			opCnt := 100000
			handlerLength := len(Heap_handlers)
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, handlerLength-1)
				handler := Heap_handlers[r]
				res := handler(heap, 1)
				if res != nil {
					deleted = append(deleted, res.([]*Item)...)
				}
			}
			HeapMustBeLegal(heap, deleted)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, heap.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("最小堆测试完毕...")
}