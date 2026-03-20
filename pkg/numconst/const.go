// Package numconst.

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
// 创建日期:2023/6/7
package numconst

type TimeUnit = int64

// 数量级常量
const (
	Thousand        = 1_000          // 千
	TenThousand     = 10_000         // 万
	HundredThousand = 100_000        // 十万
	Million         = 1_000_000      // 百万
	TenMillion      = 10_000_000     // 千万
	HundredMillion  = 100_000_000    // 亿
	Billion         = 1_000_000_000  // 十亿
)

// 存储单位常量（1024 进制）
// TB = 1_099_511_627_776，超出 int32 范围，赋值给 int 时在 32 位平台会编译报错。
// 如需兼容 32 位平台，显式使用 int64 或仅用 KB/MB/GB。
const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

// 时间相关的常量
const (
	// 基准1用作毫秒的单位
	MILLISECOND = TimeUnit(1)
	SECOND      = 1e3 * MILLISECOND
	MINUTE      = 60 * SECOND
	HOUR        = 60 * MINUTE
	DAY         = 24 * HOUR
	WEEK        = 7 * DAY
)