// Package bits.

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
// 创建日期:2023/10/19
package bits

import (
    "github.com/stormYuanYang/yytools/pkg/common/concrete/base"
)

// 检测两个整数是否具有相反的符号
// ^(异或)相同位数上的值相同得0，不同得1
// 那么具有相反符号的两个数异或之后的结果,其最高位必然是1,也就是负数
// 那么两个数字异或的结果是负数，则说明它们的符号相反
func AreSignsOpposite[T base.Integer](a T, b T) bool {
	return (a ^ b) < 0
}

// 判断一个整数是否为2的幂
// 如果一个整数a是2的幂则意味着其二进制表示中只有一个1,其他全为0
// 那么a-1的二进制表示中，对应a的二进制表示1的所有低位全是1，高位全是0
// 例如：
// a==8,二进制表示为  1000
// a-1==7,二进制表示为0111
// a&(a-1)的结果为0000,即0
// 那么当a&(a-1)时，则意味着a是2的幂
// ⚠️注意一种情况：a==0,此时a&(a-1)==0，但0不是2的幂
func IsPowerOfTwo[T base.Integer](a T) bool {
	return a != 0 && (a&(a-1) == 0)
}

// 计算一个整数的二进制表示中，有多少位1
// a & (a-1)的结果或者说目的就是清除掉a的最低位有效位
// 例如:
// 10101&10100==10100
// 10100&10011==10000
func CountingBits[T base.Integer](a T) int {
	cnt := 0
	for ; a != 0; cnt++ {
		a &= a - 1
	}
	return cnt
}