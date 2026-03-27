# TEST.md — lru

## 文件列表

| 文件 | 内容 |
|------|------|
| `lru_test.go` | 单方法单元测试（`package lru`） + 基准测试 |
| `correctness_test.go` | 正确性命题测试（`package lru_test`，6 条命题） |
| `example_test.go` | 可运行示例（`package lru_test`，go doc 展示） |

## 分层运行命令

```bash
# 快速验证（跳过大规模/压力测试）
go test -short ./pkg/ds/lru/

# 全量测试
go test ./pkg/ds/lru/

# 竞态检测（并发代码必须通过）
go test -race ./pkg/ds/lru/

# 仅正确性命题（含竞态检测）
go test -race -run TestCorrectness -v ./pkg/ds/lru/

# 基准测试（完整规模）
go test -bench=. -benchmem -benchtime=3s ./pkg/ds/lru/

# 并发基准（观察锁竞争）
go test -bench=BenchmarkConcurrent_Mixed -benchmem ./pkg/ds/lru/
```

## 正确性命题

| 命题 | 测试函数 |
|------|---------|
| LRU 淘汰顺序：始终淘汰最久未 Get/Put 的 key | `TestCorrectness_LRU_EvictionOrder` |
| TTL 惰性过期：过期后 Get/Peek/Contains 均未命中，容量槽被释放 | `TestCorrectness_TTL_LazyExpiry` |
| 容量不变量：任意操作序列后 Len() ≤ capacity | `TestCorrectness_CapacityInvariant` |
| Peek 不影响顺序：Peek 后淘汰顺序与之前一致 | `TestCorrectness_Peek_NoOrderEffect` |
| 随机混合 10 万次：与参考模型（refLRU）逐步对比 | `TestCorrectness_RandomMixed_RefModel` |
| 并发安全：20 goroutine × 5000 ops，Len ≤ capacity，无竞态 | `TestCorrectness_Concurrent` |

## 性能基准参考基线

> 环境：Apple M4，Go 1.24，`-benchtime=3s -count=1`
> 运行命令：`go test -bench=. -benchmem -benchtime=3s -count=3 ./pkg/ds/lru/`

```
BenchmarkPut/n=100          18 ns/op    0 B/op   0 allocs/op
BenchmarkPut/n=1000         21 ns/op    0 B/op   0 allocs/op
BenchmarkPut/n=10000        25 ns/op    0 B/op   0 allocs/op
BenchmarkPut/n=100000       30 ns/op    0 B/op   0 allocs/op
BenchmarkPut/n=1000000     114 ns/op    1 B/op   0 allocs/op

BenchmarkGet/n=100          15 ns/op    0 B/op   0 allocs/op
BenchmarkGet/n=1000         14 ns/op    0 B/op   0 allocs/op
BenchmarkGet/n=10000        22 ns/op    0 B/op   0 allocs/op
BenchmarkGet/n=100000       27 ns/op    0 B/op   0 allocs/op
BenchmarkGet/n=1000000      88 ns/op    0 B/op   0 allocs/op

BenchmarkMixed/n=100        16 ns/op    0 B/op   0 allocs/op
BenchmarkMixed/n=1000000    91 ns/op    0 B/op   0 allocs/op

BenchmarkConcurrent_PeekHeavy/p=1     27 ns/op
BenchmarkConcurrent_PeekHeavy/p=4     27 ns/op
BenchmarkConcurrent_PeekHeavy/p=16    37 ns/op
BenchmarkConcurrent_PeekHeavy/p=64   115 ns/op

BenchmarkConcurrent_Mixed/p=1    92 ns/op
BenchmarkConcurrent_Mixed/p=4    88 ns/op
BenchmarkConcurrent_Mixed/p=16   86 ns/op
BenchmarkConcurrent_Mixed/p=64   87 ns/op
```

**解读：**
- Put/Get/Mixed 均零分配（n ≤ 100k），n=1M 时 cache miss 导致延迟升高 3–4×。
- `Concurrent_Mixed`（70% Get 写锁）在 p=1→p=64 下几乎不变（~92ns），所有操作均持写锁，高并发下串行化，吞吐稳定。
- `Concurrent_PeekHeavy`（90% Peek 读锁）在 p=16→p=64 出现明显劣化（27→115 ns/op），原因：10% Put 抢写锁时会阻塞所有并发 Peek，形成周期性全局暂停；并发越高，暂停影响的 goroutine 越多。
