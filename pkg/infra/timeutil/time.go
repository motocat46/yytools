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
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/overflow"
)

// 在标准库基础上，支持更大的日期单位——d（日）
// 最大时长近似290年（int64 纳秒上限）。
//
// 支持负数 day："-2d1h30m" 表示向前推 2 天 1 小时 30 分钟（即 -49.5 小时）。
// 正负号只允许出现在字符串开头，中间不允许再出现 + 或 -。
func ParseDuration(s string) (time.Duration, error) {
	if s == "" || s == "0" {
		return 0, nil
	}
	if err := validateDuration(s); err != nil {
		return 0, err
	}

	s = strings.ToLower(s)
	index := strings.IndexRune(s, 'd')
	if index == -1 {
		return time.ParseDuration(s)
	}

	left, right := s[:index], s[index+1:]
	daysPart, err := parseDaysPart(left)
	if err != nil {
		return 0, fmt.Errorf("ParseDuration(%q): %w", s, err)
	}
	d, err := combineWithRemain(daysPart, strings.HasPrefix(left, "-"), right)
	if err != nil {
		return 0, fmt.Errorf("ParseDuration(%q): %w", s, err)
	}
	return d, nil
}

// parseDaysPart 将天数字符串（如 "1.5"、"-2"）转换为纳秒整数。
func parseDaysPart(left string) (int64, error) {
	daysFloat, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return 0, err
	}
	// float64(math.MaxInt64) 因精度向上取整为 2^63，故用 >= 而非 >
	daysDur := daysFloat * float64(24*time.Hour)
	if math.IsInf(daysDur, 0) || math.IsNaN(daysDur) ||
		daysDur >= float64(math.MaxInt64) || daysDur < float64(math.MinInt64) {
		return 0, fmt.Errorf("天数溢出 int64 范围")
	}
	return int64(daysDur), nil
}

// combineWithRemain 将天数纳秒与右侧剩余时长（如 "1h30m"）合并。
// negative 为 true 时结果为 -(|days| + remain)。
func combineWithRemain(daysPart int64, negative bool, right string) (time.Duration, error) {
	if len(right) == 0 {
		return time.Duration(daysPart), nil
	}
	remain, err := time.ParseDuration(right)
	if err != nil {
		return 0, err
	}
	// 负天：结果 = daysPart - remain（daysPart 已为负，再减去正的 remain）
	// 正天：结果 = daysPart + remain
	if negative {
		result, ovf := overflow.SubInt(daysPart, int64(remain))
		if ovf {
			return 0, fmt.Errorf("合并溢出 int64 范围")
		}
		return time.Duration(result), nil
	}
	result, ovf := overflow.AddInt(daysPart, int64(remain))
	if ovf {
		return 0, fmt.Errorf("合并溢出 int64 范围")
	}
	return time.Duration(result), nil
}

// validateDuration 检查长度上限与符号位置约束。
func validateDuration(s string) error {
	if len(s) > 100 {
		return fmt.Errorf("timeutil string s is too big:%s", s)
	}
	// 正负号只允许出现在第一个字符，中间出现会产生歧义
	for i := 1; i < len(s); i++ {
		if s[i] == '-' || s[i] == '+' {
			return fmt.Errorf("invalid remain timeutil s:%s", s)
		}
	}
	return nil
}
