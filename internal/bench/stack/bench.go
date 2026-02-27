// Package benchstack.

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
package benchstack

import (
	"fmt"
	"time"

	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/ds/stack"
)

var uniq = 0

func stackPush(s *stack.Stack[int], num int) {
	for i := 0; i < num; i++ {
		uniq++
		s.Push(uniq)
	}
}

func stackPop(s *stack.Stack[int], num int) {
	for i := 0; i < num; i++ {
		if !s.Empty() {
			top := s.Top()
			oldLen := s.Length()
			elem := s.Pop()
			assert.Assert(top == elem && elem != 0)
			assert.Assert(oldLen == s.Length()+1)
		}
	}
}

func stackTop(s *stack.Stack[int], num int) {
	if s.Empty() {
		return
	}
	for i := 0; i < num; i++ {
		oldLen := s.Length()
		top := s.Top()
		assert.Assert(top != 0)
		assert.Assert(oldLen == s.Length())
	}
}

func stackEmptyCheck(s *stack.Stack[int], num int) {
	if s.Empty() {
		assert.Assert(len(s.Items) == 0)
	} else {
		assert.Assert(len(s.Items) > 0)
	}
}

func stackLengthCheck(s *stack.Stack[int], num int) {
	length := s.Length()
	assert.Assert(length == len(s.Items))
}

func stackMustLegal(s *stack.Stack[int]) {
	stackEmptyCheck(s, 1)
	stackLengthCheck(s, 1)

	for i := 0; i < len(s.Items); i++ {
		if i < len(s.Items)-1 {
			// 必须是按顺序的
			assert.Assert(s.Items[i] < s.Items[i+1])
		}
	}
}

var stackHandlers = []func(s *stack.Stack[int], num int){
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
			st := stack.NewStack[int]()
			stackPush(st, s)

			opCnt := 100000
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, len(stackHandlers)-1)
				stackHandlers[r](st, 1)
				stackEmptyCheck(st, 1)
				stackLengthCheck(st, 1)
			}
			stackMustLegal(st)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, st.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("栈测试完毕...")
}
