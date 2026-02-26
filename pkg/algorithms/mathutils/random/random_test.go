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
	"math"
	"testing"
)

func TestRandInt32(t *testing.T) {
	// 测试相等边界情况
	t.Run("相等边界", func(t *testing.T) {
		result := RandInt32(5, 5)
		if result != 5 {
			t.Errorf("RandInt32(5, 5) = %v, want 5", result)
		}
	})
	
	// 测试最大值边界
	t.Run("最大值边界", func(t *testing.T) {
		result := RandInt32(math.MaxInt32, math.MaxInt32)
		if result != math.MaxInt32 {
			t.Errorf("RandInt32(MaxInt32, MaxInt32) = %v, want %v", result, math.MaxInt32)
		}
	})
	
	// 测试范围验证
	t.Run("范围验证", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			low, high := int32(1), int32(10)
			result := RandInt32(low, high)
			if result < low || result > high {
				t.Errorf("RandInt32(%d, %d) = %d, 超出范围 [%d, %d]", low, high, result, low, high)
			}
		}
	})
	
	// 测试参数交换
	t.Run("参数自动交换", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			result := RandInt32(10, 1) // high < low，应该自动交换
			if result < 1 || result > 10 {
				t.Errorf("RandInt32(10, 1) = %d, 超出交换后范围 [1, 10]", result)
			}
		}
	})
	
	// 测试大范围（包括原来的边界情况）
	t.Run("大范围测试", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			result := RandInt32(0, math.MaxInt32)
			if result < 0 {
				t.Errorf("RandInt32(0, MaxInt32) = %d, 不应该为负数", result)
			}
		}
	})
}

func TestRandInt64(t *testing.T) {
	// 测试相等边界情况
	t.Run("相等边界", func(t *testing.T) {
		result := RandInt64(100, 100)
		if result != 100 {
			t.Errorf("RandInt64(100, 100) = %v, want 100", result)
		}
	})
	
	// 测试范围验证
	t.Run("范围验证", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			low, high := int64(1), int64(1000)
			result := RandInt64(low, high)
			if result < low || result > high {
				t.Errorf("RandInt64(%d, %d) = %d, 超出范围 [%d, %d]", low, high, result, low, high)
			}
		}
	})
	
	// 测试大范围
	t.Run("大范围测试", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			result := RandInt64(0, math.MaxInt64)
			if result < 0 {
				t.Errorf("RandInt64(0, MaxInt64) = %d, 不应该为负数", result)
			}
		}
	})
}