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

// Package unionfind 实现泛型并查集（Disjoint Set Union）。
// 支持任意 comparable 类型元素，提供 O(α) 均摊的 Union/Find/Connected/Size/Count。
// 元素首次使用时自动注册为单独组，无需显式 Add。
// 非并发安全，并发访问由调用方负责加锁。
package unionfind

// UnionFind 是泛型并查集。
// 零值不可用，必须通过 New 创建；非并发安全。
type UnionFind[T comparable] struct {
	parent map[T]T   // parent[x] = x 的父节点；根节点的父节点是自身
	size   map[T]int // size[x] 仅在 x 为根时有意义，表示组大小
	count  int       // 当前独立组数
}

// New 创建空的并查集。
func New[T comparable]() *UnionFind[T] {
	return &UnionFind[T]{
		parent: make(map[T]T),
		size:   make(map[T]int),
	}
}

// register 若 x 尚未注册，将其注册为单独组。
func (uf *UnionFind[T]) register(x T) {
	if _, ok := uf.parent[x]; !ok {
		uf.parent[x] = x
		uf.size[x] = 1
		uf.count++
	}
}

// Find 返回 x 所在组的代表元（根节点）。O(α) 均摊（迭代路径压缩）。
// x 若未注册，自动注册为单独组，返回 x 自身。
func (uf *UnionFind[T]) Find(x T) T {
	uf.register(x)
	// 第一遍：找根
	root := x
	for uf.parent[root] != root {
		root = uf.parent[root]
	}
	// 第二遍：路径压缩，将路径上所有节点直接指向根
	for uf.parent[x] != root {
		next := uf.parent[x]
		uf.parent[x] = root
		x = next
	}
	return root
}

// Union 合并 a 和 b 所在的组。
// 返回 true 表示发生了实际合并；false 表示已在同一组。
// a、b 若未注册，自动注册为单独组后再合并。
func (uf *UnionFind[T]) Union(a, b T) bool {
	ra, rb := uf.Find(a), uf.Find(b)
	if ra == rb {
		return false
	}
	// 按大小合并：小树接到大树下
	if uf.size[ra] < uf.size[rb] {
		ra, rb = rb, ra
	}
	uf.parent[rb] = ra
	uf.size[ra] += uf.size[rb]
	uf.count--
	return true
}

// Connected 报告 a 和 b 是否在同一组。O(α) 均摊。
// a、b 若未注册，自动注册为单独组（两个新单独组不连通，返回 false）。
func (uf *UnionFind[T]) Connected(a, b T) bool {
	return uf.Find(a) == uf.Find(b)
}

// Count 返回当前独立组的数量。O(1)。
func (uf *UnionFind[T]) Count() int { return uf.count }

// Size 返回 a 所在组的元素数量。O(α) 均摊。
// a 若未注册，自动注册为单独组，返回 1。
func (uf *UnionFind[T]) Size(a T) int {
	return uf.size[uf.Find(a)]
}
