// Package timingwheel_test 基准测试。
package timingwheel_test

import (
	"fmt"
	"sync/atomic"
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

			// 维持规模 n：取消一个，加一个。
			// duration 取 idx%n+1 小时，在 [1h, n*h] 循环，不随迭代单调增长，
			// 保证路由层级稳定（不因 b.N 大小影响测量结果）。
			idx := 0
			for b.Loop() {
				timers[idx%n].Cancel()
				newT, _ := tw.AfterFunc(time.Duration(idx%n+1)*time.Hour, func() {})
				timers[idx%n] = newT
				idx++
			}
		})
	}
}

// BenchmarkAfterFunc_Concurrent 多调用方并发 Add（暴露锁竞争）。
// 对比 p=1 与 p=64 的 ns/op 差距，量化同步开销。
// 工作集固定为 n 个 timer，每次迭代 Cancel 一个旧 timer、Add 一个新 timer，维持集合规模稳定。
// atomic.Pointer 保证多 goroutine 对同一 slot 的读-改-写无数据竞争；
// 极低概率的 slot 碰撞仅导致个别 timer 短暂无法追踪，不影响测量语义。
func BenchmarkAfterFunc_Concurrent(b *testing.B) {
	const n = 100_000
	for _, parallelism := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", parallelism), func(b *testing.B) {
			tw := timingwheel.New()
			tw.Start()
			defer tw.Stop()

			// 预填充 n 个远未来 timer，建立稳定规模基线
			timers := make([]atomic.Pointer[timingwheel.Timer], n)
			for i := range timers {
				t, _ := tw.AfterFunc(time.Duration(i+1)*time.Hour, func() {})
				timers[i].Store(t)
			}

			var idx atomic.Int64
			b.SetParallelism(parallelism)
			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					slot := int(idx.Add(1)-1) % n
					if old := timers[slot].Swap(nil); old != nil {
						old.Cancel()
					}
					newT, _ := tw.AfterFunc(time.Duration(slot+1)*time.Hour, func() {})
					timers[slot].Store(newT)
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
			for b.Loop() {
				timers[idx%n].Cancel()
				newT, _ := tw.AfterFunc(time.Duration(idx%n+1)*time.Hour, func() {})
				timers[idx%n] = newT
				idx++
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
				t, _ := tw.AfterFunc(time.Duration(i+1)*time.Millisecond, func() {})
				timers[i] = t
			}

			b.ResetTimer()
			b.ReportAllocs()
			idx := 0
			i := 0
			// 70% 纯 Add（不 Cancel），30% Add+Cancel（淘汰旧 timer，维持规模 n）。
			// duration 取 i%n+1 毫秒，在 [1ms, n*ms] 循环，不随迭代单调增长。
			for b.Loop() {
				if i%10 < 7 {
					tw.AfterFunc(time.Duration(i%n+1)*time.Millisecond, func() {})
				} else {
					timers[idx%n].Cancel()
					newT, _ := tw.AfterFunc(time.Duration(i%n+1)*time.Millisecond, func() {})
					timers[idx%n] = newT
					idx++
				}
				i++
			}
		})
	}
}
