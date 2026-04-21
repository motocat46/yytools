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

// Package sparsetable 提供泛型稀疏表（Sparse Table）骨架。
// 适用于幂等 merge（如区间 min/max/GCD）的静态数据场景：
// 预处理复杂度 O(n log n)，后续区间查询可做到 O(1)。
// 当前提供构造、长度查询和区间 Query；使用时需要先通过 New 初始化，
// 包本身非并发安全。
package sparsetable

import (
	"fmt"
	"math/bits"

	"github.com/motocat46/yytools/pkg/common/assert"
)

// SparseTable 是泛型稀疏表。
// 数据一旦构建后视为静态；使用前需通过 New 初始化；非并发安全。
type SparseTable[T any] struct {
	n     int
	st    [][]T
	merge func(T, T) T
}

// New 基于 data 构建稀疏表并返回结果。
//
// data 必须非空，否则触发 assert panic。构造时会把 data 复制到内部存储，
// 因此调用方后续修改原切片不会影响稀疏表内容。
// merge 表示幂等合并函数，适用于 min/max/GCD 等静态区间查询语义。
func New[T any](data []T, merge func(T, T) T) *SparseTable[T] {
	assert.Assert(len(data) > 0, "sparsetable: data 不可为空")
	assert.Assert(merge != nil, "sparsetable: merge 不能为空")

	n := len(data)
	levels := bits.Len(uint(n))
	st := make([][]T, levels)
	st[0] = make([]T, n)
	copy(st[0], data)

	for level := 1; level < levels; level++ {
		span := 1 << level
		half := span >> 1
		width := n - span + 1
		st[level] = make([]T, width)
		for i := 0; i < width; i++ {
			st[level][i] = merge(st[level-1][i], st[level-1][i+half])
		}
	}

	return &SparseTable[T]{
		n:     n,
		st:    st,
		merge: merge,
	}
}

// Len 返回原始数据长度 n。O(1)。
func (s *SparseTable[T]) Len() int {
	return s.n
}

// Query 返回区间 [l, r] 的 merge 结果。[l, r] 闭区间，0-indexed，O(1)。
// l > r 或越界触发 assert panic。
func (s *SparseTable[T]) Query(l, r int) T {
	if l < 0 || r >= s.n || l > r {
		panic(fmt.Sprintf("sparsetable: Query 参数非法，l=%d r=%d n=%d", l, r, s.n))
	}
	k := bits.Len(uint(r-l+1)) - 1
	return s.merge(s.st[k][l], s.st[k][r-(1<<k)+1])
}
