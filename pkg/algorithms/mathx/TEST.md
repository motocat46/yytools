# mathx 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `mathx_test.go` | GcdR/GcdI/Gcd 正确性（含负数 panic）；Fibonacci.Calculate 正确性与 OOM 保护；FibNTable 边界；FibNMax 各类型上界 |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/algorithms/mathx/

# 基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/algorithms/mathx/
```

## 性能基准（Apple M4，benchtime=2s）

| 操作 | ns/op | allocs/op |
|------|-------|-----------|
| FibNTable（查表） | 1.51 | 0 |
| Fibonacci.Calculate（带记忆化） | 2.04 | 0 |

## 注意

- `GcdR/GcdI` 传入负数会 panic（运行时检查，不依赖 assert build tag）
- `Fibonacci.Calculate` 传入超出类型 T 斐波那契范围的 n 会 panic（利用 `FibNMax[T]()` 约束）
