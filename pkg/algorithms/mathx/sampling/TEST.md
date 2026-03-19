# sampling 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `sampling_test.go` | SampleKDistinctFloyd 不重复性/均匀性/边界；SampleWithMinGap 间隔约束/有序性；随机大规模一致性验证；Benchmark |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/algorithms/mathx/sampling/

# 基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/algorithms/mathx/sampling/
```

## 性能基准（Apple M4，benchtime=2s，k=10，范围 [1, 1000]）

| 操作 | ns/op | allocs/op |
|------|-------|-----------|
| SampleKDistinctFloyd | 388 | 5 |
| SampleWithMinGap | 677 | 6 |

## 注意

- **平台限制**：`int(hi) - int(lo)` 使用 `int` 中间类型，32 位平台范围跨度不得超过 `math.MaxInt32`
- 固定随机种子：`rand.New(rand.NewPCG(42, 0))` 可复现测试结果
