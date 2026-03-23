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

package queue_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/ds/queue"
)

// ExampleQueue 展示基本的 FIFO 用法。
func ExampleQueue() {
	q := queue.NewQueue[int]()

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	for !q.Empty() {
		fmt.Println(q.Dequeue())
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleQueue_range 展示 Range 遍历队列而不消费元素。
func ExampleQueue_range() {
	q := queue.NewQueue[string]()
	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c")

	q.Range(func(v string) {
		fmt.Println(v)
	})
	fmt.Println("len:", q.Len()) // 遍历后元素仍在
	// Output:
	// a
	// b
	// c
	// len: 3
}
