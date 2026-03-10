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

import "github.com/motocat46/yytools/pkg/common/base"

// 计算整数绝对值(abs)
// 注意一种情况：当a是最小的整数时，-a仍然等于a
// 对最小有符号整数会发生溢出，行为与 Go 内置运算一致
func Abs[T base.Integer](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

// 计算两个数的较小值
func Min[T base.Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

// 计算两个数的较大值
func Max[T base.Ordered](a T, b T) T {
	if a > b {
		return a
	}
	return b
}