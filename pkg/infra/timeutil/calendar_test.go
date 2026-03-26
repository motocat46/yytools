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
// 创建日期:2026/3/26
package timeutil

import (
	"testing"
	"time"
)

// 固定时区，不依赖机器 Local，保证测试在任何环境下确定性通过
var (
	cst = time.FixedZone("CST", 8*60*60) // UTC+8
)

// dt 在 CST(UTC+8) 时区创建指定时刻
func dt(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, cst)
}

// dtUTC 在 UTC 时区创建指定时刻
func dtUTC(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func TestStartOfDay(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{"下午时刻", dt(2026, time.March, 26, 15, 30, 45), dt(2026, time.March, 26, 0, 0, 0)},
		{"午夜本身", dt(2026, time.March, 26, 0, 0, 0), dt(2026, time.March, 26, 0, 0, 0)},
		{"23:59:59", dt(2026, time.March, 26, 23, 59, 59), dt(2026, time.March, 26, 0, 0, 0)},
		{"UTC时区保持", dtUTC(2026, time.March, 26, 15, 30, 0), dtUTC(2026, time.March, 26, 0, 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StartOfDay(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("StartOfDay(%v) = %v, want %v", tc.in, got, tc.want)
			}
			if got.Location() != tc.in.Location() {
				t.Errorf("StartOfDay 时区不一致: got %v, want %v", got.Location(), tc.in.Location())
			}
		})
	}
}

func TestStartOfTomorrow(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{"普通日期", dt(2026, time.March, 26, 15, 30, 0), dt(2026, time.March, 27, 0, 0, 0)},
		{"月末跨月", dt(2026, time.March, 31, 12, 0, 0), dt(2026, time.April, 1, 0, 0, 0)},
		{"年末跨年", dt(2026, time.December, 31, 23, 59, 59), dt(2027, time.January, 1, 0, 0, 0)},
		{"UTC时区", dtUTC(2026, time.March, 26, 15, 0, 0), dtUTC(2026, time.March, 27, 0, 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StartOfTomorrow(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("StartOfTomorrow(%v) = %v, want %v", tc.in, got, tc.want)
			}
			if got.Location() != tc.in.Location() {
				t.Errorf("StartOfTomorrow 时区不一致: got %v, want %v", got.Location(), tc.in.Location())
			}
		})
	}
}

func TestStartOfMonth(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{"月中", dt(2026, time.March, 15, 12, 0, 0), dt(2026, time.March, 1, 0, 0, 0)},
		{"已是1日", dt(2026, time.March, 1, 0, 0, 0), dt(2026, time.March, 1, 0, 0, 0)},
		{"月末", dt(2026, time.March, 31, 23, 59, 59), dt(2026, time.March, 1, 0, 0, 0)},
		{"UTC时区", dtUTC(2026, time.March, 15, 12, 0, 0), dtUTC(2026, time.March, 1, 0, 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StartOfMonth(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("StartOfMonth(%v) = %v, want %v", tc.in, got, tc.want)
			}
			if got.Location() != tc.in.Location() {
				t.Errorf("StartOfMonth 时区不一致: got %v, want %v", got.Location(), tc.in.Location())
			}
		})
	}
}

func TestStartOfNextMonthDay(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		day  int
		want time.Time
	}{
		{"下月1日", dt(2026, time.March, 15, 12, 0, 0), 1, dt(2026, time.April, 1, 0, 0, 0)},
		{"下月15日", dt(2026, time.March, 15, 12, 0, 0), 15, dt(2026, time.April, 15, 0, 0, 0)},
		{"12月→1月", dt(2026, time.December, 15, 0, 0, 0), 1, dt(2027, time.January, 1, 0, 0, 0)},
		{"clamp：4月传31日→4月30日", dt(2026, time.March, 15, 0, 0, 0), 31, dt(2026, time.April, 30, 0, 0, 0)},
		{"clamp：2月传30日→2月28日", dt(2026, time.January, 15, 0, 0, 0), 30, dt(2026, time.February, 28, 0, 0, 0)},
		{"clamp：闰年2月传30日→2月29日", dt(2024, time.January, 15, 0, 0, 0), 30, dt(2024, time.February, 29, 0, 0, 0)},
		{"clamp：day=0→下月1日", dt(2026, time.March, 15, 0, 0, 0), 0, dt(2026, time.April, 1, 0, 0, 0)},
		{"clamp：day=-3→下月1日", dt(2026, time.March, 15, 0, 0, 0), -3, dt(2026, time.April, 1, 0, 0, 0)},
		{"UTC时区", dtUTC(2026, time.March, 15, 0, 0, 0), 1, dtUTC(2026, time.April, 1, 0, 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StartOfNextMonthDay(tc.in, tc.day)
			if !got.Equal(tc.want) {
				t.Errorf("StartOfNextMonthDay(%v, %d) = %v, want %v", tc.in, tc.day, got, tc.want)
			}
			if got.Location() != tc.in.Location() {
				t.Errorf("StartOfNextMonthDay 时区不一致: got %v, want %v", got.Location(), tc.in.Location())
			}
		})
	}
}

func TestStartOfWeekday(t *testing.T) {
	cases := []struct {
		name    string
		in      time.Time
		weekday time.Weekday
		want    time.Time
	}{
		// 2026-03-26 是周四
		{"本周一（周四→周一）", dt(2026, time.March, 26, 15, 0, 0), time.Monday, dt(2026, time.March, 23, 0, 0, 0)},
		{"本周五（周四→周五）", dt(2026, time.March, 26, 15, 0, 0), time.Friday, dt(2026, time.March, 27, 0, 0, 0)},
		{"当天本身（周四→周四）", dt(2026, time.March, 26, 15, 0, 0), time.Thursday, dt(2026, time.March, 26, 0, 0, 0)},
		{"本周日（周四→周日）", dt(2026, time.March, 26, 15, 0, 0), time.Sunday, dt(2026, time.March, 29, 0, 0, 0)},
		// 2026-03-23 是周一
		{"本周一（周一→周一）", dt(2026, time.March, 23, 0, 0, 0), time.Monday, dt(2026, time.March, 23, 0, 0, 0)},
		{"UTC时区", dtUTC(2026, time.March, 26, 15, 0, 0), time.Monday, dtUTC(2026, time.March, 23, 0, 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StartOfWeekday(tc.in, tc.weekday)
			if !got.Equal(tc.want) {
				t.Errorf("StartOfWeekday(%v, %v) = %v, want %v", tc.in, tc.weekday, got, tc.want)
			}
			if got.Location() != tc.in.Location() {
				t.Errorf("StartOfWeekday 时区不一致: got %v, want %v", got.Location(), tc.in.Location())
			}
		})
	}
}

func TestStartOfNextWeekday(t *testing.T) {
	cases := []struct {
		name    string
		in      time.Time
		weekday time.Weekday
		want    time.Time
	}{
		// 2026-03-26 是周四
		{"下周一（周四→下周一）", dt(2026, time.March, 26, 15, 0, 0), time.Monday, dt(2026, time.March, 30, 0, 0, 0)},
		{"下周五（周四→下周五）", dt(2026, time.March, 26, 15, 0, 0), time.Friday, dt(2026, time.April, 3, 0, 0, 0)},
		{"下周四（当天周几→下周同一天）", dt(2026, time.March, 26, 15, 0, 0), time.Thursday, dt(2026, time.April, 2, 0, 0, 0)},
		{"跨月（下周日）", dt(2026, time.March, 29, 10, 0, 0), time.Sunday, dt(2026, time.April, 5, 0, 0, 0)},
		{"UTC时区", dtUTC(2026, time.March, 26, 15, 0, 0), time.Monday, dtUTC(2026, time.March, 30, 0, 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StartOfNextWeekday(tc.in, tc.weekday)
			if !got.Equal(tc.want) {
				t.Errorf("StartOfNextWeekday(%v, %v) = %v, want %v", tc.in, tc.weekday, got, tc.want)
			}
			if got.Location() != tc.in.Location() {
				t.Errorf("StartOfNextWeekday 时区不一致: got %v, want %v", got.Location(), tc.in.Location())
			}
		})
	}
}

// msOf 将 time.Time 转换为毫秒时间戳（测试辅助）
func msOf(t time.Time) int64 { return t.UnixMilli() }

func TestMsVariants(t *testing.T) {
	// 输入：2026-03-26 15:30:45 CST（Unix ms = 1774510245000）
	// 期望值为独立计算的硬编码绝对时间戳，避免循环引用 time.Time 变体：
	//   StartOfDay    → 2026-03-26 00:00:00 CST = 1774454400000
	//   StartOfTomorrow → 2026-03-27 00:00:00 CST = 1774540800000
	//   StartOfMonth  → 2026-03-01 00:00:00 CST = 1772294400000
	//   StartOfWeekday(Mon) → 2026-03-23 00:00:00 CST = 1774195200000（周四往前4天）
	//   StartOfNextWeekday(Mon) → 2026-03-30 00:00:00 CST = 1774800000000（下周一）
	//   StartOfNextMonthDay(15) → 2026-04-15 00:00:00 CST = 1776182400000
	ms := msOf(dt(2026, time.March, 26, 15, 30, 45))

	cases := []struct {
		name string
		got  int64
		want int64
	}{
		{"StartOfDayMs", StartOfDayMs(ms, cst), 1774454400000},
		{"StartOfTomorrowMs", StartOfTomorrowMs(ms, cst), 1774540800000},
		{"StartOfMonthMs", StartOfMonthMs(ms, cst), 1772294400000},
		{"StartOfWeekdayMs(Monday)", StartOfWeekdayMs(ms, time.Monday, cst), 1774195200000},
		{"StartOfNextWeekdayMs(Monday)", StartOfNextWeekdayMs(ms, time.Monday, cst), 1774800000000},
		{"StartOfNextMonthDayMs(15)", StartOfNextMonthDayMs(ms, 15, cst), 1776182400000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestIsSameDay(t *testing.T) {
	cases := []struct {
		name string
		a, b time.Time
		loc  *time.Location
		want bool
	}{
		{"同一天同时刻", dt(2026, time.March, 26, 10, 0, 0), dt(2026, time.March, 26, 22, 0, 0), cst, true},
		{"不同天", dt(2026, time.March, 26, 23, 59, 59), dt(2026, time.March, 27, 0, 0, 0), cst, false},
		{"跨时区：UTC+8午夜 = UTC前一天", dt(2026, time.March, 26, 0, 0, 0), dtUTC(2026, time.March, 25, 16, 0, 0), cst, true},
		{"UTC loc：同一天", dtUTC(2026, time.March, 26, 0, 0, 0), dtUTC(2026, time.March, 26, 23, 59, 59), time.UTC, true},
		{"UTC loc：不同天", dtUTC(2026, time.March, 26, 23, 59, 59), dtUTC(2026, time.March, 27, 0, 0, 0), time.UTC, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSameDay(tc.a, tc.b, tc.loc)
			if got != tc.want {
				t.Errorf("IsSameDay(%v, %v, %v) = %v, want %v", tc.a, tc.b, tc.loc, got, tc.want)
			}
		})
	}
}

func TestIsSameWeek(t *testing.T) {
	// 2026-03-23(Mon) ~ 2026-03-29(Sun) 同一 ISO 周
	cases := []struct {
		name string
		a, b time.Time
		loc  *time.Location
		want bool
	}{
		{"同周（周一和周日）", dt(2026, time.March, 23, 0, 0, 0), dt(2026, time.March, 29, 23, 59, 59), cst, true},
		{"同周（周三和周五）", dt(2026, time.March, 25, 10, 0, 0), dt(2026, time.March, 27, 10, 0, 0), cst, true},
		{"不同周（周日→下周一）", dt(2026, time.March, 29, 23, 59, 59), dt(2026, time.March, 30, 0, 0, 0), cst, false},
		{"跨年同周（如 2026-01-01 是周四，属于 2025 第 1 周）",
			dt(2025, time.December, 29, 0, 0, 0), dt(2026, time.January, 1, 0, 0, 0), cst, true},
		{"UTC loc", dtUTC(2026, time.March, 23, 0, 0, 0), dtUTC(2026, time.March, 29, 0, 0, 0), time.UTC, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSameWeek(tc.a, tc.b, tc.loc)
			if got != tc.want {
				t.Errorf("IsSameWeek(%v, %v, %v) = %v, want %v", tc.a, tc.b, tc.loc, got, tc.want)
			}
		})
	}
}

func TestIsSameDayMs(t *testing.T) {
	cases := []struct {
		name string
		a, b time.Time
		loc  *time.Location
		want bool
	}{
		{"同天", dt(2026, time.March, 26, 10, 0, 0), dt(2026, time.March, 26, 22, 0, 0), cst, true},
		{"不同天", dt(2026, time.March, 26, 23, 59, 59), dt(2026, time.March, 27, 0, 0, 0), cst, false},
		{"UTC loc 同天", dtUTC(2026, time.March, 26, 0, 0, 0), dtUTC(2026, time.March, 26, 23, 59, 59), time.UTC, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSameDayMs(msOf(tc.a), msOf(tc.b), tc.loc)
			if got != tc.want {
				t.Errorf("IsSameDayMs(%d, %d, %v) = %v, want %v", msOf(tc.a), msOf(tc.b), tc.loc, got, tc.want)
			}
		})
	}
}

func TestDaysBetween(t *testing.T) {
	cases := []struct {
		name string
		a, b time.Time
		loc  *time.Location
		want int
	}{
		{"b 在 a 之后 3 天", dt(2026, time.March, 26, 0, 0, 0), dt(2026, time.March, 29, 0, 0, 0), cst, 3},
		{"b 在 a 之前（负数）", dt(2026, time.March, 29, 0, 0, 0), dt(2026, time.March, 26, 0, 0, 0), cst, -3},
		{"同一天（0）", dt(2026, time.March, 26, 10, 0, 0), dt(2026, time.March, 26, 23, 0, 0), cst, 0},
		{"跨月", dt(2026, time.March, 30, 0, 0, 0), dt(2026, time.April, 2, 0, 0, 0), cst, 3},
		{"跨年", dt(2026, time.December, 31, 0, 0, 0), dt(2027, time.January, 2, 0, 0, 0), cst, 2},
		// DST 安全：即使两个时刻跨越夏令时转换点，日历天数也应正确
		// 用固定时区不受 DST 影响，但计算方式（UTC midnight 差）是 DST 安全的
		{"UTC loc", dtUTC(2026, time.March, 26, 0, 0, 0), dtUTC(2026, time.March, 28, 0, 0, 0), time.UTC, 2},
		// 跨时区：a 在 CST，b 在 UTC，loc 用 CST 判断
		{"跨时区计算", dt(2026, time.March, 26, 0, 30, 0), dtUTC(2026, time.March, 25, 17, 30, 0), cst, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DaysBetween(tc.a, tc.b, tc.loc)
			if got != tc.want {
				t.Errorf("DaysBetween(%v, %v, %v) = %d, want %d", tc.a, tc.b, tc.loc, got, tc.want)
			}
		})
	}
}

func TestIsSameWeekMs(t *testing.T) {
	cases := []struct {
		name string
		a, b time.Time
		loc  *time.Location
		want bool
	}{
		{"同周", dt(2026, time.March, 23, 0, 0, 0), dt(2026, time.March, 29, 23, 59, 59), cst, true},
		{"不同周", dt(2026, time.March, 29, 23, 59, 59), dt(2026, time.March, 30, 0, 0, 0), cst, false},
		{"UTC loc", dtUTC(2026, time.March, 23, 0, 0, 0), dtUTC(2026, time.March, 29, 0, 0, 0), time.UTC, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSameWeekMs(msOf(tc.a), msOf(tc.b), tc.loc)
			if got != tc.want {
				t.Errorf("IsSameWeekMs(%d, %d, %v) = %v, want %v", msOf(tc.a), msOf(tc.b), tc.loc, got, tc.want)
			}
		})
	}
}

func TestDaysBetweenMs(t *testing.T) {
	cases := []struct {
		name string
		a, b time.Time
		loc  *time.Location
		want int
	}{
		{"b 在 a 之后", dt(2026, time.March, 26, 0, 0, 0), dt(2026, time.March, 29, 0, 0, 0), cst, 3},
		{"b 在 a 之前（负数）", dt(2026, time.March, 29, 0, 0, 0), dt(2026, time.March, 26, 0, 0, 0), cst, -3},
		{"同一天（0）", dt(2026, time.March, 26, 10, 0, 0), dt(2026, time.March, 26, 22, 0, 0), cst, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DaysBetweenMs(msOf(tc.a), msOf(tc.b), tc.loc)
			if got != tc.want {
				t.Errorf("DaysBetweenMs(%d, %d, %v) = %d, want %d", msOf(tc.a), msOf(tc.b), tc.loc, got, tc.want)
			}
		})
	}
}

func TestMsVariants_UTC(t *testing.T) {
	// 输入：2026-03-26 15:30:00 UTC（Unix ms = 1774539000000）
	// 在 CST(UTC+8) 下该时刻 = 2026-03-26 23:30:00 CST，StartOfDay → 2026-03-26 00:00:00 CST = 1774454400000
	// 在 UTC 下该时刻 = 2026-03-26 15:30:00 UTC，StartOfDay → 2026-03-26 00:00:00 UTC = 1774483200000
	// 两个期望值独立计算，不依赖 StartOfDay 的实现
	ms := msOf(dtUTC(2026, time.March, 26, 15, 30, 0)) // 1774539000000

	const wantCST int64 = 1774454400000 // 2026-03-26 00:00:00 CST
	const wantUTC int64 = 1774483200000 // 2026-03-26 00:00:00 UTC

	gotCST := StartOfDayMs(ms, cst)
	gotUTC := StartOfDayMs(ms, time.UTC)

	if gotCST != wantCST {
		t.Errorf("StartOfDayMs(ms, CST) = %d, want %d (2026-03-26 00:00:00 CST)", gotCST, wantCST)
	}
	if gotUTC != wantUTC {
		t.Errorf("StartOfDayMs(ms, UTC) = %d, want %d (2026-03-26 00:00:00 UTC)", gotUTC, wantUTC)
	}
	if gotCST == gotUTC {
		t.Errorf("CST 和 UTC 的 StartOfDayMs 不应相等，got both = %d", gotCST)
	}
}
