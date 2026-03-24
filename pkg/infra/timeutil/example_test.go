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
