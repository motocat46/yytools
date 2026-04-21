# segtree 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `segtree_test.go` | 单元测试：三种 monoid 覆盖，边界，panic 路径，NoAllocs |
| `correctness_test.go` | 正确性命题：refSegTree 参考模型随机对比 |
| `bench_test.go` | 基准：Set/Apply/Query/Mixed 各规模 |

## 运行方式

```bash
# 全量测试（含 100k 随机操作）
go test -count=1 ./pkg/ds/segtree/

# 跳过大规模测试（CI 快速通道）
go test -short -count=1 ./pkg/ds/segtree/

# 竞态检测
go test -race -count=1 ./pkg/ds/segtree/

# 只跑正确性测试
go test -count=1 -run TestCorrectness ./pkg/ds/segtree/

# 基准（3s 稳定结果）
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/segtree/
```

注：验证修复时务必加 `-count=1` 绕过测试缓存，否则看到的可能是旧结果。

## 测试用例

### 单元测试（`segtree_test.go`）

| 测试函数 | Monoid | 覆盖场景 |
|----------|--------|---------|
| `TestNew_Len` | 任意 | `Len() == n` |
| `TestNew_ZeroPanics` | 任意 | `New(0)` panic |
| `TestNew_NegativePanics` | 任意 | `New(-1)` panic |
| `TestNew_InitAllIdentity` | 求和 | 初始 QueryAll 为 identity |
| `TestSet_Basic` | 求和 | 单点赋值后 Query/QueryAll 正确 |
| `TestSet_OverwritesSameIndex` | 求和 | 同一位置二次 Set 覆盖而非累加 |
| `TestSet_SumsDifferentIndexes` | 求和 | 不同位置 Set 后全局和正确 |
| `TestSet_BoundaryFirst` | 求和 | 下标 0（最小边界） |
| `TestSet_BoundaryLast` | 求和 | 下标 n-1（最大边界） |
| `TestSet_OutOfBoundsPanics` | 求和 | i=-1、i=n 触发 panic |
| `TestQuery_Basic` | 求和 | 多种 [l,r] 组合 |
| `TestQuery_LEqualsZero` | 求和 | l=0 路径 |
| `TestQuery_LEqualsR` | 求和 | l==r 单点查询 |
| `TestQuery_InvalidPanics` | 求和 | 越界、l>r 触发 panic |
| `TestQueryAll_SingleElement` | 求和 | n=1 全区间查询 |
| `TestApply_RangeAddSum` | 求和 | 区间加后各区间和正确 |
| `TestApply_FullRange` | 求和 | Apply 全区间 |
| `TestApply_LEqualsR` | 求和 | 单点 Apply |
| `TestApply_RangeAssignMin` | 区间赋值+最小值 | 区间赋值后 min 正确 |
| `TestApply_RangeAddMax` | 区间加+最大值 | 区间加后 max 正确 |
| `TestSet_AfterApply` | 求和 | Set 完全覆盖 lazy |
| `TestApply_MultipleOverlapping` | 求和 | 多次重叠 Apply |
| `TestApply_InvalidPanics` | 求和 | 越界、l>r 触发 panic |
| `TestSet_NoAllocs` | 求和 | Set 0 allocs/op |
| `TestApply_NoAllocs` | 求和 | Apply 0 allocs/op |
| `TestQuery_NoAllocs` | 求和 | Query 0 allocs/op |

### 正确性命题测试（`correctness_test.go`）

| 测试函数 | 规模 | 覆盖场景 |
|----------|------|---------|
| `TestCorrectness_RangeAddSum` | n=200，100k 次 | Set/Apply/Query 与 refSegTree 逐步对比（求和 monoid） |
| `TestCorrectness_RangeAssignMin` | n=200，100k 次 | 同上（区间赋值+最小值 monoid） |
| `TestCorrectness_LargeRandom` | n=100/1k/10k，各 100k 次 | 大规模随机验证（-short 跳过） |
| `TestCorrectness_Stress` | n=1_000_000，100k 次 | 百万规模压力测试（-short 跳过） |

## 性能基线（Apple M4，-benchtime=3s）

```bash
go test -bench=. -benchtime=3s -benchmem -count=1 ./pkg/ds/segtree/
```

（运行后填写实际结果）

```text
BenchmarkSegTree_Set/n=100-10             78.67 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Set/n=1000-10            112.5 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Set/n=10000-10           156.6 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Set/n=100000-10          200.4 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Set/n=1000000-10         318.1 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Apply/n=100-10           122.1 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Apply/n=1000-10          201.3 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Apply/n=10000-10         290.7 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Apply/n=100000-10        399.3 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Apply/n=1000000-10       606.9 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Query/n=100-10           109.0 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Query/n=1000-10          181.4 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Query/n=10000-10         255.6 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Query/n=100000-10        370.8 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Query/n=1000000-10       532.4 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Mixed/n=100-10           119.1 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Mixed/n=1000-10          174.0 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Mixed/n=10000-10         241.7 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Mixed/n=100000-10        332.2 ns/op    0 B/op    0 allocs/op
BenchmarkSegTree_Mixed/n=1000000-10       561.9 ns/op    0 B/op    0 allocs/op
```

结论：所有操作 `0 allocs/op`，ns/op 随 n 增长呈 log 级缓慢上升（每规模倍增约增加一个 log 步）。
