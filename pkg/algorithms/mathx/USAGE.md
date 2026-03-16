# mathx 使用文档

常用数学函数集合，包含 GCD、绝对值、Min/Max，以及子模块索引。

## 函数列表

### GCD（最大公约数）

| 函数 | 实现方式 | 说明 |
|------|---------|------|
| `Gcd[T Integer](x, y T) T` | 循环（推荐）| 两个非负整数的最大公约数 |
| `GcdI[T Integer](x, y T) T` | 循环 | Gcd 的底层实现 |
| `GcdR[T Integer](x, y T) T` | 递归 | 等价实现，栈空间 O(log n) |

- 时间复杂度：O(log(min(x, y)))
- 入参必须 ≥ 0，否则触发断言 panic

### 基础运算

| 函数 | 说明 |
|------|------|
| `Abs[T Integer](a T) T` | 整数绝对值（⚠️ 对最小有符号整数会溢出，与 Go 内置行为一致）|
| `Min[T Ordered](a, b T) T` | 两数较小值 |
| `Max[T Ordered](a, b T) T` | 两数较大值 |

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/algorithms/mathx"

// GCD
mathx.Gcd(12, 8)   // 4
mathx.Gcd(100, 75) // 25

// 绝对值
mathx.Abs(-42)   // 42
mathx.Abs(int8(-128)) // ⚠️ 溢出，仍返回 -128

// Min / Max（Go 1.21+ 标准库也有，此处为泛型兼容版）
mathx.Min(3, 5)   // 3
mathx.Max(3.14, 2.71) // 3.14
```

## 子模块

| 子模块 | 包路径 | 功能 |
|--------|--------|------|
| `bits` | `mathx/bits` | 位运算工具：符号检测、2 的幂判断、统计 1 的位数 |
| `overflow` | `mathx/overflow` | 有符号整数四则运算溢出检测（加减乘除）|
| `probability_distribution` | `mathx/probability_distribution` | 静态/动态概率分布采样 |
| `random` | `mathx/random` | 随机整数生成 |
| `sampling` | `mathx/sampling` | 水塘抽样（从流式数据中等概率采样）|

各子模块有独立的 USAGE.md，请参阅对应目录。
