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

import "time"

// ── 内部通用层 ────────────────────────────────────────────────────────────────

// startOfNthDay 返回 t 所在时区当天偏移 n 天后的 00:00:00。
// n=0 今天，n=1 明天，n=-1 昨天。time.Date 自动规范化日期溢出。
func startOfNthDay(t time.Time, n int) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d+n, 0, 0, 0, 0, t.Location())
}

// ── 公开 API ──────────────────────────────────────────────────────────────────

// StartOfDay 返回 t 所在时区当天 00:00:00。
func StartOfDay(t time.Time) time.Time { return startOfNthDay(t, 0) }

// StartOfTomorrow 返回 t 所在时区明天 00:00:00。
func StartOfTomorrow(t time.Time) time.Time { return startOfNthDay(t, 1) }

// lastDayOfMonth 返回给定年月在 loc 时区下的最后一天的日数。
// 利用 time.Date 的 day=0 规范化：下月第 0 天 = 本月最后一天。
func lastDayOfMonth(y int, m time.Month, loc *time.Location) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, loc).Day()
}

// startOfNthMonthDay 返回 t 所在时区当月偏移 monthOffset 月后第 day 日的 00:00:00。
// day < 1 时 clamp 至 1，day 超出该月天数时 clamp 至月末。
func startOfNthMonthDay(t time.Time, day, monthOffset int) time.Time {
	y, m, _ := t.Date()
	loc := t.Location()
	// 偏移到目标月份的 1 日，time.Date 自动规范化（如 m+1 = 13 → 次年 1 月）
	firstOfTarget := time.Date(y, m+time.Month(monthOffset), 1, 0, 0, 0, 0, loc)
	ty, tm, _ := firstOfTarget.Date()
	last := lastDayOfMonth(ty, tm, loc)
	actualDay := day
	if actualDay < 1 {
		actualDay = 1
	}
	if actualDay > last {
		actualDay = last
	}
	return time.Date(ty, tm, actualDay, 0, 0, 0, 0, loc)
}

// StartOfMonth 返回 t 所在时区当月 1 日 00:00:00。
func StartOfMonth(t time.Time) time.Time { return startOfNthMonthDay(t, 1, 0) }

// StartOfNextMonthDay 返回 t 所在时区下个月第 day 日的 00:00:00。
// 若 day 超过下月天数，clamp 至下月最后一天（如 4 月传 31 → 4 月 30 日）。
func StartOfNextMonthDay(t time.Time, day int) time.Time { return startOfNthMonthDay(t, day, 1) }

// isoWeekday 将 time.Weekday 转换为 ISO 周序（周一=1，…，周六=6，周日=7）。
func isoWeekday(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

// startOfNthWeekday 返回 t 所在时区本周偏移 weekOffset 周后指定周几的 00:00:00。
// weekOffset=0 本周，weekOffset=1 下周。ISO 周，周一为第一天。
func startOfNthWeekday(t time.Time, weekday time.Weekday, weekOffset int) time.Time {
	currentISO := isoWeekday(t.Weekday())
	targetISO := isoWeekday(weekday)
	offset := targetISO - currentISO + weekOffset*7
	return startOfNthDay(t, offset)
}

// StartOfWeekday 返回 t 所在时区本周指定周几的 00:00:00（ISO 周，周一为第一天）。
// 若目标周几即为 t 当天，返回今天 00:00:00。
func StartOfWeekday(t time.Time, weekday time.Weekday) time.Time {
	return startOfNthWeekday(t, weekday, 0)
}

// StartOfNextWeekday 返回 t 所在时区下周指定周几的 00:00:00。
func StartOfNextWeekday(t time.Time, weekday time.Weekday) time.Time {
	return startOfNthWeekday(t, weekday, 1)
}

// ── Ms 变体（int64 毫秒时间戳，显式 loc）────────────────────────────────────

func StartOfDayMs(ms int64, loc *time.Location) int64 {
	return StartOfDay(time.UnixMilli(ms).In(loc)).UnixMilli()
}

func StartOfTomorrowMs(ms int64, loc *time.Location) int64 {
	return StartOfTomorrow(time.UnixMilli(ms).In(loc)).UnixMilli()
}

func StartOfMonthMs(ms int64, loc *time.Location) int64 {
	return StartOfMonth(time.UnixMilli(ms).In(loc)).UnixMilli()
}

func StartOfWeekdayMs(ms int64, weekday time.Weekday, loc *time.Location) int64 {
	return StartOfWeekday(time.UnixMilli(ms).In(loc), weekday).UnixMilli()
}

func StartOfNextWeekdayMs(ms int64, weekday time.Weekday, loc *time.Location) int64 {
	return StartOfNextWeekday(time.UnixMilli(ms).In(loc), weekday).UnixMilli()
}

func StartOfNextMonthDayMs(ms int64, day int, loc *time.Location) int64 {
	return StartOfNextMonthDay(time.UnixMilli(ms).In(loc), day).UnixMilli()
}

// ── 比较函数 ──────────────────────────────────────────────────────────────────

// IsSameDay 判断 a、b 在 loc 时区下是否为同一天。
func IsSameDay(a, b time.Time, loc *time.Location) bool {
	ay, am, ad := a.In(loc).Date()
	by, bm, bd := b.In(loc).Date()
	return ay == by && am == bm && ad == bd
}

func IsSameDayMs(a, b int64, loc *time.Location) bool {
	return IsSameDay(time.UnixMilli(a), time.UnixMilli(b), loc)
}

// IsSameWeek 判断 a、b 在 loc 时区下是否在同一 ISO 周（周一为第一天）。
func IsSameWeek(a, b time.Time, loc *time.Location) bool {
	ay, aw := a.In(loc).ISOWeek()
	by, bw := b.In(loc).ISOWeek()
	return ay == by && aw == bw
}

func IsSameWeekMs(a, b int64, loc *time.Location) bool {
	return IsSameWeek(time.UnixMilli(a), time.UnixMilli(b), loc)
}

// DaysBetween 返回 a 到 b 在 loc 时区下相差的日历天数（b>a 为正，b<a 为负）。
// 通过 UTC midnight 算术实现 DST 安全：UTC 无夏令时，差值整除 24h 精确。
func DaysBetween(a, b time.Time, loc *time.Location) int {
	ay, am, ad := a.In(loc).Date()
	by, bm, bd := b.In(loc).Date()
	// 构造 UTC midnight，消除 DST 影响
	aUTC := time.Date(ay, am, ad, 0, 0, 0, 0, time.UTC)
	bUTC := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	return int(bUTC.Sub(aUTC) / (24 * time.Hour))
}

func DaysBetweenMs(a, b int64, loc *time.Location) int {
	return DaysBetween(time.UnixMilli(a), time.UnixMilli(b), loc)
}
