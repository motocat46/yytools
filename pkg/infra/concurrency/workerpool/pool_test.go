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

func TestWorkerPool_Submit_正常执行(t *testing.T) {
	pool := workerpool.NewWorkerPool(2, 10)
	defer pool.Close()

	var count atomic.Int64
	for range 5 {
		if err := pool.Submit(context.Background(), func() { count.Add(1) }); err != nil {
			t.Fatalf("Submit 返回错误: %v", err)
		}
	}
	pool.Wait()
	if count.Load() != 5 {
		t.Errorf("执行数 = %d, 期望 5", count.Load())
	}
}

func TestWorkerPool_Submit_PoolClosed(t *testing.T) {
	pool := workerpool.NewWorkerPool(2, 10)
	pool.Close()

	err := pool.Submit(context.Background(), func() {})
	if err != workerpool.ErrPoolClosed {
		t.Errorf("期望 ErrPoolClosed，得到 %v", err)
	}
}

func TestWorkerPool_Submit_CtxCancelled(t *testing.T) {
	pool := workerpool.NewWorkerPool(1, 0)
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
}

func TestWorkerPool_Wait_多次调用(t *testing.T) {
	pool := workerpool.NewWorkerPool(2, 10)
	defer pool.Close()

	_ = pool.Submit(context.Background(), func() {})
	pool.Wait()
	pool.Wait()
}

func TestWorkerPool_Workers1_QueueSize0(t *testing.T) {
	pool := workerpool.NewWorkerPool(1, 0)
	defer pool.Close()

	var executed atomic.Bool
	_ = pool.Submit(context.Background(), func() { executed.Store(true) })
	pool.Wait()
	if !executed.Load() {
		t.Error("任务未执行")
	}
}

func TestWorkerPool_集成_无遗漏无重复(t *testing.T) {
	const n = 100_000
	pool := workerpool.NewWorkerPool(8, 1000)
	defer pool.Close()

	var count atomic.Int64
	for range n {
		_ = pool.Submit(context.Background(), func() { count.Add(1) })
	}
	pool.Wait()
	if count.Load() != n {
		t.Errorf("执行数 = %d, 期望 %d", count.Load(), n)
	}
}

func TestWorkerPool_压力_百万任务(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过百万级压力测试")
	}
	const n = 1_000_000
	pool := workerpool.NewWorkerPool(16, 10000)
	defer pool.Close()

	var count atomic.Int64
	for range n {
		_ = pool.Submit(context.Background(), func() { count.Add(1) })
	}
	pool.Wait()
	if count.Load() != n {
		t.Errorf("执行数 = %d, 期望 %d", count.Load(), n)
	}
}

func BenchmarkWorkerPool_Submit(b *testing.B) {
	for _, workers := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			pool := workerpool.NewWorkerPool(workers, b.N+1)
			b.ResetTimer()
			b.ReportAllocs()
			for range b.N {
				_ = pool.Submit(context.Background(), func() {})
			}
			pool.Wait()
			pool.Close()
		})
	}
}
