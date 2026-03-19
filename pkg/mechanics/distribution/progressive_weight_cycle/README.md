# progressive_weight_cycle

渐进式解锁的权重周期分布机制，专为游戏场景中"奖励随进度逐步解锁"设计。

## 核心概念

一个大周期内有若干个特殊位置，每次触发时从候选池随机选出一个奖励。候选池随进度增长：

- `Item.JoinAt`：第几次特殊抽（0-based）起该奖励才进入候选池
- `Item.Quota`：该奖励在一个周期内最多出现几次，同时决定其在候选池中的相对权重

配额耗尽的奖励自动退出候选池，权重动态调整。

## 快速上手

```go
items := []progressive_weight_cycle.Item{
    {Quota: 4, JoinAt: 0}, // 基础奖励：随时可出，权重4
    {Quota: 2, JoinAt: 2}, // 稀有奖励：第3次起解锁，权重2
    {Quota: 1, JoinAt: 5}, // 史诗奖励：第6次起解锁，权重1
    {Quota: 1, JoinAt: 8}, // 终极奖励：仅最后一次可出
}

layer := progressive_weight_cycle.NewWeightCycleLayer(items)
total := progressive_weight_cycle.TotalQuota(items) // 4+2+1+1 = 8

// 每个玩家/对象独立维护一个 State
state := progressive_weight_cycle.NewState()

// 一个周期：连续触发 total 次特殊抽
for occIdx := range total {
    idx, err := layer.Generate(state, occIdx)
    if err != nil {
        // 正常不应出现，除非所有奖励配额已耗尽
        break
    }
    fmt.Printf("occIdx=%d → items[%d]\n", occIdx, idx)
}

// 周期结束后重置，开始下一周期
state.Reset()
```

## 多玩家场景

`Layer` 持有不可变规则，可被多个 `State` 复用：

```go
layer := progressive_weight_cycle.NewWeightCycleLayer(items)

// 每个玩家各一份 State，互不影响
states := make([]*progressive_weight_cycle.State, playerCount)
for i := range states {
    states[i] = progressive_weight_cycle.NewState()
}

// 各玩家独立推进
idx, err := layer.Generate(states[playerID], occIdx)
```

## API

| 函数/方法 | 说明 |
|----------|------|
| `NewWeightCycleLayer(items)` | 构造规则层（不可变，可复用） |
| `NewState()` | 构造运行时状态（每个玩家/对象各一份） |
| `layer.Generate(state, occIdx)` | 执行一次抽取，occIdx 为 0-based 当前序号 |
| `state.Reset()` | 重置状态，开始新周期 |
| `TotalQuota(items)` | 返回一个周期内特殊位置总数（= sum of Quota） |

## 与 tiered_cycle 的关系

本包是纯"渐进式解锁权重随机"的底层实现。[`tiered_cycle`](../tiered_cycle/) 是更完整的引擎，
在普通层 + 特殊保底双层场景中内部使用本包实现特殊层。
仅需纯渐进式随机时（无普通层、无自动排期）直接使用本包。

## 更多示例

见 `example_test.go` 中的 `ExampleLayer_rewardBox`（副本通关宝箱）和 `ExampleLayer_multipleStates`（多玩家场景）。
