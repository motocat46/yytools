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
// 创建日期:2025/1/10

package snowflake

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// ---- 默认位布局常量（41+10+12，个人项目默认值）----

const (
	TimestampBits = 41
	NodeIDBits    = 10
	SequenceBits  = 12
	
	MaxTimestamp = int64((1 << TimestampBits) - 1) // 2199023255551
	MaxNodeID    = int64((1 << NodeIDBits) - 1)    // 1023
	MaxSequence  = int64((1 << SequenceBits) - 1)  // 4095
	
	// Epoch 自定义纪元：2025-01-01 00:00:00.000 UTC（毫秒）
	Epoch = int64(1735689600000)
	
	// Lua51MaxInt Lua 5.1 float64 能精确表示的最大整数（2^53）
	// 本包生成的 ID 超出此范围，与 Lua 5.1 交互时需转为字符串
	Lua51MaxInt = int64(9007199254740992)
)

// ---- Layout：位布局配置 ----

// Layout 定义雪花 ID 的位布局。
//
// 约束：TimestampBits + NodeIDBits + SequenceBits 必须等于 63（int64 去掉符号位），
// 违反约束时 [NewGeneratorWithLayout] 返回 error。
type Layout struct {
	TimestampBits int
	NodeIDBits    int
	SequenceBits  int
	// Epoch 自定义纪元（UTC 毫秒时间戳），ID 中存储的时间戳是相对此纪元的偏移量。
	// 纪元越近，可用年限越长；必须早于当前时间，否则 [NewGeneratorWithLayout] 构造时 panic。
	Epoch int64
}

// defaultLayout 是默认位布局（41+10+12，纪元 2025-01-01），仅供 NewGenerator 使用。
// 不导出，防止外部修改影响全局行为。需要查看默认值时直接使用包级常量。
var defaultLayout = Layout{
	TimestampBits: TimestampBits,
	NodeIDBits:    NodeIDBits,
	SequenceBits:  SequenceBits,
	Epoch:         Epoch,
}

// ---- 类型定义 ----

// IDParts 解码后的 ID 各字段，用于调试/日志
type IDParts struct {
	Time      time.Time // 对应的 UTC 绝对时间（毫秒精度）
	Timestamp int64     // 距自定义纪元的毫秒数
	NodeID    int64     // 节点ID
	Sequence  int64     // 序列号
}

// Generator 无锁线程安全的雪花 ID 生成器
type Generator struct {
	// state 原子变量：打包存储 (lastMs << nodeIDShift) | sequence
	// 41 + 12 = 53 位（默认布局），完全在 int64 范围内
	state int64
	nid   int64 // nodeID，构造后不变
	
	// 以下字段由 Layout 派生，构造后不变，运行时零开销
	maxTimestamp   int64 // 最大时间值
	maxNodeID      int64 // 最大节点ID
	maxSequence    int64 // 最大序列号
	timestampShift int   // = NodeIDBits + SequenceBits
	nodeIDShift    int   // = SequenceBits（同时也是 state 的打包移位数）
	epoch          int64 // 纪元起始时间戳
}

// ---- Generator 实例方法（使用实例 Layout 字段，支持任意位布局）----

func (g *Generator) currentMillis() int64 {
	ms := time.Now().UnixMilli() - g.epoch
	if ms < 0 {
		panic(fmt.Sprintf("snowflake: system clock is before epoch (ms=%d, epoch=%d)", ms, g.epoch))
	}
	if ms > g.maxTimestamp {
		panic(fmt.Sprintf("snowflake: timestamp overflow (ms=%d, max=%d)", ms, g.maxTimestamp))
	}
	return ms
}

func (g *Generator) packState(ms, seq int64) int64 {
	return (ms << g.nodeIDShift) | seq
}

func (g *Generator) unpackState(state int64) (ms, seq int64) {
	return state >> g.nodeIDShift, state & g.maxSequence
}

// ---- Generator 构造 ----

// NewGeneratorWithLayout 构造一个使用自定义位布局的 Generator。
// layout.TimestampBits + layout.NodeIDBits + layout.SequenceBits 必须等于 63。
// nodeID 必须在 [0, maxNodeID] 范围内，否则返回 error。
//
// 警告：同一进程内使用相同 nodeID 创建多个 Generator 实例会产生重复 ID，调用方负责保证 nodeID 全局唯一。
func NewGeneratorWithLayout(nodeID int32, layout Layout) (*Generator, error) {
	if layout.TimestampBits <= 0 || layout.NodeIDBits <= 0 || layout.SequenceBits <= 0 {
		return nil, fmt.Errorf("snowflake: each of TimestampBits(%d), NodeIDBits(%d), SequenceBits(%d) must be > 0",
			layout.TimestampBits, layout.NodeIDBits, layout.SequenceBits)
	}
	if layout.TimestampBits+layout.NodeIDBits+layout.SequenceBits != 63 {
		return nil, fmt.Errorf("snowflake: TimestampBits(%d)+NodeIDBits(%d)+SequenceBits(%d) must equal 63",
			layout.TimestampBits, layout.NodeIDBits, layout.SequenceBits)
	}
	maxNodeID := int64((1 << layout.NodeIDBits) - 1)
	nid := int64(nodeID)
	if nid < 0 || nid > maxNodeID {
		return nil, fmt.Errorf("snowflake: nodeID %d out of range [0, %d]", nodeID, maxNodeID)
	}
	g := &Generator{
		nid:            nid,
		maxTimestamp:   int64((1 << layout.TimestampBits) - 1),
		maxNodeID:      maxNodeID,
		maxSequence:    int64((1 << layout.SequenceBits) - 1),
		timestampShift: layout.NodeIDBits + layout.SequenceBits,
		nodeIDShift:    layout.SequenceBits,
		epoch:          layout.Epoch,
	}
	g.state = g.packState(g.currentMillis(), 0)
	return g, nil
}

// NewGenerator 构造一个使用默认位布局（41+10+12）的 Generator。
// nodeID 必须在 [0, 1023] 范围内，否则返回 error。
//
// 警告：同一进程内使用相同 nodeID 创建多个 Generator 实例会产生重复 ID，调用方负责保证 nodeID 全局唯一。
func NewGenerator(nodeID int32) (*Generator, error) {
	return NewGeneratorWithLayout(nodeID, defaultLayout)
}

// ---- Generator 核心方法 ----

// NewID 生成一个全局唯一的雪花 ID。线程安全，永不返回错误。
//
// 序号范围 [0, maxSequence]，每毫秒每节点最多 2^SequenceBits 个 ID。
// 序号耗尽时使用 runtime.Gosched() 自旋等待到下一毫秒，
// 精度优于 time.Sleep（后者在 1ms 尺度精度仅 1–5ms，会过冲导致恶性循环）。
func (g *Generator) NewID() int64 {
	for {
		old := atomic.LoadInt64(&g.state)
		oldMs, oldSeq := g.unpackState(old)
		now := g.currentMillis()
		
		var newMs, newSeq int64
		if now <= oldMs {
			// 1.同一毫秒（或时钟回拨）：沿用 lastMs，递增序号
			// 2.时钟回拨时继续用 lastMs 保证 ID 唯一性不受影响
			newMs = oldMs
			newSeq = oldSeq + 1
			if newSeq > g.maxSequence {
				// 序号耗尽：自旋等待到下一毫秒边界
				// Gosched() 单次 ~50–500ns，精确检测毫秒边界，避免过冲风险
				for g.currentMillis() <= oldMs {
					runtime.Gosched()
				}
				continue
			}
		} else {
			// 进入新的毫秒：重置序号为 0。
			// 多个 goroutine 可能同时计算出相同的 now，都尝试以 newSeq=0 CAS。
			// 只有一个成功，其余 CAS 失败后重新读取 state：此时 now <= 新 oldMs，
			// 进入上方的递增分支，行为正确——这是预期的并发路径，不是 bug。
			newMs, newSeq = now, 0
		}
		
		if atomic.CompareAndSwapInt64(&g.state, old, g.packState(newMs, newSeq)) {
			// CAS成功: 组装唯一id，直接返回
			// 这里不需要再通过位运算截断newSeq(上文已判断newSeq上限值)
			// 但要留意这个点：否则newSeq一旦超过上限会污染nid
			return (newMs << g.timestampShift) | (g.nid << g.nodeIDShift) | newSeq
		}
		// CAS失败: 被其他 goroutine 抢占，继续循环尝试（无阻塞，无sleep）
	}
}

// ParseID 解码此 Generator 的布局对应的雪花 ID 为各字段，用于调试/日志。
// id 应为本生成器 NewID() 生成的正值；传入负数或不匹配布局的值时，各字段含义未定义。
func (g *Generator) ParseID(id int64) IDParts {
	ms := id >> g.timestampShift
	nid := (id >> g.nodeIDShift) & g.maxNodeID
	seq := id & g.maxSequence
	return IDParts{
		Timestamp: ms,
		NodeID:    nid,
		Sequence:  seq,
		Time:      time.UnixMilli(ms + g.epoch).UTC(),
	}
}