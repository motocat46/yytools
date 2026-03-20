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
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// 测试期间静默 slog：panic 恢复写日志是预期行为，不应污染测试输出。
	// 本包的测试只验证 panic 捕获和返回值语义，不断言日志内容。
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

func TestSafe(t *testing.T) {
	t.Run("正常执行", func(t *testing.T) {
		called := false
		Safe(func() {
			called = true
		})

		if !called {
			t.Error("Safe 函数应该正常执行，但没有被调用")
		}
	})

	t.Run("panic 恢复", func(t *testing.T) {
		// 这个测试应该不会导致测试程序崩溃
		Safe(func() {
			panic("测试 panic")
		})
		// 如果程序能执行到这里，说明 panic 被正确捕获了
	})

	t.Run("nil 函数", func(t *testing.T) {
		// 测试传入 nil 函数的情况
		Safe(nil)
		// 应该不会崩溃
	})
}

func TestSafeExec(t *testing.T) {
	t.Run("正常执行", func(t *testing.T) {
		called := false
		tag := "test_tag"
		SafeExec(tag, func() {
			called = true
		})

		if !called {
			t.Error("SafeExec 函数应该正常执行，但没有被调用")
		}
	})

	t.Run("panic 恢复", func(t *testing.T) {
		tag := "panic_test"
		SafeExec(tag, func() {
			panic("测试 panic 恢复")
		})
		// 如果程序能执行到这里，说明 panic 被正确捕获了
	})

	t.Run("nil 函数", func(t *testing.T) {
		// 测试传入 nil 函数的情况
		SafeExec("nil_test", nil)
		// 应该不会崩溃
	})
}

func TestSafeExecErr(t *testing.T) {
	t.Run("正常执行 - 无错误", func(t *testing.T) {
		tag := "no_error"
		err := SafeExecErr(tag, func() error {
			return nil
		})

		if err != nil {
			t.Errorf("SafeExecErr 应该返回 nil，但返回了: %v", err)
		}
	})

	t.Run("正常执行 - 有错误", func(t *testing.T) {
		tag := "with_error"
		expectedErr := errors.New("测试错误")
		err := SafeExecErr(tag, func() error {
			return expectedErr
		})

		if err != expectedErr {
			t.Errorf("SafeExecErr 应该返回 %v，但返回了: %v", expectedErr, err)
		}
	})

	t.Run("panic 恢复", func(t *testing.T) {
		tag := "panic_error"
		panicMsg := "测试 panic 错误"
		err := SafeExecErr(tag, func() error {
			panic(panicMsg)
		})

		if err == nil {
			t.Fatal("SafeExecErr 在 panic 时应该返回错误，但没有")
		}
		msg := err.Error()
		if !strings.Contains(msg, tag) {
			t.Errorf("error 应包含 tag %q，实际: %s", tag, msg)
		}
		if !strings.Contains(msg, panicMsg) {
			t.Errorf("error 应包含 panic 消息 %q，实际: %s", panicMsg, msg)
		}
		if !strings.Contains(msg, "goroutine") {
			t.Errorf("error 应包含调用栈（goroutine），实际: %s", msg)
		}
	})

	t.Run("nil 函数", func(t *testing.T) {
		// 测试传入 nil 函数的情况
		err := SafeExecErr("nil_test", nil)
		if err == nil {
			t.Error("传入 nil 函数时应该返回错误")
		}
	})
}

func TestSafeExecVal(t *testing.T) {
	t.Run("正常执行 - 返回字符串", func(t *testing.T) {
		tag := "string_result"
		expected := "测试结果"
		result, err := SafeExecVal(tag, func() string {
			return expected
		})

		if err != nil {
			t.Errorf("SafeExecVal 应该返回 nil 错误，但返回了: %v", err)
		}
		if result != expected {
			t.Errorf("SafeExecVal 应该返回 %q，但返回了: %q", expected, result)
		}
	})

	t.Run("正常执行 - 返回整数", func(t *testing.T) {
		tag := "int_result"
		expected := 42
		result, err := SafeExecVal(tag, func() int {
			return expected
		})

		if err != nil {
			t.Errorf("SafeExecVal 应该返回 nil 错误，但返回了: %v", err)
		}
		if result != expected {
			t.Errorf("SafeExecVal 应该返回 %d，但返回了: %d", expected, result)
		}
	})

	t.Run("正常执行 - 返回结构体", func(t *testing.T) {
		tag := "struct_result"
		type TestStruct struct {
			Name string
			Age  int
		}
		expected := TestStruct{Name: "测试", Age: 25}

		result, err := SafeExecVal(tag, func() TestStruct {
			return expected
		})

		if err != nil {
			t.Errorf("SafeExecVal 应该返回 nil 错误，但返回了: %v", err)
		}
		if result != expected {
			t.Errorf("SafeExecVal 应该返回 %+v，但返回了: %+v", expected, result)
		}
	})

	t.Run("panic 恢复", func(t *testing.T) {
		tag := "panic_result"
		panicMsg := "测试 panic 结果"
		result, err := SafeExecVal(tag, func() string {
			panic(panicMsg)
		})

		if err == nil {
			t.Fatal("SafeExecVal 在 panic 时应该返回错误，但没有")
		}
		if result != "" {
			t.Errorf("SafeExecVal 在 panic 时应该返回零值，但返回了: %q", result)
		}
		msg := err.Error()
		if !strings.Contains(msg, tag) {
			t.Errorf("error 应包含 tag %q，实际: %s", tag, msg)
		}
		if !strings.Contains(msg, panicMsg) {
			t.Errorf("error 应包含 panic 消息 %q，实际: %s", panicMsg, msg)
		}
		if !strings.Contains(msg, "goroutine") {
			t.Errorf("error 应包含调用栈（goroutine），实际: %s", msg)
		}
	})

	t.Run("nil 函数", func(t *testing.T) {
		// 测试传入 nil 函数的情况
		var fn func() string
		result, err := SafeExecVal("nil_test", fn)
		if err == nil {
			t.Error("传入 nil 函数时应该返回错误")
		}

		// 检查返回值是否为零值
		if result != "" {
			t.Errorf("传入 nil 函数时应该返回零值，但返回了: %q", result)
		}
	})
}

func TestSafeExecValComplexTypes(t *testing.T) {
	t.Run("返回切片", func(t *testing.T) {
		tag := "slice_result"
		expected := []int{1, 2, 3, 4, 5}

		result, err := SafeExecVal(tag, func() []int {
			return expected
		})

		if err != nil {
			t.Errorf("SafeExecVal 应该返回 nil 错误，但返回了: %v", err)
		}
		if len(result) != len(expected) {
			t.Errorf("SafeExecVal 应该返回长度为 %d 的切片，但返回了长度为 %d 的切片", len(expected), len(result))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("SafeExecVal 在索引 %d 应该返回 %d，但返回了 %d", i, v, result[i])
			}
		}
	})

	t.Run("返回映射", func(t *testing.T) {
		tag := "map_result"
		expected := map[string]int{"a": 1, "b": 2, "c": 3}

		result, err := SafeExecVal(tag, func() map[string]int {
			return expected
		})

		if err != nil {
			t.Errorf("SafeExecVal 应该返回 nil 错误，但返回了: %v", err)
		}
		if len(result) != len(expected) {
			t.Errorf("SafeExecVal 应该返回长度为 %d 的映射，但返回了长度为 %d 的映射", len(expected), len(result))
		}
		for k, v := range expected {
			if result[k] != v {
				t.Errorf("SafeExecVal 在键 %q 应该返回 %d，但返回了 %d", k, v, result[k])
			}
		}
	})
}


func BenchmarkSafeExec(b *testing.B) {
	for b.Loop() {
		SafeExec("bench", func() {})
	}
}

func BenchmarkSafeExecErr(b *testing.B) {
	for b.Loop() {
		SafeExecErr("bench", func() error { return nil }) //nolint
	}
}

func BenchmarkSafeExecVal(b *testing.B) {
	for b.Loop() {
		SafeExecVal("bench", func() int { return 42 }) //nolint
	}
}
