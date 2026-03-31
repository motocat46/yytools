// Package stack.

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

// 利用数组实现先进后出的数据结构：栈

// 作者:  yangyuan
// 创建日期:2023/6/7
package stack

import (
	"slices"

	"github.com/motocat46/yytools/pkg/common/assert"
)

// IStack 是泛型 LIFO 栈的公开操作接口。
// Pop / Top 在栈为空时 panic，调用前应先检查 Empty()。
type IStack[T any] interface {
	Length() int // 栈的长度
	Empty() bool // 判断栈是否为空
	Push(item T) // 入栈
	Pop() T      // 出栈
	Top() T      // 获取栈首元素(不出栈)
}

// Stack 是基于动态数组的泛型 LIFO 栈，元素数量降至容量 1/4 以下时自动缩容，
// 最低不低于 DEFAULT_STACK_SIZE。
type Stack[T any] struct {
	Items []T
}

// 默认栈大小
const DEFAULT_STACK_SIZE = 16

// NewStack 创建一个默认初始容量（16）的空栈。
func NewStack[T any]() *Stack[T] {
	return NewStackWithSize[T](DEFAULT_STACK_SIZE)
}

// NewStackWithSize 创建指定初始容量的空栈，size 须 >= 0。
func NewStackWithSize[T any](size int) *Stack[T] {
	assert.Assert(size >= 0, "size must greater than or equl to 0,size:", size)
	items := make([]T, 0, size)
	return &Stack[T]{
		Items: items,
	}
}

/*
	实现相应的接口方法
*/

func (this *Stack[T]) Length() int {
	return len(this.Items)
}

func (this *Stack[T]) Empty() bool {
	return this.Length() == 0
}

func (this *Stack[T]) Push(item T) {
	this.Items = append(this.Items, item)
}

func (this *Stack[T]) tryShrink() {
	if len(this.Items) < cap(this.Items)/4 {
		newCap := cap(this.Items) / 2
		if newCap < DEFAULT_STACK_SIZE {
			newCap = DEFAULT_STACK_SIZE
		}
		newItems := make([]T, len(this.Items), newCap)
		n := copy(newItems, this.Items)
		assert.Assert(n == len(this.Items), "缩容不能改变元素数量!", len(this.Items), n)
		this.Items = newItems
	}
}

// 需要调用者保证(可以调用Empty()判断)，栈里还有元素可以出栈
// 栈为空时 panic；调用前应先检查 Empty()
func (this *Stack[T]) Pop() T {
	length := this.Length()
	assert.Assert(length > 0, "栈空了，无法出栈!")
	item := this.Items[length-1]
	// slices.Delete 保证尾部元素置零（避免 GC 泄漏），语义等价于手动 zero + 缩短
	// 切片容量不会因此缩小，缩容由下方 tryShrink 负责
	this.Items = slices.Delete(this.Items, length-1, length)
	// 尝试缩容
	this.tryShrink()
	return item
}

// 需要调用者保证(可以调用Empty()判断)，栈里还有元素可以查看
// 栈为空时 panic；调用前应先检查 Empty()
func (this *Stack[T]) Top() T {
	length := this.Length()
	assert.Assert(length > 0, "栈空了，无法查看!")
	return this.Items[length-1]
}