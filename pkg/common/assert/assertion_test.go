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
		// TODO: Add test cases.
		//{
		//	name: "字符串数组为空时",
		//	args: args{
		//		condition: false,
		//		strList:   nil,
		//	},
		//},
		//{
		//	name: "字符串数组长度为1",
		//	args: args{
		//		condition: false,
		//		strList:   []string{"hello"},
		//	},
		//},
		//{
		//	name: "字符串数组长度为2",
		//	args: args{
		//		condition: false,
		//		strList:   []string{"hello", "yytools"},
		//	},
		//},
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
	// 确保断言开启
	SetAssert(true)
	
	t.Run("条件为假时应该panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望Assert(false)会触发panic，但没有发生")
			} else {
				// 验证panic信息包含正确的文件和行号信息
				panicMsg := r.(string)
				if !strings.Contains(panicMsg, "assertion failed at") {
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
				// 验证多参数消息被正确格式化
				if !strings.Contains(panicMsg, "hello") || !strings.Contains(panicMsg, "119") {
					t.Errorf("panic消息没有包含期望的参数: %v", panicMsg)
				}
			}
		}()
		Assert(false, "hello", []int32{1, 2, 3}, map[int32]int32{1:2, 2:3}, 3.14, 119)
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

// 测试断言关闭时的行为
func TestAssert_Disabled(t *testing.T) {
	// 保存原始状态
	originalState := IsAssertOpen()
	defer SetAssert(originalState)
	
	// 关闭断言
	SetAssert(false)
	
	t.Run("断言关闭时不应该panic", func(t *testing.T) {
		// 这不应该panic
		Assert(false, "这个断言应该被忽略")
		Assert(false) // 无参数版本也不应该panic
	})
}

// 测试AssertFast函数
func TestAssertFast(t *testing.T) {
	SetAssert(true)
	
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