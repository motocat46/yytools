// Package progressive_weight_cycle 提供渐进式解锁的权重周期分布机制。
//
// # 核心概念
//
// 在许多游戏场景中，一个大周期内会安排若干个"特殊位置"，每个位置必须从候选池中随机选出一个
// 奖励。候选池并非一成不变：随着特殊位置序号的推进，更高级的奖励会逐步解锁并加入候选池，
// 同时每种奖励的出现次数受配额约束。这就是"渐进式解锁权重周期"的核心思想。
//
// # 关键字段
//
//   - Item.Quota：该奖励在一个周期内最多出现几次，决定了其在候选池中的权重。
//     配额耗尽后自动退出候选池。
//
//   - Item.JoinAt：该奖励从第几次特殊抽（0-based occIdx）起才进入候选池。
//     例如 JoinAt=0 表示从第 0 次起即可出现；JoinAt=3 表示从第 4 次特殊抽起才解锁。
//
//   - occIdx（传给 Layer.Generate 的参数）：当前特殊位置在本周期内的序号（0-based），
//     由调用方根据业务逻辑维护和传入。
//
// # 数据流
//
// 一次完整周期的调用过程如下：
//
//  1. 用 NewWeightCycleLayer(items) 构造 Layer（持有不可变规则，可复用）。
//  2. 用 NewState() 构造 State（持有运行时状态，每个独立对象/玩家各一份）。
//  3. 每次触发特殊位置时，以当前 occIdx 调用 Layer.Generate(state, occIdx) 获取结果。
//     occIdx 从 0 开始，每次特殊位置触发后 +1，直到 TotalQuota(items) 耗尽。
//  4. 一个周期结束后，调用 State.Reset() 重置状态，开始下一个周期。
//
// # 与 tiered_cycle 的关系
//
// tiered_cycle 是面向"普通抽 + 特殊保底"双层场景的完整引擎，内部使用本包作为特殊层实现。
// 当业务只需要纯粹的"渐进式解锁权重随机"（无需普通层、无需自动排期），可直接使用本包。
//
// # 使用示例
//
// 见 example_test.go 中的 ExampleLayer_rewardBox。
package progressive_weight_cycle
