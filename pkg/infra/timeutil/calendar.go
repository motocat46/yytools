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
