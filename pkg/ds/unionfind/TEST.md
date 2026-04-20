# unionfind 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `unionfind_test.go` | 单元测试，覆盖所有公开方法的正常/边界/错误路径 |
| `correctness_test.go` | 正确性命题测试，含参考模型 `refUF` 随机对比 + 链式/星形结构验证 |
| `bench_test.go` | 基准测试，Union/Find/Connected/混合负载各规模 |

## 运行方式

```bash
# 运行所有测试（含 100k 随机操作 + 各规模正确性验证，约 10s）
go test ./pkg/ds/unionfind/

# 跳过大规模测试（CI 快速通道，约 1s）
go test -short ./pkg/ds/unionfind/

# 竞态检测
go test -race ./pkg/ds/unionfind/

# 只跑正确性命题测试
go test -run TestCorrectness ./pkg/ds/unionfind/

# 运行 Benchmark（-benchtime=3s 使结果更稳定）
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/unionfind/

# 只跑特定操作基准
go test -bench=BenchmarkUnionFind_Find ./pkg/ds/unionfind/
```

## 正确性测试用例

### 单元测试（`unionfind_test.go`）

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestNew_Empty` | 空并查集 Count=0 |
| `TestAutoRegister_NewElement` | 首次 Find 自动注册，Count+1，Size=1 |
| `TestFind_ReturnsSelf_WhenAlone` | 单独元素 Find 返回自身 |
| `TestFind_SameRoot_AfterUnion` | Union 后同组元素 Find 返回相同根 |
| `TestFind_PathCompression` | 路径压缩后 Find 结果稳定 |
| `TestUnion_ReturnsTrueOnFirstMerge` | 首次合并返回 true |
| `TestUnion_ReturnsFalseWhenAlreadyConnected` | 已连通再次 Union 返回 false |
| `TestUnion_ReducesCount` | 每次合并 Count 减 1 |
| `TestUnion_SelfUnion` | 自身 Union 返回 false，Count 不变 |
| `TestConnected_FalseForNewElements` | 未合并的两个元素不连通 |
| `TestConnected_TrueAfterUnion` | Union 后 Connected 为 true |
| `TestConnected_Transitive` | 传递连通：a-b, b-c → a-c |
| `TestSize_SingleElement` | 单独元素 Size=1 |
| `TestSize_AfterUnion` | Union 后组内所有元素 Size 一致 |

### 正确性命题测试（`correctness_test.go`）

| 测试函数 | 规模 | 覆盖场景 |
|----------|------|---------|
| `TestCorrectness_RandomOps` | 100k 次操作，universe=100 | 随机 Union/Connected/Count/Size 与参考模型 `refUF` 逐步对比 |
| `TestCorrectness_ChainUnion` | N=1000 | 链式合并后全连通、Count=1、Size=N、传递连通抽样验证 |
| `TestCorrectness_StarUnion` | N=500 | 星形合并到中心节点后全连通、Count=1、Size=N |
| `TestCorrectness_LargeRandom` | 100k 次 × 4 种 universe 规模 | universe=10/50/200/1000，每 1000 步对账 Count 和 Size |

参考模型 `refUF` 使用 `map[int]int` 追踪每个元素的组 ID，Union 通过全表重标签（O(N)）实现——语义与 UnionFind 完全相同，实现最朴素，用于正确性对比。

`TestCorrectness_LargeRandom` 使用 `-short` 跳过（CI 快速通道）。

## Benchmark

各基准均覆盖 n=100/1000/10000/100000/1000000 五个规模，便于观察复杂度曲线。

| 函数 | 说明 |
|------|------|
| `BenchmarkUnionFind_Union` | 预注册所有节点后随机合并，规模稳定 |
| `BenchmarkUnionFind_Find` | 链式结构上随机 Find，路径压缩有工作可做 |
| `BenchmarkUnionFind_Connected` | 全连通后随机 Connected |
| `BenchmarkUnionFind_Mixed` | 混合负载：50% Union + 30% Find + 20% Connected |

## 性能基线（2026-04-20，Apple M4，-benchtime=3s）

```
BenchmarkUnionFind_Union/n=100       ~40 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Union/n=1000      ~40 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Union/n=10000     ~45 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Union/n=100000    ~51 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Union/n=1000000   ~136 ns/op  0 B/op   0 allocs/op

BenchmarkUnionFind_Find/n=100        ~22 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Find/n=1000       ~22 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Find/n=10000      ~24 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Find/n=100000     ~27 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Find/n=1000000    ~67 ns/op   0 B/op   0 allocs/op

BenchmarkUnionFind_Connected/n=100   ~42 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Connected/n=1000  ~45 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Connected/n=10000 ~48 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Connected/n=100000 ~56 ns/op  0 B/op   0 allocs/op
BenchmarkUnionFind_Connected/n=1000000 ~133 ns/op 0 B/op  0 allocs/op

BenchmarkUnionFind_Mixed/n=100       ~45 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Mixed/n=1000      ~44 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Mixed/n=10000     ~47 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Mixed/n=100000    ~53 ns/op   0 B/op   0 allocs/op
BenchmarkUnionFind_Mixed/n=1000000   ~148 ns/op  0 B/op   0 allocs/op
```

n=100 到 n=100k 变化极小，确认 O(α) 均摊。n=1M 的增幅来自 map 在大规模下的 cache miss，不影响使用。
