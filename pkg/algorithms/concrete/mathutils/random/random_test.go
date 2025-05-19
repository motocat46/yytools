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
package random

import (
	"math"
	"testing"
)

func TestRandInt32(t *testing.T) {
	type args struct {
		low  int32
		high int32
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		// TODO: Add test cases.
		{
			name: "边界情况测试1",
			args: args{
				low:  0,
				high: 0,
			},
			want: 0,
		},
		{
			name: "边界情况测试2",
			args: args{
				low:  math.MaxInt32,
				high: math.MaxInt32,
			},
			want: math.MaxInt32,
		},
		//{
		//	name: "测试3",
		//	args: args{
		//		low:  -1,
		//		high: mathutils.MaxInt32,
		//	},
		//	want: mathutils.MaxInt32,
		//},
		{
			name: "测试4",
			args: args{
				low:  1,
				high: math.MaxInt32,
			},
			want: 0,
		},
		{
			name: "测试5",
			args: args{
				low:  0,
				high: math.MaxInt32,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandInt32(tt.args.low, tt.args.high); got != tt.want {
				t.Errorf("RandInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandInt64(t *testing.T) {
	type args struct {
		low  int64
		high int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandInt64(tt.args.low, tt.args.high); got != tt.want {
				t.Errorf("RandInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}