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

---

## 五、`padTwo` 不传 `orig` 参数——错误上下文归属

**问题**：早期版本 `padTwo(s, orig string)` 接受原始输入作为第二个参数，目的是在错误信息中附上上下文。

实际出现的问题是**双重上下文**：`normalizeDate` 调用 `padTwo` 时传入自己的 `orig`，`parseDate` 调用 `normalizeDate` 时又在 `fmt.Errorf` 里 `%w` 包装并追加一遍 "原始输入"，导致错误消息里出现两次相同的原始字符串。

**修复**：`padTwo` 只报告它自己知道的信息（"数字段为空"、"数字段非数字 %q"、"数字段超过 2 位 %q"），不接触任何外层上下文。`normalizeDate` / `normalizeTime` 同样不在自己的错误里重复原始输入。**上下文只由距离调用方最近的一层（`parseDate` / `parseDateTime`）添加一次。**

原则：**错误上下文归属最外层调用者，内层函数描述它们直接知道的事实。**

---

## 六、`ParseDuration` 显式拒绝 `+` 前缀

`strconv.ParseFloat` 接受 `+` 前缀（`"+2.5"` 合法），因此原实现会静默接受 `"+2d"` 并返回正常结果，但这个行为没有在任何文档中说明。

**问题**：`+` 前缀的含义不直观（大多数时长格式不使用它），且标准库 `time.ParseDuration` 也不接受 `+`。让 `ParseDuration` 接受标准库不接受的语法会造成语义不一致。

**决策**：在 `validateDuration` 中首字符为 `+` 时直接返回错误，文档更新为"正数直接省略符号"。行为与标准库对齐，不引入无文档的特例。

---

## 七、`TestMsVariants` 使用硬编码绝对时间戳

`StartOfDayMs(ms, cst)` 的实现等价于 `StartOfDay(time.UnixMilli(ms).In(cst)).UnixMilli()`。

若测试写成：
```go
got := StartOfDayMs(ms, cst)
want := msOf(StartOfDay(base))
assert got == want
```
这只是验证"两边的代码等价"，而不是验证"结果是正确的日历天起始时间"。如果 `StartOfDay` 有 bug，两侧会同步出错，测试通过但结果是错的。

**修复**：用独立脚本 `time.Date(2026, 3, 26, 0, 0, 0, 0, cst).UnixMilli()` 计算期望值，硬编码为字面量。测试失去对实现的依赖，真正验证正确性。

同一原则也适用于 `TestMsVariants_UTC` 中的时区差异验证：CST 和 UTC 下的"今天 0 点"对应不同时间戳，两个硬编码常量的差值（`1774483200000 - 1774454400000 = 28800000 ms = 8h`）本身就是可以人工验证的事实。

---

## 八、`lastDayOfMonth` 提取为命名 helper

`startOfNthMonthDay` 原本内联 `time.Date(y, m+1, 0, ...).Day()` 来获取月末天数。该表达式不直观（用"下月第 0 天"倒推本月末），且在函数内被引用两次时会重复调用。

**提取为 `lastDayOfMonth(y, m, loc)`**：
- 命名使意图直接可读，不需要注释解释"这个 day=0 是什么含义"
- 在 `startOfNthMonthDay` 内调用一次后缓存在局部变量 `last`，避免重复调用
