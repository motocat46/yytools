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

// 作者:  yangyuan
// 创建日期:2023/7/8
package stack

import (
    "fmt"
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

var unqi = 0

func Stack_Push(stack *Stack[int], num int) {
	for i := 0; i < num; i++ {
		unqi++
		stack.Push(unqi)
	}
}

func Stack_Pop(stack *Stack[int], num int) {
	for i := 0; i < num; i++ {
		if !stack.Empty() {
			top := stack.Top()
			oldLen := stack.Length()
			elem := stack.Pop()
			assert.Assert(top == elem && elem != 0)
			assert.Assert(oldLen == stack.Length()+1)
		}
	}
}

func Stack_Top(stack *Stack[int], num int) {
	if stack.Empty() {
		return
	}
	for i := 0; i < num; i++ {
		oldLen := stack.Length()
		top := stack.Top()
		assert.Assert(top != 0)
		assert.Assert(oldLen == stack.Length())
	}
}

func Stack_EmptyCheck(stack *Stack[int], num int) {
	if stack.Empty() {
		assert.Assert(len(stack.Items) == 0)
	} else {
		assert.Assert(len(stack.Items) > 0)
	}
}

func Stack_LengthCheck(stack *Stack[int], num int) {
	length := stack.Length()
	assert.Assert(length == len(stack.Items))
}

func StackMustLegal(stack *Stack[int]) {
	Stack_EmptyCheck(stack, 1)
	Stack_LengthCheck(stack, 1)
	
	for i := 0; i < len(stack.Items); i++ {
		if i < len(stack.Items)-1 {
			// 必须是按顺序的
			assert.Assert(stack.Items[i] < stack.Items[i+1])
		}
	}
}

var Stack_Handlers = []func(stack *Stack[int], num int){
	Stack_Push,
	Stack_Pop,
	Stack_Top,
}

func StackTest(num int) {
	println("栈测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	// 起始规模
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000, 100000, 1000000}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			stack := NewStack[int]()
			Stack_Push(stack, s)

			// 十万次
			opCnt := 100000
			handlerLength := len(Stack_Handlers)
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, handlerLength-1)
				handler := Stack_Handlers[r]
				handler(stack, 1)

				Stack_EmptyCheck(stack, 1)
				Stack_LengthCheck(stack, 1)
			}
			StackMustLegal(stack)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, stack.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("栈测试完毕...")
}