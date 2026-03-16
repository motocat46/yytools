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
