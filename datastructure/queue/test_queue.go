// Package queue.

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
package queue

import (
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

// 用单调递增的变量来表示元素的顺序
var uniq = 1
var min = uniq

func QueueOp_Enqueue(queue *Queue[int], num int) {
	for i := 0; i < num; i++ {
		oldLength := queue.Len()
		
		queue.Enqueue(uniq)
		uniq++
		
		assert.Assert(oldLength == queue.Len()-1)
		assert.Assert(!queue.Empty())
	}
}

func QueueOp_Dequeue(queue *Queue[int], num int) {
	for i := 0; i < num; i++ {
		if !queue.Empty() {
			assert.Assert(queue.Len() > 0)
			
			oldLength := queue.Len()
			deleted := queue.Dequeue()
			
			assert.Assert(deleted != 0)
			assert.Assert(deleted == min)
			assert.Assert(oldLength == queue.Len()+1)
			
			min++
		} else {
			assert.Assert(queue.Len() == 0)
		}
	}
}

func QueueOp_Peek(queue *Queue[int], num int) {
	for i := 0; i < num; i++ {
		if !queue.Empty() {
			oldLength := queue.Len()
			
			item := queue.Peek()
			
			assert.Assert(item == min)
			assert.Assert(oldLength == queue.Len())
			assert.Assert(!queue.Empty())
		} else {
			assert.Assert(queue.Len() == 0)
		}
	}
}

var Queue_Handlers = []func(queue *Queue[int], num int){
	QueueOp_Enqueue,
	QueueOp_Dequeue,
	QueueOp_Peek,
}

func QueueMustBeLegal(queue *Queue[int]) {
	items := make([]interface{}, 0, queue.Len())
	queue.Range(func(item int) {
		items = append(items, item)
	})
	
	if len(items) > 0 {
		assert.Assert(min == items[0])
		for i := 0; i < len(items)-1; i++ {
			assert.Assert(items[i].(int) < items[i+1].(int), "队列必须是有先后顺序的!")
		}
	} else {
		assert.Assert(queue.Empty())
		assert.Assert(queue.Len() == 0)
	}
}

func QueueTest(num int) {
	println("队列测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	// 起始规模
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 1e4, 1e5, 1e6}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			queue := NewQueue[int]()
			// 需要重置数据起始值
			uniq = 1
			min = uniq
			QueueOp_Enqueue(queue, s)
			QueueMustBeLegal(queue)
			
			// 十万次
			opCnt := 100000
			handlerLength := len(Queue_Handlers)
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, handlerLength-1)
				handler := Queue_Handlers[r]
				handler(queue, 1)
				
			}
			QueueMustBeLegal(queue)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, queue.Len())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("队列测试完毕...")
}