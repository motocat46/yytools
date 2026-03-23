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

package random_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/random"
)

// ExampleRandInt 展示使用全局随机源生成闭区间 [low, high] 内的随机整数。
// 使用全局随机源，结果不可预测——此处仅验证范围正确。
func ExampleRandInt() {
	n := random.RandInt(1, 100)
	fmt.Println(n >= 1 && n <= 100)

	// 支持任意整数类型
	s := random.RandInt(int8(-10), int8(10))
	fmt.Println(s >= -10 && s <= 10)
	// Output:
	// true
	// true
}

// ExampleNewRand 展示创建固定种子的本地随机源——相同种子始终产生相同序列，适合测试复现。
func ExampleNewRand() {
	rng := random.NewRand(42)

	// 相同种子每次运行结果相同
	n1 := random.RandIntWith(rng, 0, 99)
	n2 := random.RandIntWith(rng, 0, 99)
	fmt.Println(n1 >= 0 && n1 <= 99)
	fmt.Println(n2 >= 0 && n2 <= 99)
	// Output:
	// true
	// true
}

// ExampleRandIntWith 展示使用指定随机源生成确定性随机整数——相同 seed + 相同调用序列 = 完全相同的结果。
func ExampleRandIntWith() {
	rng := random.NewRand(0)
	n := random.RandIntWith(rng, -5, 5)
	fmt.Println(n >= -5 && n <= 5)
	// Output:
	// true
}
