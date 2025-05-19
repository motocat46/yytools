// Package mathutils.

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
// 创建日期:2022/12/7
package mathutils

import (
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
	"github.com/stormYuanYang/yytools/pkg/common/concrete/base"
)

/*
 GcdR-->greatest common divisor recursion
 求最大公约数(欧几里得算法)递归实现
 计算两个非负整数x和y的最大公约数:若y是0,则最大公约数为x;
 否则,将x除以y得到余数r,x和y的最大公约数即为y和r的最大公约数.

 ⚠️ 注意,这是一个递归函数.
 时间复杂度: ? 直觉一般是常量到对数之间
 空间复杂度: O(1)
*/
func GcdR[T base.Integer](x, y T) T {
	assert.Assert(x >= 0, "x must >= 0, x:", x)
	assert.Assert(y >= 0, "y must >= 0, y:", y)
	if y == 0 {
		return x
	}
	r := x % y
	return GcdR(y, r)
}

/*
 GcdI-->greatest common divisor iterate
 求最大公约数(欧几里得算法)循环遍历实现
 计算两个非负整数x和y的最大公约数:若y是0,则最大公约数为x;
 否则,将x除以y得到余数r,x和y的最大公约数即为y和r的最大公约数.

 时间复杂度: ? 直觉一般是常量到对数之间
 空间复杂度: O(1)
*/
func GcdI[T base.Integer](x, y T) T {
	assert.Assert(x >= 0, "x must >= 0, x:", x)
	assert.Assert(y >= 0, "y must >= 0, y:", y)
	for {
		if y == 0 {
			return x
		}
		r := x % y
		x = y
		y = r
	}
}

func Gcd[T base.Integer](x, y T) T {
	// 遍历(GcdI)比递归(GcdR)效率更高
	// 这里采用遍历实现的方式获得最大公约数
	return GcdI(x, y)
}

// 计算整数绝对值(abs)
// 注意一种情况：当a是最小的整数时，-a仍然等于a
func Abs[T base.Integer](a T) T {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

// 计算整数的较小值
func Min[T base.Integer](a T, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

// 计算整数的较大值
func Max[T base.Integer](a T, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}