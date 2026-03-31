// Package sampling.

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
// 创建日期:2026/2/28

// Package sampling 提供通用的随机采样算法。
package sampling

import (
	"math/rand/v2"
	"slices"

	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/base"
)

// SampleKDistinctFloyd 从 [lo..hi]（含端点）中均匀采样 k 个不重复的值。
// 时间复杂度 O(k)，空间复杂度 O(k)。
// 内部以 0-indexed 虚拟空间做 Floyd 算法，出口转换为 T。
func SampleKDistinctFloyd[T base.Integer](lo, hi T, k int, r *rand.Rand) []T {
	assert.Assert(lo <= hi, "invalid range")
	m := int(hi) - int(lo) + 1
	assert.Assert(k >= 0 && k <= m, "invalid k")

	chosen := make(map[int]struct{}, k)
	c := make([]int, 0, k)

	for j := m - k; j < m; j++ {
		t := r.IntN(j + 1) // [0..j]
		if _, ok := chosen[t]; ok {
			chosen[j] = struct{}{}
			c = append(c, j)
		} else {
			chosen[t] = struct{}{}
			c = append(c, t)
		}
	}

	loInt := int(lo)
	out := make([]T, k)
	for i, v := range c {
		out[i] = T(v + loInt)
	}
	return out
}

// SampleWithMinGap 从 [L..R] 中采样 k 个值，保证任意相邻两值之间间隔至少 gap，返回升序切片。
// gap=0 时退化为普通不重复采样；k=0 时返回 nil。
// 可用空间不足（N-(k-1)*gap < k）时触发 assert。
// 时间复杂度 O(k log k)，空间复杂度 O(k)。
//
// 示例（gap=2）：
//
//	SampleWithMinGap(1, 10, 3, 2, r) 可能返回 [1, 3, 5]，相邻间隔均 >= 2
func SampleWithMinGap[T base.Integer](L, R T, k, gap int, r *rand.Rand) []T {
	if k <= 0 {
		return nil
	}
	assert.Assert(gap >= 0, "gap must be >= 0")
	N := int(R) - int(L) + 1
	assert.Assert(N > 0, "invalid [L..R]")

	M := N - (k-1)*gap
	assert.Assert(M >= k, "impossible constraints: not enough room")

	// 压缩空间为 [L .. L+M-1]，采样后还原间隔
	c := SampleKDistinctFloyd(L, L+T(M-1), k, r)
	slices.SortFunc(c, func(a, b T) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})

	a := make([]T, k)
	for i := range k {
		a[i] = c[i] + T(i*gap)
	}
	return a
}
