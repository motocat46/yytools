// Package mathx.

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
package mathx

import (
	"github.com/motocat46/yytools/pkg/common/base"
)

/*
 GcdR-->greatest common divisor recursion
 求最大公约数(欧几里得算法)递归实现
 计算两个非负整数x和y的最大公约数:若y是0,则最大公约数为x;
 否则,将x除以y得到余数r,x和y的最大公约数即为y和r的最大公约数.

 ⚠️ 注意,这是一个递归函数.
 时间复杂度: O(log(min(x, y)))
 空间复杂度: O(log(min(x, y)))  // 递归栈
*/
func GcdR[T base.Integer](x, y T) T {
	if x < 0 || y < 0 {
		panic("GcdR: 参数必须为非负整数")
	}
	if y == 0 {
		return x
	}
	r := x % y
	return GcdR[T](y, r)
}

/*
 GcdI-->greatest common divisor iterate
 求最大公约数(欧几里得算法)循环遍历实现
 计算两个非负整数x和y的最大公约数:若y是0,则最大公约数为x;
 否则,将x除以y得到余数r,x和y的最大公约数即为y和r的最大公约数.

 时间复杂度: O(log(min(x, y)))
 空间复杂度: O(1)
*/
func GcdI[T base.Integer](x, y T) T {
	if x < 0 || y < 0 {
		panic("GcdI: 参数必须为非负整数")
	}
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
	return GcdI[T](x, y)
}