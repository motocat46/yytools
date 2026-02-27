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
	"math"
	"testing"
)

// 用于验证派生类型能正确派发
type myInt int32

// --- RandInt 基础范围验证 ---

func TestRandInt_BasicRange(t *testing.T) {
	t.Run("int/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt(1, 10)
			if got < 1 || got > 10 {
				t.Fatalf("RandInt(1, 10) = %d, 超出 [1, 10]", got)
			}
		}
	})
	t.Run("int8/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt[int8](10, 100)
			if got < 10 || got > 100 {
				t.Fatalf("RandInt[int8](10, 100) = %d, 超出 [10, 100]", got)
			}
		}
	})
	t.Run("int16/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt[int16](0, 1000)
			if got < 0 || got > 1000 {
				t.Fatalf("RandInt[int16](0, 1000) = %d, 超出 [0, 1000]", got)
			}
		}
	})
	t.Run("int32/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			low, high := int32(100), int32(200)
			got := RandInt(low, high)
			if got < low || got > high {
				t.Fatalf("RandInt[int32](%d, %d) = %d, 超出范围", low, high, got)
			}
		}
	})
	t.Run("int64/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			low, high := int64(1000), int64(9999)
			got := RandInt(low, high)
			if got < low || got > high {
				t.Fatalf("RandInt[int64](%d, %d) = %d, 超出范围", low, high, got)
			}
		}
	})
	t.Run("uint8/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt[uint8](0, 200)
			if got > 200 {
				t.Fatalf("RandInt[uint8](0, 200) = %d, 超出 [0, 200]", got)
			}
		}
	})
	t.Run("uint32/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			low, high := uint32(100), uint32(500)
			got := RandInt(low, high)
			if got < low || got > high {
				t.Fatalf("RandInt[uint32](%d, %d) = %d, 超出范围", low, high, got)
			}
		}
	})
	t.Run("uint64/正常范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			low, high := uint64(1000), uint64(9999)
			got := RandInt(low, high)
			if got < low || got > high {
				t.Fatalf("RandInt[uint64](%d, %d) = %d, 超出范围", low, high, got)
			}
		}
	})
}

// --- 负数范围 ---

func TestRandInt_NegativeRange(t *testing.T) {
	t.Run("int/纯负数范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt(-100, -1)
			if got < -100 || got > -1 {
				t.Fatalf("RandInt(-100, -1) = %d, 超出 [-100, -1]", got)
			}
		}
	})
	t.Run("int/跨零范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt(-50, 50)
			if got < -50 || got > 50 {
				t.Fatalf("RandInt(-50, 50) = %d, 超出 [-50, 50]", got)
			}
		}
	})
	t.Run("int8/负数范围", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			got := RandInt[int8](-100, 100)
			if got < -100 || got > 100 {
				t.Fatalf("RandInt[int8](-100, 100) = %d, 超出 [-100, 100]", got)
			}
		}
	})
	t.Run("int64/跨零范围", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			got := RandInt[int64](-1, math.MaxInt64)
			if got < -1 {
				t.Fatalf("RandInt[int64](-1, MaxInt64) = %d, 超出范围", got)
			}
		}
	})
}

// --- 边界情况 ---

func TestRandInt_EqualBounds(t *testing.T) {
	t.Run("int/相等边界", func(t *testing.T) {
		if got := RandInt(7, 7); got != 7 {
			t.Fatalf("RandInt(7, 7) = %d, 期望 7", got)
		}
	})
	t.Run("int8/相等边界负数", func(t *testing.T) {
		if got := RandInt[int8](-5, -5); got != -5 {
			t.Fatalf("RandInt[int8](-5, -5) = %d, 期望 -5", got)
		}
	})
	t.Run("uint64/相等边界最大值", func(t *testing.T) {
		if got := RandInt[uint64](math.MaxUint64, math.MaxUint64); got != math.MaxUint64 {
			t.Fatalf("RandInt[uint64](MaxUint64, MaxUint64) = %d, 期望 MaxUint64", got)
		}
	})
}

// --- 极端边界（全类型范围，验证不 panic） ---

func TestRandInt_FullRange(t *testing.T) {
	t.Run("int8/全范围", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			got := RandInt[int8](math.MinInt8, math.MaxInt8)
			if got < math.MinInt8 || got > math.MaxInt8 {
				t.Fatalf("超出 int8 全范围: %d", got)
			}
		}
	})
	t.Run("int32/全范围", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			got := RandInt[int32](math.MinInt32, math.MaxInt32)
			if got < math.MinInt32 || got > math.MaxInt32 {
				t.Fatalf("超出 int32 全范围: %d", got)
			}
		}
	})
	t.Run("int64/全范围", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			_ = RandInt[int64](math.MinInt64, math.MaxInt64) // 不 panic 即可
		}
	})
	t.Run("uint64/全范围", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			_ = RandInt[uint64](0, math.MaxUint64) // 不 panic 即可
		}
	})
}

// --- 派生类型 ---

func TestRandInt_DerivedType(t *testing.T) {
	for i := 0; i < 200; i++ {
		got := RandInt[myInt](-50, 50)
		if got < -50 || got > 50 {
			t.Fatalf("RandInt[myInt](-50, 50) = %d, 超出 [-50, 50]", got)
		}
	}
}

// --- 确定性重放 ---

func TestNewRandReplay(t *testing.T) {
	t.Run("相同种子产生相同序列", func(t *testing.T) {
		rng1 := NewRand(42)
		rng2 := NewRand(42)
		for i := 0; i < 50; i++ {
			a := RandIntWith[int64](rng1, -1000, 1000)
			b := RandIntWith[int64](rng2, -1000, 1000)
			if a != b {
				t.Fatalf("第 %d 次调用结果不同: %d vs %d", i+1, a, b)
			}
		}
	})
	t.Run("不同种子产生不同序列", func(t *testing.T) {
		rng1 := NewRand(42)
		rng2 := NewRand(99)
		sameCount := 0
		for i := 0; i < 50; i++ {
			a := RandIntWith[int64](rng1, 0, 1000000)
			b := RandIntWith[int64](rng2, 0, 1000000)
			if a == b {
				sameCount++
			}
		}
		// 50 次调用中全部相同的概率极低（约 1/1000000^50）
		if sameCount == 50 {
			t.Fatal("不同种子产生了完全相同的 50 次序列，概率异常")
		}
	})
	t.Run("重放可覆盖多种类型", func(t *testing.T) {
		rng1 := NewRand(1234)
		rng2 := NewRand(1234)
		for i := 0; i < 20; i++ {
			a1 := RandIntWith[int32](rng1, -100, 100)
			b1 := RandIntWith[int32](rng2, -100, 100)
			a2 := RandIntWith[uint8](rng1, 0, 200)
			b2 := RandIntWith[uint8](rng2, 0, 200)
			if a1 != b1 || a2 != b2 {
				t.Fatalf("第 %d 次重放结果不一致: int32(%d vs %d), uint8(%d vs %d)",
					i+1, a1, b1, a2, b2)
			}
		}
	})
}

// --- assert 验证 ---

func TestRandInt_OutOfBound(t *testing.T) {
	t.Run("low > high 触发 assert", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		_ = RandInt(10, 1)
	})
	t.Run("int8/low > high", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		_ = RandInt[int8](5, -5)
	})
}

// --- 基准测试 ---

var benchSink int64

func BenchmarkRandInt_Global(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = RandInt[int64](0, 1000)
	}
}

func BenchmarkRandIntWith_Seeded(b *testing.B) {
	rng := NewRand(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = RandIntWith[int64](rng, 0, 1000)
	}
}
