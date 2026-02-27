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