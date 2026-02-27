// Package mathx.

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
// 创建日期:2026/2/27
package mathx

import "testing"

// ---------- GcdR / GcdI / Gcd ----------

func TestGcdR(t *testing.T) {
	tests := []struct {
		name string
		x, y int
		want int
	}{
		{"两个互质数", 7, 5, 1},
		{"有公因数", 12, 8, 4},
		{"整除关系", 9, 3, 3},
		{"两数相等", 6, 6, 6},
		{"y为0", 15, 0, 15},
		{"x为0", 0, 10, 10},
		{"两者都是0", 0, 0, 0},
		{"大数", 1000000, 999999, 1},
		{"大数有公因数", 123456, 65536, 64},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GcdR(tc.x, tc.y)
			if got != tc.want {
				t.Errorf("GcdR(%d, %d) = %d，期望 %d", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

func TestGcdI(t *testing.T) {
	tests := []struct {
		name string
		x, y int
		want int
	}{
		{"两个互质数", 7, 5, 1},
		{"有公因数", 12, 8, 4},
		{"整除关系", 9, 3, 3},
		{"两数相等", 6, 6, 6},
		{"y为0", 15, 0, 15},
		{"x为0", 0, 10, 10},
		{"两者都是0", 0, 0, 0},
		{"大数", 1000000, 999999, 1},
		{"大数有公因数", 123456, 65536, 64},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GcdI(tc.x, tc.y)
			if got != tc.want {
				t.Errorf("GcdI(%d, %d) = %d，期望 %d", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

// GcdR 与 GcdI 结果应完全一致
func TestGcd_ConsistencyWithR(t *testing.T) {
	cases := [][2]int{{12, 8}, {100, 75}, {0, 5}, {7, 7}, {123456, 65536}}
	for _, c := range cases {
		r := GcdR(c[0], c[1])
		i := GcdI(c[0], c[1])
		g := Gcd(c[0], c[1])
		if r != i || r != g {
			t.Errorf("GcdR(%d,%d)=%d GcdI=%d Gcd=%d 三者应一致", c[0], c[1], r, i, g)
		}
	}
}