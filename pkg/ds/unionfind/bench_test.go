package unionfind_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	uf "github.com/motocat46/yytools/pkg/ds/unionfind"
)

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// BenchmarkUnionFind_Union 测量 Union 的均摊代价。
// 为保持集合规模稳定，预先建立 n 个独立节点，每次迭代随机合并两个节点。
// 由于最终会全连通，通过取模重用节点范围保持操作有意义。
func BenchmarkUnionFind_Union(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			u := uf.New[int]()
			// 预注册所有节点
			for i := range n {
				u.Find(i)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				a, bv := rng.IntN(n), rng.IntN(n)
				u.Union(a, bv)
			}
		})
	}
}

// BenchmarkUnionFind_Find 测量 Find（路径压缩）的均摊代价。
func BenchmarkUnionFind_Find(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			u := uf.New[int]()
			// 建链式结构，模拟路径压缩有工作可做的场景
			for i := range n - 1 {
				u.Union(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				u.Find(rng.IntN(n))
			}
		})
	}
}

// BenchmarkUnionFind_Connected 测量 Connected 的均摊代价。
func BenchmarkUnionFind_Connected(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			u := uf.New[int]()
			for i := range n - 1 {
				u.Union(i, i+1)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				u.Connected(rng.IntN(n), rng.IntN(n))
			}
		})
	}
}

// BenchmarkUnionFind_Mixed 混合负载：50% Union，30% Find，20% Connected。
// 模拟真实使用场景，规模稳定。
func BenchmarkUnionFind_Mixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			u := uf.New[int]()
			for i := range n {
				u.Find(i)
			}
			rng := rand.New(rand.NewPCG(42, 0))
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				a, bv := rng.IntN(n), rng.IntN(n)
				switch rng.IntN(10) {
				case 0, 1, 2, 3, 4: // 50% Union
					u.Union(a, bv)
				case 5, 6, 7: // 30% Find
					u.Find(a)
				default: // 20% Connected
					u.Connected(a, bv)
				}
			}
		})
	}
}
