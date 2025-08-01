// Package queue.

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
// 创建日期:2023/6/8
package queue

import (
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
)

type IQueue interface {
	Len() int                 // 队列的长度
	Empty() bool              // 判断队列是否为空
	Enqueue(item interface{}) // 入队列
	Dequeue() interface{}     // 出队列
	Peek() interface{}        // 获取队首元素(不出队列)
}

type Queue struct {
	Items  []interface{} //队列容量(队列中元素的最大数量是len(Items)-1, 因为Tail始终指向下一个可以入队的位置)
	Length int           // 队列中的元素数量
	Head   int           // Head等于Tail时表示队列为空
	Tail   int           // Tail+1==Head时表示队列满了 Tail下标对应位置上的元素必须是空的,表示下一个可以入队的空位置(才能区分队列空和满)
}

// 默认队列大小
const DEFAULT_QUEUE_SIZE = 16

func NewQueue() *Queue {
	return NewQueueWithSize(DEFAULT_QUEUE_SIZE)
}

func NewQueueWithSize(size int) *Queue {
	assert.Assert(size >= 0, "size must >= 0,size:", size)
	items := make([]interface{}, size)
	return &Queue{
		Items: items,
		Head:  0,
		Tail:  0,
	}
}

/*
	实现相应的接口方法
*/

// 可以用该方法得到的值和Len()得到的值做比较
// 两者返回的值必须相等
func (this *Queue) calcLen() int {
	length := 0
	// 队尾在队首的右边，说明队列没有环绕
	// 直接相减即可
	if this.Tail >= this.Head {
		length = this.Tail - this.Head
	} else {
		// 环绕的情况
		// 1.左边的元素数量 加上
		// 2.右边的元素数量
		// 或者可以理解为 队列总长度减去中间空着的长度:
		// this.Capacity() - (this.Head - this.Tial)
		// 其结果是一样的
		length = this.Tail + (this.Capacity() - this.Head)
	}
	return length
}

func (this *Queue) Len() int {
	return this.Length
}

// Head和Tail指向同一个位置就说明队列是空的
func (this *Queue) Empty() bool {
	// assert.Assert(this.Length == 0)
	return this.Head == this.Tail
}

// 判断队列是否已满
// 要考虑到元素环绕的情况
// 1.当不环绕时，队首元素在0，而队尾元素在Capacity-1处;
// 2.环绕时，队尾元素就在队首元素的左边第一个，满下标:Tail+1 == Head
func (this *Queue) full() bool {
	return this.nextTail() == this.Head
}

func (this *Queue) Capacity() int {
	return len(this.Items)
}

func (this *Queue) nextTail() int {
	return (this.Tail + 1) % this.Capacity()
}

func (this *Queue) copyTo(newItems []interface{}) {
	// 依次将原数组中的元素，移动到新数组
	if this.Tail >= this.Head {
		// 没有环绕直接拷贝即可
		copy(newItems, this.Items[this.Head:this.Tail])
	} else {
		// 有环绕时
		// 则先复制右边的元素(队列前半段元素)
		// 从原数组头一直到数据最后一个元素
		count := copy(newItems, this.Items[this.Head:])
		// 再复制左边的元素(队列后半段元素)
		// 从原数组的0下标一直到Tail前
		copy(newItems[count:], this.Items[:this.Tail])
	}
}

func (this *Queue) resize(newCapacity int) {
	// 新容量
	newItems := make([]interface{}, newCapacity)
	// 元素总数量
	length := this.Len()
	// 拷贝到新数组
	this.copyTo(newItems)

	// 至此，所有元素都成功复制到新数组
	// 现在替换原数组
	this.Items = newItems
	// 更新首尾下标
	this.Head = 0
	this.Tail = length
	// 做一个断言
	assert.Assert(!this.full(), "容量改变后，一定不会满:", newCapacity, ",", length)
}

func (this *Queue) tryExpand() {
	// 如果队列已满，则需要先扩容
	if this.full() {
		// 扩容(翻倍)
		newCapacity := this.Capacity() * 2
		this.resize(newCapacity)
	}
}

func (this *Queue) Enqueue(item interface{}) {
	// 判断是否需要扩容
	this.tryExpand()
	// 设置新的队尾元素
	this.Items[this.Tail] = item
	// 队尾标志前进
	this.Tail = this.nextTail()
	// 元素数量+1
	this.Length++
}

func (this *Queue) tryShrink() {
	// 1.出队列后的元素总数不到总容量的1/4时，才进行缩容处理
	// 这样可以保证，缩容后的元素总数不到新容量的1/2
	// 2.新容量要大于等于默认的队列大小，否则容易造成很小的队列（新入队元素又容易导致扩容）
	// 尽可能避免频繁扩容、缩容
	threshold := this.Capacity() / 4
	if this.Len() < threshold {
		// 缩容（一半）
		newCapacity := this.Capacity() / 2
		if newCapacity < DEFAULT_QUEUE_SIZE {
			newCapacity = DEFAULT_QUEUE_SIZE
		}
		// 队列中最大元素数量是capacity-1
		assert.Assert(this.Len() < newCapacity, "缩容后必须要保证元素都能放得下!", this.Len(), ",", newCapacity)
		this.resize(newCapacity)
	}
}

func (this *Queue) nextHead() int {
	return (this.Head + 1) % this.Capacity()
}

// 需要调用者保证(可以调用Empty()判断)，队列里还有元素可以出队列
func (this *Queue) Dequeue() interface{} {
	// assert.Assert(!this.Empty(), "队列空了，无法出队列!")
	item := this.Items[this.Head]
	// 为了安全（避免内存泄露）
	this.Items[this.Head] = nil
	// 更新队首位置
	this.Head = this.nextHead()
	// 元素数量减一
	this.Length--
	// 判断是否需要缩容
	this.tryShrink()
	return item
}

// 需要调用者保证(可以调用Empty()判断)，队列里还有元素可以查看
func (this *Queue) Peek() interface{} {
	// assert.Assert(!this.Empty(), "队列空了，无法查看队首元素!")
	return this.Items[this.Head]
}

// 遍历队列中的所有元素
func (this *Queue) Range(f func(interface{})) {
	if this.Empty() {
		return
	}

	// 从队首开始遍历
	current := this.Head
	for {
		f(this.Items[current])
		current = (current + 1) % this.Capacity()
		if current == this.Tail {
			break
		}
	}
}
