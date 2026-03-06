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
// 创建日期:2026/2/10

package snowflake

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
	
	"github.com/stormYuanYang/yytools/pkg/common/assert"
)

// ---- 位布局常量 ----

const (
	TimestampBits = 41
	NodeIDBits    = 10
	SequenceBits  = 12
	
	// TimestampShift 时间戳左移位数
	TimestampShift = NodeIDBits + SequenceBits // 22
	// NodeIDShift nodeID 左移位数
	NodeIDShift = SequenceBits // 12
	
	MaxTimestamp = int64((1 << TimestampBits) - 1) // 2199023255551
	MaxNodeID    = int64((1 << NodeIDBits) - 1)    // 1023
	MaxSequence  = int64((1 << SequenceBits) - 1)  // 4095
	
	// Epoch 自定义纪元：2025-01-01 00:00:00.000 UTC（毫秒）
	Epoch = int64(1735689600000)
	
	// Lua51MaxInt Lua 5.1 float64 能精确表示的最大整数（2^53）
	// 本包生成的 ID 超出此范围，与 Lua 5.1 交互时需转为字符串
	Lua51MaxInt = int64(9007199254740992)
)

// ---- 类型定义 ----

// IDParts 解码后的 ID 各字段，用于调试/日志
type IDParts struct {
	Time      time.Time // 对应的 UTC 绝对时间（毫秒精度）
	Timestamp int64     // 距自定义纪元（2025-01-01）的毫秒数
	NodeID    int64
	Sequence  int64
}

// Generator 无锁线程安全的雪花 ID 生成器
type Generator struct {
	// state 原子变量：打包存储 (lastMs << SequenceBits) | sequence
	// 41 + 12 = 53 位，完全在 int64 范围内
	state int64
	nid   int64 // nodeID，构造后不变
}

// ---- 辅助函数 ----

func currentMillis() int64 {
	return time.Now().UnixMilli() - Epoch
}

func packState(ms, seq int64) int64 {
	return (ms << SequenceBits) | seq
}

func unpackState(state int64) (ms, seq int64) {
	return state >> SequenceBits, state & MaxSequence
}

// ---- Generator 构造与核心方法 ----

// NewGenerator 构造一个 Generator。nodeID 必须在 [0, 1023] 范围内，否则返回 error。
func NewGenerator(nodeID int32) (*Generator, error) {
	nid := int64(nodeID)
	if nid < 0 || nid > MaxNodeID {
		return nil, fmt.Errorf("snowflake: nodeID %d out of range [0, %d]", nodeID, MaxNodeID)
	}
	g := &Generator{nid: nid}
	// 初始化 state：以当前毫秒、序号0作为初始值
	g.state = packState(currentMillis(), 0)
	return g, nil
}

// NewID 生成一个全局唯一的雪花 ID。线程安全，永不返回错误。
//
// 当序号在同一毫秒内耗尽（>4095 per ms, 即>MaxSequence per ms, 因为范围是[0,MaxSequence],对应数量上限是MaxSequence+1）时，
// 使用 runtime.Gosched() 自旋等待到下一毫秒，
// 精度优于 time.Sleep（后者在 1ms 尺度精度仅 1–5ms，会过冲导致恶性循环）。
func (g *Generator) NewID() int64 {
	for {
		old := atomic.LoadInt64(&g.state)
		oldMs, oldSeq := unpackState(old)
		now := currentMillis()
		
		var newMs, newSeq int64
		if now <= oldMs {
			// 1.同一毫秒（或时钟回拨）：沿用 lastMs，递增序号
			// 2.时钟回拨时继续用 lastMs 保证 ID 唯一性不受影响
			newMs = oldMs
			newSeq = oldSeq + 1
			if newSeq > MaxSequence {
				// 序号耗尽：自旋等待到下一毫秒边界
				// Gosched() 单次 ~50–500ns，精确检测毫秒边界，避免过冲风险
				for currentMillis() <= oldMs {
					runtime.Gosched()
				}
				continue
			}
		} else {
			// 正常情况：进入新的毫秒，重置序号
			newMs, newSeq = now, 0
		}
		
		if atomic.CompareAndSwapInt64(&g.state, old, packState(newMs, newSeq)) {
			// CAS成功: 组装唯一id，直接返回
			// 这里不需要再通过位运算截断newSeq(上文已判断newSeq上限值)
			// 但要留意这个点：否则newSeq一旦超过上限会污染nid
			return (newMs << TimestampShift) | (g.nid << NodeIDShift) | newSeq
		}
		// CAS失败: 被其他 goroutine 抢占，继续循环尝试（无阻塞，无sleep）
	}
}

// ---- 包级便捷函数 ----

// defaultGen 包级默认生成器，由 Init 初始化
var defaultGen *Generator

// Init 初始化包级默认生成器。必须在调用 NewID() 前调用一次。
// nodeID 超出 [0, 1023] 范围则 panic（启动时编程错误，应尽早暴露）。
func Init(nodeID int32) {
	g, err := NewGenerator(nodeID)
	if err != nil {
		panic(err)
	}
	defaultGen = g
}

// NewID 使用包级默认生成器生成唯一 ID。必须先调用 Init，否则 panic。
func NewID() int64 {
	assert.Assert(defaultGen != nil, "snowflake: Init must be called before NewID")
	return defaultGen.NewID()
}

// ---- 工具函数 ----

// ParseID 解码雪花 ID 为各字段，用于调试/日志。
func ParseID(id int64) IDParts {
	ms := id >> TimestampShift
	nid := (id >> NodeIDShift) & MaxNodeID
	seq := id & MaxSequence
	return IDParts{
		Timestamp: ms,
		NodeID:    nid,
		Sequence:  seq,
		Time:      time.UnixMilli(ms + Epoch).UTC(), // 因为ms是从纪元时间戳开始的时间值，所以加上纪元时间戳就是标准时间戳值
	}
}