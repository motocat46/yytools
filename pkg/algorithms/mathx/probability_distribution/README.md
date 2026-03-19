# probability_distribution

概率分布工具包，提供基于权重的随机选择实现。

## 提供的功能

### 工具函数（无需构造）

适用于一次性查找，或候选集动态变化无法预构建的场景：

```go
// 根据权重列表随机返回下标，时间复杂度 O(n)
idx := probability_distribution.CalcIndexByWeight([]int32{30, 50, 20}, 100)

// 根据 map 权重随机返回 key，时间复杂度 O(n)
key := probability_distribution.CalcKeyByWeight(map[string]int32{"SSR": 5, "SR": 25, "R": 70}, 100)
```

### NormalMethod：二分搜索实现

构建 O(n)，生成 O(log n)，适合中等规模且重复生成的场景：

```go
weights := []int32{30, 50, 20}
gen := probability_distribution.NewNormalMethod(weights)
idx := gen.Generate() // 返回 0/1/2，概率分别为 30%/50%/20%
```

### VoseAliasMethod：Vose 别名方法

构建 O(n)，生成 O(1)，适合高频生成场景：

```go
weights := []int32{30, 50, 20}
gen := probability_distribution.NewVoseAliasMethod(weights)
idx := gen.Generate() // 同上，但生成时间复杂度 O(1)
```

### 工厂函数

```go
gen := probability_distribution.ProbFactory(probability_distribution.Normal, weights)
gen := probability_distribution.ProbFactory(probability_distribution.VoseAlias, weights)
idx := gen.Generate()
```

## 选择指南

| 场景 | 推荐 |
|------|------|
| 候选集动态变化（每次不同） | `CalcIndexByWeight` / `CalcKeyByWeight` |
| 候选集固定，生成频次低 | `NormalMethod` |
| 候选集固定，高频生成（游戏帧循环等） | `VoseAliasMethod` |

## 特殊情况

- 所有权重均为 0：视为等概率，随机返回任意下标
- 单元素：直接返回 0，无随机开销
- 权重总和为 0 但元素非零（矛盾输入）：`assert` 触发 panic

## 动态权重

需要在运行时增删权重项时，参见同包的 `DynamicWeights`（`probability_distribution_dynamic.go`），
被 `progressive_weight_cycle` 内部使用。
