# TEST.md — ring_buffer

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `ring_buffer_test.go` | 单方法、边界、集成、随机混合（10 万次）、压力（100 万次）|
| `bench_test.go` | Enqueue 纯入队基准、Enqueue+Dequeue 混合负载基准、Range 全量遍历基准 |

## 分层命令

```bash
# 快速验证（跳过压力测试）
go test -short -count=1 ./pkg/ds/ring_buffer/

# 全量（含百万压力测试）
go test -count=1 ./pkg/ds/ring_buffer/

# 基准测试
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/ring_buffer/
```

## 基准基线（Apple M4，Go 1.24）

```
BenchmarkRingBuffer_Enqueue/n=100        ~3.4 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Enqueue/n=1000       ~3.5 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Enqueue/n=10000      ~3.5 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Enqueue/n=100000     ~3.6 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Enqueue/n=1000000    ~3.8 ns/op     0 B/op   0 allocs/op

BenchmarkRingBuffer_Mixed/n=100        ~8.1 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Mixed/n=1000       ~8.1 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Mixed/n=10000      ~8.3 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Mixed/n=100000     ~8.2 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Mixed/n=1000000    ~8.5 ns/op     0 B/op   0 allocs/op

BenchmarkRingBuffer_Range/n=100        ~81 ns/op      0 B/op   0 allocs/op
BenchmarkRingBuffer_Range/n=1000       ~749 ns/op     0 B/op   0 allocs/op
BenchmarkRingBuffer_Range/n=10000      ~7424 ns/op    0 B/op   0 allocs/op
BenchmarkRingBuffer_Range/n=100000     ~74505 ns/op   0 B/op   0 allocs/op
BenchmarkRingBuffer_Range/n=1000000    ~746820 ns/op  0 B/op   0 allocs/op
```

Write 在各规模下约 3.4–3.8 ns/op，Mixed（Read+Write）约 8.1–8.5 ns/op，零内存分配，符合 O(1) 预期。Range 随规模线性增长（n=100 约 81 ns，n=1000000 约 747 µs），每元素约 0.75 ns，符合 O(n) 预期。
