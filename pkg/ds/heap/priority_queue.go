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

// 用最大堆实现的优先级队列
// 优先级数值越高的元素，优先级越高

// 作者:  yangyuan
// 创建日期:2023/6/7
package heap

import (
	"container/heap"
	"slices"
	
	"github.com/motocat46/yytools/pkg/common/assert"
)

// InterfacePriorityQueue 是基于最大堆的优先级队列接口，Priority 越大越先出队。
type InterfacePriorityQueue[T any] interface {
	PushItem(item *PriorityItem[T])
	PopItem() *PriorityItem[T]
	PeekItem() *PriorityItem[T]
	UpdatePriority(item *PriorityItem[T], newPriority int)
	Length() int
}

// PriorityItem 优先级队列元素
type PriorityItem[T any] struct {
	Data     T   // 携带的数据
	Priority int // 优先级(数值越大的越靠前,即优先级越高)
	Index    int // 在堆中的下标(需要在实现heap.Interface的方法中更新)
}

// PriorityQueue 基于堆实现的优先级队列(最大堆)
type PriorityQueue[T any] struct {
	Items []*PriorityItem[T]
}

// NewPriorityQueue 创建一个空的优先级队列（最大堆）。
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	return &PriorityQueue[T]{}
}

func (this *PriorityQueue[T]) Len() int {
	return len(this.Items)
}

func (this *PriorityQueue[T]) Less(i, j int) bool {
	return this.Items[i].Priority > this.Items[j].Priority
}

func (this *PriorityQueue[T]) Swap(i, j int) {
	this.Items[i], this.Items[j] = this.Items[j], this.Items[i]
	this.Items[i].Index = i
	this.Items[j].Index = j
}

func (this *PriorityQueue[T]) Push(x interface{}) {
	n := this.Len()
	item := x.(*PriorityItem[T])
	this.Items = append(this.Items, item)
	item.Index = n
}

func (this *PriorityQueue[T]) Pop() interface{} {
	length := this.Len()
	item := this.Items[length-1]
	item.Index = -1
	this.Items = slices.Delete(this.Items, length-1, length)
	return item
}

// PushItem 将 item 压入优先级队列，item 不可为 nil。
func (this *PriorityQueue[T]) PushItem(item *PriorityItem[T]) {
	assert.Assert(item != nil)
	heap.Push(this, item)
}

// PopItem 弹出并返回优先级最高的元素；队列为空时返回 nil。
func (this *PriorityQueue[T]) PopItem() *PriorityItem[T] {
	if this.Len() == 0 {
		return nil
	}
	return heap.Pop(this).(*PriorityItem[T])
}

// PeekItem 返回优先级最高的元素但不移除；队列为空时返回 nil。
func (this *PriorityQueue[T]) PeekItem() *PriorityItem[T] {
	if this.Len() == 0 {
		return nil
	}
	return this.Items[0]
}

// UpdatePriority 修改 item 的优先级为 newPriority 并重新调整堆。
// item 必须是当前队列中存在的元素，否则触发 assert。
func (this *PriorityQueue[T]) UpdatePriority(item *PriorityItem[T], newPriority int) {
	assert.Assert(item != nil)
	assert.Assert(item.Index >= 0 && item.Index < this.Len(), "out of range:", item.Index)
	assert.Assert(this.Items[item.Index] == item, "元素未在队列中,传入的优先级：", item.Priority)
	item.Priority = newPriority
	heap.Fix(this, item.Index)
}

func (this *PriorityQueue[T]) Length() int {
	return this.Len()
}
