// 版权所有(Copyright)[yangyuan]
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

package workerpool_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/motocat46/yytools/pkg/infra/concurrency/workerpool"
)

func TestMain(m *testing.M) {
	// 静默 slog 输出：测试只验证行为，不断言日志内容
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	goleak.VerifyTestMain(m)
}

// poolFactory 封装构造函数，让所有测试对两种锁实现均运行一遍。
type poolFactory struct {
	name string
	new  func(workers, queueSize int) *workerpool.WorkerPool
}

var allFactories = []poolFactory{
	{"RWMutex", workerpool.NewWorkerPool},
	{"Mutex", workerpool.NewWorkerPoolMutex},
}

func TestWorkerPool_Submit_Normal(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			pool := f.new(2, 10)
			defer pool.Close()

			var count atomic.Int64
			for range 5 {
				if err := pool.Submit(context.Background(), func() { count.Add(1) }); err != nil {
					t.Fatalf("Submit 返回错误: %v", err)
				}
			}
			pool.Wait()
			if count.Load() != 5 {
				t.Errorf("执行数 = %d，期望 5", count.Load())
			}
		})
	}
}

func TestWorkerPool_Submit_PoolClosed(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			pool := f.new(2, 10)
			pool.Close()

			err := pool.Submit(context.Background(), func() {})
			if err != workerpool.ErrPoolClosed {
				t.Errorf("期望 ErrPoolClosed，得到 %v", err)
			}
		})
	}
}

func TestWorkerPool_Submit_CtxCancelled(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			pool := f.new(1, 0)
			defer pool.Close()

			block := make(chan struct{})
			_ = pool.Submit(context.Background(), func() { <-block })

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()

			err := pool.Submit(ctx, func() {})
			if err == nil {
				t.Error("期望 ctx 超时错误，得到 nil")
			}
			close(block)
		})
	}
}

func TestWorkerPool_Wait_MultipleCallsAllowed(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			pool := f.new(2, 10)
			defer pool.Close()

			_ = pool.Submit(context.Background(), func() {})
			pool.Wait()
			pool.Wait()
		})
	}
}

func TestWorkerPool_Workers1_QueueSize0(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			pool := f.new(1, 0)
			defer pool.Close()

			var executed atomic.Bool
			_ = pool.Submit(context.Background(), func() { executed.Store(true) })
			pool.Wait()
			if !executed.Load() {
				t.Error("任务未执行")
			}
		})
	}
}

// TestWorkerPool_Integration_Sequential 单调用方顺序提交，验证基本正确性。
func TestWorkerPool_Integration_Sequential(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			const n = 100_000
			pool := f.new(8, 1000)
			defer pool.Close()

			var count atomic.Int64
			for range n {
				_ = pool.Submit(context.Background(), func() { count.Add(1) })
			}
			pool.Wait()
			if count.Load() != n {
				t.Errorf("执行数 = %d，期望 %d", count.Load(), n)
			}
		})
	}
}

// TestWorkerPool_Integration_Concurrent 多调用方并发提交，验证并发场景下无遗漏无重复。
// 单调用方测试无法触发 Submit 内的锁竞争路径，此测试为并发原语的必要覆盖。
func TestWorkerPool_Integration_Concurrent(t *testing.T) {
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			const (
				callers   = 20
				perCaller = 5_000
				total     = callers * perCaller
			)
			pool := f.new(8, 1000)
			defer pool.Close()

			var wg sync.WaitGroup
			var count atomic.Int64
			for range callers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for range perCaller {
						_ = pool.Submit(context.Background(), func() { count.Add(1) })
					}
				}()
			}
			wg.Wait()
			pool.Wait()
			if got := count.Load(); got != total {
				t.Errorf("执行数 = %d，期望 %d（%d 个调用方 × %d）", got, total, callers, perCaller)
			}
		})
	}
}

// TestWorkerPool_Stress_Sequential 单调用方百万任务压力测试。
func TestWorkerPool_Stress_Sequential(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			const n = 1_000_000
			pool := f.new(16, 10000)
			defer pool.Close()

			var count atomic.Int64
			for range n {
				_ = pool.Submit(context.Background(), func() { count.Add(1) })
			}
			pool.Wait()
			if count.Load() != n {
				t.Errorf("执行数 = %d，期望 %d", count.Load(), n)
			}
		})
	}
}

// TestWorkerPool_Stress_Concurrent 多调用方并发百万任务压力测试。
// 验证高并发提交下正确性不退化。
func TestWorkerPool_Stress_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发压力测试")
	}
	for _, f := range allFactories {
		t.Run(f.name, func(t *testing.T) {
			const (
				callers   = 50
				perCaller = 20_000 // 总量 = 1,000,000
			)
			pool := f.new(16, 10000)
			defer pool.Close()

			var wg sync.WaitGroup
			var count atomic.Int64
			for range callers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for range perCaller {
						_ = pool.Submit(context.Background(), func() { count.Add(1) })
					}
				}()
			}
			wg.Wait()
			pool.Wait()
			if got := count.Load(); got != callers*perCaller {
				t.Errorf("执行数 = %d，期望 %d（%d 个调用方 × %d）", got, callers*perCaller, callers, perCaller)
			}
		})
	}
}

// BenchmarkWorkerPool_Submit_Sequential RWMutex 版（默认）单调用方顺序提交，无竞争基线。
func BenchmarkWorkerPool_Submit_Sequential(b *testing.B) {
	pool := workerpool.NewWorkerPool(8, b.N+1)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = pool.Submit(context.Background(), func() {})
	}
	pool.Wait()
	pool.Close()
}

// BenchmarkWorkerPoolMutex_Submit_Sequential Mutex 版单调用方顺序提交，对比基线。
func BenchmarkWorkerPoolMutex_Submit_Sequential(b *testing.B) {
	pool := workerpool.NewWorkerPoolMutex(8, b.N+1)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = pool.Submit(context.Background(), func() {})
	}
	pool.Wait()
	pool.Close()
}

// BenchmarkWorkerPool_Submit_Concurrent RWMutex 版（默认）多调用方并发提交。
func BenchmarkWorkerPool_Submit_Concurrent(b *testing.B) {
	for _, parallelism := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", parallelism), func(b *testing.B) {
			pool := workerpool.NewWorkerPool(parallelism*2, b.N+parallelism)
			b.SetParallelism(parallelism)
			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = pool.Submit(context.Background(), func() {})
				}
			})
			pool.Wait()
			pool.Close()
		})
	}
}

// BenchmarkWorkerPoolMutex_Submit_Concurrent Mutex 版多调用方并发提交，与 RWMutex 版横向对比。
func BenchmarkWorkerPoolMutex_Submit_Concurrent(b *testing.B) {
	for _, parallelism := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", parallelism), func(b *testing.B) {
			pool := workerpool.NewWorkerPoolMutex(parallelism*2, b.N+parallelism)
			b.SetParallelism(parallelism)
			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = pool.Submit(context.Background(), func() {})
				}
			})
			pool.Wait()
			pool.Close()
		})
	}
}
