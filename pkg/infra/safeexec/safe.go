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
    "github.com/stormYuanYang/yytools/pkg/common/assert"
    "log"
)

// 最简单的安全执行函数
func Safe(f func()) {
    SafeCall("anonymous", f)
}

// SafeCall 执行一个无参数无返回值的函数，若发生 panic 会捕获并记录日志，保障后续流程继续执行。
func SafeCall(tag string, f func()) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[safeexec] panic in %s: %v", tag, r)
        }
    }()
    assert.Assert(f != nil, "[safeexec] SafeCall args f is nil")
    f()
}

// SafeExecWithError 执行带 error 返回值的函数，若 panic 则包装为 error 返回；否则原样返回 error。
func SafeExecWithError(tag string, f func() error) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic in %s: %v", tag, r)
            log.Printf("[safeexec] panic in %s: %v", tag, r)
        }
    }()
    assert.Assert(f != nil, "[safeexec] SafeCall args f is nil")
    return f()
}

// SafeExecResult 执行带返回值的函数，若 panic 则返回默认值和错误。
func SafeExecResult[T any](tag string, f func() T) (res T, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic in %s: %v", tag, r)
            log.Printf("[safeexec] panic in %s: %v", tag, r)
        }
    }()
    assert.Assert(f != nil, "[safeexec] SafeCall args f is nil")
    res = f()
    return
}