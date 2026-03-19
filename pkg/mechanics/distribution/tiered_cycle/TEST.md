# tiered_cycle 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `tiered_cycle_engine_test.go` | Engine/State 生命周期、Config 校验、周期推进、ResetCycle、NextAutoReset |
| `example_test.go` | 完整抽卡保底示例（可运行文档） |

## 分层执行命令

```bash
# 快速验证（单元 + 集成）
go test ./pkg/mechanics/distribution/tiered_cycle/

# 竞态检测
go test -race ./pkg/mechanics/distribution/tiered_cycle/

# 详细输出
go test -v ./pkg/mechanics/distribution/tiered_cycle/

# 指定子测试
go test -run TestEngine_ResetCycle ./pkg/mechanics/distribution/tiered_cycle/
```

## 关键测试场景

- `TestEngine_ConfigValidation`：非法 Config（MinInterval 过大、Quota 为空等）返回 error
- `TestEngine_CycleEnd`：周期末标记 CycleEnd=true
- `TestEngine_ResetCycle`：Reset 后 posInCycle 归零、State 清空
- `TestEngine_JoinAt`：JoinAt 约束在特殊层按序号正确解锁
- `TestEngine_MinInterval`：特殊位置间隔满足 MinInterval 约束
- `TestEngine_MultiCycle`：多周期下 sum(Quota) 累计正确

## 已知局限

- 统计分布验证（特殊位置均匀性、MinInterval 概率合规性）尚未覆盖
- 随机混合操作测试数据量未达到 10 万级别
