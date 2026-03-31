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

// Abs 返回整数 a 的绝对值。
// 当 a 为该类型的最小有符号整数（如 int8(-128)）时，-a 发生溢出回绕，
// 结果仍为最小值本身，行为与 Go 内置取负运算一致。
func Abs[T base.Integer](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

// Deprecated: Go 1.21 已内置 min() 泛型函数，新代码请直接使用内置 min()。
func Min[T base.Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

// Deprecated: Go 1.21 已内置 max() 泛型函数，新代码请直接使用内置 max()。
func Max[T base.Ordered](a T, b T) T {
	if a > b {
		return a
	}
	return b
}