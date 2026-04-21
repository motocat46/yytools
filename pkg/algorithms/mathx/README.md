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

### PowMod（快速幂取模）

`PowMod(base, exp, mod int64) int64`

计算 `(base^exp) % mod`，时间复杂度 `O(log exp)`，零分配。
要求 `mod > 0`、`exp >= 0`，并且中间乘积不溢出的前提为 `mod^2 < 2^63`。

```go
mathx.PowMod(2, 10, 1_000_000_007) // 1024
mathx.PowMod(3, 0, 7)              // 1
```

### Comb（组合数单次查询）

`Comb(n, k, mod int64) int64`

返回 `C(n,k) mod mod`，时间复杂度 `O(k)`，适合偶发单次查询。
当前实现基于 Fermat 小定理，要求：

- `mod` 为质数
- `n < mod`
- `k < 0` 或 `k > n` 时返回 `0`

```go
mathx.Comb(5, 2, 1_000_000_007)  // 10
mathx.Comb(10, 3, 1_000_000_007) // 120
```

### CombTable（组合数预计算表）

`NewCombTable(maxN int, mod int64) *CombTable`

`.C(n, k int) int64`

建表时间 `O(maxN)`，单次查询 `O(1)`，适合同一质数模下大量查询。
当前实现要求 `maxN < mod`；若需要处理 `n >= mod`，应改用 Lucas 定理。

```go
ct := mathx.NewCombTable(1000, 1_000_000_007)
ct.C(20, 10)  // 184756
ct.C(100, 3)  // 161700
```

## 子模块

| 子模块 | 功能 |
|--------|------|
| [bits](bits/README.md) | 位运算工具：符号检测、2 的幂判断、统计 1 的位数 |
| [overflow](overflow/README.md) | 有符号整数四则运算溢出检测（加减乘除）|
| [probability_distribution](probability_distribution/) | 静态/动态概率分布采样 |
| [random](random/) | 随机整数生成 |
| [sampling](sampling/README.md) | Floyd 算法不重复采样、带最小间隔采样 |
