// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/2/28

// Package tiered_cycle 实现"分层周期分布"（Tiered Cycle Distribution）。
//
// # 概念模型
//
// 分层周期分布将游戏抽卡（Gacha）等场景中的"保底机制"抽象为通用分布模型：
//
//   - 一个"大周期"由 CycleLen 次抽取组成。
//   - 每个大周期内固定包含 sum(items[i].Quota) 次"特殊抽"，
//     剩余 CycleLen - sum(items[i].Quota) 次为"标准抽"。
//   - 特殊抽的位置在每轮周期开始时随机生成（受 MinInterval 最小间隔约束），
//     兼顾随机性与均匀分布。
//   - 标准抽走权重随机；特殊抽在满足 JoinAt 门槛且 Quota 未耗尽的候选集合中按剩余配额加权随机。
//
// # 参数说明
//
//   - CycleLen：大周期长度，即一个保底周期的总抽次数。
//   - MinInterval：两个特殊位置之间的最小间隔（以抽次计），0 表示不限。
//   - SpecialItem.Quota：该奖励在一个周期内最多出现的次数。
//   - SpecialItem.JoinAt：第几次特殊抽（0-based 特殊出现序号）时该奖励才开始进入候选池。
//     这是静态配置，与奖励在大周期内的绝对位置无关；每轮动态生成的特殊分布计划
//     决定"第几次特殊抽对应哪个绝对位置"，二者相互独立。
//     例：JoinAt=0 → 从第0次特殊抽起即可出现；JoinAt=4 → 从第5次特殊抽起才能出现。
//
// # Engine 与 State 的职责分离
//
//   - Engine 保存不可变规则（Config），可被多个玩家/对象共享，无状态，线程不安全（因 rand）。
//   - State 保存单个玩家/对象的进度（当前位置、周期计划、已用配额），互相独立。
//   - 使用 Engine.Init(state) 初始化一个新 State。
//   - 使用 Engine.Next(state) 推进一次抽取。
//   - 使用 Engine.ResetCycle(state) 或 Engine.NextAutoReset(state) 管理周期重置。
//
// # 典型场景
//
// 游戏抽卡保底：100 抽为一个大周期，10 个保底位置随机分布在其中；
// 早期保底只出低星奖励，后期保底逐步解锁高星限定奖励，周期末的最后一次保底必出最高级奖励。
package tiered_cycle
