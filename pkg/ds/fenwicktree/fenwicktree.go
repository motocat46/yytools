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

// Package fenwicktree 实现泛型树状数组（Binary Indexed Tree / Fenwick Tree）。
// 支持任意 base.Number 类型，提供 O(log n) 的单点更新和前缀和 / 区间和查询。
// 对外 0-indexed，内部 1-indexed；非并发安全。
package fenwicktree

import (
	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/common/base"
)

// FenwickTree 是泛型树状数组。
// 零值不可用，必须通过 New 创建；非并发安全。
type FenwickTree[T base.Number] struct {
	tree []T // 长度 n+1，tree[0] 不使用；内部下标 1..n
	n    int
}

// New 创建容量为 n 的树状数组，n 须 > 0，否则 panic。
func New[T base.Number](n int) *FenwickTree[T] {
	assert.Assert(n > 0, "fenwicktree: n 须 > 0，实际值:", n)
	return &FenwickTree[T]{
		tree: make([]T, n+1),
		n:    n,
	}
}

// Build 基于 nums 在线性时间内构建树状数组。nums 不可为空，否则 panic。
func Build[T base.Number](nums []T) *FenwickTree[T] {
	assert.Assert(len(nums) > 0, "fenwicktree: nums 不可为空")
	n := len(nums)
	f := &FenwickTree[T]{
		tree: make([]T, n+1),
		n:    n,
	}
	for i := 1; i <= n; i++ {
		f.tree[i] += nums[i-1]
		parent := i + (i & -i)
		if parent <= n {
			f.tree[parent] += f.tree[i]
		}
	}
	return f
}

// Len 返回树状数组的容量 n。O(1)。
func (f *FenwickTree[T]) Len() int {
	return f.n
}

// Add 将下标 i 的元素加上 delta。i ∈ [0, n)，O(log n)。
func (f *FenwickTree[T]) Add(i int, delta T) {
	if i < 0 || i >= f.n {
		assert.Assert(false, "fenwicktree: Add 下标越界，i=", i, "n=", f.n)
	}
	for pos := i + 1; pos <= f.n; pos += pos & (-pos) {
		f.tree[pos] += delta
	}
}

// prefixSum 返回内部 [1..i+1] 的前缀和（不做越界检查，由公开方法负责）。
func (f *FenwickTree[T]) prefixSum(i int) T {
	var sum T
	for pos := i + 1; pos > 0; pos -= pos & (-pos) {
		sum += f.tree[pos]
	}
	return sum
}

// PrefixSum 返回 [0..i] 的前缀和。i ∈ [0, n)，O(log n)。
func (f *FenwickTree[T]) PrefixSum(i int) T {
	if i < 0 || i >= f.n {
		assert.Assert(false, "fenwicktree: PrefixSum 下标越界，i=", i, "n=", f.n)
	}
	return f.prefixSum(i)
}

// RangeSum 返回 [l..r] 的区间和。l ≤ r，均 ∈ [0, n)，O(log n)。
func (f *FenwickTree[T]) RangeSum(l, r int) T {
	if l < 0 || r >= f.n || l > r {
		assert.Assert(false, "fenwicktree: RangeSum 参数非法，l=", l, "r=", r, "n=", f.n)
	}
	if l == 0 {
		return f.prefixSum(r)
	}
	return f.prefixSum(r) - f.prefixSum(l-1)
}
