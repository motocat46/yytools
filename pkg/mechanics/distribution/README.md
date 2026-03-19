# mechanics/distribution

游戏分布机制包，提供游戏场景中的随机分布与保底机制实现。

## 子模块

| 包 | 一句话描述 |
|----|-----------|
| [progressive_weight_cycle](./progressive_weight_cycle/) | 渐进式解锁权重周期——奖励随进度逐步解锁，配额约束下的动态权重随机 |
| [tiered_cycle](./tiered_cycle/) | 双层周期引擎——普通层 + 特殊保底层，自动排期保底位置，支持 MinInterval 约束 |

## 选择指南

- **只需渐进式奖励随机**（无普通层、无自动排期）→ `progressive_weight_cycle`
- **完整抽卡/保底系统**（普通抽 + 特殊保底双层）→ `tiered_cycle`（内部使用 `progressive_weight_cycle`）
