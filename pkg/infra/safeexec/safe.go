// Package safeexec.

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
// 创建日期:2025/5/19
package safeexec

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// Safe 是 SafeExec 的简便形式，tag 固定为 "anonymous"。
// 适用于不需要区分调用来源的临时或一次性场景。
// 若需在日志中区分不同调用点，请改用 SafeExec 并传入有意义的 tag。
func Safe(f func()) {
	SafeExec("anonymous", f)
}

// SafeExec 安全执行一个无参数、无返回值的函数。
//
// 设计目标：将 panic 的传播边界限制在本函数内，保障调用方所在的主流程不受影响。
// 适用场景：独立的业务单元（如为单个玩家发奖、执行定时任务、处理外部回调），
// 任意一个单元的 panic 不应中断其他单元的执行。
//
// 参数：
//   - tag：调用方标识，用于日志定位，建议传入能反映业务语义的常量字符串，如 "giveReward"。
//   - f：待执行的函数。若为 nil，记录日志后直接返回，不会 panic。
//
// panic 处理：捕获 f 执行期间的任何 panic，将 panic 值和完整调用栈写入 slog.Default()，
// 然后正常返回，调用方无感知。
// 日志输出目标由应用层通过 slog.SetDefault() 配置，默认写入 stderr。
func SafeExec(tag string, f func()) {
	if f == nil {
		slog.Warn("[safeexec] f is nil", slog.String("tag", tag))
		return
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("[safeexec] panic",
				slog.String("tag", tag),
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()
	f()
}

// SafeExecErr 安全执行一个返回 error 的函数。
//
// 设计目标：将 panic 的传播边界限制在本函数内，同时保留 f 的正常 error 返回语义。
// 适用场景：需要区分"业务错误"与"意外崩溃"的调用点。f 返回的业务 error 原样透传给调用方；
// 若发生 panic，则将其包装为 error 返回。
//
// 参数：
//   - tag：调用方标识，用于日志定位，建议传入能反映业务语义的常量字符串。
//   - f：待执行的函数。若为 nil，直接返回 error，不会 panic。
//
// 返回值：
//   - f 正常执行：返回 f 的原始 error（可能为 nil）。
//   - f 发生 panic：返回包含 tag、panic 值和完整调用栈的 error，不额外打日志。
//     调用方从 error 中获取完整信息，自行决定如何记录（可附加 request ID 等业务上下文）。
func SafeExecErr(tag string, f func() error) (err error) {
	if f == nil {
		return fmt.Errorf("[safeexec] %s: f is nil", tag)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[safeexec] panic in %s: %v\n%s", tag, r, debug.Stack())
		}
	}()
	return f()
}

// SafeExecVal 安全执行一个带返回值的函数。
//
// 设计目标：将 panic 的传播边界限制在本函数内，同时保留 f 的返回值语义。
// 适用场景：需要获取 f 执行结果，同时防止 panic 向上传播的调用点。
//
// 参数：
//   - tag：调用方标识，用于日志定位，建议传入能反映业务语义的常量字符串。
//   - f：待执行的函数。若为 nil，直接返回零值和 error，不会 panic。
//
// 返回值：
//   - f 正常执行：返回 f 的结果和 nil error。
//   - f 发生 panic：返回 T 的零值和包含 tag、panic 值和完整调用栈的 error，不额外打日志。
//     调用方收到非 nil error 时，不应使用返回的零值 res。
func SafeExecVal[T any](tag string, f func() T) (res T, err error) {
	if f == nil {
		err = fmt.Errorf("[safeexec] %s: f is nil", tag)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[safeexec] panic in %s: %v\n%s", tag, r, debug.Stack())
		}
	}()
	res = f()
	return
}
