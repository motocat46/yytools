// Package tools.

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
// 创建日期:2022/6/15
package random

import (
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
	"github.com/stormYuanYang/yytools/pkg/common/concrete/base"
	"math"
	"math/rand"
	"reflect"
)

func RandSeed(seed int64) {
	rand.Seed(seed)
}

// 返回闭区间[low,high]中的某一个数
func RandInt8(low, high int8) int8 {
	return int8(RandInt32(int32(low), int32(high)))
}

func RandInt16(low, high int16) int16 {
	return int16(RandInt32(int32(low), int32(high)))
}

// 返回闭区间[low,high]中的某一个数
func RandInt32(low, high int32) int32 {
	assert.Assert(low >= 0, "invalid low:", low)
	assert.Assert(high >= 0, "invalid high:", high)
	if low == high {
		return low
	}
	if low > high {
		low, high = high, low
	}
	assert.Assert(!(low == 0 && high == math.MaxInt32), "low等于0时，high不能为最大值")
	n := high - low + 1
	return rand.Int31n(n) + low
}

// 返回闭区间[low,high]中的某一个数
func RandInt64(low, high int64) int64 {
	assert.Assert(low >= 0, "invalid low:", low)
	assert.Assert(high >= 0, "invalid high:", high)
	if low == high {
		return low
	}
	if low > high {
		low, high = high, low
	}
	assert.Assert(!(low == 0 && high == math.MaxInt64), "low等于0时，high不能为最大值")
	n := high - low + 1
	return rand.Int63n(n) + low
}

// 返回闭区间[low,high]中的某一个数
func RandInt(low, high int) int {
	assert.Assert(low >= 0, "invalid low:", low)
	assert.Assert(high >= 0, "invalid high:", high)
	if low == high {
		return low
	}
	if low > high {
		low, high = high, low
	}
	assert.Assert(!(low == 0 && high == math.MaxInt), "low等于0时，high不能为最大值")
	n := high - low + 1
	return rand.Intn(n) + low
}

// 泛型方法
// 使用了反射（reflection）来获取 low 的整数值，并根据其类型进行相应的计算和转换。
// 请注意，使用反射会带来一些性能开销，因此在需要高性能的场景中，可能需要考虑其他方式来处理范围随机数的生成。
// 此外，记得使用 rand.Seed 来设置随机数种子，以确保每次运行程序时都会获得不同的随机结果。
func RandInteger[T base.Integer](low, high T) T {
	// 不能通过这样获取 v := reflect.Kind(low)
	kind := reflect.ValueOf(low).Kind()
	switch kind {
	case reflect.Int8:
		return T(RandInt8(int8(low), int8(high)))
	case reflect.Int16:
		return T(RandInt16(int16(low), int16(high)))
	case reflect.Int32:
		return T(RandInt32(int32(low), int32(high)))
	case reflect.Int64:
		return T(RandInt64(int64(low), int64(high)))
	case reflect.Int:
		return T(RandInt(int(low), int(high)))
	default:
		panic("unsupported type")
	}
}

// 不使用泛型的实现方式
// 和上面泛型的实现方式进行比较:
// 1.第一种实现使用了泛型类型约束 [T base.Integer]，可以保证传入的参数 low 和 high 是整数类型。
// 	这样可以避免在运行时进行类型断言或反射操作，减少了额外的开销。
// 2.第一种实现使用了静态类型判断，根据 reflect.ValueOf(low).Kind() 的结果直接选择对应的函数。
//	这样可以在编译时确定具体的类型分支，避免了动态类型检查和转换的开销。
// 3.反射操作 reflect.ValueOf(low).Kind() 在运行时会带来一定的性能开销，包括类型转换和动态类型检查。
// 综上所述，尽管第一种实现涉及一次反射操作，但由于使用了静态类型约束和静态类型判断，以及避免了类型断言和动态类型检查的开销，其效率更高。
func RandInteger1(low, high interface{}) interface{} {
	switch low := low.(type) {
	case int8:
		return RandInt8(low, high.(int8))
	case int16:
		return RandInt16(low, high.(int16))
	case int32:
		return RandInt32(low, high.(int32))
	case int64:
		return RandInt64(low, high.(int64))
	case int:
		return RandInt(low, high.(int))
	default:
		panic("unsupported type")
	}
}

func RandFloat32() float32 {
	return rand.Float32()
}

func RandFloat64() float64 {
	return rand.Float64()
}