// Package overflow.

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
// 创建日期:2023/10/13
package overflow

// 提供对常见四则运算(加减乘除)的计算和越界检查方法

import (
	"unsafe"

	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/base"
)

// 计算a*b，并判断是否越界
// 返回值1：a*b的结果，返回值2：true->越界，false->未越界
func MulInt[T base.Signed](a, b T) (T, bool) {
	// 0和任何数的乘积都为0
	if a == 0 || b == 0 {
		return 0, false
	}
	size := unsafe.Sizeof(a)
	bits := uint(size) * 8
	if size < 8 {
		// int8/int16/int32：提升为 int64 做中间计算，检查结果是否超出 T 的范围
		r := int64(a) * int64(b)
		minT := -(int64(1) << (bits - 1))
		maxT := int64(1)<<(bits-1) - 1
		if r < minT || r > maxT {
			return T(r), true
		}
		return T(r), false
	}
	// int64/int（64位平台）：符号感知的除法边界检查
	res := a * b
	// 最高位不同的话，异或之后的结果应该小于0
	sign := (a ^ b) < 0
	minT := T(1) << (bits - 1) // 补码：T(1)<<(bits-1) 即为 MinT
	maxT := ^minT
	if sign { // 异号（一负一正），结果一定是负数
		if a < 0 && a < minT/b { // b > 0
			return res, true
		}
		if a > 0 && b < minT/a { // b < 0
			return res, true
		}
	} else { // 同号（都为正，或都为负），结果一定是正数
		limit := maxT / b
		if a < 0 && a < limit {
			return res, true
		}
		if a > 0 && a > limit {
			return res, true
		}
	}
	return res, false
}

// 计算a*b，并进行越界断言
func MulIntAssert[T base.Signed](a, b T) T {
	res, overflow := MulInt(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}

// 计算a/b，并判断是否越界
// 整数除法唯一的溢出情形是 MinT / -1（数学结果超出正数最大值）
// 返回值1：a/b的结果，返回值2：true->越界，false->未越界
func DivInt[T base.Signed](a, b T) (T, bool) {
	// 本身go语言会在除以0时，调用panic，这里不再判断
	bits := uint(unsafe.Sizeof(a)) * 8
	minT := T(1) << (bits - 1) // 补码规则：T(1)<<(bits-1) 即为 MinT
	// 负数除以负数应该是正数，但实际a/b的结果仍然是a本身
	// 因为正数最大值比负数最小值的绝对值小1
	// 根据补码规则，得到的结果仍然是a，实际上就是溢出了
	if a == minT && b == -1 {
		return a, true
	}
	return a / b, false
}

// 计算a/b，并进行越界断言
func DivIntAssert[T base.Signed](a, b T) T {
	res, overflow := DivInt(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}

// 根据补码规则，对加法是否越界进行判断
// 如果越界返回true，否则返回false
func AddInt[T base.Signed](a T, b T) (T, bool) {
	sum := a + b
	// 当a为非负数，b为非负数，此时a+b小于0，则越界
	if a >= 0 && b >= 0 && sum < 0 {
		return sum, true
	}
	// 当a为负数，b为负数，此时a+b大于等于0，则越界
	if a < 0 && b < 0 && sum >= 0 {
		return sum, true
	}
	return sum, false
}

func AddIntAssert[T base.Signed](a T, b T) T {
	res, overflow := AddInt(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}

// 根据补码规则，对减法是否越界进行判断
// 如果越界返回true，否则返回false
func SubInt[T base.Signed](a T, b T) (T, bool) {
	// 当a为负数，b为整数，此时a-b大于等于0的话，就会越界
	res := a - b
	if a < 0 && b > 0 && res >= 0 {
		return res, true
	}
	// 当a为非负数，b为负数，此时a-b小于0，则会越界
	if a >= 0 && b < 0 && res < 0 {
		return res, true
	}
	// 其他情况不会越界
	return res, false
}

func SubIntAssert[T base.Signed](a T, b T) T {
	res, overflow := SubInt(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}
