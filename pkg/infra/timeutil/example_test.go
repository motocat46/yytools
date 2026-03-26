// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package timeutil_test

import (
	"fmt"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timeutil"
)

// ExampleParseDuration 展示 ParseDuration——在标准库基础上扩展了 'd'（天）单位。
func ExampleParseDuration() {
	// 纯天数
	d, _ := timeutil.ParseDuration("2d")
	fmt.Println(d) // 48h0m0s

	// 天 + 小时
	d, _ = timeutil.ParseDuration("2d3h")
	fmt.Println(d) // 51h0m0s

	// 天 + 分钟
	d, _ = timeutil.ParseDuration("1d30m")
	fmt.Println(d) // 24h30m0s

	// 零值
	d, _ = timeutil.ParseDuration("0")
	fmt.Println(d) // 0s

	// 小数天数（1.5 天 = 36 小时）
	d, _ = timeutil.ParseDuration("1.5d")
	fmt.Println(d) // 36h0m0s

	// 负数天（向前推 1 天）
	d, _ = timeutil.ParseDuration("-1d")
	fmt.Println(d) // -24h0m0s

	// 负零天（-0d30m 应得 -30m）
	d, _ = timeutil.ParseDuration("-0d30m")
	fmt.Println(d) // -30m0s

	// 不含 'd'，直接委托给标准库
	d, _ = timeutil.ParseDuration("1h30m")
	fmt.Println(d) // 1h30m0s
	// Output:
	// 48h0m0s
	// 51h0m0s
	// 24h30m0s
	// 0s
	// 36h0m0s
	// -24h0m0s
	// -30m0s
	// 1h30m0s
}

// ExampleParse 展示 Parse 的常用格式（返回 time.Local，输出因机器时区而异，不验证 Output）。
func ExampleParse() {
	// 纯日期（返回 time.Local 当天 00:00:00）
	_, _ = timeutil.Parse("2024-01-05")
	_, _ = timeutil.Parse("2024/1/5")

	// 日期时间
	_, _ = timeutil.Parse("2024-01-05 09:05:03")
	_, _ = timeutil.Parse("2024-1-5 9:5:3")

	// 省略秒（默认 :00）
	_, _ = timeutil.Parse("2024-01-05 09:05")

	// 非法输入返回错误
	_, err := timeutil.Parse("not-a-date")
	fmt.Println(err != nil) // true
	// Output:
	// true
}

// ExampleStartOfDay 展示日历边界函数的使用（使用固定时区保证输出确定性）。
func ExampleStartOfDay() {
	loc := time.UTC
	t := time.Date(2026, time.March, 26, 15, 30, 0, 0, loc)
	fmt.Println(timeutil.StartOfDay(t))
	fmt.Println(timeutil.StartOfTomorrow(t))
	// Output:
	// 2026-03-26 00:00:00 +0000 UTC
	// 2026-03-27 00:00:00 +0000 UTC
}

// ExampleIsSameDay 展示时间比较函数的使用。
func ExampleIsSameDay() {
	loc := time.UTC
	a := time.Date(2026, time.March, 26, 10, 0, 0, 0, loc)
	b := time.Date(2026, time.March, 26, 22, 0, 0, 0, loc)
	c := time.Date(2026, time.March, 27, 0, 0, 0, 0, loc)
	fmt.Println(timeutil.IsSameDay(a, b, loc)) // 同一天
	fmt.Println(timeutil.IsSameDay(a, c, loc)) // 不同天
	fmt.Println(timeutil.DaysBetween(a, c, loc))
	// Output:
	// true
	// false
	// 1
}
