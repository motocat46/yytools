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

## Parse / ParseUnixMilli — 日期时间字符串解析

### `Parse(s string) (time.Time, error)`

解析日期或日期时间字符串，返回 `time.Time`（时区为 `time.Local`）。

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
- 返回时区为 `time.Local`；需要 UTC 时由调用方用 `t.UTC()` 转换

### `ParseUnixMilli(s string) (int64, error)`

与 `Parse` 相同，成功时返回 Unix 毫秒时间戳（`int64`）。非法输入返回 `0, err`。

### 使用示例

```go
import "github.com/motocat46/yytools/pkg/infra/timeutil"

// 纯日期
t, err := timeutil.Parse("2024-1-5")         // 2024-01-05 00:00:00 Local
t, err  = timeutil.Parse("2024/01/05")       // 同上

// 日期时间
t, err  = timeutil.Parse("2024-1-5 9:5:3")  // 2024-01-05 09:05:03 Local
t, err  = timeutil.Parse("2024-01-05 9:5")  // 2024-01-05 09:05:00 Local

// 毫秒时间戳
ms, err := timeutil.ParseUnixMilli("2024-01-05 09:05:03")

// 需要 UTC 时由调用方转换
t, err = timeutil.Parse("2024-01-05")
utc   := t.UTC()
```

**不支持：** 纯时间字符串（`"09:05:03"`）、非年月日顺序、毫秒/微秒精度。
持续时间场景请用 `ParseDuration`。
