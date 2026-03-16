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
| `SafeCall(tag string, f func())` | 需要在日志中区分调用来源 |
| `SafeExecWithError(tag string, f func() error) error` | f 有业务 error 返回，同时防止 panic |
| `SafeExecResult[T](tag string, f func() T) (T, error)` | f 有返回值，同时防止 panic |

panic 发生时：捕获 panic 值和完整调用栈，写入日志，然后**正常返回**（不再向上传播）。

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
    safeexec.SafeCall("giveReward", func() {
        giveReward(p)
    })
}

// f 有 error 返回
err := safeexec.SafeExecWithError("processOrder", func() error {
    return processOrder(orderID)
})
if err != nil {
    // 业务 error 或 panic 包装的 error
}

// f 有返回值
result, err := safeexec.SafeExecResult[int]("calcScore", func() int {
    return calcScore(data)
})
if err != nil {
    // panic 发生，result 是零值，不应使用
}
```

## nil 函数处理

- `Safe / SafeCall`：f 为 nil 时记录日志后直接返回
- `SafeExecWithError`：f 为 nil 时返回 error
- `SafeExecResult`：f 为 nil 时返回零值 + error

## 注意事项

- panic 被捕获后**写入日志**，调用方无感知；需通过日志监控发现问题
- `tag` 建议传入有业务含义的常量字符串，如 `"giveReward"`，方便日志定位
- 不适合替代正常的 error 处理流程，仅用于隔离意外崩溃
