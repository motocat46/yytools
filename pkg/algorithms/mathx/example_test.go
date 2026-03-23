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

package mathx_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/algorithms/mathx"
)

// ExampleGcd 展示求最大公约数——循环迭代实现，O(log(min(x,y))) 时间。
func ExampleGcd() {
	fmt.Println(mathx.Gcd(12, 8))   // 4
	fmt.Println(mathx.Gcd(100, 75)) // 25
	fmt.Println(mathx.Gcd(7, 0))    // 7（任意数与 0 的最大公约数为自身）
	// Output:
	// 4
	// 25
	// 7
}

// ExampleAbs 展示泛型整数绝对值。
// 注意：对最小有符号整数（如 math.MinInt8）取反会溢出，行为与 Go 内置运算一致。
func ExampleAbs() {
	fmt.Println(mathx.Abs(-5))   // 5
	fmt.Println(mathx.Abs(3))    // 3
	fmt.Println(mathx.Abs[int64](-1000000)) // 1000000
	// Output:
	// 5
	// 3
	// 1000000
}

// ExampleFibN 展示通过静态查找表 O(1) 获取第 n 个斐波那契数。
// F(0)=0, F(1)=1, F(2)=1, F(3)=2, ...
func ExampleFibN() {
	fmt.Println(mathx.FibN[int64](0))  // F(0) = 0
	fmt.Println(mathx.FibN[int64](10)) // F(10) = 55
	fmt.Println(mathx.FibN[int64](20)) // F(20) = 6765
	// Output:
	// 0
	// 55
	// 6765
}

// ExampleFibNMax 展示各整数类型支持的最大斐波那契下标。
func ExampleFibNMax() {
	fmt.Println(mathx.FibNMax[int8]())  // 11（F(11)=89 ≤ MaxInt8=127）
	fmt.Println(mathx.FibNMax[int64]()) // 92（F(92) ≤ MaxInt64）
	// Output:
	// 11
	// 92
}

// ExampleFibonacci_Calculate 展示带备忘录的斐波那契计算——支持较大值，重复调用时直接命中缓存。
// n 从 1 开始，Calculate(n) 返回 F(n-1)（即第 n 个位置的斐波那契数）。
func ExampleFibonacci_Calculate() {
	fib := mathx.NewFibMem[int64]()
	fmt.Println(fib.Calculate(1))  // F(0) = 0
	fmt.Println(fib.Calculate(2))  // F(1) = 1
	fmt.Println(fib.Calculate(10)) // F(9) = 34
	// Output:
	// 0
	// 1
	// 34
}
