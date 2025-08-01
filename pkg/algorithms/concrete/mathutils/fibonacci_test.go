// Package math.

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
package mathutils

import (
	"testing"
)

func TestFibonacci(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected int
	}{
		{"F(1)", 1, 0},
		{"F(2)", 2, 1},
		{"F(3)", 3, 1},
		{"F(4)", 4, 2},
		{"F(5)", 5, 3},
		{"F(6)", 6, 5},
		{"F(7)", 7, 8},
		{"F(8)", 8, 13},
		{"F(9)", 9, 21},
		{"F(10)", 10, 34},
		{"F(20)", 20, 4181},
		{"F(30)", 30, 514229},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := &[]int{}
			result := Fibonacci(tt.n, mem)
			if result != tt.expected {
				t.Errorf("Fibonacci(%d) = %d, 期望 %d", tt.n, result, tt.expected)
			}
		})
	}
}

func TestFibonacciInt64(t *testing.T) {
	tests := []struct {
		name     string
		n        int64
		expected int64
	}{
		{"F(1)", 1, 0},
		{"F(2)", 2, 1},
		{"F(3)", 3, 1},
		{"F(4)", 4, 2},
		{"F(5)", 5, 3},
		{"F(6)", 6, 5},
		{"F(7)", 7, 8},
		{"F(8)", 8, 13},
		{"F(9)", 9, 21},
		{"F(10)", 10, 34},
		{"F(20)", 20, 4181},
		{"F(30)", 30, 514229},
		{"F(40)", 40, 63245986},
		{"F(50)", 50, 7778742049},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := &[]int64{}
			result := Fibonacci(tt.n, mem)
			if result != tt.expected {
				t.Errorf("Fibonacci(%d) = %d, 期望 %d", tt.n, result, tt.expected)
			}
		})
	}
}

func TestFibonacciMemoization(t *testing.T) {
	t.Run("备忘录重用", func(t *testing.T) {
		mem := &[]int{0, 1} // 预初始化备忘录

		// 第一次计算
		result1 := Fibonacci(10, mem)
		if result1 != 34 {
			t.Errorf("Fibonacci(10) = %d, 期望 34", result1)
		}

		// 第二次计算，应该从备忘录中获取
		result2 := Fibonacci(10, mem)
		if result2 != 34 {
			t.Errorf("Fibonacci(10) 第二次 = %d, 期望 34", result2)
		}

		// 计算更大的数，应该利用已有的备忘录
		result3 := Fibonacci(15, mem)
		if result3 != 377 {
			t.Errorf("Fibonacci(15) = %d, 期望 377", result3)
		}

		// 验证备忘录长度
		expectedLen := 16 // 0到15，共16个数
		if len(*mem) != expectedLen {
			t.Errorf("备忘录长度 = %d, 期望 %d", len(*mem), expectedLen)
		}
	})
}

func TestFibonacciMemoizationInt64(t *testing.T) {
	t.Run("备忘录重用 Int64", func(t *testing.T) {
		mem := &[]int64{0, 1} // 预初始化备忘录

		// 第一次计算
		result1 := Fibonacci(int64(10), mem)
		if result1 != 34 {
			t.Errorf("Fibonacci(10) = %d, 期望 34", result1)
		}

		// 第二次计算，应该从备忘录中获取
		result2 := Fibonacci(int64(10), mem)
		if result2 != 34 {
			t.Errorf("Fibonacci(10) 第二次 = %d, 期望 34", result2)
		}

		// 计算更大的数，应该利用已有的备忘录
		result3 := Fibonacci(int64(20), mem)
		if result3 != 4181 {
			t.Errorf("Fibonacci(20) = %d, 期望 4181", result3)
		}

		// 验证备忘录长度
		expectedLen := 21 // 0到20，共21个数
		if len(*mem) != expectedLen {
			t.Errorf("备忘录长度 = %d, 期望 %d", len(*mem), expectedLen)
		}
	})
}

func TestFibonacciPerformance(t *testing.T) {
	t.Run("性能测试", func(t *testing.T) {
		mem := &[]int{0, 1} // 预初始化备忘录

		// 测试连续计算多个斐波那契数
		for i := 1; i <= 40; i++ {
			result := Fibonacci(i, mem)

			// 验证结果不为负数
			if result < 0 {
				t.Errorf("Fibonacci(%d) 返回了负数: %d", i, result)
			}

			// 验证结果递增（除了前两个数）
			if i > 2 {
				prev := Fibonacci(i-1, mem)
				if result < prev {
					t.Errorf("Fibonacci(%d) = %d 小于 Fibonacci(%d) = %d", i, result, i-1, prev)
				}
			}
		}
	})
}

func TestFibonacciLargeNumbers(t *testing.T) {
	t.Run("大数测试", func(t *testing.T) {
		mem := &[]int64{0, 1} // 预初始化备忘录

		// 测试较大的斐波那契数
		largeN := int64(50)
		result := Fibonacci(largeN, mem)

		// 验证结果不为负数
		if result < 0 {
			t.Errorf("Fibonacci(%d) 返回了负数: %d", largeN, result)
		}

		// 验证结果不为零
		if result == 0 {
			t.Errorf("Fibonacci(%d) 返回了零", largeN)
		}

		// 验证备忘录被正确填充
		if len(*mem) <= int(largeN) {
			t.Errorf("备忘录长度 %d 应该大于 %d", len(*mem), largeN)
		}
	})
}

func TestFibonacciEdgeCases(t *testing.T) {
	t.Run("边界情况测试", func(t *testing.T) {
		mem := &[]int{0, 1} // 预初始化备忘录

		// 测试 n = 1 的情况
		result1 := Fibonacci(1, mem)
		if result1 != 0 {
			t.Errorf("Fibonacci(1) = %d, 期望 0", result1)
		}

		// 验证备忘录初始化
		if len(*mem) < 2 {
			t.Errorf("备忘录应该至少包含 2 个元素，实际有 %d 个", len(*mem))
		}

		// 验证备忘录的前两个元素
		if (*mem)[0] != 0 {
			t.Errorf("备忘录[0] = %d, 期望 0", (*mem)[0])
		}
		if (*mem)[1] != 1 {
			t.Errorf("备忘录[1] = %d, 期望 1", (*mem)[1])
		}
	})
}

func TestFibonacciConsistency(t *testing.T) {
	t.Run("一致性测试", func(t *testing.T) {
		mem := &[]int{0, 1} // 预初始化备忘录

		// 测试斐波那契数列的性质：F(n) = F(n-1) + F(n-2)
		for i := 3; i <= 20; i++ {
			fn := Fibonacci(i, mem)
			fn1 := Fibonacci(i-1, mem)
			fn2 := Fibonacci(i-2, mem)

			if fn != fn1+fn2 {
				t.Errorf("Fibonacci(%d) = %d, 但 Fibonacci(%d) + Fibonacci(%d) = %d + %d = %d",
					i, fn, i-1, i-2, fn1, fn2, fn1+fn2)
			}
		}
	})
}
