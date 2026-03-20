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
	"strings"
	"testing"
)

func TestAssert(t *testing.T) {
	type args struct {
		condition bool
		strList   []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "条件为真时",
			args: args{
				condition: true,
				strList:   []interface{}{"hello", "yytools"},
			},
		},
		{
			name: "条件为真时 strList为空",
			args: args{
				condition: true,
				strList:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Assert(tt.args.condition, tt.args.strList...)
		})
	}
}

// 测试断言失败时的panic行为
func TestAssert_Panic(t *testing.T) {
	t.Run("条件为假时应该panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望Assert(false)会触发panic，但没有发生")
			} else {
				panicMsg := r.(string)
				if !strings.Contains(panicMsg, "assertion failed") {
					t.Errorf("panic消息格式不正确: %v", panicMsg)
				}
			}
		}()
		Assert(false, "测试消息")
	})

	t.Run("条件为假_多参数", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望Assert(false)会触发panic，但没有发生")
			} else {
				panicMsg := r.(string)
				if !strings.Contains(panicMsg, "hello") || !strings.Contains(panicMsg, "119") {
					t.Errorf("panic消息没有包含期望的参数: %v", panicMsg)
				}
			}
		}()
		Assert(false, "hello", []int32{1, 2, 3}, map[int32]int32{1: 2, 2: 3}, 3.14, 119)
	})

	t.Run("条件为假_无参数", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望Assert(false)会触发panic，但没有发生")
			}
		}()
		Assert(false)
	})
}

// 测试AssertFast函数
func TestAssertFast(t *testing.T) {
	t.Run("条件为真时正常运行", func(t *testing.T) {
		AssertFast(true) // 不应该panic
	})

	t.Run("条件为假时应该panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望AssertFast(false)会触发panic，但没有发生")
			}
		}()
		AssertFast(false)
	})
}
