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
	"fmt"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/motocat46/yytools/pkg/algorithms/mathx"
)

// bigComb 用 math/big 精确计算 C(n,k) mod p，作为参考模型。
func bigComb(n, k, p int64) int64 {
	if k < 0 || k > n {
		return 0
	}
	fn := bigFactorial(n)
	fk := bigFactorial(k)
	fnk := bigFactorial(n - k)
	denom := new(big.Int).Mul(fk, fnk)
	result := new(big.Int).Div(fn, denom)
	return result.Mod(result, big.NewInt(p)).Int64()
}

func bigFactorial(n int64) *big.Int {
	result := big.NewInt(1)
	for i := int64(2); i <= n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

func TestComb_Basic(t *testing.T) {
	cases := []struct {
		n    int64
		k    int64
		mod  int64
		want int64
	}{
		{5, 2, pm, 10},
		{5, 0, pm, 1},
		{5, 5, pm, 1},
		{10, 3, pm, 120},
		{0, 0, pm, 1},
		{1, 1, pm, 1},
		{20, 10, pm, 184756},
	}
	for _, tc := range cases {
		got := mathx.Comb(tc.n, tc.k, tc.mod)
		if got != tc.want {
			t.Errorf("Comb(%d,%d,%d) = %d，期望 %d", tc.n, tc.k, tc.mod, got, tc.want)
		}
	}
}

func TestComb_Symmetry(t *testing.T) {
	for n := int64(0); n <= 30; n++ {
		for k := int64(0); k <= n; k++ {
			a := mathx.Comb(n, k, pm)
			b := mathx.Comb(n, n-k, pm)
			if a != b {
				t.Errorf("Comb(%d,%d) = %d，Comb(%d,%d) = %d，应相等", n, k, a, n, n-k, b)
			}
		}
	}
}

func TestComb_KGreaterN(t *testing.T) {
	cases := []struct {
		n int64
		k int64
	}{
		{5, 6},
		{0, 1},
		{3, 4},
	}
	for _, tc := range cases {
		if got := mathx.Comb(tc.n, tc.k, pm); got != 0 {
			t.Errorf("Comb(%d,%d,%d) = %d，期望 0", tc.n, tc.k, pm, got)
		}
	}
}

func TestComb_NegativeK(t *testing.T) {
	if got := mathx.Comb(5, -1, pm); got != 0 {
		t.Errorf("Comb(5,-1,%d) = %d，期望 0", pm, got)
	}
}

func TestComb_PanicNegN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Comb 负 n 应 panic，但未 panic")
		}
	}()
	_ = mathx.Comb(-1, 0, pm)
}

func TestComb_AgainstBigInt(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过 bigInt 正确性验证")
	}
	rng := rand.New(rand.NewPCG(42, 0))
	for range 10_000 {
		n := int64(rng.IntN(201))
		k := int64(rng.IntN(int(n) + 1))
		got := mathx.Comb(n, k, pm)
		want := bigComb(n, k, pm)
		if got != want {
			t.Fatalf("Comb(%d,%d,%d) = %d，期望 %d（bigInt）", n, k, pm, got, want)
		}
	}
}

func BenchmarkComb(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = mathx.Comb(1000, 500, pm)
	}
}

func TestCombTable_Basic(t *testing.T) {
	ct := mathx.NewCombTable(100, pm)
	cases := []struct {
		n    int
		k    int
		want int64
	}{
		{5, 2, 10},
		{5, 0, 1},
		{5, 5, 1},
		{10, 3, 120},
		{0, 0, 1},
		{1, 1, 1},
		{20, 10, 184756},
	}
	for _, tc := range cases {
		got := ct.C(tc.n, tc.k)
		if got != tc.want {
			t.Errorf("CombTable.C(%d,%d) = %d，期望 %d", tc.n, tc.k, got, tc.want)
		}
	}
}

func TestCombTable_Symmetry(t *testing.T) {
	ct := mathx.NewCombTable(30, pm)
	for n := 0; n <= 30; n++ {
		for k := 0; k <= n; k++ {
			a := ct.C(n, k)
			b := ct.C(n, n-k)
			if a != b {
				t.Errorf("C(%d,%d) = %d，C(%d,%d) = %d，应相等", n, k, a, n, n-k, b)
			}
		}
	}
}

func TestCombTable_KGreaterN(t *testing.T) {
	ct := mathx.NewCombTable(10, pm)
	if got := ct.C(5, 6); got != 0 {
		t.Errorf("C(5,6) = %d，期望 0", got)
	}
	if got := ct.C(5, -1); got != 0 {
		t.Errorf("C(5,-1) = %d，期望 0", got)
	}
}

func TestCombTable_PanicOutOfBounds(t *testing.T) {
	ct := mathx.NewCombTable(10, pm)
	defer func() {
		if r := recover(); r == nil {
			t.Error("C(11,0) 应 panic，但未 panic")
		}
	}()
	_ = ct.C(11, 0)
}

func TestCombTable_AgainstComb(t *testing.T) {
	ct := mathx.NewCombTable(200, pm)
	rng := rand.New(rand.NewPCG(44, 0))
	for range 10_000 {
		n := int64(rng.IntN(201))
		k := int64(rng.IntN(int(n) + 1))
		got := ct.C(int(n), int(k))
		want := mathx.Comb(n, k, pm)
		if got != want {
			t.Fatalf("CombTable.C(%d,%d) = %d，Comb(%d,%d,%d) = %d", n, k, got, n, k, pm, want)
		}
	}
}

func TestCombTable_AgainstBigInt(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过 bigInt 正确性验证")
	}
	ct := mathx.NewCombTable(200, pm)
	rng := rand.New(rand.NewPCG(45, 0))
	for range 10_000 {
		n := int64(rng.IntN(201))
		k := int64(rng.IntN(int(n) + 1))
		got := ct.C(int(n), int(k))
		want := bigComb(n, k, pm)
		if got != want {
			t.Fatalf("n=%d k=%d: CombTable = %d，bigInt = %d", n, k, got, want)
		}
	}
}

func BenchmarkCombTable_Build(b *testing.B) {
	sizes := []int{100, 1_000, 10_000, 100_000, 1_000_000}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = mathx.NewCombTable(n, pm)
			}
		})
	}
}

func BenchmarkCombTable_Query(b *testing.B) {
	ct := mathx.NewCombTable(10_000, pm)
	rng := rand.New(rand.NewPCG(42, 0))
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		n := rng.IntN(10_001)
		k := rng.IntN(n + 1)
		_ = ct.C(n, k)
	}
}
