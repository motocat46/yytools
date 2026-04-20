package slidingwindow_test

import (
	"fmt"
	"testing"

	sw "github.com/motocat46/yytools/pkg/ds/slidingwindow"
)

var benchSizes = []int{100, 1000, 10000, 100000}

// BenchmarkWindow_Add 稳态 Add（窗口已满，每次 Add 淘汰一个旧元素）
func BenchmarkWindow_Add(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			w := sw.New[int](n)
			for i := range n {
				w.Add(i) // 预填充到满
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := n
			for b.Loop() {
				w.Add(i)
				i++
			}
		})
	}
}

// BenchmarkWindow_Max O(1) 均摊 Max 查询
func BenchmarkWindow_Max(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			w := sw.New[int](n)
			for i := range n {
				w.Add(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = w.Max()
			}
		})
	}
}

// BenchmarkWindow_Min O(1) 均摊 Min 查询
func BenchmarkWindow_Min(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			w := sw.New[int](n)
			for i := range n {
				w.Add(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = w.Min()
			}
		})
	}
}

// BenchmarkWindow_Sum O(1) Sum 查询
func BenchmarkWindow_Sum(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			w := sw.New[int](n)
			for i := range n {
				w.Add(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = w.Sum()
			}
		})
	}
}

// BenchmarkWindow_Mixed Add + Max + Min 混合负载（模拟真实使用）
// 操作比例：8 次 Add，1 次 Max，1 次 Min
func BenchmarkWindow_Mixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			w := sw.New[int](n)
			for i := range n {
				w.Add(i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			i := n
			for b.Loop() {
				w.Add(i)
				i++
				if i%10 == 0 {
					_ = w.Max()
					_ = w.Min()
				}
			}
		})
	}
}
