# TEST.md — workerpool 测试说明

## 测试文件列表

| 文件 | 覆盖范围 |
|------|---------|
| `pool_test.go` | WorkerPool 单方法测试、集成测试、压力测试、基准测试 |
| `pipeline_test.go` | Pipeline 单方法测试、集成测试 |

### pool_test.go 覆盖点

| 测试 | 层次 | 说明 |
|------|------|------|
| `TestWorkerPool_Submit_正常执行` | 单方法 | 5 个任务全部执行，计数正确 |
| `TestWorkerPool_Submit_PoolClosed` | 单方法 | Close 后 Submit 返回 ErrPoolClosed |
| `TestWorkerPool_Submit_CtxCancelled` | 单方法 | 队列满时 ctx 超时返回错误 |
| `TestWorkerPool_Wait_多次调用` | 单方法 | Wait 幂等，多次调用不 panic |
| `TestWorkerPool_Workers1_QueueSize0` | 边界 | 1 worker + 0 队列容量正常工作 |
| `TestWorkerPool_集成_无遗漏无重复` | 集成 | 10 万任务全部执行，无遗漏无重复 |
| `TestWorkerPool_压力_百万任务` | 压力 | 100 万任务正确性验证（默认跳过，`-short` 模式下跳过）|
| `BenchmarkWorkerPool_Submit` | 基准 | workers=1/10/100/1000 的 Submit 吞吐量 |

### pipeline_test.go 覆盖点

| 测试 | 层次 | 说明 |
|------|------|------|
| `TestPipeline_正常处理` | 单方法 | 5 个元素全部转换，结果数正确 |
| `TestPipeline_错误传递` | 单方法 | fn 返回错误时，Result.Err 正确传递 |
| `TestPipeline_输入关闭后输出自动关闭` | 单方法 | 空输入 channel 关闭后输出 channel 立即关闭 |
| `TestPipeline_集成_大规模` | 集成 | 10 万元素全部处理，结果数正确 |

### Goroutine 泄漏检测

`TestMain` 使用 `goleak.VerifyTestMain`，所有测试结束后自动验证无残留 goroutine。

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

### 压力测试（百万任务，默认跳过）

```bash
go test ./pkg/infra/concurrency/workerpool/ -run 压力 -v -timeout 120s
```

> 不加 `-short` 标志时自动运行；加 `-short` 时跳过。

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

> 运行环境：Apple M4 / Go 1.24.4 / darwin arm64

```
BenchmarkWorkerPool_Submit/workers=1-10      51515196    70.03 ns/op    0 B/op    0 allocs/op
BenchmarkWorkerPool_Submit/workers=10-10     45740446    93.43 ns/op    0 B/op    0 allocs/op
BenchmarkWorkerPool_Submit/workers=100-10    46546642    92.93 ns/op    0 B/op    0 allocs/op
BenchmarkWorkerPool_Submit/workers=1000-10   45642415    96.02 ns/op    0 B/op    0 allocs/op
```

Submit 无内存分配（0 allocs/op）。单 worker 约 70 ns，多 worker 约 93-96 ns（mutex 竞争开销约 +25 ns）。

运行以下命令获取基线，填入上方：

```bash
go test -bench=BenchmarkWorkerPool_Submit -benchtime=3s -benchmem -count=3 \
    ./pkg/infra/concurrency/workerpool/ 2>&1 | tee bench_baseline.txt
```
