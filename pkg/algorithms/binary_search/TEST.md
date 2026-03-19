# binary_search 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `binary_search_test.go` | BinarySearch/LeftBound/RightBound/SearchBound/SearchBoundOpt 边界用例；10 万次随机大规模一致性验证（对比 stdlib sort.SearchInts）；多规模 Benchmark |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/algorithms/binary_search/

# 跳过大规模测试（CI 快速通道）
go test -short ./pkg/algorithms/binary_search/

# 随机大规模一致性测试（~2s）
go test -run TestBinarySearch_RandomLarge ./pkg/algorithms/binary_search/

# 基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/algorithms/binary_search/
```

## 性能基准（Apple M4，benchtime=2s）

O(log n) 复杂度曲线：n 增大 10 倍，耗时约增加 3~4 ns（符合预期）。

| 操作 | n=100 | n=1K | n=10K | n=100K | n=1M |
|------|-------|------|-------|--------|------|
| BinarySearch (ns/op) | 8.0 | 11.9 | 25.5 | 39.4 | 45.6 |
| LeftBound (ns/op) | 6.4 | 8.8 | 14.5 | 27.3 | 32.9 |
| SearchBoundOpt (ns/op) | 9.2 | 12.5 | 22.4 | 33.4 | 40.4 |

所有操作 **0 allocs/op**。
