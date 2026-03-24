# timeutil 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、为什么扩展 'd' 单位而不直接换算后传给标准库

**标准库 `time.ParseDuration` 不支持 'd'（天）**，但业务代码频繁需要"3 天后过期"这类表达。

直接让调用方自己计算 `3 * 24 * time.Hour` 有两个问题：
1. **散落重复代码**：每处用到"天"的地方都要手写乘法。
2. **溢出无提示**：大天数乘以 `24 * time.Hour` 的纳秒值可能溢出 int64，调用方通常不检查。

本包的选择：
- 仅添加 `d` 单位扩展，其余全部委托给标准库。
- 在天数与小时合并时显式做溢出检测（借助 `overflow.MulInt` / `AddInt`），最大支持约 290 年。

---

## 二、支持小数天数（如 `1.5d` = 36h）

标准库 `time.ParseDuration` 本身支持小数，如 `"1.5h"` = 90 分钟。本包最初用 `strconv.Atoi` 解析天数，只接受整数，与标准库行为不一致。

修复：改用 `strconv.ParseFloat`，天数乘以 `float64(24*time.Hour)` 后转换为 `int64` 纳秒。

浮点溢出检测比整数溢出更复杂：
- `float64(math.MaxInt64)` 因浮点精度向上取整为 2^63，因此检查用 `>= float64(math.MaxInt64)`（不是 `>`）
- `float64(math.MinInt64)` 恰好等于 -2^63，可精确表示，检查用 `< float64(math.MinInt64)`

---

## 三、负数天的处理与 `-0d` 的特殊情形

`-2d1h30m` 表示向前推 2 天 1 小时 30 分钟（结果为 -49.5 小时）。

**符号检测用字符串前缀，而非数值比较。**

最初实现用 `days < 0`（整数版本）或 `daysFloat < 0`（浮点版本）判断方向。两者都有盲区：`"-0d30m"` 中 `-0.0 < 0` 在 IEEE 754 下为 `false`，导致 `-0d30m` 错误返回 `+30m`。

修复：改用 `strings.HasPrefix(left, "-")`，直接检测字符串层面的负号，对 `-0` 也能正确判断。

```
"-2d1h30m" → left="-2", right="1h30m"
daysPart = int64(-2 * 24h) = -172800000000000
remain = +1h30m
HasPrefix("-2", "-") = true → result = daysPart - remain = -49h30m ✓

"-0d30m" → left="-0", right="30m"
daysPart = int64(-0.0 * 24h) = 0
HasPrefix("-0", "-") = true → result = 0 - 30m = -30m ✓（原来错误返回 +30m）
```

若 `right` 部分已携带负号（如 `-2d-1h`），中间的 `-` 被提前拒绝，不会进入解析逻辑。

---

## 四、为什么用 float 乘法代替 `overflow.MulInt`

原实现：`overflow.MulInt(int64(time.Hour)*24, int64(days))` 需要两步 int64 乘法，且要求 `days` 是整数。

新实现：`daysFloat * float64(24*time.Hour)` 一步完成，支持小数天，再统一做 float 范围检查后 `int64(...)` 截断。`overflow.MulInt` 只剩 `AddInt` / `SubInt` 用于合并天数与剩余时长，职责更清晰。
