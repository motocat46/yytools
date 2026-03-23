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

package slicex_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/slicex"
)

// ExampleMinInSlice 展示在有序/无序切片中查找最小值及其下标。
func ExampleMinInSlice() {
	nums := []int{3, 1, 4, 1, 5, 9, 2}
	idx, val := slicex.MinInSlice(nums)
	fmt.Println(idx, val) // 1 1（第一个最小值）
	// Output:
	// 1 1
}

// ExampleMaxInSlice 展示查找最大值及其下标。
func ExampleMaxInSlice() {
	nums := []int{3, 1, 4, 1, 5, 9, 2}
	idx, val := slicex.MaxInSlice(nums)
	fmt.Println(idx, val) // 5 9
	// Output:
	// 5 9
}

// ExampleMinBy 展示按自定义规则查找"最优"元素。
// better(a, b) == true 表示 a 优于 b（应当替换当前最优）。
func ExampleMinBy() {
	type Player struct {
		Name  string
		Score int
	}
	players := []Player{
		{"alice", 1500},
		{"bob", 800},
		{"carol", 1200},
	}

	// 找积分最低的玩家
	idx, p := slicex.MinBy(players, func(a, b Player) bool {
		return a.Score < b.Score // a 积分更低则 a 更"优"
	})
	fmt.Println(idx, p.Name, p.Score)
	// Output:
	// 1 bob 800
}

// ExampleMaxBy 展示按自定义规则查找最高分玩家。
func ExampleMaxBy() {
	type Player struct {
		Name  string
		Score int
	}
	players := []Player{
		{"alice", 1500},
		{"bob", 800},
		{"carol", 1800},
	}

	idx, p := slicex.MaxBy(players, func(a, b Player) bool {
		return a.Score > b.Score
	})
	fmt.Println(idx, p.Name)
	// Output:
	// 2 carol
}

// ExampleMinInSliceOK 展示空切片安全版本——返回 ok=false 而非 panic。
func ExampleMinInSliceOK() {
	var empty []int
	_, _, ok := slicex.MinInSliceOK(empty)
	fmt.Println(ok) // false，不 panic
	// Output:
	// false
}
