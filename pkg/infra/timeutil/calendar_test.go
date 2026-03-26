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
