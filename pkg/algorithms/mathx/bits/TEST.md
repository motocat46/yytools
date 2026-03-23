# bits 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `bits_test.go` | AreSignsOpposite、IsPowerOfTwo、CountingBits 的正常值、边界值、多类型（int/int64）、一致性 |
| `example_test.go` | 可运行的使用示例（`go test -run Example`） |

## 测试设计要点

- **AreSignsOpposite**：覆盖正负、同正、同负、零与正/负、零与零、int32/int64 最大边界值；注意 `(0, -1)` 返回 true（0 的符号位为 0，-1 的符号位为 1）
- **IsPowerOfTwo**：覆盖 2^0 ~ 2^60、0（特殊情形，0 & -1 == 0 但不是 2 的幂）、负数（负数的 a-1 回绕导致 a&(a-1) 非零）
- **CountingBits**：覆盖 0 ~ 1024、255（全 1）、负数补码（int8(-1)=64 bits all-1）
- **一致性测试**：验证"IsPowerOfTwo(n) ⇒ CountingBits(n)==1"的不变量

## 运行方式

```bash
# 快速验证
go test ./pkg/algorithms/mathx/bits/

# 含示例
go test -run Example ./pkg/algorithms/mathx/bits/
```

## 基准测试

本包为纯位运算，每次操作 O(1) / O(popcount)，不设基准基线。
