# tiered_cycle 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `tiered_cycle_engine_test.go` | Engine/State 生命周期、Config 校验、周期推进、ResetCycle、NextAutoReset |
| `correctness_test.go` | 并发正确性命题：Engine 共享无 race、相同种子序列独立 |
| `example_test.go` | 完整抽卡保底示例（可运行文档） |

## 分层执行命令

```bash
# 快速验证（单元 + 集成）
go test ./pkg/mechanics/distribution/tiered_cycle/

# 竞态检测
go test -race ./pkg/mechanics/distribution/tiered_cycle/

# 并发正确性命题（必须 -race）
go test -race -run TestCorrectness -v ./pkg/mechanics/distribution/tiered_cycle/

# 详细输出
go test -v ./pkg/mechanics/distribution/tiered_cycle/

# 指定子测试
go test -run TestEngine_ResetCycle ./pkg/mechanics/distribution/tiered_cycle/
```

## 关键测试场景

- `TestEngine_Config_Validation`：非法 Config（MinInterval 过大、Quota 为空等）返回 error；有特殊层但 r=nil 时 NewState 应 panic
- `TestEngine_FullCycle_CycleEnd`：周期末标记 CycleEnd=true
- `TestEngine_ResetCycle_StateCleared`：Reset 后 posInCycle 归零、State 清空
- `TestEngine_FullCycle_JoinAtEnforced`：JoinAt 约束在特殊层按序号正确解锁
- `TestEngine_FullCycle_QuotaEnforced`：Quota 每周期不超额
- `TestEngine_Replay`：相同种子产生完全相同的 plan 位置（可重放性）
- `TestCorrectness_EngineShared_StateIsolated_NoDataRace`：多 goroutine 共享 Engine，各持独立 State，-race 无 data race
- `TestCorrectness_SameSeed_SameSequence`：相同种子并发运行产生相同序列（互不干扰）

## 已知局限

- 统计分布验证（特殊位置均匀性、MinInterval 概率合规性）尚未覆盖
- 随机混合操作测试数据量未达到 10 万级别
