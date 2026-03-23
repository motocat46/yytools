# progressive_weight_cycle 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md，概念说明见 doc.go。

---

## 一、V1 → V2：为什么引入 DynamicWeights

### V1：每次 Generate 重新构建权重表

```go
// V1 核心逻辑（specialCycleCore）
for i, item := range items {
    diff := item.Quota - used[int32(i)]
    if specialOccIdx >= item.JoinAt && diff > 0 {
        weightMap[i] = diff
        totalWeight += diff
    }
}
selectedIndex = pd.CalcKeyByWeight(weightMap, totalWeight)
```

每次调用都要遍历所有 items，重新构建 `weightMap` 并计算 `totalWeight`。时间复杂度 O(n)，但有一个隐性开销：每次都分配新的 `weightMap`（map 分配不便宜）。

### V2：增量维护 DynamicWeights

```go
// V2 核心逻辑（specialCycleCoreV2）
for i, item := range items {
    if specialOccIdx >= item.JoinAt && !unlocked[i] {
        unlocked[i] = true
        dw.AddWeight(i, item.Quota)  // 只在解锁时加入，不重复
    }
}
return dw.Generate(), nil
```

`DynamicWeights` 内部维护总权重和权重映射，`Generate` 直接使用。每次调用只需扫描"尚未解锁"的候选项（随着周期推进，未解锁项越来越少），配额耗尽时 `DynamicWeights` 自动移除该项。

**关键优势**：解锁是单向的（只解锁，不撤销），`Unlocked` map 确保每个奖励最多加入一次，避免重复 AddWeight。

### 为什么保留 V1？

`specialCycleCore`（V1）保留在代码中，标注为"保留用于对比测试"。对比测试意义：对随机性算法，V2 必须与 V1 产生统计上等价的结果分布，单独跑随机混合测试时用 V1 作参考模型。

---

## 二、JoinAt 语义：为什么是"特殊出现序号"而非"普通抽计数器"

`JoinAt` 的单位是**当前周期内第几次特殊出现**（0-based `occIdx`），不是总普通抽次数或绝对计数器。

**原因**：`progressive_weight_cycle` 是纯粹的"渐进式权重随机"层，它不感知"普通层"的存在，也不关心普通抽发生了多少次。调用方负责维护 `occIdx`，可以将任何进度信号（抽数、时间、关卡）映射到 `occIdx` 后传入。这让本包对调用方解耦，不绑定特定的进度计数语义。

在 `tiered_cycle` 中，`tiered_cycle` 内部负责维护 `occIdx`（每触发一次特殊位置就递增），对外完全屏蔽这个细节。

---

## 三、Layer / State 分离

```
Layer（不可变规则）          State（运行时状态）
├── Items []Item         ├── Dw *DynamicWeights
└── TotalQuota int32     └── Unlocked map[int]bool
```

**Layer 是无状态的规则描述**，一经构造不再变化。同一套规则可被多个 State（多个玩家）复用，无并发问题。

**State 是每个独立对象的运行时快照**，包含：
- `Dw`：当前候选池及各奖励剩余配额
- `Unlocked`：已解锁（已加入 DW）的奖励下标集合，防止重复加入

周期结束时只需重置 State，Layer 不需要任何操作。

---

## 四、为什么 State 不直接存 `used map[int32]int32`（V1 方案）

V1 用 `used` map 记录每个奖励已出现次数，每次 Generate 时计算 `Quota - used[i]` 得到剩余权重。

V2 改为直接让 `DynamicWeights` 管理剩余权重：`AddWeight(i, item.Quota)` 后，每次 `Generate` 内部自动扣减权重，配额归零时自动移除。这样 State 不需要再维护 `used` map，减少了一层间接状态，逻辑更集中。

---

## 五、与 tiered_cycle 的边界

| 包 | 职责 | 适用场景 |
|----|------|---------|
| `progressive_weight_cycle` | 纯渐进式权重随机，无层级概念 | 只需特殊分布，不需要普通层 |
| `tiered_cycle` | 双层引擎（普通层 + 特殊保底），内部用本包实现特殊层 | 完整的"普通抽 + 保底"场景 |

本包不引用 `tiered_cycle`，依赖方向是单向的：`tiered_cycle → progressive_weight_cycle`。
