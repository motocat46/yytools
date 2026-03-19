# slicex 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `slicex_test.go` | MinInSliceOK/MaxInSliceOK/MinBy/MaxBy 空切片/单元素/重复最值/负数；Benchmark |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/slicex/

# 基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/slicex/
```

## 性能基准（Apple M4，benchtime=2s，n=1000）

| 操作 | ns/op | allocs/op |
|------|-------|-----------|
| MinInSlice | 1673 | 1 |
| MaxInSlice | 1662 | 1 |
