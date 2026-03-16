//go:build ignore

// Package main.

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
// 创建日期:2025/6/1
package main

import (
	"container/list"
	"fmt"
	"sync"
	
	"github.com/motocat46/yytools/pkg/common/assert"
)

// Target 实现一个简单的
// 1.不保证消息顺序的;
// 2.size无限制的;
// 通道。
type UnboundedChannelV1 struct {
	mutex   sync.Mutex
	list    *list.List // buffer
	channel chan int   // 通道
}

func NewUnboundedChannelV1() *UnboundedChannelV1 {
	return &UnboundedChannelV1{
		list:    list.New(),
		channel: make(chan int, 10),
	}
}

// 往通道发消息(不保证消息的先后顺序)
func (uc *UnboundedChannelV1) Send(msg int) {
	assert.Assert(uc.channel != nil)
	assert.Assert(uc.list != nil)
	
	uc.mutex.Lock()
	if len(uc.channel) == cap(uc.channel) {
		// 通道满了，先暂存list
		uc.list.PushFront(msg)
		uc.mutex.Unlock()
		return
	}
	// 现在channel未满
	uc.channelSend(msg)
	// 必须在发送到通道后解锁，因为发送数据到通道后，通道就可能满
	// 必须要保证上面的是否满的判断和添加元素的操作为原子性操作。
	uc.mutex.Unlock()
}

func (uc *UnboundedChannelV1) channelSend(msg int) {
	// 复制消息到通道
	select {
	case uc.channel <- msg:
		fmt.Printf("发送消息:%d到通道\n", msg)
	default:
		fmt.Printf("丢弃消息:%d\n", msg)
	}
}

// 从通道接收消息
func (uc *UnboundedChannelV1) Receive() int {
	assert.Assert(uc.channel != nil)
	assert.Assert(uc.list != nil)
	uc.mutex.Lock()
	if e := uc.list.Back(); e != nil {
		val := uc.list.Remove(e)
		uc.mutex.Unlock()
		// 不考虑消息顺序，直接从缓存list中返回
		return val.(int)
	}
	uc.mutex.Unlock()
	msg := <-uc.channel
	return msg
}

func (uc *UnboundedChannelV1) Close() {
	assert.Assert(uc.channel != nil)
	close(uc.channel)
}

func (uc *UnboundedChannelV1) PrintList() {
	uc.mutex.Lock()
	for e := uc.list.Front(); e != nil; e = e.Next() {
		fmt.Printf("%d\t", e.Value)
	}
	fmt.Println()
	uc.mutex.Unlock()
}