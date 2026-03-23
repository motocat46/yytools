# random 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `random_test.go` | 范围正确性、负数范围、边界值、全类型范围、派生类型、确定性重放、assert 触发、基准测试 |
| `example_test.go` | 可运行的使用示例（`go test -run Example`） |

## 测试设计要点

### 范围正确性
所有有符号（int8/int16/int32/int64）与无符号（uint8/uint32/uint64）类型均有独立测试，每种各重复 200 次。

### 关键边界场景
| 场景 | 说明 |
|------|------|
| `low == high` | 始终返回该值，不进入随机逻辑 |
| 全类型范围 `[MinT, MaxT]` | 验证不 panic，覆盖 int64/uint64 的溢出边界（uint64 全范围 n=0 特判）|
| 跨零范围 `[-1, MaxInt64]` | 验证有符号 64 位的 uint64 补数算术路径 |
| 派生类型 `type myInt int32` | 验证泛型派发不依赖具体类型名 |

### 确定性重放
- 相同种子 + 相同调用序列 → 完全相同的结果（测试 50 次）
- 不同种子 → 50 次调用不全相同（概率约 1/10^300）
- 跨类型重放：int32 和 uint8 交叉调用，同种子两组结果一致

### assert 触发
`low > high` 时触发 assert（panic），验证 int 和 int8 两种类型。

## 运行方式

```bash
# 快速验证
go test ./pkg/algorithms/mathx/random/

# 含示例
go test -run Example ./pkg/algorithms/mathx/random/

# 基准测试
go test -bench=. -benchmem ./pkg/algorithms/mathx/random/
```

## 基准基线（Apple M4，2026-03）

| 基准 | ns/op | allocs/op |
|------|-------|-----------|
| `BenchmarkRandInt_Global` | ~7.7 ns | 0 |
| `BenchmarkRandIntWith_Seeded` | ~5.2 ns | 0 |

全局随机源（`RandInt`）略慢于本地实例（`RandIntWith`），因全局源内部有互斥保护。
两者均零内存分配。
