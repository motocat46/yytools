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
//
// 作者:  yangyuan

package mathx_test

import (
	"testing"

	"github.com/motocat46/yytools/pkg/algorithms/mathx"
)

const pm int64 = 1_000_000_007

func TestPowMod_Basic(t *testing.T) {
	cases := []struct {
		base int64
		exp  int64
		mod  int64
		want int64
	}{
		{2, 10, pm, 1024},
		{2, 0, pm, 1},
		{0, 5, pm, 0},
		{1, 1_000_000, pm, 1},
		{3, 0, 7, 1},
		{3, 1, 7, 3},
		{3, 3, 7, 6},
		{2, 30, pm, 73_741_817},
	}
	for _, tc := range cases {
		got := mathx.PowMod(tc.base, tc.exp, tc.mod)
		if got != tc.want {
			t.Errorf("PowMod(%d,%d,%d) = %d，期望 %d", tc.base, tc.exp, tc.mod, got, tc.want)
		}
	}
}

func TestPowMod_ModOne(t *testing.T) {
	if got := mathx.PowMod(5, 3, 1); got != 0 {
		t.Errorf("PowMod(5,3,1) = %d，期望 0", got)
	}
}

func TestPowMod_Fermat(t *testing.T) {
	for _, a := range []int64{2, 3, 100, 999_999} {
		got := mathx.PowMod(a, pm-1, pm)
		if got != 1 {
			t.Errorf("PowMod(%d,%d,%d) = %d，期望 1", a, pm-1, pm, got)
		}
	}
}

func TestPowMod_PanicNegExp(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("PowMod 负指数应 panic，但未 panic")
		}
	}()
	_ = mathx.PowMod(2, -1, pm)
}

func TestPowMod_PanicZeroMod(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("PowMod mod=0 应 panic，但未 panic")
		}
	}()
	_ = mathx.PowMod(2, 3, 0)
}

func BenchmarkPowMod(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = mathx.PowMod(123456789, 987654321, pm)
	}
}
