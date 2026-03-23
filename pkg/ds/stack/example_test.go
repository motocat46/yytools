// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stack_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/ds/stack"
)

// ExampleStack 展示基本的 LIFO 用法。
func ExampleStack() {
	s := stack.NewStack[int]()

	s.Push(1)
	s.Push(2)
	s.Push(3)

	for !s.Empty() {
		fmt.Println(s.Pop())
	}
	// Output:
	// 3
	// 2
	// 1
}

// ExampleStack_top 展示 Top 查看栈顶而不弹出。
func ExampleStack_top() {
	s := stack.NewStack[string]()

	s.Push("a")
	s.Push("b")

	fmt.Println(s.Top())    // 查看，不弹出
	fmt.Println(s.Length()) // 仍有 2 个元素
	// Output:
	// b
	// 2
}
