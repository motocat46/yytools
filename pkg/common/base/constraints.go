// Package base.

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

// Package base provides common type constraints used across yytools.

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
package base

import "cmp"

// Signed 约束底层类型为有符号整数的类型，包含 int、int8、int16、int32、int64
// 及其底层类型相同的自定义类型（如 type MyInt int）。
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned 约束底层类型为无符号整数的类型，包含 uint、uint8、uint16、uint32、uint64、uintptr
// 及其底层类型相同的自定义类型。
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer 约束底层类型为任意整数（有符号或无符号）的类型，是 Signed 与 Unsigned 的联合。
type Integer interface {
	Signed | Unsigned
}

// Float 约束底层类型为浮点数的类型，包含 float32、float64 及其底层类型相同的自定义类型。
type Float interface {
	~float32 | ~float64
}

// Number 约束底层类型为任意数值（整数或浮点数）的类型，是 Integer 与 Float 的联合。
type Number interface {
	Integer | Float
}

// Ordered 是 cmp.Ordered 的别名，约束支持 <、<=、>、>= 运算符的类型，
// 包含所有有符号整数、无符号整数、浮点数及 string。
type Ordered = cmp.Ordered