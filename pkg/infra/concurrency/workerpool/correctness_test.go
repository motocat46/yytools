// 正确性专项验证测试
// 针对 WorkerPool 和 Pipeline 的核心并发正确性命题，而非功能覆盖。

package workerpool_test

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/infra/concurrency/workerpool"
)

// ─── 命题 1：精确一次执行 ────────────────────────────────────────────────────
//
// 不变量：Submit 返回 nil 的任务，最终恰好执行一次。
// 在 Submit 和 Close 并发时验证。

func TestCorrectness_ExactlyOnce_并发SubmitClose(t *testing.T) {
	const (
		submitters = 20
		perWorker  = 5_000
	)

	pool := workerpool.NewWorkerPool(8, 100)

	var submitted atomic.Int64
	var executed atomic.Int64

	var wg sync.WaitGroup
	for range submitters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range perWorker {
				err := pool.Submit(context.Background(), func() {
					executed.Add(1)
				})
				if err == nil {
					submitted.Add(1)
				}
			}
		}()
	}

	// 短暂并发后关闭
	time.Sleep(2 * time.Millisecond)
	pool.Close()
	wg.Wait()

	// Close 返回时所有任务已完成，精确计数
	if got, want := executed.Load(), submitted.Load(); got != want {
		t.Errorf("执行数 %d ≠ 成功提交数 %d（精确一次保证被违反）", got, want)
	}
}

// ─── 命题 2：Close 真正等待 ───────────────────────────────────────────────────
//
// 不变量：Close 返回后，所有任务的副作用（包括最后一条语句）必须已完成。

func TestCorrectness_CloseWaitsForCompletion(t *testing.T) {
	pool := workerpool.NewWorkerPool(4, 50)

	const n = 1_000
	results := make([]int64, n)

	for i := range n {
		i := i
		_ = pool.Submit(context.Background(), func() {
			// 模拟有多步副作用的任务，验证 Close 等到最后一步
			runtime.Gosched()
			atomic.StoreInt64(&results[i], 1)
		})
	}

	pool.Close() // 必须等所有任务完成

	// Close 返回后立即检查，不允许有任何 0 值
	for i, v := range results {
		if atomic.LoadInt64(&v) != 1 {
			t.Errorf("Close 返回后 results[%d] 仍为 0，说明任务未完成", i)
		}
	}
}

// ─── 命题 3：Task panic 不死锁 ────────────────────────────────────────────────
//
// 不变量：任务内部 panic 不导致 Wait/Close 永久阻塞。

func TestCorrectness_TaskPanic_NotDeadlock(t *testing.T) {
	pool := workerpool.NewWorkerPool(2, 10)

	var normalExecuted atomic.Int64

	// 提交一个会 panic 的任务
	_ = pool.Submit(context.Background(), func() {
		panic("intentional panic in task")
	})

	// 提交若干正常任务
	for range 5 {
		_ = pool.Submit(context.Background(), func() {
			normalExecuted.Add(1)
		})
	}

	// Close 必须在合理时间内返回，不死锁
	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()

	select {
	case <-done:
		// 正常退出
	case <-time.After(5 * time.Second):
		t.Fatal("Close 超时（5s），疑似 task panic 导致 wg 泄漏，Wait 死锁")
	}

	// panic 不应影响后续正常任务的执行
	if normalExecuted.Load() != 5 {
		t.Errorf("正常任务执行数 = %d，期望 5", normalExecuted.Load())
	}
}

// ─── 命题 4：Submit/Close 高压无 panic ────────────────────────────────────────
//
// 验证 mutex 保护下，Submit 与 Close 并发不触发任何 panic。
// （原 bug：wg.Add 在 closed 检查之后，close(queue) 后 Submit 写入已关闭 channel）

func TestCorrectness_SubmitClose_NoPanic(t *testing.T) {
	// 多轮，每轮大量并发
	for round := range 10 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("第 %d 轮发生 panic: %v", round, r)
				}
			}()

			pool := workerpool.NewWorkerPool(4, 10)
			var count atomic.Int64

			// 并发 submit
			var wg sync.WaitGroup
			for range 50 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for range 100 {
						if pool.Submit(context.Background(), func() {
							count.Add(1)
						}) != nil {
							return // pool 已关闭，停止提交
						}
					}
				}()
			}

			// 随机时刻关闭
			time.Sleep(time.Duration(round) * time.Microsecond)
			pool.Close()
			wg.Wait()
		}()
	}
}

// ─── 命题 5：Pipeline 结果完整性 ──────────────────────────────────────────────
//
// 不变量：每个输入元素恰好产生一个输出结果。

func TestCorrectness_Pipeline_ResultCompleteness(t *testing.T) {
	const n = 50_000

	p := workerpool.NewPipeline(8, 200, func(i int) (int, error) {
		return i * 2, nil
	})
	defer p.Close()

	in := make(chan int, 500)
	go func() {
		for i := range n {
			in <- i
		}
		close(in)
	}()

	var count int
	var sum int64
	for r := range p.Process(in) {
		if r.Err != nil {
			t.Fatalf("意外错误: %v", r.Err)
		}
		count++
		sum += int64(r.Value)
	}

	if count != n {
		t.Errorf("输出数 %d ≠ 输入数 %d", count, n)
	}
	// 验证结果正确性：sum = 2 * (0+1+...+(n-1)) = n*(n-1)
	expected := int64(n) * int64(n-1)
	if sum != expected {
		t.Errorf("结果总和 %d ≠ 期望 %d", sum, expected)
	}
}

// ─── 命题 6：Pipeline 输出顺序无关，数量严格匹配 ────────────────────────────

func TestCorrectness_Pipeline_NoResultLost_NoDuplicate(t *testing.T) {
	const n = 10_000

	p := workerpool.NewPipeline(4, 100, func(i int) (int, error) {
		return i, nil
	})
	defer p.Close()

	in := make(chan int, 100)
	go func() {
		for i := range n {
			in <- i
		}
		close(in)
	}()

	seen := make([]atomic.Int32, n)
	for r := range p.Process(in) {
		if r.Err != nil {
			t.Fatalf("意外错误: %v", r.Err)
		}
		seen[r.Value].Add(1)
	}

	for i, v := range seen {
		if cnt := v.Load(); cnt != 1 {
			t.Errorf("元素 %d 出现 %d 次（期望恰好 1 次）", i, cnt)
		}
	}
}
