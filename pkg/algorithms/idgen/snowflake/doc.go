// Copyright [yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/3/3

// Package snowflake 提供无锁线程安全的雪花算法唯一 ID 生成器，专为游戏服务器设计。
//
// # 默认位布局（63位，符号位永远为0）
//
//	63                                              0
//	+------------------------------------------+----------+-------------+
//	|    41-bit 毫秒时间戳                       | 10-bit   |  12-bit     |
//	|    since 2025-01-01 00:00:00.000 UTC      | nodeID   |  sequence   |
//	+------------------------------------------+----------+-------------+
//
//   - 41位毫秒时间戳：从自定义纪元（2025-01-01）起，有效期约69年（到2094年）
//   - 10位 nodeID：最多1024个节点（0–1023）
//   - 12位 sequence：每毫秒每节点最多4096个ID（约4,096,000个/秒）
//
// # 典型使用流程
//
//	g, err := snowflake.NewGenerator(nodeID)
//	if err != nil { panic(err) }
//	id := g.NewID()
//	parts := g.ParseID(id)
//
// # 自定义位布局
//
// 不同项目可能需要不同的位分配（如更多节点或更高序号上限）。通过 [Layout] 传入自定义布局：
//
//	layout := snowflake.Layout{
//	    TimestampBits: 40,
//	    NodeIDBits:    12,  // 最多 4096 个节点
//	    SequenceBits:  11,  // 每毫秒每节点最多 2048 个 ID
//	    Epoch:         snowflake.Epoch,
//	}
//	g, err := snowflake.NewGeneratorWithLayout(nodeID, layout)
//
// 约束：TimestampBits + NodeIDBits + SequenceBits 必须等于 63。
// 自定义布局的 ID 须用 [Generator.ParseID] 解码，不能用包级 [ParseID]（后者固定使用默认布局）。
//
// # 解码调试
//
//	parts := g.ParseID(id)
//	// parts.Timestamp  距纪元毫秒数
//	// parts.NodeID     节点ID
//	// parts.Sequence   毫秒内序号
//	// parts.Time       对应的 UTC 绝对时间
//
// # 错误处理
//
// [Generator.NewID] 中的 [currentMillis] 时钟越界检测会无条件 panic，始终有效。
//
// # Lua 5.1 注意事项
//
// Lua 5.1 使用 double（float64）表示所有数值，能精确表示的最大整数为
// 2^53 = 9007199254740992（常量 Lua51MaxInt）。
// 本包生成的 ID 最大约 9.2×10^18，超出此范围。
// 与 Lua 5.1 交互时务必将 ID 转为字符串传输，避免精度丢失。
package snowflake
