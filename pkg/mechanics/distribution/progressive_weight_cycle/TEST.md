# progressive_weight_cycle 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `progressive_weight_cycle_test.go` | TotalQuota、State Reset、JoinAt 解锁约束、Quota 耗尽、V1/V2 行为一致性 |
| `example_test.go` | 副本宝箱场景、多玩家独立 State 场景（可运行文档） |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/mechanics/distribution/progressive_weight_cycle/

# 竞态检测
go test -race ./pkg/mechanics/distribution/progressive_weight_cycle/

# 详细输出
go test -v ./pkg/mechanics/distribution/progressive_weight_cycle/
```

## 关键测试场景

- `TestTotalQuota`：空/单/多 Item 的 Quota 累加正确
- `TestState_Reset`：Reset 后 Dw 和 Unlocked 归零
- `TestLayer_Generate_JoinAt`：各 occIdx 阶段候选池正确（V2 实现）
- `TestLayer_Generate_QuotaExhausted`：Quota 耗尽后正确返回 error
- `TestV1V2Consistency`：V1（每次重建）与 V2（动态权重）行为完全一致

## 已知局限

- 暂无大规模随机混合测试（测试规模在 5-20 次，低于项目 10 万规范）
