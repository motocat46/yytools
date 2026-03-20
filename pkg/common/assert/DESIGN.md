# assert 设计记录

## V1：带 file:line 前缀的可关闭断言（2022）

**目标**：对 `panic` 进行封装，提供 Assert 语义，并在 panic 消息中自动附加调用处的文件路径和行号。

**实现**：
- `isAssertOpen` 全局变量控制开关
- `SetAssert(bool)` / `IsAssertOpen()` 提供运行时控制
- `getPrefix()` 通过 `runtime.Caller(2)` 获取调用处的 `file:line`，拼入 panic 消息

**设计约束**：`runtime.Caller(2)` 假设调用链固定为 `getPrefix → Assert → 调用方`，Assert 不能被包装，否则行号指向包装层而非真实调用处。

---

## V2：Build tag 控制开关（2025/5）

**动机**：在 `cmd/demo/main.go` 中需要显式调用 `SetAssert(true)` 才能开启断言，遗漏时断言静默失效。改为 build tag 控制默认值：
- 无 tag（正常编译）：`assertion_on.go` 设置 `isAssertOpen = true`
- `-tags assertion_off`：`assertion_off.go` 设置 `isAssertOpen = false`

**遗留问题**：可关闭机制开始引发代码审查噪音——每次发现用 `assert.Assert` 守护安全关键路径时，都要将其改为无条件 `panic`，原因是"如果断言被关闭，此处会静默失效"。

---

## V3：永远开启，纯语法糖（2026/3）—— 当前版本

**决策过程**：

代码审查中反复出现如下模式：
1. 发现 `assert.Assert(...)` 守护某个安全关键检查
2. 判断"如果断言被关闭，此处会产生静默错误或 nil dereference"
3. 将其改为无条件 `panic`

这个循环在 Stack、Queue、GcdR、RadixSort、Fibonacci、sorted_set 等多处发生，是明显的设计信号。

**根因分析**：

1. **file:line 前缀是冗余的**。`safeexec` 的所有 `recover()` 均配套 `debug.Stack()`，完整调用链已记录。Go 的 panic 自带栈帧，`runtime.Caller(2)` 获取的信息已包含在内，且使栈帧更嘈杂。

2. **可关闭的价值主张站不住脚**。
   - *性能理由*：Assert 守护的条件（`length > 0`、`n >= 1`、`ptr != nil`）全是 O(1) 比较，分支预测后开销可忽略不计。
   - *信任前提*：关闭的前提是"代码已经过充分测试"——但如果代码正确，断言永远不触发，关闭没有意义；如果代码有 bug，关闭断言让 bug 在生产静默传播，代价更高。

3. **Assert 的核心价值是语法糖**。`assert.Assert(cond, msg)` 用正向条件表达契约，比 `if !cond { panic(msg) }` 更直觉，这一点值得保留。

**变更内容**：
- 删除 `isAssertOpen`、`SetAssert`、`IsAssertOpen`、`getPrefix`
- 删除 `assertion_on.go`、`assertion_off.go`
- `Assert` 和 `AssertFast` 直接无条件检查条件，触发即 panic
- panic 消息不再附加 `file:line` 前缀（Go 栈帧已提供）

**Assert vs panic 的使用边界**（由此次重构确立）：

| 场景 | 使用 |
|------|------|
| 调用方违反契约（空栈出栈、负数索引等） | `assert.Assert` |
| 运行时边界保护（时间戳溢出、类型范围超限） | 无条件 `panic` |
| 外部输入校验 | 返回 `error`，不用 assert 也不用 panic |
