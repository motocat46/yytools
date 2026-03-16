# bits 使用文档

位运算工具函数，基于位操作技巧实现，性能优于通用算法。

## API

| 函数 | 说明 | 时间复杂度 |
|------|------|-----------|
| `AreSignsOpposite[T Integer](a, b T) bool` | 判断两整数是否符号相反 | O(1) |
| `IsPowerOfTwo[T Integer](a T) bool` | 判断整数是否为 2 的幂 | O(1) |
| `CountingBits[T Integer](a T) int` | 统计整数二进制表示中 1 的个数 | O(k)，k 为 1 的位数 |

类型约束：`T` 满足 `base.Integer`（有符号或无符号整数）。

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/algorithms/mathx/bits"

// 符号相反判断
bits.AreSignsOpposite(3, -5)  // true
bits.AreSignsOpposite(3, 5)   // false
bits.AreSignsOpposite(-3, -5) // false

// 2 的幂判断
bits.IsPowerOfTwo(8)   // true  (1000)
bits.IsPowerOfTwo(6)   // false (0110)
bits.IsPowerOfTwo(0)   // false（特殊处理）
bits.IsPowerOfTwo(1)   // true  (0001)

// 统计 1 的个数（Hamming Weight）
bits.CountingBits(7)   // 3  (0111)
bits.CountingBits(255) // 8  (11111111)
bits.CountingBits(0)   // 0
```

## 实现原理

- `AreSignsOpposite`：利用异或最高位（符号位），异或结果 < 0 则符号相反
- `IsPowerOfTwo`：`a & (a-1)` 清除最低有效位，2 的幂只有一个 1，结果必为 0
- `CountingBits`：Brian Kernighan 算法，每次 `a &= a-1` 清除一个 1，循环次数等于 1 的个数
