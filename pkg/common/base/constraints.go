// Package common.

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
// 创建日期:2023/7/27
package base

import "cmp"

// 在 Go 的 泛型类型约束（type constraints） 中，~int 里的 ~ 不是取反，也不是位运算，它有一个非常关键但容易被忽略的语义：
// ~T 表示：所有“底层类型（underlying type）是 T 的类型”。
// 也就是说：
// ~int
// 匹配的不是 只有 int，而是：
// int
// 以及 所有底层类型是 int 的自定义类型
// 例如：
// type MyInt int
// type Age int
// type Score int
// 这些类型的 underlying type 都是 int。
// 所以：
// ~int
// 允许：
// int
// MyInt
// Age
// Score
// 但 不允许
// int32
// uint
// float64
// 因为它们的 underlying type 不是 int。

// 有符号的整数
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// 无符号整数
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// 整数(包含有符号和无符号)
type Integer interface {
	Signed | Unsigned
}

// 浮点数
type Float interface {
	~float32 | ~float64
}

// 数字（包含整数和浮点数）
type Number interface {
	Integer | Float
}

// go标准库已经定义有Ordered这里直接复用
type Ordered = cmp.Ordered