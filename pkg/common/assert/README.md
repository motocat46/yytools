# assert 使用文档

运行时断言框架，用于表达"绝不应该出现的情况"。断言失败时触发 `panic`，触发即意味着调用方存在 bug，需在开发期修复。

**断言不可关闭，始终生效。**

## API

| 函数 | 说明 |
|------|------|
| `Assert(condition bool, list ...interface{})` | 条件为 false 时 panic，附带自定义消息 |
| `AssertFast(cond bool)` | 条件为 false 时 panic，无附加消息 |

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/common/assert"

// 带消息的断言
assert.Assert(n > 0, "n must be positive, got:", n)

// 快速断言（无消息）
assert.AssertFast(ptr != nil)
```

## Assert vs AssertFast

- `Assert`：可附带多个参数作为错误描述，panic 信息更详细，适合调试
- `AssertFast`：无附加信息，适合条件简单、意图自明的场景

## 设计原则

Assert 是 `if !condition { panic(...) }` 的语法糖，核心价值在于：
- 正向表达契约（"这里必须满足 X"），比取反的 `if` 更符合直觉
- 统一风格，代码中所有"不该发生"的检查一眼可辨

**何时用 Assert：** 调用方违反契约时（如空栈出栈、负数索引），触发意味着调用方有 bug。

**何时用普通 panic：** 运行时边界保护（如参数超出类型范围、时间戳溢出），语义上属于运行时防护而非契约断言时，直接用 `panic` 更明确。

## 注意事项

- 断言失败会触发 `panic`，若有 `recover` 中间件，确保同时记录 `debug.Stack()` 以获取完整调用链
- 不要用于校验用户输入或外部数据——那是错误处理，应返回 `error`
