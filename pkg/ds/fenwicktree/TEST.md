# fenwicktree 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `fenwicktree_test.go` | 单元测试：New/Add/PrefixSum/RangeSum 正常/边界/panic 路径 |
| `correctness_test.go` | 正确性命题：refFenwick 参考模型随机对比 |
| `bench_test.go` | 基准：Build/Add/PrefixSum/RangeSum/Mixed 各规模 |

## 运行方式

```bash
# 全量测试（含 100k 随机操作）
go test ./pkg/ds/fenwicktree/

# 跳过大规模测试（CI 快速通道）
go test -short ./pkg/ds/fenwicktree/

# 竞态检测
go test -race ./pkg/ds/fenwicktree/

# 只跑正确性测试
go test -run TestCorrectness ./pkg/ds/fenwicktree/

# 基准（3s 稳定结果）
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/fenwicktree/

# 只跑 Build benchmark
go test -run '^$' -bench 'BenchmarkFenwickTree_Build' -benchtime=1s -benchmem ./pkg/ds/fenwicktree/
```

## 测试用例

### 单元测试（`fenwicktree_test.go`）

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestNew_Len` | `Len() == n` |
| `TestNew_ZeroPanics` | `New(0)` panic |
| `TestNew_NegativePanics` | `New(-1)` panic |
| `TestAdd_PrefixSum_Single` | 单点 Add 后前后位置的前缀和 |
| `TestAdd_PrefixSum_Accumulate` | 逐位加入后所有前缀和正确 |
| `TestAdd_NegativeDelta` | 负数 delta 正确累减 |
| `TestAdd_BoundaryFirst` | 下标 0（最小边界） |
| `TestAdd_BoundaryLast` | 下标 n-1（最大边界） |
| `TestAdd_OutOfBoundsPanics` | `i=-1`、`i=n` 触发 panic |
| `TestPrefixSum_OutOfBoundsPanics` | `i=-1`、`i=n` 触发 panic |
| `TestRangeSum_Basic` | 多种 `[l,r]` 组合 |
| `TestRangeSum_LEqualsZero` | `l=0` 的特殊路径 |
| `TestRangeSum_InvalidPanics` | `l<0`、`r>=n`、`l>r` 触发 panic |
| `TestFloatType` | `float64` 类型正确工作 |
| `TestBuild_Basic` | 从数组建树后的 PrefixSum/RangeSum 正确 |
| `TestBuild_Single` | 单元素数组建树 |
| `TestBuild_EmptyPanics` | 空切片触发 panic |
| `TestBuild_DoesNotAliasInput` | 不保留输入切片引用 |
| `TestBuild_FloatType` | `float64` 路径正确 |

### 正确性命题测试（`correctness_test.go`）

| 测试函数 | 规模 | 覆盖场景 |
|----------|------|---------|
| `TestCorrectness_RandomOps` | `n=200`，100k 次 | Add/PrefixSum/RangeSum 与参考模型逐步对比 |
| `TestCorrectness_AllSameIndex` | `n=10`，1000 次 | 同一下标反复 Add 累加正确 |
| `TestCorrectness_LargeRandom` | `n=100/1k/10k`，各 100k 次 | 大规模随机验证（`-short` 跳过） |
| `TestCorrectness_Stress` | `n=1_000_000`，100k 次 | 百万规模压力测试（`-short` 跳过） |

## 性能基线（Apple M4，-benchtime=3s）

运行以下命令获取基线，结果记录于此：

```bash
go test -bench=. -benchtime=3s -benchmem -count=1 ./pkg/ds/fenwicktree/
```

实际结果：

```text
BenchmarkFenwickTree_Build/n=100-10           17113827    213.6 ns/op      928 B/op        2 allocs/op
BenchmarkFenwickTree_Build/n=1000-10           1550138    2327 ns/op      8224 B/op        2 allocs/op
BenchmarkFenwickTree_Build/n=10000-10           148032    23803 ns/op    81952 B/op        2 allocs/op
BenchmarkFenwickTree_Build/n=100000-10           13396   267898 ns/op   802861 B/op        2 allocs/op
BenchmarkFenwickTree_Build/n=1000000-10           1400  2633600 ns/op  8003627 B/op        2 allocs/op
BenchmarkFenwickTree_Add/n=100-10            295087257      12.23 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Add/n=1000-10           255153306      14.07 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Add/n=10000-10          223718505      16.04 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Add/n=100000-10         198156633      18.01 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Add/n=1000000-10        171309344      21.09 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_PrefixSum/n=100-10      305738563      11.58 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_PrefixSum/n=1000-10     262297557      13.63 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_PrefixSum/n=10000-10    237746649      15.26 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_PrefixSum/n=100000-10   209711200      17.28 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_PrefixSum/n=1000000-10  170076980      20.92 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_RangeSum/n=100-10       174401592      20.58 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_RangeSum/n=1000-10      152594862      23.52 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_RangeSum/n=10000-10     135698119      26.49 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_RangeSum/n=100000-10    121115737      29.71 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_RangeSum/n=1000000-10   100000000      33.30 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Mixed/n=100-10          149725310      24.09 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Mixed/n=1000-10         135339206      26.55 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Mixed/n=10000-10        125464915      28.73 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Mixed/n=100000-10       100000000      31.03 ns/op        0 B/op        0 allocs/op
BenchmarkFenwickTree_Mixed/n=1000000-10      107866162      33.27 ns/op        0 B/op        0 allocs/op
```

结论：`Build` 随 `n` 呈线性增长，分配来自新树对象和内部数组本身；`Add/PrefixSum/RangeSum/Mixed` 均为 `0 allocs/op`，`ns/op` 随 `n` 呈对数级缓慢上升，符合树状数组的预期复杂度。
