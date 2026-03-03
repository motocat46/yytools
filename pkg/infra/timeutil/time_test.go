// Package timeutil.

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
// 创建日期:2025/1/7
package timeutil

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		hasError bool
	}{
		// 基本测试用例
		{"空字符串", "", 0, false},
		{"零值", "0", 0, false},
		{"标准时间格式 - 秒", "30s", 30 * time.Second, false},
		{"标准时间格式 - 分钟", "5m", 5 * time.Minute, false},
		{"标准时间格式 - 小时", "2h", 2 * time.Hour, false},
		{"标准时间格式 - 组合", "1h30m15s", 1*time.Hour + 30*time.Minute + 15*time.Second, false},
		{"负数", "-30s", -30 * time.Second, false},
		{"负数组合", "-1h30m", -1*time.Hour - 30*time.Minute, false},

		// 天数测试用例
		{"1天", "1d", 24 * time.Hour, false},
		{"2天", "2d", 48 * time.Hour, false},
		{"负数天数", "-1d", -24 * time.Hour, false},
		{"天数加标准格式", "1d30m", 24*time.Hour + 30*time.Minute, false},
		{"天数加标准格式 - 负数", "-1d30m", -24*time.Hour - 30*time.Minute, false},
		{"标准格式加天数", "1d30m", 24*time.Hour + 30*time.Minute, false},

		// 错误测试用例
		{"无效格式", "invalid", 0, true},
		{"无效天数", "ad", 0, true},
		{"中间有负号", "1d-30m", 0, true},
		{"中间有正号", "1d+30m", 0, true},
		{"字符串过长", "1d" + string(make([]byte, 101)), 0, true},
		{"无效组合", "1d30x", 0, true},
		{"溢出测试", "1000000d", 0, true}, // 会导致溢出
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("ParseDuration(%q) 应该返回错误，但没有", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseDuration(%q) 返回了意外的错误: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseDuration(%q) = %v, 期望 %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestParseDurationEdgeCases(t *testing.T) {
	// 测试边界情况
	t.Run("零值变体", func(t *testing.T) {
		durations := []string{"0", "", "0s", "0m", "0h", "0d"}
		for _, d := range durations {
			result, err := ParseDuration(d)
			if err != nil {
				t.Errorf("ParseDuration(%q) 返回了错误: %v", d, err)
			}
			if result != 0 {
				t.Errorf("ParseDuration(%q) = %v, 期望 0", d, result)
			}
		}
	})

	t.Run("大数值测试", func(t *testing.T) {
		// 测试接近但不超过限制的值
		largeDay := "10000d"
		result, err := ParseDuration(largeDay)
		if err != nil {
			t.Errorf("ParseDuration(%q) 返回了错误: %v", largeDay, err)
		}
		expected := 10000 * 24 * time.Hour
		if result != expected {
			t.Errorf("ParseDuration(%q) = %v, 期望 %v", largeDay, result, expected)
		}
	})
}

func TestParseDurationPerformance(t *testing.T) {
	// 性能测试
	inputs := []string{
		"1d", "1h30m", "30s", "1d1h1m1s",
		"-1d", "-1h30m", "-30s", "-1d1h1m1s",
	}

	for i := 0; i < 1000; i++ {
		for _, input := range inputs {
			_, err := ParseDuration(input)
			if err != nil {
				t.Errorf("性能测试中 ParseDuration(%q) 返回了错误: %v", input, err)
			}
		}
	}
}
