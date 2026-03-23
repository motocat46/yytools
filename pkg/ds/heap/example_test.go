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

package heap_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/ds/heap"
)

// ExampleHeap 展示最小堆：按 Weight 从小到大弹出。
func ExampleHeap() {
	h := heap.NewHeap[string]()

	h.PushItem(&heap.Item[string]{Data: "task-c", Weight: 3})
	h.PushItem(&heap.Item[string]{Data: "task-a", Weight: 1})
	h.PushItem(&heap.Item[string]{Data: "task-b", Weight: 2})

	for h.Length() > 0 {
		item := h.PopItem()
		fmt.Println(item.Data, item.Weight)
	}
	// Output:
	// task-a 1
	// task-b 2
	// task-c 3
}

// ExamplePriorityQueue 展示优先级队列：动态更新优先级。
func ExamplePriorityQueue() {
	pq := heap.NewPriorityQueue[string]()

	normal := &heap.PriorityItem[string]{Data: "normal-task", Priority: 5}
	urgent := &heap.PriorityItem[string]{Data: "urgent-task", Priority: 10}

	pq.PushItem(normal)
	pq.PushItem(urgent)

	// urgent 先出
	fmt.Println(pq.PopItem().Data)

	// 提升 normal 的优先级
	pq.PushItem(normal)
	pq.UpdatePriority(normal, 20)
	fmt.Println(pq.PopItem().Data)
	// Output:
	// urgent-task
	// normal-task
}
