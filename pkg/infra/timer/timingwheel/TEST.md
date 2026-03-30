# TimingWheel 测试说明

## 快速验证

```bash
# 功能 + 竞态检测
go test -race -count=1 ./pkg/infra/timer/timingwheel/

# 正确性命题专项（-race 必须通过）
go test -race -run TestCorrectness -v ./pkg/infra/timer/timingwheel/

# 跳过大规模测试（CI 快速通道）
go test -race -short ./pkg/infra/timer/timingwheel/
```

## 基准测试

```bash
go test -bench=. -count=3 -benchtime=3s -benchmem ./pkg/infra/timer/timingwheel/
```

## 性能基线（Apple M4，2026-03-30，`-count=3 -benchtime=3s`）

```
# BenchmarkAfterFunc_Sequential：单调用方 Add（稳定规模，duration 循环于 [1h, n*h]）
BenchmarkAfterFunc_Sequential/n=100        ~38 ns/op     64 B/op    1 allocs/op
BenchmarkAfterFunc_Sequential/n=1000       ~36 ns/op     64 B/op    1 allocs/op
BenchmarkAfterFunc_Sequential/n=10000      ~49 ns/op    100 B/op    1 allocs/op
BenchmarkAfterFunc_Sequential/n=100000     ~49 ns/op    107 B/op    1 allocs/op
BenchmarkAfterFunc_Sequential/n=1000000    ~53 ns/op    109 B/op    1 allocs/op

# BenchmarkAfterFunc_Concurrent：多调用方并发 Add/Cancel（工作集 n=100,000）
# p=1 vs p=64 差距极小（~183 ns vs ~171 ns），说明锁竞争可控
BenchmarkAfterFunc_Concurrent/p=1          ~187 ns/op   106 B/op    1 allocs/op
BenchmarkAfterFunc_Concurrent/p=4          ~175 ns/op   112 B/op    1 allocs/op
BenchmarkAfterFunc_Concurrent/p=16         ~168 ns/op   111 B/op    1 allocs/op
BenchmarkAfterFunc_Concurrent/p=64         ~170 ns/op   108 B/op    1 allocs/op

# BenchmarkCancel_Sequential：O(1) Cancel（稳定规模，duration 循环于 [1h, n*h]）
BenchmarkCancel_Sequential/n=1000          ~41 ns/op     64 B/op    1 allocs/op
BenchmarkCancel_Sequential/n=10000         ~49 ns/op    104 B/op    1 allocs/op
BenchmarkCancel_Sequential/n=100000        ~53 ns/op    107 B/op    1 allocs/op
BenchmarkCancel_Sequential/n=1000000       ~54 ns/op    107 B/op    1 allocs/op

# BenchmarkMixed_AddCancel：混合负载（70% 纯 Add / 30% Add+Cancel）
# n 越大，ns/op 越低：大 n 时 timer 层级更高，路由开销更均匀
BenchmarkMixed_AddCancel/n=1000            ~160 ns/op   110 B/op    1 allocs/op
BenchmarkMixed_AddCancel/n=10000           ~121 ns/op    97 B/op    1 allocs/op
BenchmarkMixed_AddCancel/n=100000           ~52 ns/op    65 B/op    1 allocs/op
BenchmarkMixed_AddCancel/n=1000000          ~48 ns/op    64 B/op    1 allocs/op
```

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `constants_test.go` | 层级常量正确性、bucket 初始状态 |
| `bucket_test.go` | bucket Add/Flush/isFirst 正确性 |
| `routing_test.go` | addInternal 层级路由、advanceClock、Cancel |
| `timingwheel_test.go` | 功能：one-shot/repeating/cancel/Stop 排水/多层降级/10万并发 |
| `correctness_test.go` | 正确性命题：精确一次/等待语义/异常隔离/并发安全/repeating |
| `bench_test.go` | 基准：sequential/concurrent/cancel/混合负载 |
