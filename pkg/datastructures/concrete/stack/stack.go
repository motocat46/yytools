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
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
)

type IStack interface {
	Length() int           // 栈的长度
	Empty() bool           // 判断栈是否为空
	Push(item interface{}) // 入栈
	Pop() interface{}      // 出栈
	Top() interface{}      // 获取栈首元素(不出栈)
}

type Stack struct {
	Items []interface{}
}

// 默认栈大小
const DEFAULT_STACK_SIZE = 16

func NewStack() *Stack {
	return NewStackWithSize(DEFAULT_STACK_SIZE)
}

func NewStackWithSize(size int) *Stack {
	assert.Assert(size >= 0, "size must greater than or equl to 0,size:", size)
	items := make([]interface{}, 0, size)
	return &Stack{
		Items: items,
	}
}

/*
	实现相应的接口方法
*/

func (this *Stack) Length() int {
	return len(this.Items)
}

func (this *Stack) Empty() bool {
	return this.Length() == 0
}

func (this *Stack) Push(item interface{}) {
	this.Items = append(this.Items, item)
}

func (this *Stack) tryShrink() {
	if len(this.Items) < cap(this.Items)/4 {
		newCap := cap(this.Items) / 2
		if newCap < DEFAULT_STACK_SIZE {
			newCap = DEFAULT_STACK_SIZE
		}
		newItems := make([]interface{}, len(this.Items), newCap)
		n := copy(newItems, this.Items)
		assert.Assert(n == len(this.Items), "缩容不能改变元素数量!", len(this.Items), n)
		this.Items = newItems
	}
}

// 需要调用者保证(可以调用Empty()判断)，栈里还有元素可以出栈
func (this *Stack) Pop() interface{} {
	length := this.Length()
	assert.Assert(length > 0, "栈空了，无法出栈!")
	item := this.Items[length-1]
	this.Items[length-1] = nil // 为了安全（避免内存泄露）
	// 切面赋值的效率如何？手动决定何时缩容呢？
	// 效率很高相当于直接操作数组下标 但是其切片的容量是不会减小的
	// 也就是说当pop足够多元素后,切片所对应的cap是不会减小的，就会浪费很多空间
	// 这时就需要手动实现缩容了
	this.Items = this.Items[:length-1]
	// 尝试缩容
	this.tryShrink()
	return item
}

// 需要调用者保证(可以调用Empty()判断)，栈里还有元素可以查看
func (this *Stack) Top() interface{} {
	length := this.Length()
	assert.Assert(length > 0, "栈空了，无法出栈!")
	return this.Items[length-1]
}
