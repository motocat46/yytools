# timeutil 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `time_test.go` | `ParseDuration`：标准格式（秒/分/小时/组合）、天数扩展格式（1d/负数天/天+标准组合）、错误格式（无效输入/天数越界/非法字符）；边界：零值变体（"0"/"0s"/"0d"/""）、10000d 大数值；性能：1000 次×8 种格式的正确性验证 |
| `parse_test.go` | `normalizeDate`/`normalizeTime` 单元测试（正常路径 + 全错误路径）；`Parse` 合法输入（10 组格式变体）、非法输入（19 条清单）、时区验证；`ParseUnixMilli` 合法/非法各 7+ 组；跨格式一致性（3 组 × 4 变体）；边界值（月末/闰年/时间极值，含具体值断言）；Roundtrip 100,000 组；RandomVariants 100,000 组；基准测试 5 组 |
| `calendar_test.go` | `StartOfDay/Tomorrow/Month`：正常、边界（月末/年末/已是0点）、UTC 时区；`StartOfNextMonthDay`：含上界 clamp（4月→30、2月非闰年→28、闰年→29）和下界 clamp（day=0/day=-3→1日）；`StartOfWeekday/NextWeekday`：周四起算各方向、当天命中、跨月；Ms 变体（6 个）：硬编码绝对时间戳验证（独立于 time.Time 变体）、CST/UTC 时区差异验证；`IsSameDay/Week`：同天/不同天、跨时区、跨年 ISO 周；`IsSameDayMs/IsSameWeekMs/DaysBetweenMs`：正/负/零、UTC、table-driven；`DaysBetween`：正/负/零、跨月、跨年、UTC、跨时区输入 |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/infra/timeutil/

# 竞态检测
go test -race ./pkg/infra/timeutil/

# 基准测试（Parse）
go test -bench=BenchmarkParse -benchtime=3s -benchmem ./pkg/infra/timeutil/

# 基准测试（ParseDuration）
go test -bench=BenchmarkParseDuration -benchtime=3s -benchmem ./pkg/infra/timeutil/
```

## 注意

- `ParseDuration` 在标准 `time.ParseDuration` 基础上扩展了 `d`（天）单位支持
- 字符串长度硬限制（> 100 字符报错），防止异常输入
- 溢出检测：1000000d 因超出 `time.Duration`（int64，纳秒）范围会返回错误
- 基准测试见下方基线数字
- 日历工具测试使用固定时区（CST UTC+8 和 time.UTC），不依赖机器 Local，保证跨环境确定性

---

## ParseDuration 基准基线（2026-03-26）

**环境：** darwin arm64（Apple M4），`-benchtime=3s`

| 基准 | ns/op | B/op | allocs/op |
|------|-------|------|-----------|
| BenchmarkParseDuration_NoDay | 47 | 0 | 0 |
| BenchmarkParseDuration_WithDay | 47 | 0 | 0 |
| BenchmarkParseDuration_NegativeDay | 31 | 0 | 0 |

---

## Parse / ParseUnixMilli 基准基线（2026-03-26）

**环境：** darwin arm64（Apple M4），`-benchtime=3s`

| 基准 | ns/op | B/op | allocs/op |
|------|-------|------|-----------|
| BenchmarkParse_Date | 128 | 80 | 3 |
| BenchmarkParse_DateShort | 150 | 80 | 5 |
| BenchmarkParse_DateTime | 268 | 176 | 6 |
| BenchmarkParse_DateTimeShort | 319 | 224 | 12 |
| BenchmarkParseUnixMilli | 266 | 176 | 6 |
