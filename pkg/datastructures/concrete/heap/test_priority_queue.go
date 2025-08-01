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

// 优先级队列本质上就是个最大堆

// 作者:  yangyuan
// 创建日期:2023/7/11
package heap

import (
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

func PriorityQueueOp_PushItem(pq InterfacePriorityQueue, num int) interface{} {
	for i := 0; i < num; i++ {
		randomPriority := random.RandInt(1, 100000)
		one := &PriorityItem{
			Data:     nil,
			Priority: randomPriority,
		}
		pq.PushItem(one)

		assert.Assert(pq.Length() > 0)
	}
	return nil
}

func PriorityQueueOp_PopItem(pq InterfacePriorityQueue, num int) interface{} {
	res := make([]*PriorityItem, 0, num)
	for i := 0; i < num; i++ {
		if pq.Length() > 0 {
			tmp := pq.PeekItem()

			item := pq.PopItem()
			assert.Assert(item == tmp)

			res = append(res, item)
		}
	}
	// 必须是从大到小的
	for i := 0; i < len(res)-1; i++ {
		assert.Assert(res[i].Priority >= res[i+1].Priority)
	}
	return res
}

func PriorityQueueOp_PeekItem(pq InterfacePriorityQueue, num int) interface{} {
	oldLength := pq.Length()
	if oldLength > 0 {
		item := pq.PeekItem()
		assert.Assert(oldLength == pq.Length())
		return item
	}
	return nil
}

func PriorityQueueOp_UpdatePriority(pq InterfacePriorityQueue, num int) interface{} {
	for i := 0; i < num; i++ {
		oldLen := pq.Length()
		if oldLen > 0 {
			randomIndex := random.RandInt(0, pq.Length()-1)
			q := pq.(*PriorityQueue)
			randomPriority := random.RandInt(1, 100000)

			item := q.Items[randomIndex]


			pq.UpdatePriority(item, randomPriority)
			assert.Assert(oldLen == pq.Length())
		}
	}
	return nil
}

var PriorityQueue_handlers = []func(heap InterfacePriorityQueue, num int) interface{}{
	PriorityQueueOp_PushItem,
	PriorityQueueOp_PopItem,
	PriorityQueueOp_PeekItem,
	PriorityQueueOp_UpdatePriority,
}

func PriorityQueueMustBeLegal(pq InterfacePriorityQueue) {
	var deleted []*PriorityItem
	items := PriorityQueueOp_PopItem(pq, pq.Length()).([]*PriorityItem)
	assert.Assert(pq.Length() == 0)
	deleted = append(deleted, items...)
	// 必须是从大到小的
	for i := 0; i < len(deleted)-1; i++ {
		assert.Assert(deleted[i].Priority >= deleted[i+1].Priority)
	}
}

func PriorityQueueTest(num int) {
	println("优先级队列测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	// 起始规模
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000, 100000, 1000000}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			var pq InterfacePriorityQueue = NewPriorityQueue()
			PriorityQueueOp_PushItem(pq, s)

			// 十万次
			opCnt := 100000
			handlerLength := len(PriorityQueue_handlers)
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, handlerLength-1)
				handler := PriorityQueue_handlers[r]
				handler(pq, 1)
			}
			PriorityQueueMustBeLegal(pq)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, pq.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("优先级队列测试完毕...")
}