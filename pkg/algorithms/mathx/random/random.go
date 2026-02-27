// Package random.

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
	"math/rand/v2"
	"unsafe"
	
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/common/base"
)

// randSource 是内部随机源抽象
// *rand.Rand（本地实例）和全局函数包装器均实现此接口，
// 使 RandInt 与 RandIntWith 共用同一套核心逻辑。
type randSource interface {
	Int64N(int64) int64
	Uint64N(uint64) uint64
	Uint64() uint64
}

// globalSource 将 math/rand/v2 包级函数（goroutine 安全）封装为 randSource
type globalSource struct{}

func (globalSource) Int64N(n int64) int64    { return rand.Int64N(n) }
func (globalSource) Uint64N(n uint64) uint64 { return rand.Uint64N(n) }
func (globalSource) Uint64() uint64          { return rand.Uint64() }

var globalSrc randSource = globalSource{}

// RandInt 返回闭区间 [low, high] 内均匀分布的随机整数。
// 使用全局随机源（goroutine 安全，自动以 OS 熵初始化，不支持确定性重放）。
// 支持所有整数类型及其派生类型（如 type Score int8）。
// low > high 时触发 assert。
func RandInt[T base.Integer](low, high T) T {
	assert.Assert(low <= high)
	if low == high {
		return low
	}
	return randIntImpl[T](globalSrc, low, high)
}

// RandIntWith 使用指定随机源返回闭区间 [low, high] 内的随机整数。
// 配合 NewRand(seed) 可实现确定性重放：相同 seed + 相同调用序列 = 完全相同的结果。
// rng 非 goroutine 安全，请勿在多 goroutine 间共享同一实例。
func RandIntWith[T base.Integer](rng *rand.Rand, low, high T) T {
	assert.AssertFast(rng != nil)
	assert.Assert(low <= high)
	if low == high {
		return low
	}
	return randIntImpl[T](rng, low, high)
}

// NewRand 创建使用固定种子的本地随机数生成器。
// 相同 seed 始终产生完全相同的随机序列，适用于单元测试、仿真复现等场景。
// 返回的 *rand.Rand 非 goroutine 安全，需要并发使用时由调用方加锁。
func NewRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 0))
}

// randIntImpl 是内部核心实现。
// 通过 unsafe.Sizeof + signed 检测将泛型 T 派发到具体宽度的处理函数，
// 该检测在编译期即可确定，开销为零。
func randIntImpl[T base.Integer](src randSource, low, high T) T {
	var zero T
	signed := T(0)-T(1) < 0
	size := unsafe.Sizeof(zero)
	
	if signed {
		switch size {
		case 1:
			return T(signedRandN(int64(int8(low)), int64(int8(high)), src))
		case 2:
			return T(signedRandN(int64(int16(low)), int64(int16(high)), src))
		case 4:
			return T(signedRandN(int64(int32(low)), int64(int32(high)), src))
		case 8:
			return T(signedRand64(int64(low), int64(high), src))
		}
	} else {
		switch size {
		case 1:
			return T(unsignedRandN(uint64(uint8(low)), uint64(uint8(high)), src))
		case 2:
			return T(unsignedRandN(uint64(uint16(low)), uint64(uint16(high)), src))
		case 4:
			return T(unsignedRandN(uint64(uint32(low)), uint64(uint32(high)), src))
		case 8:
			return T(unsignedRand64(uint64(low), uint64(high), src))
		}
	}
	panic("unsupported integer type")
}

// signedRandN 处理 int8/int16/int32：提升到 int64 做中间计算，不可能溢出。
// 最大范围 n = MaxInt32 - MinInt32 + 1 = 4294967296，int64 可完整容纳。
func signedRandN(low, high int64, src randSource) int64 {
	n := high - low + 1
	return src.Int64N(n) + low
}

// signedRand64 处理 int64：用 uint64 2补数算术计算范围，避免有符号溢出。
// 例如 low=-1, high=MaxInt64 时，n = MaxInt64+2 有符号溢出，但 uint64 计算正确。
// 最终通过 int64(uint64(low) + r) 而非 low + int64(r) 避免加法溢出。
func signedRand64(low, high int64, src randSource) int64 {
	n := uint64(high) - uint64(low) + 1
	if n == 0 {
		// 全范围 [MinInt64, MaxInt64]：任意 uint64 位模式都是有效的 int64
		return int64(src.Uint64())
	}
	return int64(uint64(low) + src.Uint64N(n))
}

// unsignedRandN 处理 uint8/uint16/uint32：提升到 uint64 做中间计算，不可能溢出。
// 最大范围 n = MaxUint32 + 1 = 4294967296，uint64 可完整容纳。
func unsignedRandN(low, high uint64, src randSource) uint64 {
	n := high - low + 1
	return src.Uint64N(n) + low
}

// unsignedRand64 处理 uint64：检测全范围溢出（[0, MaxUint64] 时 n 溢出为 0）。
func unsignedRand64(low, high uint64, src randSource) uint64 {
	n := high - low + 1
	if n == 0 {
		// 全范围 [0, MaxUint64]
		return src.Uint64()
	}
	return src.Uint64N(n) + low
}

// RandFloat64 返回 [0.0, 1.0) 内均匀分布的随机浮点数。
func RandFloat64() float64 {
	return rand.Float64()
}