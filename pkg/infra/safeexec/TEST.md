# safeexec 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `safe_test.go` | `Safe`/`SafeExec`：正常执行、panic 恢复、nil 函数；`SafeExecErr`：无错误/有错误/panic 转 error/nil 函数；`SafeExecVal`：基础类型/结构体/切片/map 返回值、panic 转零值+error；基准测试三个有返回值变体 |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/infra/safeexec/

# 竞态检测
go test -race ./pkg/infra/safeexec/

# 基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/infra/safeexec/
```

## 性能基准（Apple M4，benchtime=2s）

| 操作 | ns/op | B/op | allocs/op |
|------|-------|------|-----------|
| SafeExec（无返回值） | 2.963 | 0 | 0 |
| SafeExecErr（返回 error） | 3.266 | 0 | 0 |
| SafeExecVal（返回泛型值） | 3.795 | 0 | 0 |

正常路径零内存分配；panic 路径因 `runtime.Stack` 会有分配，不在基准中体现（非热路径）。

## 注意

- 测试期间用 `TestMain` 将 `slog.Default()` 重定向到 `io.Discard`，避免 panic 恢复日志污染测试输出
- `Safe`/`SafeExec` 在 panic 时写 `slog.Error`；`SafeExecErr`/`SafeExecVal` 在 panic 时**不打日志**，将 panic 信息嵌入 error 返回（log-or-return 原则，详见 DESIGN.md）
