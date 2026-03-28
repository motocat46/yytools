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
package timecond_test

import (
	"math"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timecond"
	"github.com/motocat46/yytools/pkg/infra/timeutil"
)

// ── 辅助函数 ──────────────────────────────────────────────────────────────────

func mustParse(t *testing.T, op timecond.Op, value string) *timecond.TimeCondition {
	t.Helper()
	c, err := timecond.Parse(op, value)
	if err != nil {
		t.Fatalf("Parse(%s, %q) failed: %v", op, value, err)
	}
	return c
}

func mustParseMs(t *testing.T, value string) int64 {
	t.Helper()
	tm, err := timeutil.Parse(value)
	if err != nil {
		t.Fatalf("timeutil.Parse(%q) failed: %v", value, err)
	}
	return tm.UnixMilli()
}

// ── Parse 错误路径 ────────────────────────────────────────────────────────────

func TestParse_Error(t *testing.T) {
	cases := []struct {
		name  string
		op    timecond.Op
		value string
	}{
		{"unsupported op", timecond.Op(99), ""},
		{"OpLT invalid time", timecond.OpLT, "not-a-time"},
		{"OpGE invalid time", timecond.OpGE, "not-a-time"},
		{"OpWithin missing comma", timecond.OpWithin, "2024-03-15 10:00:00"},
		{"OpWithin invalid first", timecond.OpWithin, "bad,2024-03-15 10:00:00"},
		{"OpWithin invalid second", timecond.OpWithin, "2024-03-15 10:00:00,bad"},
		{"OpRelLT invalid duration", timecond.OpRelLT, "not-a-duration"},
		{"OpRelGE invalid duration", timecond.OpRelGE, "not-a-duration"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := timecond.Parse(tc.op, tc.value)
			if err == nil {
				t.Errorf("Parse(%s, %q): got nil error, want error", tc.op, tc.value)
			}
		})
	}
}

func TestParse_OpAlways_ValueIgnored(t *testing.T) {
	// OpAlways 不解析 value，任何值都不应报错
	for _, v := range []string{"", "anything", "2024-03-15 10:00:00"} {
		if _, err := timecond.Parse(timecond.OpAlways, v); err != nil {
			t.Errorf("Parse(OpAlways, %q): got error %v, want nil", v, err)
		}
	}
}

// ── Check：OpAlways ───────────────────────────────────────────────────────────

func TestCheck_OpAlways(t *testing.T) {
	cond := mustParse(t, timecond.OpAlways, "")
	cases := []struct{ subjectMs, nowMs int64 }{
		{0, 0},
		{-1, 0},
		{math.MaxInt64, math.MaxInt64},
		{math.MinInt64, math.MaxInt64},
	}
	for _, tc := range cases {
		if !cond.Check(tc.subjectMs, tc.nowMs) {
			t.Errorf("OpAlways.Check(%d, %d) = false, want true", tc.subjectMs, tc.nowMs)
		}
	}
}

// ── Check：OpLT ──────────────────────────────────────────────────────────────

func TestCheck_OpLT(t *testing.T) {
	const timeStr = "2024-03-15 10:00:00"
	cond := mustParse(t, timecond.OpLT, timeStr)
	threshold := mustParseMs(t, timeStr)

	cases := []struct {
		name      string
		subjectMs int64
		want      bool
	}{
		{"小于阈值", threshold - 1, true},
		{"等于阈值（边界，不满足）", threshold, false},
		{"大于阈值", threshold + 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cond.Check(tc.subjectMs, 0)
			if got != tc.want {
				t.Errorf("OpLT.Check(%d, 0) = %v, want %v（threshold=%d）",
					tc.subjectMs, got, tc.want, threshold)
			}
		})
	}
}

// ── Check：OpGE ──────────────────────────────────────────────────────────────

func TestCheck_OpGE(t *testing.T) {
	const timeStr = "2024-03-15 10:00:00"
	cond := mustParse(t, timecond.OpGE, timeStr)
	threshold := mustParseMs(t, timeStr)

	cases := []struct {
		name      string
		subjectMs int64
		want      bool
	}{
		{"小于阈值", threshold - 1, false},
		{"等于阈值（边界，满足）", threshold, true},
		{"大于阈值", threshold + 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cond.Check(tc.subjectMs, 0)
			if got != tc.want {
				t.Errorf("OpGE.Check(%d, 0) = %v, want %v（threshold=%d）",
					tc.subjectMs, got, tc.want, threshold)
			}
		})
	}
}

// ── Check：OpWithin ───────────────────────────────────────────────────────────

func TestCheck_OpWithin(t *testing.T) {
	const loStr = "2024-03-15 10:00:00"
	const hiStr = "2024-03-20 10:00:00"
	cond := mustParse(t, timecond.OpWithin, loStr+","+hiStr)
	lo := mustParseMs(t, loStr)
	hi := mustParseMs(t, hiStr)

	cases := []struct {
		name      string
		subjectMs int64
		want      bool
	}{
		{"小于 lo（区间外）", lo - 1, false},
		{"等于 lo（左闭，满足）", lo, true},
		{"区间中间", (lo + hi) / 2, true},
		{"等于 hi-1（区间内）", hi - 1, true},
		{"等于 hi（右开，不满足）", hi, false},
		{"大于 hi（区间外）", hi + 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cond.Check(tc.subjectMs, 0)
			if got != tc.want {
				t.Errorf("OpWithin.Check(%d, 0) = %v, want %v（lo=%d, hi=%d）",
					tc.subjectMs, got, tc.want, lo, hi)
			}
		})
	}
}

func TestCheck_OpWithin_AutoSort(t *testing.T) {
	// 配置中 t1 > t2，应自动排序，结果与 t1 < t2 一致
	const loStr = "2024-03-15 10:00:00"
	const hiStr = "2024-03-20 10:00:00"
	condNormal := mustParse(t, timecond.OpWithin, loStr+","+hiStr)
	condReversed := mustParse(t, timecond.OpWithin, hiStr+","+loStr)

	lo := mustParseMs(t, loStr)
	hi := mustParseMs(t, hiStr)
	mid := (lo + hi) / 2

	for _, ms := range []int64{lo - 1, lo, mid, hi - 1, hi, hi + 1} {
		a := condNormal.Check(ms, 0)
		b := condReversed.Check(ms, 0)
		if a != b {
			t.Errorf("自动排序不一致：subject=%d, normal=%v, reversed=%v", ms, a, b)
		}
	}
}

// ── Check：OpRelLT ────────────────────────────────────────────────────────────

func TestCheck_OpRelLT(t *testing.T) {
	cond := mustParse(t, timecond.OpRelLT, "1h") // (nowMs - subjectMs) < 1h
	oneHourMs := time.Hour.Milliseconds()
	subject := int64(0)

	cases := []struct {
		name  string
		nowMs int64
		want  bool
	}{
		{"经过时长 < 1h", subject + oneHourMs - 1, true},
		{"经过时长 = 1h（边界，不满足）", subject + oneHourMs, false},
		{"经过时长 > 1h", subject + oneHourMs + 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cond.Check(subject, tc.nowMs)
			if got != tc.want {
				t.Errorf("OpRelLT.Check(subject=%d, now=%d) = %v, want %v",
					subject, tc.nowMs, got, tc.want)
			}
		})
	}
}

// ── Check：OpRelGE ────────────────────────────────────────────────────────────

func TestCheck_OpRelGE(t *testing.T) {
	cond := mustParse(t, timecond.OpRelGE, "1h") // (nowMs - subjectMs) >= 1h
	oneHourMs := time.Hour.Milliseconds()
	subject := int64(0)

	cases := []struct {
		name  string
		nowMs int64
		want  bool
	}{
		{"经过时长 < 1h", subject + oneHourMs - 1, false},
		{"经过时长 = 1h（边界，满足）", subject + oneHourMs, true},
		{"经过时长 > 1h", subject + oneHourMs + 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cond.Check(subject, tc.nowMs)
			if got != tc.want {
				t.Errorf("OpRelGE.Check(subject=%d, now=%d) = %v, want %v",
					subject, tc.nowMs, got, tc.want)
			}
		})
	}
}

// ── 绝对 Op 不受 nowMs 影响 ───────────────────────────────────────────────────

func TestCheck_AbsOp_NowMsIgnored(t *testing.T) {
	const timeStr = "2024-03-15 10:00:00"
	threshold := mustParseMs(t, timeStr)
	subject := threshold - 1 // 小于阈值：OpLT 期望 true，OpGE 期望 false
	nowValues := []int64{0, threshold, threshold * 2, math.MinInt64, math.MaxInt64}

	cases := []struct {
		op   timecond.Op
		want bool
	}{
		{timecond.OpLT, true},
		{timecond.OpGE, false},
	}
	for _, tc := range cases {
		cond := mustParse(t, tc.op, timeStr)
		for _, nowMs := range nowValues {
			got := cond.Check(subject, nowMs)
			if got != tc.want {
				t.Errorf("%s.Check(subject=%d, now=%d) = %v，want %v（nowMs 不应影响绝对条件）",
					tc.op, subject, nowMs, got, tc.want)
			}
		}
	}

	// OpWithin 同样不受 nowMs 影响，lo==hi 退化区间期望 false
	condWithin := mustParse(t, timecond.OpWithin, timeStr+","+timeStr)
	for _, nowMs := range nowValues {
		if got := condWithin.Check(threshold, nowMs); got {
			t.Errorf("OpWithin(lo==hi).Check(subject=%d, now=%d) = true，want false",
				threshold, nowMs)
		}
	}
}

// ── OpWithin 退化区间（lo == hi）─────────────────────────────────────────────

func TestCheck_OpWithin_Degenerate(t *testing.T) {
	// lo == hi 时区间为空 [lo, lo)，任何值都不满足
	const timeStr = "2024-03-15 10:00:00"
	cond := mustParse(t, timecond.OpWithin, timeStr+","+timeStr)
	threshold := mustParseMs(t, timeStr)

	for _, subject := range []int64{threshold - 1, threshold, threshold + 1} {
		if got := cond.Check(subject, 0); got {
			t.Errorf("OpWithin(lo==hi).Check(%d, 0) = true，want false（空区间）", subject)
		}
	}
}

// ── OpRelLT/GE 负数经过时长（nowMs < subjectMs）──────────────────────────────

func TestCheck_Rel_NegativeElapsed(t *testing.T) {
	// nowMs < subjectMs 时经过时长为负数
	// OpRelLT：负数 < 任何正时长 → true
	// OpRelGE：负数 >= 任何正时长 → false
	condLT := mustParse(t, timecond.OpRelLT, "1h")
	condGE := mustParse(t, timecond.OpRelGE, "1h")

	subject := int64(1000)
	now := int64(0) // now < subject，经过时长 = -1000ms

	if got := condLT.Check(subject, now); !got {
		t.Errorf("OpRelLT.Check(subject=%d, now=%d)（负数经过时长）= false，want true", subject, now)
	}
	if got := condGE.Check(subject, now); got {
		t.Errorf("OpRelGE.Check(subject=%d, now=%d)（负数经过时长）= true，want false", subject, now)
	}
}

// ── Op.String ─────────────────────────────────────────────────────────────────

func TestOp_String(t *testing.T) {
	cases := []struct {
		op   timecond.Op
		want string
	}{
		{timecond.OpAlways, "OpAlways"},
		{timecond.OpLT, "OpLT"},
		{timecond.OpGE, "OpGE"},
		{timecond.OpWithin, "OpWithin"},
		{timecond.OpRelLT, "OpRelLT"},
		{timecond.OpRelGE, "OpRelGE"},
		{timecond.Op(99), "Op(99)"},
	}
	for _, tc := range cases {
		if got := tc.op.String(); got != tc.want {
			t.Errorf("Op(%d).String() = %q, want %q", int(tc.op), got, tc.want)
		}
	}
}

// ── Op 访问器 ─────────────────────────────────────────────────────────────────

func TestTimeCondition_Op(t *testing.T) {
	for _, op := range []timecond.Op{
		timecond.OpAlways, timecond.OpLT, timecond.OpGE,
		timecond.OpWithin, timecond.OpRelLT, timecond.OpRelGE,
	} {
		value := ""
		switch op {
		case timecond.OpLT, timecond.OpGE:
			value = "2024-03-15 10:00:00"
		case timecond.OpWithin:
			value = "2024-03-15 10:00:00,2024-03-20 10:00:00"
		case timecond.OpRelLT, timecond.OpRelGE:
			value = "1h"
		}
		cond := mustParse(t, op, value)
		if got := cond.Op(); got != op {
			t.Errorf("TimeCondition.Op() = %s, want %s", got, op)
		}
	}
}
