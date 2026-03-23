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

package bits_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/bits"
)

// ExampleAreSignsOpposite 展示通过异或最高位检测两个整数是否符号相反。
func ExampleAreSignsOpposite() {
	fmt.Println(bits.AreSignsOpposite(3, -5))  // true
	fmt.Println(bits.AreSignsOpposite(-2, -8)) // false（同负）
	fmt.Println(bits.AreSignsOpposite(4, 7))  // false（同正）
	fmt.Println(bits.AreSignsOpposite(0, -1)) // true（0 的符号位为 0，-1 的符号位为 1，两者相反）
	// Output:
	// true
	// false
	// false
	// true
}

// ExampleIsPowerOfTwo 展示通过位与技巧 a&(a-1)==0 判断 2 的幂。
func ExampleIsPowerOfTwo() {
	fmt.Println(bits.IsPowerOfTwo(8))  // true  (0b1000)
	fmt.Println(bits.IsPowerOfTwo(16)) // true  (0b10000)
	fmt.Println(bits.IsPowerOfTwo(6))  // false (0b0110，有两个 1-bit)
	fmt.Println(bits.IsPowerOfTwo(0))  // false（0 不是 2 的幂）
	// Output:
	// true
	// true
	// false
	// false
}

// ExampleCountingBits 展示统计整数二进制补码中 1-bit 的数量（Brian Kernighan 算法）。
func ExampleCountingBits() {
	fmt.Println(bits.CountingBits(7))  // 3（0b0111）
	fmt.Println(bits.CountingBits(10)) // 2（0b1010）
	fmt.Println(bits.CountingBits(0))  // 0
	// Output:
	// 3
	// 2
	// 0
}
