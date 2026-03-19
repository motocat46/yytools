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
// 创建日期:2023/6/7
package heap

import (
	"container/heap"

	"github.com/motocat46/yytools/pkg/common/assert"
)

// MaxHeap 最大堆
type MaxHeap[T any] struct {
	Heap[T]
}

func NewMaxHeap[T any]() *MaxHeap[T] {
	return &MaxHeap[T]{
		Heap: *NewHeap[T](),
	}
}

func (this *MaxHeap[T]) Less(i, j int) bool {
	return this.Items[i].Weight > this.Items[j].Weight
}

func (this *MaxHeap[T]) PushItem(item *Item[T]) {
	assert.Assert(item != nil)
	heap.Push(this, item)
}

func (this *MaxHeap[T]) PopItem() *Item[T] {
	if this.Len() == 0 {
		return nil
	}
	return heap.Pop(this).(*Item[T])
}
