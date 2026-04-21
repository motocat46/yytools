# Test

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `sparsetable_test.go` | New/Query/Len，min/max/gcd 三种 merge，全区间/单元素/同值/升序/降序，panic 路径 |
| `correctness_test.go` | 参考模型对比（n=1000，100k 次），大规模随机（n≤10k），压力（n=1M，100k 次） |
| `bench_test.go` | Build/Query 规模曲线（100～1M），SparseTable O(1) vs 朴素 O(n) 对比 |

## 运行命令

```bash
# 快速验证（跳过大规模测试）
go test -short -count=1 ./pkg/ds/sparsetable/

# 全量正确性（含大规模）
go test -count=1 -timeout=120s ./pkg/ds/sparsetable/

# 竞态检测（虽为非并发结构，仍应通过）
go test -race -count=1 ./pkg/ds/sparsetable/

# 基准测试
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/sparsetable/

# 仅 O(1) vs O(n) 对比基准
go test -bench=BenchmarkSparseTable_QueryVsNaive -benchtime=3s -benchmem ./pkg/ds/sparsetable/
```

## 性能基线

> 环境：Apple M4，`go test -bench=. -benchtime=3s -benchmem ./pkg/ds/sparsetable/`

| 操作 | n=1k | n=10k | n=100k | n=1M |
|------|------|-------|--------|------|
| Build (ns/op) | 17608 | 240062 | 3357393 | 38886618 |
| Query (ns/op) | 4.370 | 5.241 | 6.388 | 7.776 |
| Query allocs/op | 0 | 0 | 0 | 0 |
| SparseTable O(1) vs Naive O(n) @ n=100k | 6.539 ns/op | — | — | 18856 ns/op |
