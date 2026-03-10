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

// Item 堆元素
type Item[T any] struct {
	Data   T   // 携带的数据
	Weight int // 权重值（决定堆元素的顺序）
}

type InterfaceHeap[T any] interface {
	Length() int
	PushItem(item *Item[T])
	PopItem() *Item[T]
	PeekItem() *Item[T]
}

/*
	堆(最小堆)
	本质上是个数组
	利用二叉堆的性质
	通过golang提供的堆的接口和实现的方法
*/
type Heap[T any] struct {
	Items []*Item[T]
}

// NewHeap new heap
func NewHeap[T any]() *Heap[T] {
	return &Heap[T]{}
}

/*
实现golang关于堆的接口
*/
func (this *Heap[T]) Len() int {
	return len(this.Items)
}

func (this *Heap[T]) Less(i, j int) bool {
	// 这里的比较，决定了该堆是个最小堆
	return this.Items[i].Weight < this.Items[j].Weight
}

func (this *Heap[T]) Swap(i, j int) {
	this.Items[i], this.Items[j] = this.Items[j], this.Items[i]
}

func (this *Heap[T]) Push(x interface{}) {
	this.Items = append(this.Items, x.(*Item[T]))
}

// 根据堆的原理，首位的元素会被交换到最后一位
func (this *Heap[T]) Pop() interface{} {
	length := len(this.Items)          // 获取堆长度
	item := this.Items[length-1]       // 取最后一个元素
	this.Items[length-1] = nil         // 避免内存泄露
	this.Items = this.Items[:length-1] // 堆的长度减一
	return item
}

/*
	实现golang关于堆的接口结束
*/

/*
	自定义的一些方法
 	不能直接使用Push和Pop(是为了实现go提供的堆的接口,不应直接调用)
	使用者应该使用PushItem和PopItem替代
*/

func (this *Heap[T]) Length() int {
	return this.Len()
}

func (this *Heap[T]) PushItem(item *Item[T]) {
	assert.Assert(item != nil)
	heap.Push(this, item)
}

func (this *Heap[T]) PopItem() *Item[T] {
	return heap.Pop(this).(*Item[T])
}

func (this *Heap[T]) PeekItem() *Item[T] {
	return this.Items[0]
}