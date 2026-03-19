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

// 获取断言调用者的文件名和行号。
// 调用链固定为：getPrefix → Assert/AssertFast → 调用方，因此 skip=2。
// Assert 设计为直接调用，不应被包装——包装会导致行号指向包装层而非真实调用处。
func getPrefix() string {
	_, file, line, _ := runtime.Caller(2)

	// 这里使用strings.Builder手动拼接字符串效率更高
	prefixBuilder := strings.Builder{}
	prefixBuilder.WriteString("assertion failed at ")
	prefixBuilder.WriteString(file)
	prefixBuilder.WriteRune(':')
	prefixBuilder.WriteString(strconv.Itoa(line))
	return prefixBuilder.String()
}

// 断言
// 当断言失败时，调用panic
func Assert(condition bool, list ...interface{}) {
	if isAssertOpen && !condition {
		prefix := getPrefix()
		if len(list) == 0 {
			panic(prefix)
		}
		// 1.使用Builder拼接字符串效率更高，但是不方便;list传入的参数类型是不确定的，手动处理很麻烦。
		// 又因只有断言失败才会执行，这里直接使用fmt.Sprint(list...)来处理
		// 2.展开list参数，拥有更好的打印格式
		panic(fmt.Sprintf("%s - %s", prefix, fmt.Sprint(list...)))
	}
}

// 快速断言，不传入参数
func AssertFast(cond bool) {
	if isAssertOpen && !cond {
		prefix := getPrefix()
		panic(prefix)
	}
}