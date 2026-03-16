# numconst 使用文档

常用整型常量，避免在代码中散落魔法数字。

## 数量级常量

```go
Thousand        = 1_000          // 千
TenThousand     = 10_000         // 万
HundredThousand = 100_000        // 十万
Million         = 1_000_000      // 百万
TenMillion      = 10_000_000     // 千万
HundredMillion  = 100_000_000    // 亿
Billion         = 1_000_000_000  // 十亿
```

```go
if score > 1*numconst.HundredMillion { ... }
var maxGold int64 = 500 * numconst.TenThousand
```

## 存储单位常量（1024 进制）

```go
KB = 1024
MB = 1024 * KB
GB = 1024 * MB
TB = 1024 * GB
```

```go
const maxUpload = 10 * numconst.MB
buf := make([]byte, 4*numconst.KB)
```

## 时间常量（基准单位：毫秒）

```go
MILLISECOND = 1           // 1 毫秒
SECOND      = 1_000       // 1 秒
MINUTE      = 60_000      // 1 分钟
HOUR        = 3_600_000   // 1 小时
DAY         = 86_400_000  // 1 天
WEEK        = 604_800_000 // 1 周
```

类型别名：`type TimeUnit = int64`

```go
// 计算过期时间戳（毫秒）
expireAt := time.Now().UnixMilli() + 7*numconst.DAY

// 判断是否超时
if time.Now().UnixMilli()-createAt > 30*numconst.MINUTE { ... }

// 与 time.Duration 互换
duration := time.Duration(numconst.HOUR) * time.Millisecond
```

## 注意事项

- 时间常量单位是**毫秒**，不是 `time.Duration`（纳秒）
- 与标准库互转：`time.Duration(n * numconst.SECOND) * time.Millisecond`
