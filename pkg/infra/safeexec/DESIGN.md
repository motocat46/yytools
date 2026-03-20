# safeexec 设计演进记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## V1：log.Printf + SafeCall/SafeExecWithError/SafeExecResult

初始实现。

**问题：**

1. **日志耦合**：`log.Printf` 绑定标准库 logger，无法接入业务层的 zap / zerolog 等日志框架，生产环境 panic 日志与业务日志分散在不同 sink，难以关联。
2. **有返回值的函数也在内部打日志**：`SafeExecWithError` 和 `SafeExecResult` 捕获 panic 后既打日志又返回 error，违反"log or return"原则——调用方持有更多上下文（request ID、user ID），由调用方统一记录更合理，内部打日志反而制造重复。
3. **命名不统一**：`SafeCall` vs `SafeExec` 前缀混用；`SafeExecWithError` 偏长，`SafeExecResult` 语义模糊。

---

## V2：slog.Default() + SafeExec/SafeExecErr/SafeExecVal（当前版本）

### 决策一：用 slog.Default() 替代 log.Printf

`slog.Default()` 是标准库（Go 1.21+）提供的全局 logger 桥接点。

- **零外部依赖**：库本身不引入 zap/zerolog，保持轻量
- **应用层一次注入**：`slog.SetDefault(slog.New(zapslog.NewHandler(...)))` 在程序启动时配置一次，所有 `SafeExec` 的 panic 日志自动走业务 logger
- **默认行为友好**：未配置时写 stderr，纯文本格式，适合开发调试，无需任何配置

### 决策二：有返回值的函数不在内部打日志（log-or-return 原则）

| 函数 | 有无返回值 | panic 处理方式 | 理由 |
|------|-----------|--------------|------|
| `Safe` / `SafeExec` | 无 | `slog.Error` | 无返回值，不打日志调用方完全无感知，必须在内部记录 |
| `SafeExecErr` / `SafeExecVal` | 有 | 嵌入 error 返回 | 调用方收到 error 后有完整上下文，由调用方决定如何记录，避免重复打日志 |

panic error 格式：`[safeexec] panic in <tag>: <panic值>\n<完整调用栈>`，调用方从 error.Error() 可获取全部信息。

### 决策三：命名规范

- 所有函数统一 `SafeExec` 前缀
- 后缀遵循 Go 惯用法：
  - 无后缀：无返回值，最简形式
  - `Err`：返回 `error`（比 `WithError` 短，与 `fmt.Errorf` 命名风格一致）
  - `Val`：返回泛型值 + error（比 `Result` 更直观表达"有值"）
- `Safe` 作为 `SafeExec("anonymous", f)` 的快捷方式保留

### 决策四：f == nil 的处理

| 函数 | f == nil 时 |
|------|------------|
| `Safe` / `SafeExec` | `slog.Warn` 后直接返回，不 panic |
| `SafeExecErr` | 返回 error，不 panic |
| `SafeExecVal` | 返回零值 + error，不 panic |

`f == nil` 是调用方 bug，不应让调用方感知不到（静默忽略），也不该用 panic 中断主流程——用 Warn 而非 Error，因为函数体未执行，没有 panic 也没有业务错误，只是一个"空操作"警告。
