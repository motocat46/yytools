# snowflake

无锁线程安全的雪花算法唯一 ID 生成器，专为游戏服务器设计。

## 位布局

```
63                                              0
+------------------------------------------+----------+-------------+
|    41-bit 毫秒时间戳                       | 10-bit   |  12-bit     |
|    since 2025-01-01 00:00:00.000 UTC      | nodeID   |  sequence   |
+------------------------------------------+----------+-------------+
```

| 字段 | 位数 | 范围 | 说明 |
|------|------|------|------|
| 时间戳 | 41 | 0 ~ 2199023255551 | 距纪元毫秒数，有效期约 69 年（到 2094 年） |
| nodeID | 10 | 0 ~ 1023 | 最多 1024 个节点 |
| sequence | 12 | 0 ~ 4095 | 每毫秒每节点最多 4096 个 ID |

## 快速使用

```go
// 启动时初始化一次（nodeID 为本节点唯一编号，0–1023）
snowflake.Init(nodeID)

// 任意位置生成唯一 ID
id := snowflake.NewID()

// 调试：解码 ID 各字段
parts := snowflake.ParseID(id)
fmt.Println(parts.Timestamp, parts.NodeID, parts.Sequence, parts.Time)
```

多节点/测试场景直接构造实例：

```go
g, err := snowflake.NewGenerator(nodeID)
id := g.NewID()
```

## 运行测试

### 日常开发 / CI（推荐，~7 秒）

```bash
go test ./pkg/algorithms/snowflake/... -v -race -short
```

- `-race`：开启竞态检测，验证并发安全
- `-short`：跳过耗时慢测试（千万级、500万单线程、30秒长跑）

### 上线前完整压测（~45 秒）

```bash
go test ./pkg/algorithms/snowflake/... -v -timeout 120s
```

不加 `-short`，运行全部测试，包括：
- 千万级并发唯一性（~3s）
- 500 万单线程 + 自然序号耗尽 1220 次（~1.25s）
- 30 秒持续运行，生成约 1 亿个 ID（~33s）

### 压测覆盖点一览

| 测试函数 | 数据量 | 验证内容 |
|----------|--------|---------|
| `TestStress_Concurrent_Uniqueness` | 100 万（-race）| 100 goroutine 并发，全局无重复 |
| `TestStress_TenMillion_Uniqueness` | **1000 万** | 千万级唯一性，自然序号耗尽约 2441 次 |
| `TestStress_SequenceExhaustion` | — | 强制触发序号耗尽自旋，验证跨毫秒边界 |
| `TestStress_BitFields_AllValid` | 1 万 | 每个 ID 的符号位/NodeID/Timestamp/Sequence 范围 |
| `TestStress_ClockRollback_Simulation` | 100 | 模拟 NTP 时钟回拨，ID 唯一且不回退 |
| `TestStress_MultiNode_Isolation` | 10 万 | 两节点 ID 空间无交集，字段编码正确 |
| `TestStress_IDMonotonicity_AcrossMsBoundary` | — | 跨毫秒边界严格单调递增 |
| `TestStress_HighFreqSingleThread` | **500 万** | 单线程，自然序号耗尽约 1220 次，严格单调 |
| `TestStress_ConcurrentHighFreq_WithRace` | 80 万（-race）| 高频并发，竞态检测 |
| `TestStress_LongRunning` | **~1 亿**（30s）| 持续运行，统计吞吐量，序号耗尽约 2.6 万次 |

## 运行基准测试

### 关键：`-bench` 默认会同时运行所有 Test*

Go 的设计哲学是"先验证正确性，再测性能"，因此：

```
go test -bench=.       → 先跑全部 Test*，再跑 Benchmark*
go test（不加 -bench）  → 只跑 Test*，不跑 Benchmark*
```

**只跑 Benchmark，不跑 Test，需要加 `-run=^$`：**

```bash
# 标准写法：-run=^$ 匹配空字符串，跳过所有 Test*
go test ./pkg/algorithms/snowflake/... -bench=. -run=^$ -benchmem
```

```bash
# 加 -benchtime 延长采样时间，结果更稳定
go test ./pkg/algorithms/snowflake/... -bench=. -run=^$ -benchmem -benchtime=3s

# 加 -count 多轮运行，消除偶发抖动
go test ./pkg/algorithms/snowflake/... -bench=. -run=^$ -benchmem -benchtime=3s -count=3

# 只跑指定 Benchmark
go test ./pkg/algorithms/snowflake/... -bench=BenchmarkGenerator_NewID -run=^$
```

### 输出解读

```
BenchmarkGenerator_NewID-10    14760691    244.1 ns/op    0 B/op    0 allocs/op
│                    │              │            │           │              │
│                    │              │            │           └─ 每次调用堆内存分配次数
│                    │              │            └─ 每次调用分配字节数（0 = 零分配）
│                    │              └─ 每次操作耗时（纳秒）
│                    └─ 共执行了多少次迭代
└─ 函数名，-10 表示使用了 10 个 CPU（GOMAXPROCS）
```

`-benchtime=Ns` 的含义：让 benchmark 循环体**累计运行满 N 秒**，框架自动调整迭代次数。时间越长，均值越稳定。默认 1s。

### 基准测试结果（Apple M4）

| Benchmark | ns/op | 吞吐量 | 内存分配 |
|-----------|-------|--------|---------|
| `NewID`（单线程） | 244 ns | ~410 万/秒 | 0 allocs |
| `NewID_Parallel`（多线程） | ~274 ns | ~365 万/秒 | 0 allocs |
| `ParseID` | 3 ns | — | 0 allocs |

零内存分配意味着热路径完全不触发 GC，适合游戏服高频调用。

## 命令速查

```bash
# 日常开发
go test ./pkg/algorithms/snowflake/... -v -race -short

# 上线前
go test ./pkg/algorithms/snowflake/... -v -timeout 120s

# 只跑 Benchmark
go test ./pkg/algorithms/snowflake/... -bench=. -run=^$ -benchmem
```
