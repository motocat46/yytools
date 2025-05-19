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
		{
			name: "条件为假时",
			args: args{
				condition: false,
				strList:   []interface{}{"hello", []int32{1, 2, 3}, map[int32]int32{1:2, 2:3}, 3.14, 119},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Assert(tt.args.condition, tt.args.strList...)
		})
	}
}