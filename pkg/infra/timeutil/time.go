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
	"strconv"
	"strings"
	"time"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/overflow"
)

// 在标准库基础上，支持更大的日期单位——d（日）
// 最大时长近似290年
func ParseDuration(s string) (time.Duration, error) {
	// 支持特殊格式：空或者0，都返回0
	if s == "" || s == "0" {
		return 0, nil
	}

	// 这里额外做个长度判断(理论上输入的时间长度字符串不会超过这个长度)
	// 避免处理错误的输入数据
	if len(s) > 100 {
		return 0, fmt.Errorf("timeutil string s is too big:%s", s)
	}

	// 因为在库函数的基础上增加了天数的支持,就要在判断一下字符串中间是否还有正负号
	// 剩下的检查交给库函数
	for i := 1; i < len(s); i++ {
		// -和+只允许出现在第一个字符
		if s[i] == '-' || s[i] == '+' {
			return 0, fmt.Errorf("invalid remain timeutil s:%s", s)
		}
	}

	// 处理一下，将可能的大写字母转换成小写字母(增加容错率)
	s = strings.ToLower(s)
	// 判断一下是否有天的时间单位
	index := strings.IndexRune(s, 'd')
	if index == -1 {
		// 没有单位'd'，则直接调用系统原生函数
		return time.ParseDuration(s)
	}

	left, right := s[:index], s[index+1:]
	days, err := strconv.Atoi(left)
	if err != nil {
		return 0, err
	}

	daysPart, ovf := overflow.MulInt(int64(time.Hour)*24, int64(days))
	if ovf {
		return 0, fmt.Errorf("invalid duration overflow:%s", s)
	}

	if len(right) == 0 {
		return time.Duration(daysPart), nil
	}

	remain, err := time.ParseDuration(right)
	if err != nil {
		return 0, err
	}

	// 要考虑负数的情况(这种情况下，传到系统函数时，就没有负数标志了)
	if days < 0 {
		result, ovf := overflow.SubInt(daysPart, int64(remain))
		if ovf {
			return 0, fmt.Errorf("invalid duration overflow:%s", s)
		}
		return time.Duration(result), nil
	}

	result, ovf := overflow.AddInt(daysPart, int64(remain))
	if ovf {
		return 0, fmt.Errorf("invalid duration overflow:%s", s)
	}
	return time.Duration(result), nil
}
