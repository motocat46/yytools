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

package safeexec_test

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/motocat46/yytools/pkg/infra/safeexec"
)

func init() {
	// 示例中静默 slog，避免 panic 恢复日志污染 Output 验证
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// ExampleSafe 展示最简形式：执行可能 panic 的函数，自动恢复。
func ExampleSafe() {
	safeexec.Safe(func() {
		panic("something went wrong")
	})
	fmt.Println("程序正常继续")
	// Output:
	// 程序正常继续
}

// ExampleSafeExecErr 展示带 tag 的执行：panic 转换为 error 返回，调用方决定如何处理。
func ExampleSafeExecErr() {
	err := safeexec.SafeExecErr("handler", func() error {
		panic("unexpected nil pointer")
	})

	if err != nil {
		fmt.Println("捕获到错误（panic 被转换为 error）")
	}
	// Output:
	// 捕获到错误（panic 被转换为 error）
}

// ExampleSafeExecVal 展示泛型返回值：panic 时返回零值 + error。
func ExampleSafeExecVal() {
	result, err := safeexec.SafeExecVal("compute", func() int {
		return 42
	})

	if err == nil {
		fmt.Println(result)
	}
	// Output:
	// 42
}
