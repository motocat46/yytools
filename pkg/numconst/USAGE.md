# numconst 使用文档

时间相关的整型常量，基准单位为**毫秒**。避免在代码中散落魔法数字。

## 常量列表

```go
MILLISECOND = 1          // 1 毫秒
SECOND      = 1_000      // 1 秒
MINUTE      = 60_000     // 1 分钟
HOUR        = 3_600_000  // 1 小时
DAY         = 86_400_000 // 1 天
WEEK        = 604_800_000// 1 周
```

类型别名：`type TimeUnit = int64`

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/numconst"

// 计算过期时间戳（毫秒）
expireAt := time.Now().UnixMilli() + 7*numconst.DAY

// 判断是否过期
if time.Now().UnixMilli()-createAt > 30*numconst.MINUTE {
    // 超过 30 分钟
}

// 与 time.Duration 互换
duration := time.Duration(numconst.HOUR) * time.Millisecond
```

## 注意事项

- 单位是**毫秒**，不是 `time.Duration`（纳秒）
- 与标准库互转：`time.Duration(n * numconst.SECOND) * time.Millisecond`
- 适合存储时间戳差值、配置过期时长等业务场景
