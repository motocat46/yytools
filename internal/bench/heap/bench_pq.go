// Package benchheap.

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
package benchheap

import (
	"fmt"
	"time"

	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/ds/heap"
)

func priorityQueueOpPushItem(pq heap.InterfacePriorityQueue[any], num int) interface{} {
	for i := 0; i < num; i++ {
		randomPriority := random.RandInt(1, 100000)
		one := &heap.PriorityItem[any]{
			Data:     nil,
			Priority: randomPriority,
		}
		pq.PushItem(one)
		assert.Assert(pq.Length() > 0)
	}
	return nil
}

func priorityQueueOpPopItem(pq heap.InterfacePriorityQueue[any], num int) interface{} {
	res := make([]*heap.PriorityItem[any], 0, num)
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

func priorityQueueOpPeekItem(pq heap.InterfacePriorityQueue[any], num int) interface{} {
	oldLength := pq.Length()
	if oldLength > 0 {
		item := pq.PeekItem()
		assert.Assert(oldLength == pq.Length())
		return item
	}
	return nil
}

func priorityQueueOpUpdatePriority(pq heap.InterfacePriorityQueue[any], num int) interface{} {
	for i := 0; i < num; i++ {
		oldLen := pq.Length()
		if oldLen > 0 {
			randomIndex := random.RandInt(0, pq.Length()-1)
			q := pq.(*heap.PriorityQueue[any])
			randomPriority := random.RandInt(1, 100000)
			item := q.Items[randomIndex]
			pq.UpdatePriority(item, randomPriority)
			assert.Assert(oldLen == pq.Length())
		}
	}
	return nil
}

var priorityQueueHandlers = []func(pq heap.InterfacePriorityQueue[any], num int) interface{}{
	priorityQueueOpPushItem,
	priorityQueueOpPopItem,
	priorityQueueOpPeekItem,
	priorityQueueOpUpdatePriority,
}

func priorityQueueMustBeLegal(pq heap.InterfacePriorityQueue[any]) {
	var deleted []*heap.PriorityItem[any]
	items := priorityQueueOpPopItem(pq, pq.Length()).([]*heap.PriorityItem[any])
	assert.Assert(pq.Length() == 0)
	deleted = append(deleted, items...)
	for i := 0; i < len(deleted)-1; i++ {
		assert.Assert(deleted[i].Priority >= deleted[i+1].Priority)
	}
}

func PriorityQueueTest(num int) {
	println("优先级队列测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000, 100000, 1000000}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			var pq heap.InterfacePriorityQueue[any] = heap.NewPriorityQueue[any]()
			priorityQueueOpPushItem(pq, s)

			opCnt := 100000
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, len(priorityQueueHandlers)-1)
				priorityQueueHandlers[r](pq, 1)
			}
			priorityQueueMustBeLegal(pq)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, pq.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("优先级队列测试完毕...")
}
