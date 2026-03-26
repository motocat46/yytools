# TEST.md — workerpool 测试说明

## 测试文件列表

| 文件 | 覆盖范围 |
|------|---------|
| `pool_test.go` | WorkerPool 单方法测试、集成测试（顺序/并发）、压力测试（顺序/并发）、基准测试 |
| `correctness_test.go` | WorkerPool 并发正确性命题测试；Pipeline 结果完整性命题测试 |
| `pipeline_test.go` | Pipeline 单方法测试、集成测试 |

所有 WorkerPool 测试均通过 `allFactories`（RWMutex 和 Mutex 两种实现）参数化运行，确保两种锁实现行为一致。

---

### pool_test.go 覆盖点

| 测试 | 层次 | 说明 |
|------|------|------|
| `TestWorkerPool_Submit_Normal` | 单方法 | 5 个任务全部执行，计数正确 |
| `TestWorkerPool_Submit_PoolClosed` | 单方法 | Close 后 Submit 返回 ErrPoolClosed |
| `TestWorkerPool_Submit_CtxCancelled` | 单方法 | 队列满时 ctx 超时返回错误 |
| `TestWorkerPool_Wait_MultipleCallsAllowed` | 单方法 | Wait 幂等，多次调用不 panic |
| `TestWorkerPool_Workers1_QueueSize0` | 边界 | 1 worker + 0 队列容量正常工作 |
| `TestWorkerPool_Integration_Sequential` | 集成 | 10 万任务，单调用方顺序提交，无遗漏无重复 |
| `TestWorkerPool_Integration_Concurrent` | 集成 | 10 万任务，20 调用方并发提交，无遗漏无重复 |
| `TestWorkerPool_Stress_Sequential` | 压力 | 100 万任务，单调用方，正确性验证（`-short` 跳过）|
| `TestWorkerPool_Stress_Concurrent` | 压力 | 100 万任务，50 调用方并发，正确性验证（`-short` 跳过）|

### correctness_test.go 覆盖点

| 测试 | 命题 | 说明 |
|------|------|------|
| `TestCorrectness_ExactlyOnce_ConcurrentSubmitClose` | 精确一次执行 | 并发 Submit + Close 竞争，已提交任务恰好执行一次 |
| `TestCorrectness_CloseWaitsForCompletion` | 等待语义 | Close 返回后所有任务副作用已完成 |
| `TestCorrectness_TaskPanic_NotDeadlock` | 异常隔离 | 单个任务 panic 不阻塞整体，后续任务正常执行 |
| `TestCorrectness_SubmitClose_NoPanic` | 并发安全 | 大量并发 Submit/Close 不触发 panic 或数据竞争 |
| `TestCorrectness_Close_Idempotent_Concurrent` | Close 幂等 | 多 goroutine 并发调用 Close 不 panic |
| `TestCorrectness_Pipeline_ResultCompleteness` | 结果完整性 | Pipeline 每个输入恰好产生一个输出 |
| `TestCorrectness_Pipeline_MultiProducer` | 多生产者并发安全 | 10 生产者 × 5000 = 5 万，输出总数严格等于输入总数 |
| `TestCorrectness_Pipeline_NoResultLost_NoDuplicate` | 无丢失无重复 | 大规模并发下 Pipeline 结果数精确等于输入数 |

### pipeline_test.go 覆盖点

| 测试 | 层次 | 说明 |
|------|------|------|
| `TestPipeline_正常处理` | 单方法 | 5 个元素全部转换，结果数正确 |
| `TestPipeline_错误传递` | 单方法 | fn 返回错误时，Result.Err 正确传递 |
| `TestPipeline_输入关闭后输出自动关闭` | 单方法 | 空输入 channel 关闭后输出 channel 立即关闭 |
| `TestPipeline_集成_大规模` | 集成 | 10 万元素全部处理，结果数正确 |

### Goroutine 泄漏检测

`TestMain` 使用 `goleak.VerifyTestMain`，所有测试结束后自动验证无残留 goroutine。
`TestMain` 同时将 `slog.Default()` 重定向到 `io.Discard`，避免 panic 恢复日志污染测试输出。

---

## 分层测试命令

### 快速验证（日常开发）

```bash
go test ./pkg/infra/concurrency/workerpool/
```

### 含竞态检测（并发代码必须通过）

```bash
go test -race ./pkg/infra/concurrency/workerpool/
```

### 正确性命题测试

```bash
go test -race -run TestCorrectness -v ./pkg/infra/concurrency/workerpool/
```

### 压力测试（百万任务，默认跳过）

```bash
go test -run Stress -v -timeout 120s ./pkg/infra/concurrency/workerpool/
```

### 基准测试

```bash
go test -bench=. -benchtime=3s -benchmem ./pkg/infra/concurrency/workerpool/
```

多次采样（用于 benchstat 对比）：

```bash
go test -bench=. -count=5 -benchtime=3s -benchmem ./pkg/infra/concurrency/workerpool/
```

---

## 基准基线数字

> 运行环境：Apple M4 / Go 1.24 / darwin arm64 / benchtime=3s count=3

### Sequential（单调用方，无竞争基线）

```
BenchmarkWorkerPool_Submit_Sequential-10       ~101 ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPoolMutex_Submit_Sequential-10  ~100 ns/op   0 B/op   0 allocs/op
```

单调用方下 RWMutex 与 Mutex 吞吐基本相同（~100 ns/op），均无内存分配。

### Concurrent（多调用方，暴露锁竞争）

```
BenchmarkWorkerPool_Submit_Concurrent/p=1-10    ~97  ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPool_Submit_Concurrent/p=4-10    ~95  ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPool_Submit_Concurrent/p=16-10   ~98  ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPool_Submit_Concurrent/p=64-10   ~95  ns/op   0 B/op   0 allocs/op

BenchmarkWorkerPoolMutex_Submit_Concurrent/p=1-10   ~134 ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPoolMutex_Submit_Concurrent/p=4-10   ~131 ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPoolMutex_Submit_Concurrent/p=16-10  ~128 ns/op   0 B/op   0 allocs/op
BenchmarkWorkerPoolMutex_Submit_Concurrent/p=64-10  ~137 ns/op   0 B/op   0 allocs/op
```

**结论**：并发场景下 RWMutex 约 ~96 ns/op，Mutex 约 ~133 ns/op，RWMutex 吞吐高约 **28%**。
RWMutex 随并发度增加吞吐基本持平（竞争可控）；Mutex 各并发度下吞吐相近（独占锁，竞争明显）。