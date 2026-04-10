# tiered_cycle 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `tiered_cycle_engine_test.go` | Engine/State 生命周期、Config 校验、周期推进、ResetCycle、NextAutoReset |
| `correctness_test.go` | 并发正确性命题：Engine 共享无 race、相同种子序列独立 |
| `largescale_test.go` | 大规模不变量：100k 抽取验证 plan 约束和 Quota 约束；1M 压力测试 |
| `example_test.go` | 完整抽卡保底示例（可运行文档） |

## 分层执行命令

```bash
# 快速验证（单元 + 集成，跳过压力测试）
go test -short ./pkg/mechanics/distribution/tiered_cycle/

# 完整验证（含 1M 压力测试，约 0.1 秒）
go test -count=1 ./pkg/mechanics/distribution/tiered_cycle/

# 竞态检测
go test -race ./pkg/mechanics/distribution/tiered_cycle/

# 并发正确性命题（必须 -race）
go test -race -run TestCorrectness -v ./pkg/mechanics/distribution/tiered_cycle/

# 大规模不变量（100k 抽取）
go test -count=1 -run TestEngine_LargeScale -v ./pkg/mechanics/distribution/tiered_cycle/

# 压力测试（1M 抽取）
go test -count=1 -run TestEngine_Stress -v ./pkg/mechanics/distribution/tiered_cycle/

# 详细输出
go test -v ./pkg/mechanics/distribution/tiered_cycle/
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
- `TestEngine_LargeScale_CycleInvariants`：1000 周期（100k 抽取），每轮验证 plan 排序/MinInterval/Quota 约束
- `TestEngine_Stress_CycleInvariants`：10000 周期（1M 抽取），同上，-short 跳过

## 已知局限

- 统计分布验证（特殊位置均匀性、MinInterval 概率合规性）尚未覆盖
