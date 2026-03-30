# timeutil 使用文档

对 `time.ParseDuration` 的扩展，增加对 `d`（天）单位的支持。

## API

### `ParseDuration(s string) (time.Duration, error)`

解析时间字符串，支持标准库的所有单位（ns、us、ms、s、m、h），额外支持 `d`（天）。

**支持的格式：**

| 输入 | 含义 |
|------|------|
| `""` 或 `"0"` | 0 |
| `"1d"` | 24 小时 |
| `"2d12h"` | 2 天 12 小时 |
| `"-1d"` | 负 1 天 |
| `"-2d6h30m"` | 负 2 天 6 小时 30 分 |
| `"1h30m"` | 标准库格式，直接透传 |

**限制：**
- 字符串长度不超过 100 字符
- 最大时长约 290 年
- `+`/`-` 符号只允许出现在首位
- 大小写不敏感（`1D` 等同于 `1d`）

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/infra/timeutil"

// 解析天数
d, err := timeutil.ParseDuration("7d")       // 7 * 24 * time.Hour
d, err  = timeutil.ParseDuration("1d12h")    // 36 * time.Hour
d, err  = timeutil.ParseDuration("-3d")      // 负 3 天

// 读取配置文件中的过期时间
expireDuration, err := timeutil.ParseDuration(config.TokenExpire)
if err != nil {
    // 格式不合法
}
expireAt := time.Now().Add(expireDuration)
```

## 与标准库的区别

标准库 `time.ParseDuration` 最大单位是 `h`（小时），配置中写 `168h` 表示一周不够直观。本函数支持 `7d`，可读性更好。

---

## Parse / ParseInLoc — 日期时间字符串解析

### `ParseInLoc(s string, loc *time.Location) (time.Time, error)`

解析日期或日期时间字符串，以 `loc` 指定的时区返回 `time.Time`。
用于服务器时区与业务时区不一致的场景（如服务器在 UTC+0，业务使用 Asia/Shanghai）。

### `Parse(s string) (time.Time, error)`

解析同 `ParseInLoc`，时区为 `time.Local`。语义见 `ParseInLoc`。

**支持的格式：**

| 类型 | 示例 |
|------|------|
| 纯日期 | `"2024-01-05"`、`"2024/1/5"` |
| 日期时间（含秒）| `"2024-01-05 09:05:03"`、`"2024/1/5 9:5:3"` |
| 日期时间（省秒）| `"2024-01-05 09:05"`、`"2024/1/5 9:5"` |

**规则：**
- 分隔符 `-` 或 `/` 均可，不允许在同一字符串中混用
- 月、日、时、分、秒可省略前导零（只能 1 位或 2 位）
- 输入前后空白和中间多余空白自动处理
- 省略秒时默认为 `0`

### `ParseUnixMilliInLoc(s string, loc *time.Location) (int64, error)`

解析同 `ParseInLoc`，成功时返回 Unix 毫秒时间戳。语义见 `ParseInLoc`。

### `ParseUnixMilli(s string) (int64, error)`

解析同 `Parse`，成功时返回 Unix 毫秒时间戳。非法输入返回 `0, err`。

### 使用示例

```go
import (
    "time"
    "github.com/motocat46/yytools/pkg/infra/timeutil"
)

// 业务时区（如服务器在 UTC+0，业务使用上海时区）
shanghai, _ := time.LoadLocation("Asia/Shanghai")
t, err := timeutil.ParseInLoc("2024-03-15 10:00:00", shanghai)
// → 2024-03-15 10:00:00 CST（UTC 02:00:00）

// 默认 time.Local
t, err = timeutil.Parse("2024-1-5")         // 2024-01-05 00:00:00 Local
t, err = timeutil.Parse("2024-1-5 9:5:3")  // 2024-01-05 09:05:03 Local
t, err = timeutil.Parse("2024-01-05 9:5")  // 2024-01-05 09:05:00 Local

// 毫秒时间戳
ms, err := timeutil.ParseUnixMilli("2024-01-05 09:05:03")
ms, err  = timeutil.ParseUnixMilliInLoc("2024-01-05 09:05:03", shanghai)
```

**不支持：** 纯时间字符串（`"09:05:03"`）、非年月日顺序、毫秒/微秒精度。
持续时间场景请用 `ParseDuration`。

---

## 日历边界与时间比较工具

### 边界计算

#### `time.Time` 变体（使用 `t.Location()`）

| 函数 | 说明 |
|------|------|
| `StartOfDay(t)` | t 所在时区当天 00:00:00 |
| `StartOfTomorrow(t)` | t 所在时区明天 00:00:00 |
| `StartOfMonth(t)` | t 所在时区当月 1 日 00:00:00 |
| `StartOfWeekday(t, weekday)` | t 所在时区本周指定周几 00:00:00（ISO 周，周一为第一天；目标即当天时返回今天 00:00:00） |
| `StartOfNextWeekday(t, weekday)` | t 所在时区下周指定周几 00:00:00 |
| `StartOfNextMonthDay(t, day)` | t 所在时区下月第 day 日 00:00:00；day < 1 时 clamp 至 1，day 超出月末时 clamp 至月末（如 4 月传 31 → 4 月 30 日） |

#### `int64` 毫秒时间戳变体（显式传入 `*time.Location`）

函数名在上表基础上加 `Ms` 后缀，参数 `ms int64` 替换 `t time.Time`，末尾追加 `loc *time.Location`，返回值为 `int64`（毫秒时间戳）。

| 函数 | 签名 |
|------|------|
| `StartOfDayMs` | `(ms int64, loc *time.Location) int64` |
| `StartOfTomorrowMs` | `(ms int64, loc *time.Location) int64` |
| `StartOfMonthMs` | `(ms int64, loc *time.Location) int64` |
| `StartOfWeekdayMs` | `(ms int64, weekday time.Weekday, loc *time.Location) int64` |
| `StartOfNextWeekdayMs` | `(ms int64, weekday time.Weekday, loc *time.Location) int64` |
| `StartOfNextMonthDayMs` | `(ms int64, day int, loc *time.Location) int64` |

### 时间比较

| 函数 | 签名 | 说明 |
|------|------|------|
| `IsSameDay` | `(a, b time.Time, loc *time.Location) bool` | a、b 在 loc 时区下是否同一天 |
| `IsSameWeek` | `(a, b time.Time, loc *time.Location) bool` | a、b 在 loc 时区下是否同一 ISO 周 |
| `DaysBetween` | `(a, b time.Time, loc *time.Location) int` | a→b 的日历天数（b>a 为正，DST 安全） |
| `IsSameDayMs` | `(a, b int64, loc *time.Location) bool` | 同 IsSameDay，毫秒时间戳入参 |
| `IsSameWeekMs` | `(a, b int64, loc *time.Location) bool` | 同 IsSameWeek，毫秒时间戳入参 |
| `DaysBetweenMs` | `(a, b int64, loc *time.Location) int` | 同 DaysBetween，毫秒时间戳入参 |

> **注意：** 比较函数和 `Ms` 变体均要求显式传入 `*time.Location`，不依赖机器 Local——这对全球同服游戏、多时区数据处理等场景尤为重要。

### 使用示例

```go
import (
    "time"
    "github.com/motocat46/yytools/pkg/infra/timeutil"
)

cst := time.FixedZone("CST", 8*60*60) // UTC+8
now := time.Now().In(cst)

// 边界计算
sod   := timeutil.StartOfDay(now)               // 今天 00:00:00 CST
tom   := timeutil.StartOfTomorrow(now)          // 明天 00:00:00 CST
mon   := timeutil.StartOfWeekday(now, time.Monday)     // 本周一 00:00:00 CST
nxtMon := timeutil.StartOfNextWeekday(now, time.Monday) // 下周一 00:00:00 CST
next15 := timeutil.StartOfNextMonthDay(now, 15)         // 下月 15 日 00:00:00 CST

// Ms 变体（输入/输出均为毫秒时间戳）
ms    := now.UnixMilli()
sodMs := timeutil.StartOfDayMs(ms, cst)                   // 今天 00:00:00 CST（毫秒）
monMs := timeutil.StartOfWeekdayMs(ms, time.Monday, cst)  // 本周一 00:00:00 CST（毫秒）

// 时间比较
a, b := now, now.Add(2*24*time.Hour)
timeutil.IsSameDay(a, b, cst)      // false
timeutil.IsSameWeek(a, b, cst)     // 取决于 now 是否在周末附近
timeutil.DaysBetween(a, b, cst)    // 2
```
