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
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/common/base"
	"math"
)

// 计算a*b，并判断是否越界
// 返回值1：a*b的结果，返回值2：true->越界，false->未越界
func MulInt32(a int32, b int32) (int32, bool) {
	// 0和任何数的乘积都为0
	if a == 0 || b == 0 {
		return 0, false
	}
	res := a * b
	// 最高位不同的话，异或之后的结果应该小于0
	sign := (a ^ b) < 0
	// 小心a或者b为math.MinInt32的情况
	if sign { // 异号(一负一正) 结果一定是负数
		if a < 0 && a < math.MinInt32/b { // b > 0
			return res, true
		}
		if a > 0 && b < math.MinInt32/a { // b < 0
			return res, true
		}
	} else { // 同号(都为正，或都为负) 结果一定是正数
		limit := math.MaxInt32 / b
		if a < 0 && a < limit {
			return res, true
		}
		if a > 0 && a > limit {
			return res, true
		}
	}
	return res, false
}

// func MulInteger[T base.Integer](a T, b T) (T, bool) {
//	// 0和任何数的乘积都为0
//	if a == 0 || b == 0 {
//		return 0, false
//	}
//
//	var minInt T
//	var maxInt T
//	// 不能通过这样获取 v := reflect.Kind(low)
//	kind := reflect.ValueOf(a).Kind()
//	switch kind {
//	case reflect.Int32:
//		minInt = math.MinInt32
//		maxInt = math.MaxInt32
//	case reflect.Int64:
//		minInt = T(int64(math.MinInt64))
//		maxInt = T(int64(math.MaxInt64))
//	case reflect.Int:
//		minInt = math.MinInt
//		maxInt = math.MaxInt
//	default:
//		panic("unsupported type")
//	}
//
//	res := a*b
//	// 最高位不同的话，异或之后的结果应该小于0
//	sign := (a^b)<0
//	// 小心a或者b为math.MinInt32的情况
//	if sign { // 异号(一负一正) 结果一定是负数
//		if a < 0 && a < minInt/b {//b > 0
//			return res, true
//		}
//		if a > 0 && b < maxInt/a { // b < 0
//			return res, true
//		}
//	} else { // 同号(都为正，或都为负) 结果一定是正数
//		limit := maxInt/b
//		if a < 0 && a < limit {
//			return res, true
//		}
//		if a > 0 && a > limit {
//			return res, true
//		}
//	}
//	return res, false
// }

// 计算a*b，并进行越界断言
func MulInt32Assert(a int32, b int32) int32 {
	res, overflow := MulInt32(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}

// 计算a/b,并判断是否越界
func DivInt32(a int32, b int32) (int32, bool) {
	// 本身go语言会在除以0时，调用panic,这里不再判断
	// if b == 0 {
	//	// 除0错误，结果趋近于无穷，必然越界
	//	return math.MaxInt32, true
	// }
	
	// 负数除以负数应该是正数，但实际a/b的结果仍然是a本身
	// 因为正数最大值比负数最小值的绝对值小1
	// 根据补码规则，得到的结果仍然是a
	// 所以实际上就是溢出了
	if a == math.MinInt32 && b == -1 {
		return a, true
	}
	return a / b, false
}

func DivInt32Assert(a int32, b int32) int32 {
	res, overflow := DivInt32(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}

// 根据补码规则，对加法是否越界进行判断
// 如果越界返回true，否则返回false
func AddInt[T base.Integer](a T, b T) (T, bool) {
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

func AddIntAssert[T base.Integer](a T, b T) T {
	res, overflow := AddInt(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}

// 根据补码规则，对减法是否越界进行判断
// 如果越界返回true，否则返回false
func SubInt[T base.Integer](a T, b T) (T, bool) {
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

func SubIntAssert[T base.Integer](a T, b T) T {
	res, overflow := SubInt(a, b)
	assert.Assert(!overflow, a, b, res)
	return res
}