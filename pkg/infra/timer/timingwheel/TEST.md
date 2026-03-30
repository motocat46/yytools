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

## 性能基线（Apple M4，2026-03-30）

```
BenchmarkAfterFunc_Sequential/n=1000    ~39 ns/op    64 B/op    1 allocs/op
BenchmarkAfterFunc_Sequential/n=10000   ~44 ns/op    64 B/op    1 allocs/op
BenchmarkAfterFunc_Sequential/n=100000  ~42 ns/op    64 B/op    1 allocs/op
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
