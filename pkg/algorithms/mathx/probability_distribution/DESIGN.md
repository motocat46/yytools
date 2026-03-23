# probability_distribution 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、两种离散分布实现：NormalMethod vs VoseAliasMethod

| 维度 | NormalMethod | VoseAliasMethod |
|------|-------------|-----------------|
| 初始化 | O(n) 构建前缀和 | O(n) 构建概率表 + 别名表 |
| 生成 | O(log n)（二分搜索）| O(1)（两次随机 + 查表）|
| 空间 | O(n) | O(n) |
| 权重更新 | 需要重建，O(n) | 需要重建，O(n) |
| 适用场景 | 中等频率生成、权重偶尔更新 | 极高频率生成、权重固定 |

两种实现均通过 `ProbFactory` 工厂统一构造，调用方无需感知内部差异。

**为什么保留两种实现？**

VoseAlias 生成 O(1) 更优，但初始化代码更复杂、调试更难。NormalMethod 代码更直观，适合权重频繁变化（每次变化都需重建）或单次使用的场景，调试友好。

---

## 二、NormalMethod 的二分搜索优化

朴素实现（`CalcIndexByWeight`）线性遍历，O(n)。`NormalMethod` 预计算**前缀和数组**，生成时用二分搜索找"随机值落入哪个区间"，O(log n)。

关键：使用左边界二分——找满足 `target ≤ weightsSum[i]` 的最小 i。

**全零权重**的特殊情形：前缀和全为 0，退化为等概率随机（直接随机下标），不 panic。

---

## 三、VoseAliasMethod 的别名构造

核心思路（Vose 1991）：将每个元素的概率等比放大 n 倍（n = 元素数），使平均概率恰为 1。然后通过"大池补小池"将每个格子恰好填满概率 1，同时记录别名。生成时只需两次随机：

1. 随机选一个格子下标 i
2. 随机浮点数 p ∈ [0, 1)：若 `p < prob[i]` 返回 i，否则返回 `alias[i]`

数值稳定性：更新剩余概率时用 `prob[g] = prob[g] + prob[l] - 1`（等价于 `prob[g] - (1 - prob[l])`），减少浮点累积误差。

---

## 四、DynamicWeights 的设计

`DynamicWeights` 解决的问题：从一组带权重的 key 中**不重复地**按权重抽取，每次抽取后对应权重减少，直到耗尽。

不复用 NormalMethod/VoseAlias 的原因：这两者面向静态权重，每次抽取后需整体重建 O(n)。DynamicWeights 直接维护权重 map，每次 Generate 仅线性遍历 + 更新当前 key 权重，虽然单次 O(n)，但总轮数受总权重限制，整体行为更可预测。
