# timecond 测试说明

## 运行方式

```bash
# 标准运行（绕过缓存）
go test -count=1 ./pkg/infra/timecond/...

# 详细输出
go test -count=1 -v ./pkg/infra/timecond/...
```

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `timecond_test.go` | Parse 错误路径、各 Op 的 Check 边界、nowMs 对绝对条件无影响、OpWithin 退化区间、OpRelLT/GE 负数经过时长、Op.String()、Op() 访问器 |

## 关键边界场景

| 场景 | 测试函数 |
|------|---------|
| Parse 所有错误路径（非法 Op、非法时间、非法时长、OpWithin 格式错误） | `TestParse_Error` |
| OpAlways 忽略 value | `TestParse_OpAlways_ValueIgnored` |
| OpLT/GE 等于阈值时的边界行为 | `TestCheck_OpLT`、`TestCheck_OpGE` |
| OpWithin 左闭右开边界 | `TestCheck_OpWithin` |
| OpWithin t1/t2 顺序填反时自动排序 | `TestCheck_OpWithin_AutoSort` |
| OpWithin 退化区间（lo == hi，永远 false） | `TestCheck_OpWithin_Degenerate` |
| nowMs 不影响绝对条件（OpLT/GE/Within） | `TestCheck_AbsOp_NowMsIgnored` |
| OpRelLT/GE 等于时长时的边界行为 | `TestCheck_OpRelLT`、`TestCheck_OpRelGE` |
| OpRelLT/GE nowMs < subjectMs（负数经过时长） | `TestCheck_Rel_NegativeElapsed` |
