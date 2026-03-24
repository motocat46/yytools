# workerpool — 固定大小 goroutine 池 + 泛型 Pipeline

## 功能简介

`workerpool` 解决两类常见并发问题：

- **WorkerPool**：限制并发数，防止无限创建 goroutine 压垮系统。任务通过有界队列进入，队列满时 `Submit` 自动阻塞（背压），调用方可通过 `ctx` 设置超时或取消。
- **Pipeline**：对 `WorkerPool` 的薄封装，将"从 channel 读取 → 并发处理 → 结果写入 channel"这一固定模式结构化，免去手工管理 WaitGroup 和结果收集的样板代码。

## WorkerPool 快速上手

```go
import "github.com/motocat46/yytools/pkg/infra/concurrency/workerpool"

// 创建 4 个 worker，队列容量 100
pool := workerpool.NewWorkerPool(4, 100)
defer pool.Close() // 等待所有任务完成后退出 worker

for _, item := range items {
    item := item
    if err := pool.Submit(ctx, func() {
        process(item)
    }); err != nil {
        // ctx 被取消，或 pool 已关闭
        log.Printf("submit failed: %v", err)
        break
    }
}

pool.Wait() // 等待所有已提交任务完成（Close 内部也会调用，可按需显式等待）
```

## Pipeline 快速上手

```go
import "github.com/motocat46/yytools/pkg/infra/concurrency/workerpool"

// 4 个 worker，队列 100，fn 将 int 转为 string
p := workerpool.NewPipeline(4, 100, func(n int) (string, error) {
    return strconv.Itoa(n), nil
})
defer p.Close()

// 准备输入 channel（必须在某处 close(in)，Process 才会结束）
in := make(chan int, 100)
go func() {
    for _, n := range numbers {
        in <- n
    }
    close(in)
}()

// 消费输出 channel（必须全部消费完，否则 worker 阻塞）
for r := range p.Process(in) {
    if r.Err != nil {
        log.Printf("处理出错: %v", r.Err)
        continue
    }
    fmt.Println(r.Value)
}
```

## API 说明

### WorkerPool

#### `NewWorkerPool(workers, queueSize int) *WorkerPool`

创建固定大小的 goroutine 池，同时启动 `workers` 个 worker goroutine。

| 参数 | 说明 |
|------|------|
| `workers` | 并发 goroutine 数，必须 > 0 |
| `queueSize` | 待执行队列容量，0 表示无缓冲（每次 Submit 都直接阻塞到有 worker 空闲） |

#### `Submit(ctx context.Context, task func()) error`

提交任务到队列。

- 队列满时阻塞，直到有空位或 `ctx` 取消。
- pool 已关闭时立即返回 `ErrPoolClosed`。
- `ctx` 取消时返回 `ctx.Err()`（`context.Canceled` 或 `context.DeadlineExceeded`）。

#### `Wait()`

等待所有已提交任务完成。可多次调用，幂等。

#### `Close()`

关闭 pool：不再接受新任务，等待已提交任务全部完成，然后退出所有 worker goroutine。幂等，多次调用安全。

---

### Pipeline

#### `NewPipeline[T, R any](workers, queueSize int, fn func(T) (R, error)) *Pipeline[T, R]`

创建泛型 Pipeline。`workers`、`queueSize` 透传给内部 WorkerPool；`fn` 为每个元素的处理函数。

#### `Process(in <-chan T) <-chan Result[R]`

消费输入 channel，并发执行 `fn`，结果写入返回的输出 channel。输入 channel 关闭且所有任务完成后，输出 channel 自动关闭。

返回的 `Result[R]`：

```go
type Result[R any] struct {
    Value R
    Err   error
}
```

#### `Close()`

关闭内部 WorkerPool，释放所有 worker goroutine。

---

## 常见误用

### 误用一：忘记 `defer pool.Close()`

```go
// ❌ worker goroutine 泄漏
pool := workerpool.NewWorkerPool(4, 100)
pool.Submit(ctx, task)
pool.Wait()
// worker 依然在后台阻塞等待，永不退出

// ✅ 正确
pool := workerpool.NewWorkerPool(4, 100)
defer pool.Close()
```

### 误用二：Pipeline 未消费完输出 channel 就调用 Close

```go
// ❌ 提前 Close，worker 写 out channel 时阻塞，导致死锁
out := p.Process(in)
p.Close() // 此时 out 可能还没消费完

// ✅ 先消费完输出，再 Close（或用 defer）
out := p.Process(in)
for r := range out { // 消费完毕，out 自动关闭
    _ = r
}
p.Close()

// ✅ 或直接 defer，确保顺序正确
defer p.Close()
for r := range p.Process(in) { ... }
```

### 误用三：输入 channel 不关闭

```go
// ❌ Process 内部用 range in 读取，若 in 不关闭，goroutine 永久阻塞
out := p.Process(in) // in 永远不关闭 → out 永远不关闭 → range out 死锁

// ✅ 生产者负责在写完后 close(in)
```

### 误用四：忽略 Submit 的返回值

```go
// ❌ ctx 取消后继续提交任务，静默丢失错误
pool.Submit(ctx, task)

// ✅ 检查错误，ctx 取消时停止提交
if err := pool.Submit(ctx, task); err != nil {
    break
}
```
