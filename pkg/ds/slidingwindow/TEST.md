# slidingwindow 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `slidingwindow_test.go` | 单元测试，覆盖所有公开方法的正常/边界/错误路径 |
| `correctness_test.go` | 正确性命题测试，含参考模型 `refWindow` 随机对比 + 多种数据分布验证 |
| `bench_test.go` | 基准测试，Add/Max/Min/Sum/混合负载各规模 |

## 运行方式

```bash
# 运行所有测试（含 100k 随机操作，约 5s）
go test ./pkg/ds/slidingwindow/

# 跳过大规模测试（CI 快速通道，约 1s）
go test -short ./pkg/ds/slidingwindow/

# 竞态检测
go test -race ./pkg/ds/slidingwindow/

# 只跑正确性命题测试
go test -run TestCorrectness ./pkg/ds/slidingwindow/

# 运行 Benchmark（-benchtime=3s 使结果更稳定）
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/slidingwindow/
```

## 正确性测试用例

### 单元测试（`slidingwindow_test.go`）

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestNew_InitialState` | 初始 Len=0，Empty=true，Full=false |
| `TestNew_ZeroSizePanics` | size=0 触发 panic |
| `TestAdd_Sum_UnderCapacity` | 未满窗口的 Sum 正确 |
| `TestAdd_Sum_ExactCapacity` | 恰好填满时 Sum 正确 |
| `TestAdd_Sum_OverCapacity_Evicts` | 超容淘汰最旧元素，Sum 正确 |
| `TestAvg_Basic` | 基本 Avg 计算 |
| `TestAvg_AfterEviction` | 淘汰后 Avg 正确 |
| `TestAvg_EmptyPanics` | 空窗口调用 Avg panic |
| `TestMax_Basic` | 多元素中取最大值 |
| `TestMax_AfterEviction` | 最大值元素被淘汰后 Max 更新 |
| `TestMax_EmptyPanics` | 空窗口调用 Max panic |
| `TestMin_Basic` | 多元素中取最小值 |
| `TestMin_AfterEviction` | 最小值元素被淘汰后 Min 更新 |
| `TestMin_EmptyPanics` | 空窗口调用 Min panic |

### 正确性命题测试（`correctness_test.go`）

| 测试函数 | 规模 | 覆盖场景 |
|----------|------|---------|
| `TestCorrectness_RandomOps` | 100k 次 Add，窗口大小随机 [1,200] | 随机值与参考模型 `refWindow` 逐步对比 Sum/Max/Min |
| `TestCorrectness_AllDuplicates` | 窗口大小 50，加入 200 个相同值 | 全相同值时 Max=Min=Sum/N |
| `TestCorrectness_StrictlyDecreasing` | 窗口大小 10 | 严格递减序列，Max 始终是最新值，Min 是最旧值（边界） |
| `TestCorrectness_StrictlyIncreasing` | 窗口大小 10 | 严格递增序列，Min 始终是最新值，Max 是最旧值（边界） |
| `TestCorrectness_SizeOne` | 窗口大小 1，100 次 Add | 单元素窗口每次 Add 后 Max=Min=Sum=最新值 |
| `TestCorrectness_NegativeValues` | 窗口大小 5，负数和零 | 负数、零混合，Max/Min/Sum 与参考模型一致 |
| `TestCorrectness_FloatWindow` | 窗口大小 10，float64 | 浮点类型正确工作 |
| `TestCorrectness_LargeRandom` | 100k 次 × 多种窗口大小 | 大规模随机验证（使用 `-short` 跳过） |

参考模型 `refWindow` 使用 `[]int` 全量扫描，语义与 Window 完全相同，实现最朴素（O(N) Max/Min），用于逐步对比正确性。

## Benchmark

各基准均覆盖 n=10/100/1000/10000/100000 五个窗口规模（n 为窗口大小），窗口预填满后测量稳态性能。

| 函数 | 说明 |
|------|------|
| `BenchmarkWindow_Add` | 稳态（窗口满）均摊 Add 代价 |
| `BenchmarkWindow_Max` | O(1) Max 查询代价 |
| `BenchmarkWindow_Min` | O(1) Min 查询代价 |
| `BenchmarkWindow_Sum` | O(1) Sum 查询代价 |
| `BenchmarkWindow_Mixed` | 混合负载：Add + Max + Min + Sum |

## 性能基线（2026-04-20，Apple M4，-benchtime=3s）

```
BenchmarkWindow_Add/n=10     ~7 ns/op   0 B/op   0 allocs/op
BenchmarkWindow_Add/n=100    ~7 ns/op   0 B/op   0 allocs/op
BenchmarkWindow_Add/n=1000   ~7 ns/op   0 B/op   0 allocs/op
BenchmarkWindow_Add/n=10000  ~7 ns/op   0 B/op   0 allocs/op
BenchmarkWindow_Add/n=100000 ~7 ns/op   0 B/op   0 allocs/op

BenchmarkWindow_Max/n=10     ~2.3 ns/op 0 B/op   0 allocs/op
BenchmarkWindow_Max/n=100000 ~2.3 ns/op 0 B/op   0 allocs/op

BenchmarkWindow_Min/n=10     ~2.3 ns/op 0 B/op   0 allocs/op
BenchmarkWindow_Min/n=100000 ~2.3 ns/op 0 B/op   0 allocs/op

BenchmarkWindow_Sum/n=10     ~2 ns/op   0 B/op   0 allocs/op
BenchmarkWindow_Sum/n=100000 ~2 ns/op   0 B/op   0 allocs/op
```

Add/Max/Min/Sum 各规模 ns/op 完全一致，确认 O(1)，0 allocs/op。
