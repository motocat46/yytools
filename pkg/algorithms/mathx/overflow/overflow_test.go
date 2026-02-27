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

func TestMulInt(t *testing.T) {
	// int8 用例
	int8Tests := []struct {
		name     string
		a, b     int8
		expected int8
		overflow bool
	}{
		{"正常乘法", 2, 3, 6, false},
		{"零乘任何数", 0, 100, 0, false},
		{"任何数乘零", 100, 0, 0, false},
		{"负数乘法", -2, 3, -6, false},
		{"负数乘法2", 2, -3, -6, false},
		{"负数乘法3", -2, -3, 6, false},
		{"MaxInt8乘1", math.MaxInt8, 1, math.MaxInt8, false},
		{"MinInt8乘1", math.MinInt8, 1, math.MinInt8, false},
		{"正数溢出", math.MaxInt8, 2, 0, true},
		{"负数溢出", math.MinInt8, 2, 0, true},
	}
	for _, tt := range int8Tests {
		t.Run("int8_"+tt.name, func(t *testing.T) {
			result, overflow := MulInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("MulInt[int8](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("MulInt[int8](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// int16 用例
	int16Tests := []struct {
		name     string
		a, b     int16
		expected int16
		overflow bool
	}{
		{"正常乘法", 100, 200, 20000, false},
		{"零乘任何数", 0, 1000, 0, false},
		{"负数乘法", -100, 200, -20000, false},
		{"MaxInt16乘1", math.MaxInt16, 1, math.MaxInt16, false},
		{"MinInt16乘1", math.MinInt16, 1, math.MinInt16, false},
		{"正数溢出", math.MaxInt16, 2, 0, true},
		{"负数溢出", math.MinInt16, 2, 0, true},
	}
	for _, tt := range int16Tests {
		t.Run("int16_"+tt.name, func(t *testing.T) {
			result, overflow := MulInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("MulInt[int16](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("MulInt[int16](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// int32 用例（等价原 TestMulInt32 全部用例）
	int32Tests := []struct {
		name     string
		a, b     int32
		expected int32
		overflow bool
	}{
		{"正常乘法", 2, 3, 6, false},
		{"零乘任何数", 0, 5, 0, false},
		{"任何数乘零", 5, 0, 0, false},
		{"负数乘法", -2, 3, -6, false},
		{"负数乘法2", 2, -3, -6, false},
		{"负数乘法3", -2, -3, 6, false},
		{"MaxInt32乘1", math.MaxInt32, 1, math.MaxInt32, false},
		{"1乘MaxInt32", 1, math.MaxInt32, math.MaxInt32, false},
		{"MinInt32乘1", math.MinInt32, 1, math.MinInt32, false},
		{"1乘MinInt32", 1, math.MinInt32, math.MinInt32, false},
		{"正数溢出", math.MaxInt32, 2, 0, true},
		{"负数溢出", math.MinInt32, 2, 0, true},
		{"大数乘法", 1000000, 1000000, 0, true},
		{"负数大数乘法", -1000000, 1000000, 0, true},
	}
	for _, tt := range int32Tests {
		t.Run("int32_"+tt.name, func(t *testing.T) {
			result, overflow := MulInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("MulInt[int32](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("MulInt[int32](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// int64 用例
	int64Tests := []struct {
		name     string
		a, b     int64
		expected int64
		overflow bool
	}{
		{"正常乘法", 2, 3, 6, false},
		{"零乘任何数", 0, math.MaxInt64, 0, false},
		{"任何数乘零", math.MaxInt64, 0, 0, false},
		{"负数乘法", -2, 3, -6, false},
		{"负数乘法2", 2, -3, -6, false},
		{"负数乘法3", -2, -3, 6, false},
		{"MaxInt64乘1", math.MaxInt64, 1, math.MaxInt64, false},
		{"MinInt64乘1", math.MinInt64, 1, math.MinInt64, false},
		{"正数溢出", math.MaxInt64, 2, 0, true},
		{"负数溢出", math.MinInt64, 2, 0, true},
		{"大数乘法溢出", 1000000000, 10000000000, 0, true},
		{"MinInt64乘-1溢出", math.MinInt64, -1, 0, true},
	}
	for _, tt := range int64Tests {
		t.Run("int64_"+tt.name, func(t *testing.T) {
			result, overflow := MulInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("MulInt[int64](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("MulInt[int64](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMulIntAssert(t *testing.T) {
	t.Run("正常乘法", func(t *testing.T) {
		result := MulIntAssert(int32(2), int32(3))
		if result != 6 {
			t.Errorf("MulIntAssert[int32](2, 3) = %d, 期望 6", result)
		}
	})

	t.Run("溢出时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MulIntAssert 在溢出时应该 panic")
			}
		}()
		MulIntAssert(int32(math.MaxInt32), int32(2))
	})
}

func TestDivInt(t *testing.T) {
	// int8 用例
	int8Tests := []struct {
		name     string
		a, b     int8
		expected int8
		overflow bool
	}{
		{"正常除法", 6, 2, 3, false},
		{"负数除法", -6, 2, -3, false},
		{"负数除法2", 6, -2, -3, false},
		{"负数除法3", -6, -2, 3, false},
		{"零除任何数", 0, 5, 0, false},
		{"MaxInt8除1", math.MaxInt8, 1, math.MaxInt8, false},
		{"MinInt8除1", math.MinInt8, 1, math.MinInt8, false},
		{"MinInt8除-1溢出", math.MinInt8, -1, 0, true},
	}
	for _, tt := range int8Tests {
		t.Run("int8_"+tt.name, func(t *testing.T) {
			result, overflow := DivInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("DivInt[int8](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("DivInt[int8](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// int32 用例（等价原 TestDivInt32 全部用例）
	int32Tests := []struct {
		name     string
		a, b     int32
		expected int32
		overflow bool
	}{
		{"正常除法", 6, 2, 3, false},
		{"负数除法", -6, 2, -3, false},
		{"负数除法2", 6, -2, -3, false},
		{"负数除法3", -6, -2, 3, false},
		{"零除任何数", 0, 5, 0, false},
		{"MaxInt32除1", math.MaxInt32, 1, math.MaxInt32, false},
		{"MinInt32除1", math.MinInt32, 1, math.MinInt32, false},
		{"MinInt32除-1溢出", math.MinInt32, -1, 0, true},
	}
	for _, tt := range int32Tests {
		t.Run("int32_"+tt.name, func(t *testing.T) {
			result, overflow := DivInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("DivInt[int32](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("DivInt[int32](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// int64 用例
	int64Tests := []struct {
		name     string
		a, b     int64
		expected int64
		overflow bool
	}{
		{"正常除法", 100, 7, 14, false},
		{"负数除法", -100, 7, -14, false},
		{"零除任何数", 0, math.MaxInt64, 0, false},
		{"MaxInt64除1", math.MaxInt64, 1, math.MaxInt64, false},
		{"MinInt64除1", math.MinInt64, 1, math.MinInt64, false},
		{"MinInt64除-1溢出", math.MinInt64, -1, 0, true},
	}
	for _, tt := range int64Tests {
		t.Run("int64_"+tt.name, func(t *testing.T) {
			result, overflow := DivInt(tt.a, tt.b)
			if overflow != tt.overflow {
				t.Errorf("DivInt[int64](%d, %d) overflow = %t, 期望 %t", tt.a, tt.b, overflow, tt.overflow)
			}
			if !tt.overflow && result != tt.expected {
				t.Errorf("DivInt[int64](%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestDivIntAssert(t *testing.T) {
	t.Run("正常除法", func(t *testing.T) {
		result := DivIntAssert(int32(6), int32(2))
		if result != 3 {
			t.Errorf("DivIntAssert[int32](6, 2) = %d, 期望 3", result)
		}
	})

	t.Run("溢出时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("DivIntAssert 在溢出时应该 panic")
			}
		}()
		DivIntAssert(int32(math.MinInt32), int32(-1))
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
			_, overflow := MulInt(a, b)
			_ = overflow
		}
	})

	t.Run("加法性能测试", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			a := i % 1000
			b := i % 1000
			_, overflow := AddInt(a, b)
			_ = overflow
		}
	})

	t.Run("减法性能测试", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			a := i % 1000
			b := i % 1000
			_, overflow := SubInt(a, b)
			_ = overflow
		}
	})
}

func TestOverflowEdgeCases(t *testing.T) {
	t.Run("边界值测试", func(t *testing.T) {
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
			_, mulOverflow := MulInt(tc.a, tc.b)
			_ = mulOverflow

			if tc.b != 0 {
				_, divOverflow := DivInt(tc.a, tc.b)
				_ = divOverflow
			}
		}
	})
}
