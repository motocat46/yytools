// Package tools.

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
// 创建日期:2022/6/15

package assert

import "fmt"

// Assert 断言条件为真，否则 panic。
// 用于表达"绝不应该出现的情况"——触发即意味着调用方存在 bug，需在开发期修复。
// 不可关闭，始终生效。
func Assert(condition bool, list ...interface{}) {
	if !condition {
		if len(list) == 0 {
			panic("assertion failed")
		}
		panic(fmt.Sprintf("assertion failed - %s", fmt.Sprint(list...)))
	}
}

// AssertFast 断言 cond 为真，否则 panic，语义见 Assert。
// 与 Assert 的区别：不接受附加消息，panic 信息固定为 "assertion failed"；
// 适合内层热路径——调用约定已在外层注释中说明、无需额外上下文的场合。
func AssertFast(cond bool) {
	if !cond {
		panic("assertion failed")
	}
}
