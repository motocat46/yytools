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
package mathx

import "github.com/motocat46/yytools/pkg/common/assert"

// Comb 返回组合数 C(n, k) mod mod，适合同一参数只查一次的场景。
// k < 0 或 k > n 返回 0；n 必须 >= 0；mod 必须为大于 1 的质数，且要求 n < mod。
// 该实现基于 Fermat 小定理求逆元，时间复杂度 O(k)。
//
// 示例：
//
//	Comb(5, 2, 1_000_000_007)  // 10
//	Comb(10, 0, 1_000_000_007) // 1
func Comb(n, k, mod int64) int64 {
	validateCombInputs(n, mod)
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1 % mod
	}

	k = normalizeCombK(n, k)
	num, den := buildCombFraction(n, k, mod)
	return num * PowMod(den, mod-2, mod) % mod
}

func validateCombInputs(n, mod int64) {
	assert.AssertFast(n >= 0)
	assert.AssertFast(mod > 1)
	assert.AssertFast(n < mod)
}

func normalizeCombK(n, k int64) int64 {
	if k > n-k {
		return n - k
	}
	return k
}

func buildCombFraction(n, k, mod int64) (int64, int64) {
	num := int64(1)
	den := int64(1)
	for i := int64(0); i < k; i++ {
		num = num * ((n - i) % mod) % mod
		den = den * (i + 1) % mod
	}
	return num, den
}

// CombTable 预计算 [0,maxN] 范围内的阶乘与逆阶乘，适合同一 mod 下高频查询。
// 零值不可用，必须通过 NewCombTable 创建；并发读安全，构造后不应修改内部切片。
type CombTable struct {
	fact    []int64
	invFact []int64
	mod     int64
	maxN    int
}

// NewCombTable 返回组合数查询表，建表时间 O(maxN)，单次查询 O(1)。
// maxN 必须 >= 0 且 maxN < mod；mod 必须为大于 1 的质数。
//
// 示例：
//
//	ct := NewCombTable(1000, 1_000_000_007)
//	ct.C(20, 10) // 184756
func NewCombTable(maxN int, mod int64) *CombTable {
	validateCombTableInputs(maxN, mod)

	fact := buildFactorials(maxN, mod)
	invFact := buildInverseFactorials(fact, mod)

	return &CombTable{
		fact:    fact,
		invFact: invFact,
		mod:     mod,
		maxN:    maxN,
	}
}

// C 返回 C(n, k) mod mod。
// k < 0 或 k > n 返回 0；n 必须满足 0 <= n <= maxN，否则 panic。
func (c *CombTable) C(n, k int) int64 {
	assert.AssertFast(c != nil)
	assert.AssertFast(n >= 0 && n <= c.maxN)
	if k < 0 || k > n {
		return 0
	}
	return c.fact[n] * c.invFact[k] % c.mod * c.invFact[n-k] % c.mod
}

func validateCombTableInputs(maxN int, mod int64) {
	assert.AssertFast(maxN >= 0)
	assert.AssertFast(mod > 1)
	assert.AssertFast(int64(maxN) < mod)
}

func buildFactorials(maxN int, mod int64) []int64 {
	fact := make([]int64, maxN+1)
	fact[0] = 1
	for i := 1; i <= maxN; i++ {
		fact[i] = fact[i-1] * int64(i) % mod
	}
	return fact
}

func buildInverseFactorials(fact []int64, mod int64) []int64 {
	maxN := len(fact) - 1
	invFact := make([]int64, maxN+1)
	invFact[maxN] = PowMod(fact[maxN], mod-2, mod)
	for i := maxN - 1; i >= 0; i-- {
		invFact[i] = invFact[i+1] * int64(i+1) % mod
	}
	return invFact
}
