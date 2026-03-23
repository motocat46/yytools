# overflow 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、双 API 设计：`(T, bool)` + `Assert` 变体

每种运算提供两个版本：

| 函数 | 返回溢出标志 | 适用场景 |
|------|------------|---------|
| `MulInt` / `AddInt` / ... | `(result T, overflow bool)` | 调用方需要区分溢出/正常并分别处理 |
| `MulIntAssert` / `AddIntAssert` / ... | `result T`（溢出时 panic）| 溢出即调用方 bug，用 assert 直接暴露 |

`Assert` 变体是对 `(T, bool)` 变体的薄封装，逻辑零重复：

```go
func MulIntAssert[T base.Signed](a, b T) T {
    res, overflow := MulInt(a, b)
    assert.Assert(!overflow, a, b, res)
    return res
}
```

**为什么不只提供 assert 版本？**
某些场景（如协议解析、用户输入处理）溢出是预期的合法情况，调用方需要知道"溢出了"并走不同分支，不应 panic。提供 `(T, bool)` 版本让这些场景也能使用本包。

---

## 二、`MulInt` 的实现策略：按位宽分支

**小类型（int8/int16/int32）**：提升到 int64 做中间计算，再检查结果是否超出原类型范围。实现简单，精确。

**int64（64 位平台的 int 也走此路径）**：无法再提升，改用**除法边界检查**：
- 溢出的充要条件：`|a| > MaxT / |b|`
- 分符号讨论避免带符号整数的陷阱（`MaxT / -b` 方向取反）

**为什么不用 `math/bits.Mul64`？**
`bits.Mul64` 返回 `(hi, lo uint64)` 需要再做符号转换，代码更复杂。目前的除法检查已足够准确且更直接。

---

## 三、`DivInt` 的唯一溢出情形

整数除法只有一种情况会溢出：`MinT / -1`。

数学上 `|MinT| = MaxT + 1`，超出正数范围，补码运算回绕，CPU 实际返回 `MinT` 本身（即结果是错的但不报错）。本包显式检测并返回溢出标志。

其他除法情形（包括除以 0）由 Go 运行时处理（panic），本包不重复处理。

---

## 四、`AddInt` / `SubInt` 的符号检测

加法溢出的充要条件（补码规则）：
- 两个**非负数**相加结果为**负数** → 正向溢出
- 两个**负数**相加结果为**非负数** → 负向溢出
- 一正一负：永远不会溢出

`SubInt(a, b)` 转化为符号等价判断，与 `AddInt` 对称，不重用 `AddInt` 实现以保持各运算的逻辑独立。
