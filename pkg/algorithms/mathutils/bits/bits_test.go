// Package bits.

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
// 创建日期:2023/10/19
package bits

import (
	"testing"
)

func TestAreSignsOpposite(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected bool
	}{
		// 正数和负数
		{"正数负数", 5, -3, true},
		{"负数正数", -5, 3, true},
		{"正数正数", 5, 3, false},
		{"负数负数", -5, -3, false},
		{"零正数", 0, 5, false},
		{"零负数", 0, -5, true},
		{"正数零", 5, 0, false},
		{"负数零", -5, 0, true},
		{"零零", 0, 0, false},
		// 边界值测试
		{"最大正数最小负数", 2147483647, -2147483648, true},
		{"最小负数最大正数", -2147483648, 2147483647, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AreSignsOpposite(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("AreSignsOpposite(%d, %d) = %t, 期望 %t", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestAreSignsOppositeInt64(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int64
		expected bool
	}{
		// 正数和负数
		{"正数负数", 5, -3, true},
		{"负数正数", -5, 3, true},
		{"正数正数", 5, 3, false},
		{"负数负数", -5, -3, false},
		// 边界值测试
		{"最大正数最小负数", 9223372036854775807, -9223372036854775808, true},
		{"最小负数最大正数", -9223372036854775808, 9223372036854775807, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AreSignsOpposite(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("AreSignsOpposite(%d, %d) = %t, 期望 %t", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestIsPowerOfTwo(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected bool
	}{
		// 2的幂
		{"2^0", 1, true},
		{"2^1", 2, true},
		{"2^2", 4, true},
		{"2^3", 8, true},
		{"2^4", 16, true},
		{"2^5", 32, true},
		{"2^10", 1024, true},
		{"2^20", 1048576, true},
		{"2^30", 1073741824, true},

		// 不是2的幂
		{"0", 0, false},
		{"3", 3, false},
		{"5", 5, false},
		{"6", 6, false},
		{"7", 7, false},
		{"9", 9, false},
		{"10", 10, false},
		{"15", 15, false},
		{"负数", -4, false},
		{"负数", -8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPowerOfTwo(tt.input)
			if result != tt.expected {
				t.Errorf("IsPowerOfTwo(%d) = %t, 期望 %t", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPowerOfTwoInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected bool
	}{
		// 2的幂
		{"2^0", 1, true},
		{"2^1", 2, true},
		{"2^10", 1024, true},
		{"2^30", 1073741824, true},
		{"2^40", 1099511627776, true},
		{"2^50", 1125899906842624, true},
		{"2^60", 1152921504606846976, true},

		// 不是2的幂
		{"0", 0, false},
		{"3", 3, false},
		{"负数", -4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPowerOfTwo(tt.input)
			if result != tt.expected {
				t.Errorf("IsPowerOfTwo(%d) = %t, 期望 %t", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountingBits(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"0", 0, 0},
		{"1", 1, 1},
		{"2", 2, 1},
		{"3", 3, 2},
		{"4", 4, 1},
		{"5", 5, 2},
		{"6", 6, 2},
		{"7", 7, 3},
		{"8", 8, 1},
		{"9", 9, 2},
		{"10", 10, 2},
		{"15", 15, 4},
		{"16", 16, 1},
		{"255", 255, 8},
		{"256", 256, 1},
		{"1023", 1023, 10},
		{"1024", 1024, 1},
		// 负数测试
		{"-1", -1, 64}, // 在64位系统上，-1的所有位都是1
		{"-2", -2, 63},
		{"-3", -3, 63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountingBits(tt.input)
			if result != tt.expected {
				t.Errorf("CountingBits(%d) = %d, 期望 %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountingBitsInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int
	}{
		{"0", 0, 0},
		{"1", 1, 1},
		{"255", 255, 8},
		{"1023", 1023, 10},
		{"9223372036854775807", 9223372036854775807, 63}, // 最大正数
		{"-1", -1, 64}, // 在64位系统上，-1的所有位都是1
		{"-2", -2, 63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountingBits(tt.input)
			if result != tt.expected {
				t.Errorf("CountingBits(%d) = %d, 期望 %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountingBitsPerformance(t *testing.T) {
	// 性能测试
	t.Run("大量计算", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			result := CountingBits(i)
			// 验证结果不为负数
			if result < 0 {
				t.Errorf("CountingBits(%d) 返回了负数: %d", i, result)
			}
			// 验证结果不超过位数
			if result > 64 {
				t.Errorf("CountingBits(%d) 返回了超过位数的值: %d", i, result)
			}
		}
	})
}

func TestCountingBitsEdgeCases(t *testing.T) {
	t.Run("边界值测试", func(t *testing.T) {
		// 测试各种边界值
		testCases := []int{
			0, 1, -1, 2147483647, -2147483648, // int32 边界
		}

		for _, input := range testCases {
			result := CountingBits(input)
			if result < 0 {
				t.Errorf("CountingBits(%d) 返回了负数: %d", input, result)
			}
		}
	})
}

func TestBitsOperationsConsistency(t *testing.T) {
	t.Run("一致性测试", func(t *testing.T) {
		// 测试位操作的一致性
		for i := 1; i <= 1000; i++ {
			// 如果 i 是 2 的幂，那么它应该只有一个位为 1
			if IsPowerOfTwo(i) {
				bitCount := CountingBits(i)
				if bitCount != 1 {
					t.Errorf("数字 %d 是 2 的幂，但 CountingBits 返回 %d", i, bitCount)
				}
			}
		}
	})
}
