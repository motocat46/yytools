# TEST.md — snowflake

## 测试文件

| 文件 | 内容 |
|------|------|
| `snowflake_test.go` | 单元测试：参数校验、单调性、唯一性、并发安全、ParseID 往返、Init/NewID panic 路径 |
| `snowflake_stress_test.go` | 压力测试：百万并发唯一性、序号耗尽自旋、时钟回拨模拟、位字段完整性、多节点隔离、跨毫秒边界单调性、长时间运行 |

## 分层测试命令

### 日常开发 / CI（推荐，~7 秒）

```bash
go test ./pkg/algorithms/idgen/snowflake/... -v -race -short
```

- `-race`：开启竞态检测
- `-short`：跳过耗时慢测试（千万级、500 万单线程、30 秒长跑）

### 上线前完整压测（~45 秒）

```bash
go test ./pkg/algorithms/idgen/snowflake/... -v -timeout 120s
```

包含：千万级并发唯一性（~3s）、500 万单线程（~1.25s）、30 秒持续运行（~33s）。

## 压测覆盖一览

| 测试函数 | 数据量 | -short 跳过 | 验证内容 |
|----------|--------|------------|---------|
| `TestStress_Concurrent_Uniqueness` | 100 万（-race）| 否 | 100 goroutine 并发，全局无重复 |
| `TestStress_TenMillion_Uniqueness` | **1000 万** | 是 | 千万级唯一性，自然序号耗尽约 2441 次 |
| `TestStress_SequenceExhaustion` | — | 否 | 强制触发序号耗尽自旋，验证跨毫秒边界 |
| `TestStress_BitFields_AllValid` | 1 万 | 否 | 符号位/NodeID/Timestamp/Sequence 范围 |
| `TestStress_ClockRollback_Simulation` | 100 | 否 | NTP 时钟回拨，ID 唯一且不回退 |
| `TestStress_MultiNode_Isolation` | 10 万 | 否 | 两节点 ID 空间无交集，字段编码正确 |
| `TestStress_IDMonotonicity_AcrossMsBoundary` | — | 否 | 跨毫秒边界严格单调递增 |
| `TestStress_HighFreqSingleThread` | **500 万** | 是 | 单线程，自然序号耗尽约 1220 次，严格单调 |
| `TestStress_ConcurrentHighFreq_WithRace` | 80 万（-race）| 否 | 高频并发，竞态检测 |
| `TestStress_LongRunning` | **~1 亿**（30s）| 是 | 持续运行，统计吞吐量，序号耗尽约 2.6 万次 |

## 基准测试

```bash
# 只跑 Benchmark
go test ./pkg/algorithms/idgen/snowflake/... -bench=. -run=^$ -benchmem -benchtime=3s
```

### 基线数字（Apple M4，2026-03-18）

| Benchmark | ns/op | 吞吐量 | 内存分配 |
|-----------|-------|--------|---------|
| `NewID`（单线程） | 244 ns | ~410 万/秒 | 0 allocs |
| `NewID_Parallel`（多线程，10 核） | 266 ns | ~375 万/秒 | 0 allocs |
| `ParseID` | 3 ns | — | 0 allocs |

零内存分配：热路径完全不触发 GC，适合游戏服高频调用。
