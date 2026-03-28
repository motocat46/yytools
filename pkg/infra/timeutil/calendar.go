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
//
// 示例（Asia/Shanghai）：
//
//	t = 2024-03-15 14:30:00 CST → 2024-03-15 00:00:00 CST
func StartOfDay(t time.Time) time.Time { return startOfNthDay(t, 0) }

// StartOfTomorrow 返回 t 所在时区明天 00:00:00。
//
// 示例（Asia/Shanghai）：
//
//	t = 2024-03-15 14:30:00 CST → 2024-03-16 00:00:00 CST
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
//
// 示例（Asia/Shanghai）：
//
//	t = 2024-03-15 14:30:00 CST → 2024-03-01 00:00:00 CST
func StartOfMonth(t time.Time) time.Time { return startOfNthMonthDay(t, 1, 0) }

// StartOfNextMonthDay 返回 t 所在时区下个月第 day 日的 00:00:00。
// 若 day < 1，clamp 至 1；若 day 超过下月天数，clamp 至下月最后一天。
//
// 示例（Asia/Shanghai）：
//
//	t = 2024-03-15，day=10  → 2024-04-10 00:00:00 CST
//	t = 2024-03-15，day=31  → 2024-04-30 00:00:00 CST（4 月只有 30 天，clamp）
//	t = 2024-01-31，day=31  → 2024-02-29 00:00:00 CST（2024 年闰年，clamp）
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
// 若目标周几即为 t 当天，返回今天 00:00:00（不跳到下周）。
//
// 示例（Asia/Shanghai，t = 2024-03-13 周三）：
//
//	weekday=Wednesday → 2024-03-13 00:00:00 CST（当天）
//	weekday=Monday    → 2024-03-11 00:00:00 CST（本周一，往前）
//	weekday=Friday    → 2024-03-15 00:00:00 CST（本周五，往后）
func StartOfWeekday(t time.Time, weekday time.Weekday) time.Time {
	return startOfNthWeekday(t, weekday, 0)
}

// StartOfNextWeekday 返回 t 所在时区下周指定周几的 00:00:00（ISO 周，周一为第一天）。
// 无论 t 是星期几，始终返回下一个自然周（+7 天）的目标周几。
//
// 示例（Asia/Shanghai，t = 2024-03-13 周三）：
//
//	weekday=Monday → 2024-03-18 00:00:00 CST（下周一）
//	weekday=Friday → 2024-03-22 00:00:00 CST（下周五）
func StartOfNextWeekday(t time.Time, weekday time.Weekday) time.Time {
	return startOfNthWeekday(t, weekday, 1)
}

// ── Ms 变体（int64 毫秒时间戳，显式 loc）────────────────────────────────────
//
// 以下函数是对应 time.Time 版本的毫秒时间戳封装：
// 入参 ms 为 Unix 毫秒时间戳，loc 指定计算所用时区，返回值同为 Unix 毫秒时间戳。
// 语义与对应的 time.Time 版本完全一致，详见各函数的文档注释。

// StartOfDayMs 返回 ms 在 loc 时区下当天 00:00:00 的毫秒时间戳。
func StartOfDayMs(ms int64, loc *time.Location) int64 {
	return StartOfDay(time.UnixMilli(ms).In(loc)).UnixMilli()
}

// StartOfTomorrowMs 返回 ms 在 loc 时区下明天 00:00:00 的毫秒时间戳。
func StartOfTomorrowMs(ms int64, loc *time.Location) int64 {
	return StartOfTomorrow(time.UnixMilli(ms).In(loc)).UnixMilli()
}

// StartOfMonthMs 返回 ms 在 loc 时区下当月 1 日 00:00:00 的毫秒时间戳。
func StartOfMonthMs(ms int64, loc *time.Location) int64 {
	return StartOfMonth(time.UnixMilli(ms).In(loc)).UnixMilli()
}

// StartOfWeekdayMs 返回 ms 在 loc 时区下本周指定周几 00:00:00 的毫秒时间戳。
// 语义见 StartOfWeekday。
func StartOfWeekdayMs(ms int64, weekday time.Weekday, loc *time.Location) int64 {
	return StartOfWeekday(time.UnixMilli(ms).In(loc), weekday).UnixMilli()
}

// StartOfNextWeekdayMs 返回 ms 在 loc 时区下下周指定周几 00:00:00 的毫秒时间戳。
// 语义见 StartOfNextWeekday。
func StartOfNextWeekdayMs(ms int64, weekday time.Weekday, loc *time.Location) int64 {
	return StartOfNextWeekday(time.UnixMilli(ms).In(loc), weekday).UnixMilli()
}

// StartOfNextMonthDayMs 返回 ms 在 loc 时区下下个月第 day 日 00:00:00 的毫秒时间戳。
// 语义见 StartOfNextMonthDay。
func StartOfNextMonthDayMs(ms int64, day int, loc *time.Location) int64 {
	return StartOfNextMonthDay(time.UnixMilli(ms).In(loc), day).UnixMilli()
}

// ── 比较函数 ──────────────────────────────────────────────────────────────────

// IsSameDay 判断 a、b 在 loc 时区下是否为同一天（年、月、日均相同）。
// 两个参数自身携带的时区信息会被忽略，统一用 loc 换算日历日期后比较。
//
// 示例（loc = Asia/Shanghai，UTC+8）：
//
//	a = 2024-03-15 23:00:00 UTC（= 2024-03-16 07:00:00 CST）
//	b = 2024-03-16 01:00:00 UTC（= 2024-03-16 09:00:00 CST）
//	→ true（CST 下同一天）
func IsSameDay(a, b time.Time, loc *time.Location) bool {
	ay, am, ad := a.In(loc).Date()
	by, bm, bd := b.In(loc).Date()
	return ay == by && am == bm && ad == bd
}

// IsSameDayMs 判断毫秒时间戳 a、b 在 loc 时区下是否为同一天。语义见 IsSameDay。
func IsSameDayMs(a, b int64, loc *time.Location) bool {
	return IsSameDay(time.UnixMilli(a), time.UnixMilli(b), loc)
}

// IsSameWeek 判断 a、b 在 loc 时区下是否在同一 ISO 周（周一为第一天）。
// ISO 周以周一为起点，周日为终点；跨年边界时，12 月末可能属于下一年的第 1 周，
// 1 月初可能属于上一年的最后一周——本函数同时比较 ISO 年和周序号，可正确处理跨年边界。
//
// 示例（loc = Asia/Shanghai）：
//
//	a = 2024-03-11（周一），b = 2024-03-17（周日）→ true（同一 ISO 周）
//	a = 2024-03-17（周日），b = 2024-03-18（周一）→ false（不同 ISO 周）
//	a = 2023-12-31（周日，属于 2023 年第 52 周），b = 2024-01-01（属于 2024 年第 1 周）→ false
func IsSameWeek(a, b time.Time, loc *time.Location) bool {
	ay, aw := a.In(loc).ISOWeek()
	by, bw := b.In(loc).ISOWeek()
	return ay == by && aw == bw
}

// IsSameWeekMs 判断毫秒时间戳 a、b 在 loc 时区下是否在同一 ISO 周。语义见 IsSameWeek。
func IsSameWeekMs(a, b int64, loc *time.Location) bool {
	return IsSameWeek(time.UnixMilli(a), time.UnixMilli(b), loc)
}

// DaysBetween 返回 a 到 b 在 loc 时区下相差的日历天数。
//
// 符号规则：b 在 a 之后为正，b 在 a 之前为负，同一天为 0。
//
// 计算的是"日历天数差"，而非"小时数 ÷ 24"：
// 只要两个时刻在 loc 时区下落在相邻的日历日，差值就是 1，与具体时刻无关。
//
// 示例（loc = Asia/Shanghai）：
//
//	a = 2024-03-15 23:59:59 CST，b = 2024-03-16 00:00:01 CST → 1（相邻日历日）
//	a = 2024-03-16 00:00:01 CST，b = 2024-03-15 23:59:59 CST → -1
//	a = 2024-03-15 00:00:00 CST，b = 2024-03-15 23:59:59 CST → 0（同一天）
//	a = 2024-03-13 CST，        b = 2024-03-20 CST            → 7
//
// DST 安全：通过在 UTC 下做 midnight 差值算术实现，不受夏令时切换影响。
func DaysBetween(a, b time.Time, loc *time.Location) int {
	ay, am, ad := a.In(loc).Date()
	by, bm, bd := b.In(loc).Date()
	// 构造 UTC midnight，消除 DST 影响
	aUTC := time.Date(ay, am, ad, 0, 0, 0, 0, time.UTC)
	bUTC := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	return int(bUTC.Sub(aUTC) / (24 * time.Hour))
}

// DaysBetweenMs 返回毫秒时间戳 a 到 b 在 loc 时区下的日历天数差。语义见 DaysBetween。
func DaysBetweenMs(a, b int64, loc *time.Location) int {
	return DaysBetween(time.UnixMilli(a), time.UnixMilli(b), loc)
}
