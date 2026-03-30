# timecond

基于配置的时间条件判断：解析 `(Op, value)` 字符串，返回可复用的 `TimeCondition`，调用 `CheckNow` 判断给定时间戳是否满足条件。

## 快速上手

```go
// 1. 解析配置（通常在初始化阶段执行一次）
cond, err := timecond.Parse(timecond.OpGE, "2024-03-15 10:00:00")
if err != nil { ... }

// 2. 运行时判断（生产代码）
if cond.CheckNow(playerRegisterTs) {
    // 满足条件
}

// 3. 测试中精确控制当前时间
if cond.Check(playerRegisterTs, fixedNowMs) {
    // 满足条件
}
```

## Op 类型

| Op | 语义 | value 格式 |
|----|------|-----------|
| `OpAlways` | 无条件成立 | 忽略，传 `""` |
| `OpLT` | `subjectMs < absTs` | 时间字符串，见 `timeutil.Parse` |
| `OpGE` | `subjectMs >= absTs` | 时间字符串，见 `timeutil.Parse` |
| `OpWithin` | `lo <= subjectMs < hi`（左闭右开） | `"t1,t2"`，自动排序 |
| `OpRelLT` | `(nowMs - subjectMs) < relDur` | 时长字符串，见 `timeutil.ParseDuration` |
| `OpRelGE` | `(nowMs - subjectMs) >= relDur` | 时长字符串，见 `timeutil.ParseDuration` |

## API

### `ParseInLoc(op Op, value string, loc *time.Location) (*TimeCondition, error)`

解析配置字符串，以 `loc` 指定的时区解释绝对时间字符串，返回可复用的条件对象。
用于服务器时区与业务时区不一致的场景（如服务器在 UTC+0，业务使用 Asia/Shanghai）。
`OpRelLT` / `OpRelGE` 使用时长，`loc` 对其无影响。

### `Parse(op Op, value string) (*TimeCondition, error)`

解析同 `ParseInLoc`，以 `time.Local` 解释绝对时间字符串。语义见 `ParseInLoc`。
通常在服务启动或配置加载时调用一次，之后多次复用。

### `(*TimeCondition).CheckNow(subjectMs int64) bool`

以当前系统时间判断 `subjectMs` 是否满足条件。生产代码的首选调用方式。

### `(*TimeCondition).Check(subjectMs, nowMs int64) bool`

以指定的 `nowMs` 判断 `subjectMs` 是否满足条件。用于测试或需要精确控制当前时间的场景。

- `nowMs`：仅 `OpRelLT` / `OpRelGE` 使用，其他 Op 忽略

### `(*TimeCondition).Op() Op`

返回条件的运算符类型，可用于日志和调试。

## 边界说明

- `OpLT`：等于阈值时返回 `false`（严格小于）
- `OpGE`：等于阈值时返回 `true`（大于等于）
- `OpWithin`：左闭右开 `[lo, hi)`，等于 `lo` 满足，等于 `hi` 不满足
- `OpWithin` 的 `value` 中两个时间顺序填反时自动排序，不报错
- `OpRelLT/GE` 中 `nowMs` 参数对 `OpAlways/LT/GE/Within` 无影响，可传 `0`
- 当 `nowMs < subjectMs` 时经过时长为负数：`OpRelLT` 返回 `true`，`OpRelGE` 返回 `false`
- `TimeCondition` 零值（未经 `Parse` 初始化）的 `op` 为 0，对应 `OpAlways`，`Check` 始终返回 `true`；**始终通过 `Parse` 创建，不要使用零值**

## 业务层对接示例

业务层维护自身枚举到 `timecond.Op` 的映射，`timecond` 不感知业务常量：

```go
var opMap = map[int32]timecond.Op{
    ActvRegisterTsTriggerNo:    timecond.OpAlways,
    ActvRegisterTsTriggerLTAbs: timecond.OpLT,
    ActvRegisterTsTriggerGEAbs: timecond.OpGE,
    ActvRegisterTsTriggerLTRel: timecond.OpRelLT,
    ActvRegisterTsTriggerGERel: timecond.OpRelGE,
    ActvOpenTsTriggerGEAbs:     timecond.OpGE,
    ActvOpenTsTriggerLTAbs:     timecond.OpLT,
    ActvOpenTsTriggerWithinAbs: timecond.OpWithin,
}

op, ok := opMap[cfgType]
if !ok {
    return fmt.Errorf("unsupported type: %d", cfgType)
}
cond, err := timecond.Parse(op, cfgValue)
```
