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

import (
	"math"
	"testing"
)

func TestMulInt32(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int32
		expected int32
		overflow bool
	}{
		// 正常情况
		{"正常乘法", 2, 3, 6, false},
		{"零乘任何数", 0, 5, 0, false},
		{"任何数乘零", 5, 0, 0, false},
		{"负数乘法", -2, 3, -6, false},
		{"负数乘法2", 2, -3, -6, false},
		{"负数乘法3", -2, -3, 6, false},

		// 边界情况
		{"最大正数乘1", math.MaxInt32, 1, math.MaxInt32, false},
		{"1乘最大正数", 1, math.MaxInt32, math.MaxInt32, false},
		{"最小负数乘1", math.MinInt32, 1, math.MinInt32, false},
		{"1乘最小负数", 1, math.MinInt32, math.MinInt32, false},

		// 溢出情况
		{"正数溢出", math.MaxInt32, 2, 0, true},
		{"负数溢出", math.MinInt32, 2, 0, true},
		{"大数乘法", 1000000, 1000000, 0, true},
		{"负数大数乘法", -1000000, 1000000, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, overflow := MulInt32(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("MulInt32(%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("MulInt32(%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMulInt32Assert(t *testing.T) {
	t.Run("正常乘法", func(t *testing.T) {
		result := MulInt32Assert(2, 3)
		if result != 6 {
			t.Errorf("MulInt32Assert(2, 3) = %d, 期望 6", result)
		}
	})

	t.Run("溢出时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MulInt32Assert 在溢出时应该 panic")
			}
		}()
		MulInt32Assert(math.MaxInt32, 2)
	})
}

func TestDivInt32(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int32
		expected int32
		overflow bool
	}{
		// 正常情况
		{"正常除法", 6, 2, 3, false},
		{"负数除法", -6, 2, -3, false},
		{"负数除法2", 6, -2, -3, false},
		{"负数除法3", -6, -2, 3, false},
		{"零除任何数", 0, 5, 0, false},

		// 边界情况
		{"最大正数除1", math.MaxInt32, 1, math.MaxInt32, false},
		{"最小负数除1", math.MinInt32, 1, math.MinInt32, false},

		// 溢出情况
		{"最小负数除-1", math.MinInt32, -1, 0, true}, // 这是唯一会溢出的除法情况
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, overflow := DivInt32(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("DivInt32(%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("DivInt32(%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestDivInt32Assert(t *testing.T) {
	t.Run("正常除法", func(t *testing.T) {
		result := DivInt32Assert(6, 2)
		if result != 3 {
			t.Errorf("DivInt32Assert(6, 2) = %d, 期望 3", result)
		}
	})

	t.Run("溢出时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("DivInt32Assert 在溢出时应该 panic")
			}
		}()
		DivInt32Assert(math.MinInt32, -1)
	})
}

func TestAddInt(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
		overflow bool
	}{
		// 正常情况
		{"正常加法", 2, 3, 5, false},
		{"负数加法", -2, -3, -5, false},
		{"正负加法", 5, -3, 2, false},
		{"负正加法", -5, 3, -2, false},
		{"零加法", 0, 5, 5, false},
		{"加法零", 5, 0, 5, false},

		// 边界情况
		{"最大正数加0", math.MaxInt, 0, math.MaxInt, false},
		{"0加最大正数", 0, math.MaxInt, math.MaxInt, false},
		{"最小负数加0", math.MinInt, 0, math.MinInt, false},
		{"0加最小负数", 0, math.MinInt, math.MinInt, false},

		// 溢出情况
		{"正数溢出", math.MaxInt, 1, 0, true},
		{"负数溢出", math.MinInt, -1, 0, true},
		{"大数加法", math.MaxInt, math.MaxInt, 0, true},
		{"大负数加法", math.MinInt, math.MinInt, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, overflow := AddInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("AddInt(%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("AddInt(%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestAddIntAssert(t *testing.T) {
	t.Run("正常加法", func(t *testing.T) {
		result := AddIntAssert(2, 3)
		if result != 5 {
			t.Errorf("AddIntAssert(2, 3) = %d, 期望 5", result)
		}
	})

	t.Run("溢出时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("AddIntAssert 在溢出时应该 panic")
			}
		}()
		AddIntAssert(math.MaxInt, 1)
	})
}

func TestSubInt(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
		overflow bool
	}{
		// 正常情况
		{"正常减法", 5, 3, 2, false},
		{"负数减法", -5, -3, -2, false},
		{"正负减法", 5, -3, 8, false},
		{"负正减法", -5, 3, -8, false},
		{"零减法", 0, 5, -5, false},
		{"减法零", 5, 0, 5, false},

		// 边界情况
		{"最大正数减0", math.MaxInt, 0, math.MaxInt, false},
		{"最小负数减0", math.MinInt, 0, math.MinInt, false},

		// 溢出情况
		{"负数溢出", math.MinInt, 1, 0, true},
		{"正数溢出", math.MaxInt, -1, 0, true},
		{"大数减法", math.MinInt, math.MaxInt, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, overflow := SubInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("SubInt(%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("SubInt(%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSubIntAssert(t *testing.T) {
	t.Run("正常减法", func(t *testing.T) {
		result := SubIntAssert(5, 3)
		if result != 2 {
			t.Errorf("SubIntAssert(5, 3) = %d, 期望 2", result)
		}
	})

	t.Run("溢出时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("SubIntAssert 在溢出时应该 panic")
			}
		}()
		SubIntAssert(math.MinInt, 1)
	})
}

func TestOverflowPerformance(t *testing.T) {
	t.Run("乘法性能测试", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			a := int32(i % 1000)
			b := int32(i % 1000)
			_, overflow := MulInt32(a, b)
			// 只检查不崩溃，不检查具体结果
			_ = overflow
		}
	})

	t.Run("加法性能测试", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			a := i % 1000
			b := i % 1000
			_, overflow := AddInt(a, b)
			// 只检查不崩溃，不检查具体结果
			_ = overflow
		}
	})

	t.Run("减法性能测试", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			a := i % 1000
			b := i % 1000
			_, overflow := SubInt(a, b)
			// 只检查不崩溃，不检查具体结果
			_ = overflow
		}
	})
}

func TestOverflowEdgeCases(t *testing.T) {
	t.Run("边界值测试", func(t *testing.T) {
		// 测试各种边界值组合
		testCases := []struct {
			a, b int32
		}{
			{math.MaxInt32, math.MaxInt32},
			{math.MinInt32, math.MinInt32},
			{math.MaxInt32, math.MinInt32},
			{math.MinInt32, math.MaxInt32},
			{1, math.MaxInt32},
			{math.MaxInt32, 1},
			{-1, math.MinInt32},
			{math.MinInt32, -1},
		}

		for _, tc := range testCases {
			// 测试乘法
			_, mulOverflow := MulInt32(tc.a, tc.b)
			_ = mulOverflow

			// 测试除法（避免除零）
			if tc.b != 0 {
				_, divOverflow := DivInt32(tc.a, tc.b)
				_ = divOverflow
			}
		}
	})
}
