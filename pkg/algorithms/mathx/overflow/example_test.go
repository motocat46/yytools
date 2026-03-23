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

package overflow_test

import (
	"fmt"
	"math"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/overflow"
)

// ExampleMulInt 展示乘法溢出检测——返回 (结果, 是否溢出)。
// 适合需要自行处理溢出的场景。
func ExampleMulInt() {
	// 正常计算
	result, overflowed := overflow.MulInt(int32(1000), int32(2000))
	fmt.Println(result, overflowed) // 2000000 false

	// int32 溢出：1000000 * 3000 = 3e9 > MaxInt32(≈2.1e9)
	_, overflowed = overflow.MulInt(int32(1_000_000), int32(3000))
	fmt.Println(overflowed) // true
	// Output:
	// 2000000 false
	// true
}

// ExampleMulIntAssert 展示乘法溢出断言——溢出直接 panic。
// 适合"溢出即 bug"的场景，省去每次检查 bool 的代码。
func ExampleMulIntAssert() {
	result := overflow.MulIntAssert(int64(1_000_000), int64(1_000_000))
	fmt.Println(result) // 1000000000000
	// Output:
	// 1000000000000
}

// ExampleDivInt 展示除法溢出检测——唯一的溢出情形是 MinInt / -1。
func ExampleDivInt() {
	// 正常除法
	result, overflowed := overflow.DivInt(int32(100), int32(3))
	fmt.Println(result, overflowed) // 33 false

	// int32 唯一溢出：MinInt32 / -1 数学结果超出 MaxInt32
	_, overflowed = overflow.DivInt(int32(math.MinInt32), int32(-1))
	fmt.Println(overflowed) // true
	// Output:
	// 33 false
	// true
}

// ExampleAddInt 展示加法溢出检测。
func ExampleAddInt() {
	result, overflowed := overflow.AddInt(int8(100), int8(50))
	fmt.Println(result, overflowed) // -106 true（int8 最大 127）
	// Output:
	// -106 true
}
