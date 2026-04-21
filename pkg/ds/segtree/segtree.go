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

// Package segtree 提供 ACL 风格的泛型线段树骨架。
// 当前已实现构造、长度查询、单点赋值和全量查询；区间查询与更复杂更新在后续任务补齐。
package segtree

import "github.com/motocat46/yytools/pkg/common/assert"

// SegTree 是 ACL 风格泛型线段树（带 lazy 传播）。
// 零值不可用，必须通过 New 创建；非并发安全。
type SegTree[T, L any] struct {
	n        int
	tree     []T
	lazy     []L
	identity T
	merge    func(T, T) T
	lazyZero L
	apply    func(T, L, int) T
	compose  func(L, L) L
}

// New 创建容量为 n 的线段树。
//
// 参数语义：
//   - identity：T 的单位元（sum→0, min→MaxInt, max→MinInt）；merge(identity, x)==x 须成立
//   - merge：合并两段结果，须满足结合律
//   - lazyZero：lazy 的零值，表示"无操作"；apply(val, lazyZero, size)==val 须成立；compose(lazyZero, x)==x 须成立
//   - apply：将 lazy 作用到节点值；第三个参数为区间长度（sum 场景需用到）
//   - compose：组合两个 lazy，新 lazy 在前，旧 lazy 在后
//
// n 须 > 0，否则 assert panic。
func New[T, L any](
	n int,
	identity T,
	merge func(T, T) T,
	lazyZero L,
	apply func(T, L, int) T,
	compose func(L, L) L,
) *SegTree[T, L] {
	assert.Assert(n > 0, "segtree: n 须 > 0，实际值:", n)
	assert.Assert(merge != nil, "segtree: merge 不能为空")
	assert.Assert(apply != nil, "segtree: apply 不能为空")
	assert.Assert(compose != nil, "segtree: compose 不能为空")
	s := &SegTree[T, L]{
		n:        n,
		tree:     make([]T, 4*n),
		lazy:     make([]L, 4*n),
		identity: identity,
		merge:    merge,
		lazyZero: lazyZero,
		apply:    apply,
		compose:  compose,
	}
	for i := range s.tree {
		s.tree[i] = identity
	}
	for i := range s.lazy {
		s.lazy[i] = lazyZero
	}
	return s
}

// Len 返回容量 n。O(1)。
func (s *SegTree[T, L]) Len() int { return s.n }

// QueryAll 返回全区间 [0, n) 的合并结果。O(1)。
func (s *SegTree[T, L]) QueryAll() T { return s.tree[1] }

// pushup 用左右子节点更新当前节点。
func (s *SegTree[T, L]) pushup(v int) {
	s.tree[v] = s.merge(s.tree[2*v], s.tree[2*v+1])
}

// pushdown 将当前节点 lazy 下传给左右子节点，然后将自身 lazy 清零。
// lsize/rsize 为左右子区间长度。始终下传（不判零值），依赖 lazyZero 为 apply 单位元。
func (s *SegTree[T, L]) pushdown(v, lsize, rsize int) {
	s.tree[2*v] = s.apply(s.tree[2*v], s.lazy[v], lsize)
	s.tree[2*v+1] = s.apply(s.tree[2*v+1], s.lazy[v], rsize)
	s.lazy[2*v] = s.compose(s.lazy[v], s.lazy[2*v])
	s.lazy[2*v+1] = s.compose(s.lazy[v], s.lazy[2*v+1])
	s.lazy[v] = s.lazyZero
}

// Set 将下标 i 的元素赋值为 val。i ∈ [0, n)，O(log n)。
// 越界触发 assert panic。
func (s *SegTree[T, L]) Set(i int, val T) {
	if i < 0 || i >= s.n {
		assert.Assert(false, "segtree: Set 下标越界，i=", i, "n=", s.n)
	}
	s.setAt(1, 0, s.n-1, i, val)
}

// setAt 在区间 [l, r] 内递归定位下标 i，并用 val 覆盖叶子后向上回收。
func (s *SegTree[T, L]) setAt(v, l, r, i int, val T) {
	if l == r {
		s.tree[v] = val
		s.lazy[v] = s.lazyZero
		return
	}
	mid := (l + r) / 2
	s.pushdown(v, mid-l+1, r-mid)
	if i <= mid {
		s.setAt(2*v, l, mid, i, val)
	} else {
		s.setAt(2*v+1, mid+1, r, i, val)
	}
	s.pushup(v)
}

// Apply 对区间 [l, r] 内每个元素应用 lazy。l ≤ r，均 ∈ [0, n)，O(log n)。
// 越界或 l > r 触发 assert panic。
func (s *SegTree[T, L]) Apply(l, r int, lazy L) {
	if l < 0 || r >= s.n || l > r {
		assert.Assert(false, "segtree: Apply 参数非法，l=", l, "r=", r, "n=", s.n)
	}
	s.applyRange(1, 0, s.n-1, l, r, lazy)
}

func (s *SegTree[T, L]) applyRange(v, nodeL, nodeR, l, r int, lazy L) {
	if l <= nodeL && nodeR <= r {
		s.tree[v] = s.apply(s.tree[v], lazy, nodeR-nodeL+1)
		s.lazy[v] = s.compose(lazy, s.lazy[v])
		return
	}
	mid := (nodeL + nodeR) / 2
	s.pushdown(v, mid-nodeL+1, nodeR-mid)
	if l <= mid {
		s.applyRange(2*v, nodeL, mid, l, r, lazy)
	}
	if r > mid {
		s.applyRange(2*v+1, mid+1, nodeR, l, r, lazy)
	}
	s.pushup(v)
}

// Query 返回区间 [l, r] 的合并结果。l ≤ r，均 ∈ [0, n)，O(log n)。
// 越界或 l > r 触发 assert panic。
func (s *SegTree[T, L]) Query(l, r int) T {
	if l < 0 || r >= s.n || l > r {
		assert.Assert(false, "segtree: Query 参数非法，l=", l, "r=", r, "n=", s.n)
	}
	return s.queryRange(1, 0, s.n-1, l, r)
}

func (s *SegTree[T, L]) queryRange(v, nodeL, nodeR, l, r int) T {
	if l <= nodeL && nodeR <= r {
		return s.tree[v]
	}
	mid := (nodeL + nodeR) / 2
	s.pushdown(v, mid-nodeL+1, nodeR-mid)
	if r <= mid {
		return s.queryRange(2*v, nodeL, mid, l, r)
	}
	if l > mid {
		return s.queryRange(2*v+1, mid+1, nodeR, l, r)
	}
	return s.merge(
		s.queryRange(2*v, nodeL, mid, l, r),
		s.queryRange(2*v+1, mid+1, nodeR, l, r),
	)
}
