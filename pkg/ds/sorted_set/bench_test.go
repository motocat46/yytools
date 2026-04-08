package sorted_set

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// ---- Benchmark ----
//
// 运行方式：
//   go test -bench=. -benchmem ./pkg/ds/sorted_set/
//   go test -bench=BenchmarkSortedSet_Mixed/n=10000 -benchtime=5s ./pkg/ds/sorted_set/
//   go test -bench=. -count=3 ./pkg/ds/sorted_set/

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// --- 单操作基准（稳定集合规模，测单次 O(log n) 代价）---
//
// 所有单操作基准均在预填充的稳定集合上运行，集合规模在整个基准过程中保持不变，
// 确保每次迭代的操作代价反映的是"规模为 n 时的成本"，而非建集合的摊销成本。

// BenchmarkSortedSet_Get 测量 O(1) 哈希查找
func BenchmarkSortedSet_Get(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				ss.Get(i%size + 1)
				i++
			}
		})
	}
}

// BenchmarkSortedSet_GetRank 测量 O(log n) 排名查询
func BenchmarkSortedSet_GetRank(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				ss.GetRank(i%size + 1)
				i++
			}
		})
	}
}

// BenchmarkSortedSet_UpdateScore 测量 O(log n) 分数更新
// UpdateScore 不改变集合大小，天然适合稳定集合测量
func BenchmarkSortedSet_UpdateScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				// 使用随机 score 触发真实的重排序路径（删除重插）
				ss.UpdateScore(i%size+1, float64((i*7+3)%size))
				i++
			}
		})
	}
}

// BenchmarkSortedSet_Insert 测量 O(log n) 插入
// 每次插入后删除 rank=1 的元素以维持集合规模，防止 n 随 b.N 增长导致后期成本失真
func BenchmarkSortedSet_Insert(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			newKey := size + 1
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				oldest := ss.GetByRank(1)
				ss.Delete(oldest.Key)
				ss.Insert(&NodeData[int, int]{Key: newKey, Score: float64(newKey % size), Val: newKey})
				newKey++
			}
		})
	}
}

// BenchmarkSortedSet_GetRangeByRank 测量 O(log n + k) 排名范围查询（k = n/2）
func BenchmarkSortedSet_GetRangeByRank(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := size / 2
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				ss.GetRangeByRank(1, mid)
			}
		})
	}
}

// BenchmarkSortedSet_GetRangeByScore 测量 O(log n + k) 分数范围查询（k ≈ n/2）
func BenchmarkSortedSet_GetRangeByScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := float64(size / 2)
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				ss.GetRangeByScore(1, false, mid, false)
			}
		})
	}
}

// BenchmarkSortedSet_GetMin 测量 O(1) 最小值访问（Head.Levels[0].Forward）
func BenchmarkSortedSet_GetMin(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				ss.GetMin()
			}
		})
	}
}

// BenchmarkSortedSet_GetMax 测量 O(1) 最大值访问（Tail 指针）
func BenchmarkSortedSet_GetMax(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				ss.GetMax()
			}
		})
	}
}

// BenchmarkSortedSet_CountByScore 测量 O(log n + k) 范围计数（k ≈ n/2，无切片分配）
// 与 BenchmarkSortedSet_GetRangeByScore 对比，可量化零分配带来的收益
func BenchmarkSortedSet_CountByScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := float64(size / 2)
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				ss.CountByScore(1, false, mid, false)
			}
		})
	}
}

// --- 混合负载基准（模拟真实排行榜场景）---
//
// BenchmarkSortedSet_Mixed 在稳定规模的集合上混合执行多种操作，
// 模拟排行榜典型负载：
//
//	50% UpdateScore（玩家分数更新，最高频）
//	25% GetRank（查询自己排名）
//	15% GetRangeByRank(1,10)（查看排行榜前 10）
//	 5% GetRangeByScore（按积分段查询）
//	 5% Insert+Delete（玩家进出场，维持集合规模）
//
// 该基准反映的是整体吞吐量，而非单一操作的延迟。
func BenchmarkSortedSet_Mixed(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			rng := rand.New(rand.NewPCG(42, 0))
			nextKey := size + 1
			scoreRange := float64(size * 10)

			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				key := rng.IntN(size) + 1
				op := rng.IntN(100)
				switch {
				case op < 50: // UpdateScore 50%
					ss.UpdateScore(key, rng.Float64()*scoreRange)
				case op < 75: // GetRank 25%
					ss.GetRank(key)
				case op < 90: // GetRangeByRank(1,10) 15%
					ss.GetRangeByRank(1, 10)
				case op < 95: // GetRangeByScore 5%
					lo := rng.Float64() * scoreRange / 2
					ss.GetRangeByScore(lo, false, lo+scoreRange/4, false)
				default: // Insert+Delete 5%（维持规模）
					oldest := ss.GetByRank(1)
					if oldest != nil {
						ss.Delete(oldest.Key)
					}
					ss.Insert(&NodeData[int, int]{
						Key:   nextKey,
						Score: rng.Float64() * scoreRange,
						Val:   nextKey,
					})
					nextKey++
				}
			}
		})
	}
}
