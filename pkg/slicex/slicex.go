// Package slicex.

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
package slicex

import (
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/base"
)

// 与 slices.Min 不同，本函数额外返回最小值的下标
// 保证稳定查找在切片中第一个最小的元素
// ⚠️ 若 T 包含浮点类型且切片中存在 NaN，则结果不满足全序语义（与 Go 的 <、> 比较规则一致）
func MinInSlice[T base.Ordered](s []T) (int, T) {
	idx, v, ok := MinInSliceOK[T](s)
	assert.Assert(ok, "len(s):", len(s))
	return idx, v
}

// 与 slices.Max 不同，本函数额外返回最大值的下标
// 保证稳定查找在切片中第一个最大的元素
// ⚠️ 若 T 包含浮点类型且切片中存在 NaN，则结果不满足全序语义（与 Go 的 <、> 比较规则一致）
func MaxInSlice[T base.Ordered](s []T) (int, T) {
	idx, v, ok := MaxInSliceOK[T](s)
	assert.Assert(ok, "len(s):", len(s))
	return idx, v
}

// MinInSliceOK 在切片中查找最小元素；当切片为空时返回 ok=false。
// 返回值含义：最小元素的下标、最小元素的值、是否找到（切片是否非空）
//
// ⚠️ 若 T 包含浮点类型且切片中存在 NaN，则结果不满足全序语义（与 Go 的 <、> 比较规则一致）
func MinInSliceOK[T base.Ordered](s []T) (int, T, bool) {
	return opInSliceFuncOK(s, func(a T, b T) bool {
		return a < b
	})
}

// MaxInSliceOK 在切片中查找最大元素；当切片为空时返回 ok=false。
// 返回值含义：最大元素的下标、最大元素的值、是否找到（切片是否非空）
//
// ⚠️ 若 T 包含浮点类型且切片中存在 NaN，则结果不满足全序语义（与 Go 的 <、> 比较规则一致）
func MaxInSliceOK[T base.Ordered](s []T) (int, T, bool) {
	return opInSliceFuncOK(s, func(a T, b T) bool {
		return a > b
	})
}

// MinBy 在切片中查找"最优"元素；当切片为空时触发断言。
// better(a, b) == true 表示 a 比 b 更优（应当替换当前最优值）。
//
// 该函数保证稳定：当存在多个并列"最优"元素时，返回第一个"最优"元素的下标。
func MinBy[T any](s []T, better func(a, b T) bool) (int, T) {
	idx, v, ok := MinByOK(s, better)
	assert.Assert(ok, "len(s):", len(s))
	return idx, v
}

// MaxBy 与 MinBy 语义对称：同样使用 better(a, b) 判断 a 是否优于 b；当切片为空时触发断言。
// 在实际使用中，MaxBy 常用于传入 "a > b" 或 "score(a) > score(b)" 这类规则。
//
// 该函数保证稳定：当存在多个并列"最优"元素时，返回第一个"最优"元素的下标。
func MaxBy[T any](s []T, better func(a, b T) bool) (int, T) {
	idx, v, ok := MaxByOK(s, better)
	assert.Assert(ok, "len(s):", len(s))
	return idx, v
}

// MinByOK 在切片中查找"最优"元素；当切片为空时返回 ok=false。
// 返回值含义：最优元素的下标、最优元素的值、是否找到（切片是否非空）
//
// 该函数保证稳定：当存在多个并列"最优"元素时，返回第一个"最优"元素的下标。
func MinByOK[T any](s []T, better func(a, b T) bool) (int, T, bool) {
	return opInSliceByOK(s, better)
}

// MaxByOK 在切片中查找"最优"元素；当切片为空时返回 ok=false。
// 返回值含义：最优元素的下标、最优元素的值、是否找到（切片是否非空）
//
// 该函数保证稳定：当存在多个并列"最优"元素时，返回第一个"最优"元素的下标。
func MaxByOK[T any](s []T, better func(a, b T) bool) (int, T, bool) {
	return opInSliceByOK(s, better)
}

// 特化版本，元素本身可比较
func opInSliceFuncOK[T base.Ordered](s []T, compare func(a T, b T) bool) (int, T, bool) {
	return opInSliceByOK(s, compare)
}

// 在切片中查找的通用模式(一定会遍历所有元素)，元素的比较规则由传入的better函数决定
// better函数的第一个参数为被遍历的元素，第二个参数为当前最优值
func opInSliceByOK[T any](s []T, better func(a, b T) bool) (int, T, bool) {
	if len(s) == 0 {
		var zero T
		return 0, zero, false
	}
	index := 0
	val := s[0]
	for i := 1; i < len(s); i++ {
		v := s[i]
		if better(v, val) {
			index = i
			val = v
		}
	}
	return index, val, true
}
