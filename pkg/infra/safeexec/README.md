# safeexec 使用文档

将 `panic` 的传播边界限制在单个函数调用内，保障主流程不受影响。适合需要"单个任务失败不影响整体"的场景。

## 适用场景

- 为多个用户逐一执行操作（奖励发放、通知推送），任意一个 panic 不中断其他用户
- 执行第三方回调或插件逻辑
- 定时任务中的独立业务单元

## API

| 函数 | 适用场景 |
|------|---------|
| `Safe(f func())` | 最简用法，tag 固定为 "anonymous" |
| `SafeExec(tag string, f func())` | 需要在日志中区分调用来源 |
| `SafeExecErr(tag string, f func() error) error` | f 有业务 error 返回，同时防止 panic |
| `SafeExecVal[T](tag string, f func() T) (T, error)` | f 有返回值，同时防止 panic |

panic 发生时行为因函数而异：

| 函数 | panic 处理方式 |
|------|--------------|
| `Safe` / `SafeExec` | 通过 `slog.Default()` 记录 panic 值和完整调用栈，**正常返回** |
| `SafeExecErr` / `SafeExecVal` | 将 panic 值和完整调用栈**嵌入返回的 error**，由调用方决定如何记录 |

有返回值的函数不在内部打日志，是因为调用方持有更多上下文（如 request ID、user ID），由调用方统一记录更合理，也避免重复打日志。

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/infra/safeexec"

// 最简用法
safeexec.Safe(func() {
    doSomething()
})

// 带 tag（推荐，日志可定位来源）
for _, player := range players {
    p := player
    safeexec.SafeExec("giveReward", func() {
        giveReward(p)
    })
}

// f 有 error 返回
err := safeexec.SafeExecErr("processOrder", func() error {
    return processOrder(orderID)
})
if err != nil {
    // 业务 error 或 panic 包装的 error
}

// f 有返回值
result, err := safeexec.SafeExecVal[int]("calcScore", func() int {
    return calcScore(data)
})
if err != nil {
    // panic 发生，result 是零值，不应使用
}
```

## nil 函数处理

- `Safe / SafeExec`：f 为 nil 时记录日志后直接返回
- `SafeExecErr`：f 为 nil 时返回 error
- `SafeExecVal`：f 为 nil 时返回零值 + error

## 日志配置

`Safe` / `SafeExec` 的 panic 日志通过 `slog.Default()` 输出。

**默认行为**：未配置时写入 stderr，格式为纯文本，适合开发调试。

**生产环境**：建议在程序启动时将业务日志注入 `slog.Default()`，配置一次后全局生效，无需修改任何调用处：

```go
// 以 zap 为例，需引入 go.uber.org/zap/exp/zapslog
import (
    "log/slog"
    "go.uber.org/zap/exp/zapslog"
)

func main() {
    zapLogger, _ := zap.NewProduction()
    slog.SetDefault(slog.New(zapslog.NewHandler(zapLogger.Core())))

    // 此后所有 SafeExec 的 panic 日志自动走 zap
}
```

**其他日志库**：只要实现了 `slog.Handler` 接口即可接入，标准库 `slog.NewJSONHandler` / `slog.NewTextHandler` 也可直接使用。

## 注意事项

- panic 被捕获后写入 `slog.Default()`，调用方无感知；需通过日志监控发现问题
- `tag` 建议传入有业务含义的常量字符串，如 `"giveReward"`，方便日志定位
- 不适合替代正常的 error 处理流程，仅用于隔离意外崩溃
