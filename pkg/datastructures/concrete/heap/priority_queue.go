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

// 用最大堆实现的优先级队列的
// 优先级数值越高的元素，优先级越高

// 作者:  yangyuan
// 创建日期:2023/6/7
package heap

import (
    "container/heap"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
)

type InterfacePriorityQueue interface {
	PushItem(item *PriorityItem)
	PopItem() *PriorityItem
	PeekItem() *PriorityItem
	UpdatePriority(item *PriorityItem, newPriority int)
	Length() int
}

// PriorityItem 优先级队列元素
type PriorityItem struct {
	Data     interface{} // 携带的数据
	Priority int         // 优先级(数值越大的越靠前,即优先级越高)
	Index    int         // 在堆中的下标(需要在实现heap.Interface的方法中更新)
}

/*
	基于堆实现的优先级队列(最大堆)
	本质上是个数组
	利用二叉堆的性质
	通过golang提供的堆的接口和实现的方法
*/
type PriorityQueue struct {
	Items []*PriorityItem
}

// NewHeap new heap
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

/*
	实现golang关于堆的接口
*/
func (this *PriorityQueue) Len() int {
	return len(this.Items)
}

func (this *PriorityQueue) Less(i, j int) bool {
	// 这里的比较，决定了该堆是个最大堆
	// 优先级数值越大的越靠前
	return this.Items[i].Priority > this.Items[j].Priority
}

func (this *PriorityQueue) Swap(i, j int) {
	// 交换元素位置
	this.Items[i], this.Items[j] = this.Items[j], this.Items[i]
	// 同时也要更新元素对应的索引位置
	this.Items[i].Index = i
	this.Items[j].Index = j
}

func (this *PriorityQueue) Push(x interface{}) {
	n := this.Len()
	item := x.(*PriorityItem)
	// 新元素进入堆中，肯定是添加到最后一位
	// 然后通过up方法去提升其位置(如果可以的话)
	this.Items = append(this.Items, item)
	item.Index = n
}

// 根据堆的原理，首位的元素会被交换到最后一位
func (this *PriorityQueue) Pop() interface{} {
	length := this.Len() // 获取堆长度
	
	item := this.Items[length-1] // 取最后一个元素
	item.Index = -1              // 为了安全(不再引用数组内下标)
	
	this.Items[length-1] = nil         // 避免内存泄露
	this.Items = this.Items[:length-1] // 堆的长度减一
	return item
}

/*
	实现golang关于堆的接口结束
*/

/*
	自定义的一些方法 使用者不应该用上面这些方法
    使用下面这些方法(最好通过接口访问)
*/

// push元素到优先级队列中
func (this *PriorityQueue) PushItem(item *PriorityItem) {
	assert.Assert(item != nil)
	heap.Push(this, item)
}

//取出优先级最高的元素
func (this *PriorityQueue) PopItem() *PriorityItem {
	return heap.Pop(this).(*PriorityItem)
}

func (this *PriorityQueue) PeekItem() *PriorityItem {
	return this.Items[0]
}

// 更新元素的优先级;重新调节堆内元素的顺序
func (this *PriorityQueue) UpdatePriority(item *PriorityItem, newPriority int) {
	assert.Assert(item != nil)
	assert.Assert(item.Index >= 0 && item.Index < this.Len(), "out of range :", item.Index)
	assert.Assert(this.Items[item.Index] == item, "元素未在队列中,传入的优先级：", item.Priority)
	
	// 设置新的优先级
	item.Priority = newPriority
	// 重新建立堆的顺序
	heap.Fix(this, item.Index)
}

func (this *PriorityQueue) Length() int {
	return this.Len()
}