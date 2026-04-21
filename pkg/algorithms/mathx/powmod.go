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

// PowMod 返回 (base^exp) % mod，使用二进制快速幂，时间复杂度 O(log exp)。
// mod 必须 > 0，exp 必须 >= 0；mod=1 时结果恒为 0。
// 中间乘积基于 int64 直接相乘，要求 mod^2 < 2^63；负 base 会先规范到 [0, mod)。
//
// 示例：
//
//	PowMod(2, 10, 1_000_000_007) // 1024
//	PowMod(3, 0, 7)              // 1
func PowMod(base, exp, mod int64) int64 {
	assert.AssertFast(mod > 0)
	assert.AssertFast(exp >= 0)

	if mod == 1 {
		return 0
	}

	result := int64(1)
	base %= mod
	if base < 0 {
		base += mod
	}

	for exp > 0 {
		if exp&1 == 1 {
			result = result * base % mod
		}
		base = base * base % mod
		exp >>= 1
	}
	return result
}
