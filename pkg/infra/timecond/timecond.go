// Package timecond 提供基于配置的时间条件判断：解析 (Op, value) 字符串，
// 返回可复用的 TimeCondition，调用 Check 判断给定时间戳是否满足条件。

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
// 创建日期:2026/3/28
package timecond

import (
	"fmt"
	"strings"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timeutil"
)

// Op 表示时间条件的比较运算符。
type Op int

const (
	OpAlways Op = iota // 无条件成立，value 忽略
	OpLT               // subjectMs < absTs
	OpGE               // subjectMs >= absTs
	OpWithin           // absRange[0] <= subjectMs < absRange[1]（左闭右开）
	OpRelLT            // (nowMs - subjectMs) < relDur
	OpRelGE            // (nowMs - subjectMs) >= relDur
)

func (o Op) String() string {
	switch o {
	case OpAlways:
		return "OpAlways"
	case OpLT:
		return "OpLT"
	case OpGE:
		return "OpGE"
	case OpWithin:
		return "OpWithin"
	case OpRelLT:
		return "OpRelLT"
	case OpRelGE:
		return "OpRelGE"
	default:
		return fmt.Sprintf("Op(%d)", int(o))
	}
}

// TimeCondition 表示一个可复用的时间条件，由 Parse 创建后可多次调用 Check。
type TimeCondition struct {
	op       Op
	absTs    int64         // OpLT, OpGE：绝对时间戳（毫秒）
	absRange [2]int64      // OpWithin：[lo, hi)，lo <= subject < hi
	relDur   time.Duration // OpRelLT, OpRelGE：相对时长
}

// Op 暴露条件的运算符类型，供调用方在日志、序列化或调试时使用。
func (c *TimeCondition) Op() Op { return c.op }

// Parse 解析 (op, value) 配置，返回可复用的 TimeCondition。
// op 不在已知枚举范围内，或 value 格式不符合对应 Op 的要求时，返回非 nil error。
//
// value 格式：
//   - OpAlways：忽略，传空字符串即可
//   - OpLT / OpGE：时间字符串，格式同 timeutil.Parse（如 "2024-03-15 10:00:00"）
//   - OpWithin：两个时间字符串用英文逗号分隔（如 "2024-03-15 10:00:00,2024-03-20 10:00:00"），
//     自动排序，语义为左闭右开 [t1, t2)
//   - OpRelLT / OpRelGE：时长字符串，格式同 timeutil.ParseDuration（如 "7d"、"24h30m"）
func Parse(op Op, value string) (*TimeCondition, error) {
	switch op {
	case OpAlways:
		return &TimeCondition{op: op}, nil
	case OpLT, OpGE:
		ms, err := parseAbsMs(value)
		if err != nil {
			return nil, err
		}
		return &TimeCondition{op: op, absTs: ms}, nil
	case OpWithin:
		lo, hi, err := parseAbsRange(value)
		if err != nil {
			return nil, err
		}
		return &TimeCondition{op: op, absRange: [2]int64{lo, hi}}, nil
	case OpRelLT, OpRelGE:
		dur, err := timeutil.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("timecond: parse duration %q: %w", value, err)
		}
		return &TimeCondition{op: op, relDur: dur}, nil
	default:
		return nil, fmt.Errorf("timecond: unsupported %s", op)
	}
}

// Check 判断 subjectMs 是否满足条件，满足返回 true。
//
//   - subjectMs：被判断的时间戳（如注册时间、开服时间），单位毫秒
//   - nowMs：当前时间戳，单位毫秒；仅 OpRelLT / OpRelGE 使用，其他 Op 忽略
//
// 当 nowMs < subjectMs 时，经过时长为负数：OpRelLT 返回 true（负数 < 任何正时长），
// OpRelGE 返回 false。
func (c *TimeCondition) Check(subjectMs, nowMs int64) bool {
	switch c.op {
	case OpAlways:
		return true
	case OpLT:
		return subjectMs < c.absTs
	case OpGE:
		return subjectMs >= c.absTs
	case OpWithin:
		return c.absRange[0] <= subjectMs && subjectMs < c.absRange[1]
	case OpRelLT:
		return nowMs-subjectMs < c.relDur.Milliseconds()
	case OpRelGE:
		return nowMs-subjectMs >= c.relDur.Milliseconds()
	default:
		panic(fmt.Sprintf("timecond: unknown %s", c.op))
	}
}

// parseAbsMs 将时间字符串解析为 Unix 毫秒时间戳。
func parseAbsMs(value string) (int64, error) {
	t, err := timeutil.Parse(value)
	if err != nil {
		return 0, fmt.Errorf("timecond: parse time %q: %w", value, err)
	}
	return t.UnixMilli(), nil
}

// parseAbsRange 将 "t1,t2" 格式解析为左闭右开区间 [lo, hi)，lo > hi 时自动交换。
// 用 SplitN(..., 2) 强制只拆成两段，保证格式为恰好一个逗号分隔的两个时间。
func parseAbsRange(value string) (lo, hi int64, err error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("timecond: OpWithin expects \"t1,t2\", got %q", value)
	}
	lo, err = parseAbsMs(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	hi, err = parseAbsMs(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi, nil
}
