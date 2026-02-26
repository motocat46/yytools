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
	"time"

	"github.com/stormYuanYang/yytools/pkg/algorithms/mathutils/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
)

var uniq = 0

func stackPush(stack *Stack[int], num int) {
	for i := 0; i < num; i++ {
		uniq++
		stack.Push(uniq)
	}
}

func stackPop(stack *Stack[int], num int) {
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

func stackTop(stack *Stack[int], num int) {
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

func stackEmptyCheck(stack *Stack[int], num int) {
	if stack.Empty() {
		assert.Assert(len(stack.Items) == 0)
	} else {
		assert.Assert(len(stack.Items) > 0)
	}
}

func stackLengthCheck(stack *Stack[int], num int) {
	length := stack.Length()
	assert.Assert(length == len(stack.Items))
}

func stackMustLegal(stack *Stack[int]) {
	stackEmptyCheck(stack, 1)
	stackLengthCheck(stack, 1)

	for i := 0; i < len(stack.Items); i++ {
		if i < len(stack.Items)-1 {
			// 必须是按顺序的
			assert.Assert(stack.Items[i] < stack.Items[i+1])
		}
	}
}

var stackHandlers = []func(stack *Stack[int], num int){
	stackPush,
	stackPop,
	stackTop,
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
			stackPush(stack, s)

			opCnt := 100000
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, len(stackHandlers)-1)
				stackHandlers[r](stack, 1)
				stackEmptyCheck(stack, 1)
				stackLengthCheck(stack, 1)
			}
			stackMustLegal(stack)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, stack.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("栈测试完毕...")
}
