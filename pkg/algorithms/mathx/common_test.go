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

package mathx

import (
	"testing"
)

// ---------- Abs ----------

func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		a    int
		want int
	}{
		{"正数不变", 5, 5},
		{"负数取反", -5, 5},
		{"零", 0, 0},
		{"较大正数", 1000, 1000},
		{"较大负数", -1000, 1000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Abs(tc.a)
			if got != tc.want {
				t.Errorf("Abs(%d) = %d，期望 %d", tc.a, got, tc.want)
			}
		})
	}
}

func TestAbs_Int32(t *testing.T) {
	if got := Abs(int32(-42)); got != 42 {
		t.Errorf("Abs[int32](-42) = %d，期望 42", got)
	}
}

// ---------- Min ----------

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"a小于b", 3, 5, 3},
		{"a大于b", 7, 2, 2},
		{"a等于b", 4, 4, 4},
		{"负数", -3, -5, -5},
		{"负正混合", -1, 1, -1},
		{"零参与比较", 0, 5, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Min(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Min(%d, %d) = %d，期望 %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestMin_Float64(t *testing.T) {
	if got := Min(1.5, 2.5); got != 1.5 {
		t.Errorf("Min[float64](1.5, 2.5) = %f，期望 1.5", got)
	}
}

// ---------- Max ----------

func TestMax(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"a大于b", 7, 2, 7},
		{"a小于b", 3, 5, 5},
		{"a等于b", 4, 4, 4},
		{"负数", -3, -5, -3},
		{"负正混合", -1, 1, 1},
		{"零参与比较", 0, 5, 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Max(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Max(%d, %d) = %d，期望 %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestMax_Float64(t *testing.T) {
	if got := Max(1.5, 2.5); got != 2.5 {
		t.Errorf("Max[float64](1.5, 2.5) = %f，期望 2.5", got)
	}
}