# timeutil 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `time_test.go` | `ParseDuration`：标准格式（秒/分/小时/组合）、天数扩展格式（1d/负数天/天+标准组合）、错误格式（无效输入/天数越界/非法字符）；边界：零值变体（"0"/"0s"/"0d"/""）、10000d 大数值；性能：1000 次×8 种格式的正确性验证 |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/infra/timeutil/

# 竞态检测
go test -race ./pkg/infra/timeutil/
```

## 注意

- `ParseDuration` 在标准 `time.ParseDuration` 基础上扩展了 `d`（天）单位支持
- 字符串长度硬限制（> 100 字符报错），防止异常输入
- 溢出检测：1000000d 因超出 `time.Duration`（int64，纳秒）范围会返回错误
- 无基准测试——字符串解析热路径性能由调用方场景决定
