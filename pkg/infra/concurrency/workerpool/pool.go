// Package workerpool 提供有界 goroutine 池，限制并发数并支持任务队列。

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/3/24
package workerpool

import (
	"context"
	"errors"
	"sync"

	"github.com/motocat46/yytools/pkg/common/assert"
	"github.com/motocat46/yytools/pkg/infra/safeexec"
)

// ErrPoolClosed 在 pool 已关闭后调用 Submit 时返回。
var ErrPoolClosed = errors.New("workerpool: pool is closed")

// WorkerPool 是固定大小的 goroutine 池。
//
// 用法：
//
//	pool := NewWorkerPool(4, 100)
//	defer pool.Close()
//
//	if err := pool.Submit(ctx, func() { doWork() }); err != nil {
//	    // ctx 取消或 pool 已关闭
//	}
//	pool.Wait()
type WorkerPool struct {
	queue  chan func()
	wg     sync.WaitGroup
	locker submitLocker // 抽象锁，保护 closed 标志与 wg.Add 的原子性
	closed bool
	once   sync.Once
	stop   chan struct{} // close 广播通知所有 worker 退出
}

// NewWorkerPool 创建固定大小的 worker pool，使用 RWMutex（默认）。
// Submit 持读锁（并发提交不互斥），Close 持写锁（独占），并发吞吐优于 Mutex 约 20%。
// workers：并发 goroutine 数；queueSize：待执行队列容量（提供背压）。
func NewWorkerPool(workers, queueSize int) *WorkerPool {
	return newWorkerPool(workers, queueSize, &rwMutexLocker{})
}

// NewWorkerPoolMutex 创建使用 Mutex 的 worker pool，保留用于性能对比。
func NewWorkerPoolMutex(workers, queueSize int) *WorkerPool {
	return newWorkerPool(workers, queueSize, &mutexLocker{})
}

// newWorkerPool 注入 locker 并创建 worker pool，启动 workers 个 goroutine 消费任务队列。
func newWorkerPool(workers, queueSize int, locker submitLocker) *WorkerPool {
	assert.Assert(workers > 0, "workerpool: workers 必须大于 0")
	assert.Assert(queueSize >= 0, "workerpool: queueSize 不能为负数")
	p := &WorkerPool{
		queue:  make(chan func(), queueSize),
		stop:   make(chan struct{}),
		locker: locker,
	}
	for range workers {
		go p.run()
	}
	return p
}

// run 是每个 worker goroutine 的主循环，从队列取任务执行，收到 stop 信号后退出。
func (p *WorkerPool) run() {
	for {
		select {
		case task := <-p.queue:
			safeexec.SafeExec("workerpool.task", task) // task panic 隔离并记录日志，worker goroutine 继续服务
			p.wg.Done()                                // 依赖 safeexec 保证 task panic 不穿透；若穿透则 wg.Done 被跳过，Close 将永久阻塞
		case <-p.stop:
			return
		}
	}
}

// Submit 提交任务。队列满时阻塞，直到有空位或 ctx 取消。
// pool 已关闭时返回 ErrPoolClosed。
//
// locker 保证"检查 closed"与"wg.Add(1)"是原子操作，
// 从而杜绝 Close 在两者之间插入、导致 wg.Wait() 提前返回的竞态。
func (p *WorkerPool) Submit(ctx context.Context, task func()) error {
	p.locker.lockSubmit()
	if p.closed {
		p.locker.unlockSubmit()
		return ErrPoolClosed
	}
	p.wg.Add(1)
	p.locker.unlockSubmit()

	select {
	case p.queue <- task:
		return nil
	case <-ctx.Done():
		p.wg.Done()
		return ctx.Err()
	}
}

// Wait 等待所有已提交任务完成。可多次调用。
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// Close 关闭 pool，不再接受新任务，等待所有已提交任务完成后退出 worker。
// 幂等：多次调用安全，只有第一次有效。
func (p *WorkerPool) Close() {
	p.once.Do(func() {
		// 持锁设置 closed，与 Submit 的 check+Add 互斥，
		// 保证之后不会再有新的 wg.Add(1)。
		p.locker.lockClose()
		p.closed = true
		p.locker.unlockClose()
		// 等待所有已提交（wg.Add 已完成）的任务执行完毕。
		p.wg.Wait()
		// 所有任务已处理完毕，通知 worker 退出。
		close(p.stop)
	})
}
