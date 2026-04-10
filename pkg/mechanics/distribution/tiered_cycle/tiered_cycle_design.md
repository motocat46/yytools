# tiered_cycle 设计决策记录

作者: yangyuan
日期: 2026-03
模块: pkg/mechanics/distribution/tiered_cycle

⸻

## 一、背景

tiered_cycle 是一个用于实现 分层周期分布机制 的组件。

典型使用场景包括：
•	抽奖系统
•	掉落系统
•	刷怪/事件触发
•	周期性奖励控制

核心特点：
•	在一个周期（cycle）内
•	不同层（Layer）按照规则触发
•	每一层有自己的算法与状态

当前实现包含两层：
1.	Standard Layer（普通层）
•	普通权重随机分布
2.	Special Layer（特殊层）
•	带约束的特殊奖励分布
•	支持 quota / joinAt / interval 等规则

⸻

## 二、核心设计目标

在设计 tiered_cycle 时，主要考虑以下目标：
1.	易用性
2.	稳定性
3.	可测试性
4.	可维护性

而不是一开始就追求：
•	任意层数
•	完全可插拔机制
•	框架级扩展能力

⸻

## 三、当前架构（v1）

当前版本采用 固定两层架构：

TieredEngine
├── StandardLayer
└── SpecialLayer

状态结构：

State
├── posInCycle
├── StandardLayerState
└── SpecialLayerState

职责划分：

Engine 负责
•	周期推进（posInCycle）
•	调度触发哪个 Layer
•	周期结束判断
•	调用 Layer 的算法

Layer 负责
•	自身算法
•	自身配置解释
•	自身运行时状态
•	重置逻辑

这种结构的特点是：
•	Engine 不关心 Layer 内部实现
•	Layer 独立演进

⸻

## 四、为什么没有做“可插拔 Layer”

在设计过程中曾讨论过：

是否允许用户自由替换 Layer 实现？

即：

Engine
├── Layer interface
├── CustomLayerA
├── CustomLayerB

但最终决定 暂不实现。

原因如下。

⸻

1 状态结构会变复杂

当前 State 结构是：

State
├── StandardLayerState
└── SpecialLayerState

如果 Layer 可以自由替换：
•	不同 Layer 会有不同 State
•	Engine 就不能固定 State 结构

必须改成：

State
├── layerStates []any

或者：

map[string]any

这会带来问题：
•	类型安全下降
•	调试困难
•	API 复杂度上升

⸻

2 需要解决状态序列化问题

公共库必须考虑：
•	state 如何持久化
•	state 如何升级
•	如何做版本迁移
•	如何做问题回放

这些问题在 v1 中是完全不需要处理的。

⸻

3 没有第二种实现需求

目前只有一种 StandardLayer 和 SpecialLayer 实现。

如果现在就做：
•	Layer 接口
•	可插拔 State
•	自定义调度

属于 基于假设的设计。

工程实践表明：

没有第二个真实实现时，很难设计出正确抽象。

⸻

## 五、设计决策

最终确定：

v1 采用固定两层结构

优点：
•	API 简单
•	易于理解
•	易于测试
•	使用成本低

当前设计：

Engine
├── StandardLayer
└── SpecialLayer

State：

State
├── StandardLayerState
└── SpecialLayerState


⸻

## 六、未来演进策略

如果未来出现以下情况：
1.	出现第二种 StandardLayer 实现
2.	出现新的特殊层算法
3.	Layer 需要不同 State 结构

则可以引入：

EngineV2

新的架构可能是：

EngineV2
├── []Layer
└── []LayerState

Layer 负责：

Layer
├── NewState()
├── Reset()
└── Generate()

Engine 只负责：
•	调度
•	周期推进
•	触发 Layer

这种模式属于：

可插拔 Layer + 可插拔 State

但这属于 架构升级，而不是简单重构。

因此：

v2 应在真实需求出现时再实现。

⸻

## 七、设计原则总结

本模块遵循以下原则：

1 先做稳定工具，再做通用框架

v1 的目标是：

一个可靠、稳定、简单的工具

而不是一个高度抽象的框架。

⸻

2 没有第二实现，不做通用抽象

只有当出现：
•	第二种 Layer
•	第二种 State 结构

才值得抽象接口。

⸻

3 优先保证 API 稳定

公共库的第一目标是：

稳定

而不是：

可扩展

⸻

## 八、一句话总结

tiered_cycle 当前版本选择：

优先提供一个稳定好用的工具（v1），而不是过早设计可插拔框架。

当真实需求出现时，再通过 EngineV2 引入更通用的架构。

⸻

## 九、rand 归属决策（2026-04）

### 问题

v1 原始设计将 `*rand.Rand` 存入 `Engine.r`，并在 `ConfigBase.R` 接收。
`Engine` 的设计注释已声明"持有不可变规则"，但 `e.r` 是可变状态，与该不变量矛盾。
多 goroutine 共享同一 `Engine` 并发调用 `ResetCycle` 时会产生 data race。

### 决策

**rand 移入 `State`，通过 `engine.NewState(r)` 工厂方法创建。**

判断依据：

- rand 用于生成"本玩家本轮特殊计划"，计划本身是玩家状态，生成工具（rand）也属于玩家状态
- 每个玩家拥有独立随机序列，A 的运气不受 B 操作次数影响
- Engine 彻底不持有可变状态，真正 goroutine-safe
- State 自包含：PCG rand 的内部状态（两个 uint64）可序列化，便于存档

### 影响

- `ConfigBase.R` 字段删除（breaking change）
- `NewState()` 独立函数删除，改为 `Engine.NewState(r *rand.Rand)` 工厂方法
- nil rand 校验从 `New()` 移至 `NewState()`（nil rand + 有特殊层 → panic）
- 现有调用方需将 rand 从 Config 移到 `eng.NewState(r)` 调用处