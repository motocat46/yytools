# WorkerPool 设计文档

**日期：** 2026-03-24
**状态：** 待实现
**路径：** `pkg/infra/concurrency/workerpool`

---

## 为什么需要 WorkerPool

Go 的 goroutine 很轻量，但"轻量"不等于"无限制"。常见问题：

- **无限制并发**：对每个请求/任务直接 `go func()`，在高负载下 goroutine 数量爆炸，内存耗尽
- **重复造轮子**：每个项目都要自己实现 "带队列的 goroutine 池"，代码散落各处，质量参差不齐
- **缺乏背压**：没有队列上限，上游生产速度远超下游处理速度时，任务在内存中无限堆积

WorkerPool 提供一个**经过测试、行为可预期的并发调度原语**，让调用方专注业务逻辑，不需要每次重新解决这些基础问题。

Pipeline 在此之上再封装一层 channel I/O，适配"数据流处理"场景，进一步降低使用成本。

---

## API 设计

### WorkerPool（Submit 型）

```go
// NewWorkerPool 创建固定大小的 worker pool。
// workers：并发 worker 数量；queueSize：待执行队列容量（提供背压）。
func NewWorkerPool(workers, queueSize int) *WorkerPool

// Submit 提交任务。队列满时阻塞，直到有空位或 ctx 取消。
// pool 已关闭时返回 error。
func (p *WorkerPool) Submit(ctx context.Context, task func()) error

// Wait 等待所有已提交任务完成。可多次调用。
func (p *WorkerPool) Wait()

// Close 关闭 pool，不再接受新任务，等待已提交任务全部执行完毕后返回。
func (p *WorkerPool) Close()
```

### Pipeline（流水线型）

```go
// NewPipeline 创建泛型 pipeline。fn 为每个元素的处理函数。
func NewPipeline[T, R any](workers, queueSize int, fn func(T) (R, error)) *Pipeline[T, R]

// Process 消费输入 channel，并发执行 fn，结果写入输出 channel。
// 输入 channel 关闭后，所有任务完成，输出 channel 自动关闭。
func (p *Pipeline[T, R]) Process(in <-chan T) <-chan Result[R]

// Result 携带处理结果或错误。
type Result[R any] struct {
    Value R
    Err   error
}
```

---

## 关键设计决策

### 1. Submit 接收 ctx，不强迫阻塞
调用方可以通过 ctx 设置超时或取消，适应不同调用场景，不绑架调用方。

### 2. queueSize 有界
有界队列提供背压：生产速度超过消费速度时，Submit 阻塞，而不是无限堆积任务。不使用 `unbounded_channel`，防止内存无上限增长。

### 3. fn 签名为 `func(T) (R, error)`
现实中的处理函数几乎都可能出错。若只支持无 error 版本，调用方被迫自己包一层。输出 `Result[R]` 让调用方决定如何处理错误（忽略、记录、终止），不替调用方做决定。

### 4. Pipeline 是 WorkerPool 的薄封装
Pipeline 不重新实现 goroutine 管理，只负责 channel I/O 适配。职责分离，内部逻辑不重复。

### 5. Wait 和 Close 分离
- `Wait`：等待已提交任务完成，可多次调用，pool 仍可继续接受任务
- `Close`：终止接受 + 等待排空，只能调用一次

---

## 文件结构

```
pkg/infra/concurrency/workerpool/
├── pool.go          # WorkerPool 实现
├── pipeline.go      # Pipeline + Result 实现
├── pool_test.go     # WorkerPool 测试
├── pipeline_test.go # Pipeline 测试
├── README.md        # 使用文档
├── DESIGN.md        # 设计决策记录
└── TEST.md          # 测试分层与基准基线
```

`pkg/infra/concurrency/README.md` 导航索引需同步更新。

---

## 测试策略

| 层次 | 内容 |
|------|------|
| 单方法 | Submit/Wait/Close 正常路径、边界（workers=1、queueSize=0）、关闭后提交返回 error |
| 集成 | 提交 N 个任务 → Wait → 验证全部执行、无遗漏、无重复 |
| 并发安全 | 多 goroutine 并发 Submit，`-race` 验证无数据竞争 |
| Pipeline | 输入 channel 关闭后输出自动关闭；fn 返回 error 时 Result.Err 正确传递 |
| 压力 | 100 万任务，验证正确性与无内存泄漏（goleak） |
| Benchmark | workers=1/10/100/1000，混合 Submit+Wait 负载 |

**强制约束：**
- 必须通过 `go test -race`
- 用 `goleak` 验证 Close 后所有 worker goroutine 退出，无泄漏
- 大规模测试用 `testing.Short()` 跳过，适配 CI
