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

## PowMod / Comb / CombTable

| 文件 | 覆盖范围 |
|------|---------|
| `powmod_test.go` | PowMod 基本幂运算、`mod=1`、Fermat 小定理、负指数与零模 panic 路径 |
| `combination_test.go` | Comb 基本/对称性/越界/panic；CombTable 基本/对称性/越界/panic；`CombTable` 对 `Comb` 与 `math/big` 参考模型一致性验证 |

## 运行命令

```bash
# 快速验证
go test -short -count=1 ./pkg/algorithms/mathx/

# 全量（含 bigInt 对比）
go test -count=1 -timeout=120s ./pkg/algorithms/mathx/

# 竞态检测
go test -race -count=1 -timeout=120s ./pkg/algorithms/mathx/

# 新增组合数相关基准
go test -run '^$' -bench='BenchmarkPowMod|BenchmarkComb|BenchmarkCombTable_Build|BenchmarkCombTable_Query' -benchtime=3s -benchmem ./pkg/algorithms/mathx/
```

## PowMod / Comb / CombTable 性能基线（Apple M4，benchtime=3s）

| 操作 | ns/op | allocs/op |
|------|-------|-----------|
| PowMod | 83.95 | 0 |
| Comb(1000,500) | 1956 | 0 |
| CombTable Build n=1k | 9492 | 3 |
| CombTable Build n=1M | 8746147 | 3 |
| CombTable Query | 6.749 | 0 |

## 新增模块注意

- `PowMod`、`Comb`、`CombTable` 的热路径使用 `assert.AssertFast`，基准目标为零分配查询
- `Comb` 当前仅支持 `n < mod`
- `CombTable` 当前仅支持 `maxN < mod`；若需要 `n >= mod` 的组合数模质数查询，应使用 Lucas 定理
