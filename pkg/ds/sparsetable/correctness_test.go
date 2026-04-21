// Copyright [yangyuan]
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

package sparsetable_test

import (
	"math/rand/v2"
	"testing"

	"github.com/motocat46/yytools/pkg/ds/sparsetable"
)

// refQuery 是区间查询的朴素参考模型，直接线性扫描 [l, r]。
func refQuery[T any](data []T, merge func(T, T) T, l, r int) T {
	acc := data[l]
	for i := l + 1; i <= r; i++ {
		acc = merge(acc, data[i])
	}
	return acc
}

func TestCorrectness_Min(t *testing.T) {
	const n = 1_000

	rng := rand.New(rand.NewPCG(42, 0))
	data := make([]int, n)
	for i := range data {
		data[i] = rng.IntN(2_000_001) - 1_000_000
	}
	st := sparsetable.New(data, minInt)

	for i := 0; i < 100_000; i++ {
		l, r := rng.IntN(n), rng.IntN(n)
		if l > r {
			l, r = r, l
		}
		got := st.Query(l, r)
		want := refQuery(data, minInt, l, r)
		if got != want {
			t.Fatalf("Min 对拍失败：op=%d, l=%d, r=%d, got=%d, want=%d, data[%d:%d]=%v",
				i, l, r, got, want, l, r+1, data[l:r+1])
		}
	}
}

func TestCorrectness_Max(t *testing.T) {
	const n = 1_000

	rng := rand.New(rand.NewPCG(43, 0))
	data := make([]int, n)
	for i := range data {
		data[i] = rng.IntN(2_000_001) - 1_000_000
	}
	st := sparsetable.New(data, maxInt)

	for i := 0; i < 100_000; i++ {
		l, r := rng.IntN(n), rng.IntN(n)
		if l > r {
			l, r = r, l
		}
		got := st.Query(l, r)
		want := refQuery(data, maxInt, l, r)
		if got != want {
			t.Fatalf("Max 对拍失败：op=%d, l=%d, r=%d, got=%d, want=%d, data[%d:%d]=%v",
				i, l, r, got, want, l, r+1, data[l:r+1])
		}
	}
}

func TestCorrectness_GCD(t *testing.T) {
	const n = 1_000

	rng := rand.New(rand.NewPCG(44, 0))
	data := make([]int, n)
	for i := range data {
		data[i] = rng.IntN(9_999) + 1
	}
	st := sparsetable.New(data, gcd)

	for i := 0; i < 100_000; i++ {
		l, r := rng.IntN(n), rng.IntN(n)
		if l > r {
			l, r = r, l
		}
		got := st.Query(l, r)
		want := refQuery(data, gcd, l, r)
		if got != want {
			t.Fatalf("GCD 对拍失败：op=%d, l=%d, r=%d, got=%d, want=%d, data[%d:%d]=%v",
				i, l, r, got, want, l, r+1, data[l:r+1])
		}
	}
}

func TestCorrectness_LargeRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机验证")
	}

	sizes := []int{100, 1_000, 10_000}
	for _, n := range sizes {
		seedA, seedB := uint64(n), uint64(2026)
		rng := rand.New(rand.NewPCG(seedA, seedB))
		data := make([]int, n)
		for i := range data {
			data[i] = rng.IntN(2_000_001) - 1_000_000
		}
		st := sparsetable.New(data, minInt)

		for i := 0; i < 1_000; i++ {
			l, r := rng.IntN(n), rng.IntN(n)
			if l > r {
				l, r = r, l
			}
			got := st.Query(l, r)
			want := refQuery(data, minInt, l, r)
			if got != want {
				t.Fatalf("LargeRandom 对拍失败：n=%d, op=%d, seedA=%d, seedB=%d, l=%d, r=%d, got=%d, want=%d, data[%d:%d]=%v",
					n, i, seedA, seedB, l, r, got, want, l, r+1, data[l:r+1])
			}
		}
	}
}

func TestCorrectness_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模压力测试")
	}

	const n = 1_000_000

	rng := rand.New(rand.NewPCG(777, 0))
	data := make([]int, n)
	minVal, maxVal := 0, 0
	for i := range data {
		v := rng.IntN(2_000_001) - 1_000_000
		data[i] = v
		if i == 0 || v < minVal {
			minVal = v
		}
		if i == 0 || v > maxVal {
			maxVal = v
		}
	}
	st := sparsetable.New(data, minInt)

	for i := 0; i < 100_000; i++ {
		l, r := rng.IntN(n), rng.IntN(n)
		if l > r {
			l, r = r, l
		}
		got := st.Query(l, r)
		if got < minVal || got > maxVal {
			t.Fatalf("Stress 结果越界：op=%d, l=%d, r=%d, got=%d, dataMin=%d, dataMax=%d",
				i, l, r, got, minVal, maxVal)
		}
	}
}
