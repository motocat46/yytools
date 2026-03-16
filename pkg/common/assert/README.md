# assert 使用文档

运行时断言框架，用于在开发期快速发现逻辑错误。断言失败时触发 `panic`，并附带文件路径和行号。

## API

| 函数 | 说明 |
|------|------|
| `SetAssert(open bool)` | 全局开关，开启或关闭所有断言 |
| `IsAssertOpen() bool` | 查询当前断言是否开启 |
| `Assert(condition bool, list ...interface{})` | 条件为 false 时 panic，附带自定义消息 |
| `AssertFast(cond bool)` | 条件为 false 时 panic，无附加消息（性能稍好） |

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/common/assert"

// 开启断言（默认状态由 build tag 决定）
assert.SetAssert(true)

// 带消息的断言
assert.Assert(n > 0, "n must be positive, got:", n)

// 快速断言（无消息）
assert.AssertFast(ptr != nil)
```

## 开关控制

断言默认状态由 build tag 决定：
- 无 tag（正常编译）：由 `assertion_on.go` 或 `assertion_off.go` 控制初始值
- 可在运行时通过 `SetAssert(false)` 关闭，适合生产环境

```go
// 生产环境关闭断言
func init() {
    assert.SetAssert(false)
}
```

## Assert vs AssertFast

- `Assert`：可附带多个参数作为错误描述，panic 信息更详细，适合调试
- `AssertFast`：无附加信息，适合热路径中的简单条件检查

## 注意事项

- 断言失败会触发 `panic`，需由上层 `recover` 处理或让程序崩溃
- 生产环境建议关闭断言，避免意外 panic
- 断言用于检查**不可能发生的逻辑错误**，不用于校验用户输入
