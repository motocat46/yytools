// Package benchqueue.

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
package benchqueue

import (
	"fmt"

	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/ds/queue"
)

// 用单调递增的变量来表示元素的顺序
var uniq = 1
var min = uniq

func queueOpEnqueue(q *queue.Queue[int], num int) {
	for i := 0; i < num; i++ {
		oldLength := q.Len()
		q.Enqueue(uniq)
		uniq++
		assert.Assert(oldLength == q.Len()-1)
		assert.Assert(!q.Empty())
	}
}

func queueOpDequeue(q *queue.Queue[int], num int) {
	for i := 0; i < num; i++ {
		if !q.Empty() {
			assert.Assert(q.Len() > 0)
			oldLength := q.Len()
			deleted := q.Dequeue()
			assert.Assert(deleted != 0)
			assert.Assert(deleted == min)
			assert.Assert(oldLength == q.Len()+1)
			min++
		} else {
			assert.Assert(q.Len() == 0)
		}
	}
}

func queueOpPeek(q *queue.Queue[int], num int) {
	for i := 0; i < num; i++ {
		if !q.Empty() {
			oldLength := q.Len()
			item := q.Peek()
			assert.Assert(item == min)
			assert.Assert(oldLength == q.Len())
			assert.Assert(!q.Empty())
		} else {
			assert.Assert(q.Len() == 0)
		}
	}
}

var queueHandlers = []func(q *queue.Queue[int], num int){
	queueOpEnqueue,
	queueOpDequeue,
	queueOpPeek,
}

func queueMustBeLegal(q *queue.Queue[int]) {
	items := make([]interface{}, 0, q.Len())
	q.Range(func(item int) {
		items = append(items, item)
	})

	if len(items) > 0 {
		assert.Assert(min == items[0])
		for i := 0; i < len(items)-1; i++ {
			assert.Assert(items[i].(int) < items[i+1].(int), "队列必须是有先后顺序的!")
		}
	} else {
		assert.Assert(q.Empty())
		assert.Assert(q.Len() == 0)
	}
}

func QueueTest(num int) {
	println("队列测试开始...")
	// 起始规模
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 1e4, 1e5, 1e6}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			q := queue.NewQueue[int]()
			// 需要重置数据起始值
			uniq = 1
			min = uniq
			queueOpEnqueue(q, s)
			queueMustBeLegal(q)

			opCnt := 100000
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, len(queueHandlers)-1)
				queueHandlers[r](q, 1)
			}
			queueMustBeLegal(q)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, q.Len())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("队列测试完毕...")
}
