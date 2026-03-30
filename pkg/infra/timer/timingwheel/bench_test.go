// Package timingwheel_test 基准测试。
package timingwheel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/infra/timer/timingwheel"
)

// BenchmarkAfterFunc_Sequential 单调用方 Add 基线（无锁竞争）
func BenchmarkAfterFunc_Sequential(b *testing.B) {
	for _, n := range []int{100, 1000, 10_000, 100_000, 1_000_000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tw := timingwheel.New()
			tw.Start()
			defer tw.Stop()

			// 预填充 n 个远未来 timer（不会触发）
			timers := make([]*timingwheel.Timer, n)
			for i := 0; i < n; i++ {
				t, _ := tw.AfterFunc(time.Duration(i+1)*time.Hour, func() {})
				timers[i] = t
			}
			b.ResetTimer()
			b.ReportAllocs()

			// 维持规模 n：取消一个，加一个
			idx := 0
			i := 0
			for b.Loop() {
				timers[idx%n].Cancel()
				newT, _ := tw.AfterFunc(time.Duration(n+i+1)*time.Millisecond, func() {})
				timers[idx%n] = newT
				idx++
				i++
			}
		})
	}
}

// BenchmarkAfterFunc_Concurrent 多调用方并发 Add（暴露锁竞争）。
// 对比 p=1 与 p=64 的 ns/op 差距，量化同步开销。
func BenchmarkAfterFunc_Concurrent(b *testing.B) {
	for _, parallelism := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", parallelism), func(b *testing.B) {
			tw := timingwheel.New()
			tw.Start()
			defer tw.Stop()

			b.SetParallelism(parallelism)
			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					tw.AfterFunc(time.Duration(i+1)*time.Hour, func() {})
					i++
				}
			})
		})
	}
}

// BenchmarkCancel_Sequential Cancel 的 O(1) 代价（稳定规模 n，取消一个同步补一个）
func BenchmarkCancel_Sequential(b *testing.B) {
	for _, n := range []int{1000, 10_000, 100_000, 1_000_000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tw := timingwheel.New()
			tw.Start()
			defer tw.Stop()

			timers := make([]*timingwheel.Timer, n)
			for i := range timers {
				t, _ := tw.AfterFunc(time.Duration(i+1)*time.Hour, func() {})
				timers[i] = t
			}

			b.ResetTimer()
			b.ReportAllocs()
			idx := 0
			i := 0
			for b.Loop() {
				timers[idx%n].Cancel()
				newT, _ := tw.AfterFunc(time.Duration(n+i+1)*time.Hour, func() {})
				timers[idx%n] = newT
				idx++
				i++
			}
		})
	}
}

// BenchmarkMixed_AddCancel 混合负载：70% 纯 Add，30% Add+Cancel（模拟游戏业务）
func BenchmarkMixed_AddCancel(b *testing.B) {
	for _, n := range []int{1000, 10_000, 100_000, 1_000_000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tw := timingwheel.New()
			tw.Start()
			defer tw.Stop()

			timers := make([]*timingwheel.Timer, n)
			for i := 0; i < n; i++ {
				t, _ := tw.AfterFunc(time.Duration(n+i+1)*time.Millisecond, func() {})
				timers[i] = t
			}

			b.ResetTimer()
			b.ReportAllocs()
			idx := 0
			i := 0
			// 70% 纯 Add（不 Cancel），30% Add+Cancel（淘汰旧 timer，维持规模 n）
			for b.Loop() {
				if i%10 < 7 {
					tw.AfterFunc(time.Duration(n+i+1)*time.Millisecond, func() {})
				} else {
					timers[idx%n].Cancel()
					newT, _ := tw.AfterFunc(time.Duration(n+i+1)*time.Millisecond, func() {})
					timers[idx%n] = newT
					idx++
				}
				i++
			}
		})
	}
}
