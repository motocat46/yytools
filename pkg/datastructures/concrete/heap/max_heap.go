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
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
)

/*
 最大堆
*/
type MaxHeap struct {
	Heap
}

func NewMaxHeap() *MaxHeap {
	return &MaxHeap{
		Heap: *NewHeap(),
	}
}

func (this *MaxHeap) Less(i, j int) bool {
	// 这里的比较，决定了该堆是个最大堆
	return this.Items[i].Weight > this.Items[j].Weight
}

func (this *MaxHeap) PushItem(item *Item) {
	assert.Assert(item != nil)
	heap.Push(this, item)
}

func (this *MaxHeap) PopItem() *Item {
	return heap.Pop(this).(*Item)
}