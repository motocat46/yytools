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

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func SetAssert(open bool) {
	isAssertOpen = open
}

func IsAssertOpen() bool {
	return isAssertOpen
}

// 获取调用者的文件名和行号。
// extraSkip：在固定的 getPrefix→Assert→调用方 三层之外，额外跳过的帧数。
// 直接调用 Assert 时传 0；每多一层包装函数就加 1。
func getPrefix(extraSkip int) string {
	// 调用链：getPrefix(skip=0) → Assert(skip=1) → 调用方(skip=2)
	_, file, line, _ := runtime.Caller(2 + extraSkip)

	prefixBuilder := strings.Builder{}
	prefixBuilder.WriteString("assertion failed at ")
	prefixBuilder.WriteString(file)
	prefixBuilder.WriteRune(':')
	prefixBuilder.WriteString(strconv.Itoa(line))
	return prefixBuilder.String()
}

// 断言
// 当断言失败时，调用 panic
func Assert(condition bool, list ...interface{}) {
	if isAssertOpen && !condition {
		prefix := getPrefix(0)
		if len(list) == 0 {
			panic(prefix)
		}
		// list 参数类型不确定，直接用 fmt.Sprint；断言失败才执行，不在热路径上
		panic(fmt.Sprintf("%s - %s", prefix, fmt.Sprint(list...)))
	}
}

// 快速断言，不传入参数
func AssertFast(cond bool) {
	if isAssertOpen && !cond {
		panic(getPrefix(0))
	}
}

// AssertSkip 与 Assert 相同，但允许调用方指定额外跳过的栈帧数。
// 用于将 Assert 包装在辅助函数中的场景：
//
//	func myCheck(cond bool) {
//	    assert.AssertSkip(cond, 1, "message") // 跳过 myCheck 这一层
//	}
func AssertSkip(condition bool, skip int, list ...interface{}) {
	if isAssertOpen && !condition {
		prefix := getPrefix(skip + 1) // +1 跳过 AssertSkip 本身
		if len(list) == 0 {
			panic(prefix)
		}
		panic(fmt.Sprintf("%s - %s", prefix, fmt.Sprint(list...)))
	}
}