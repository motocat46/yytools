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

package snowflake

import (
	"sync"
	"testing"
)

// TestNewGenerator_InvalidNodeID 验证非法 nodeID 返回 error
func TestNewGenerator_InvalidNodeID(t *testing.T) {
	cases := []int32{-1, -100, 1024, 2000}
	for _, nodeID := range cases {
		_, err := NewGenerator(nodeID)
		if err == nil {
			t.Errorf("NewGenerator(%d) expected error, got nil", nodeID)
		}
	}
	// 边界合法值不应返回 error
	for _, nodeID := range []int32{0, 1, 1023} {
		_, err := NewGenerator(nodeID)
		if err != nil {
			t.Errorf("NewGenerator(%d) unexpected error: %v", nodeID, err)
		}
	}
}

// TestGenerator_NewID_Monotonic 验证单 goroutine 连续生成的 ID 单调递增
func TestGenerator_NewID_Monotonic(t *testing.T) {
	g, _ := NewGenerator(1)
	prev := g.NewID()
	for range 10000 {
		id := g.NewID()
		if id <= prev {
			t.Fatalf("ID not monotonic: prev=%d, cur=%d", prev, id)
		}
		prev = id
	}
}

// TestGenerator_NewID_Uniqueness 验证单线程大量生成无重复
func TestGenerator_NewID_Uniqueness(t *testing.T) {
	const n = 100000
	g, _ := NewGenerator(0)
	seen := make(map[int64]struct{}, n)
	for i := range n {
		id := g.NewID()
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID %d at iteration %d", id, i)
		}
		seen[id] = struct{}{}
	}
}

// TestGenerator_NewID_Concurrent 验证多 goroutine 并发生成全局唯一（配合 -race）
func TestGenerator_NewID_Concurrent(t *testing.T) {
	const (
		goroutines = 20
		perG       = 5000
		total      = goroutines * perG
	)
	g, _ := NewGenerator(7)
	ids := make([]int64, total)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		offset := i * perG
		go func(off int) {
			defer wg.Done()
			for j := 0; j < perG; j++ {
				ids[off+j] = g.NewID()
			}
		}(offset)
	}
	wg.Wait()

	seen := make(map[int64]struct{}, total)
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID found in concurrent test: %d", id)
		}
		seen[id] = struct{}{}
	}
}

// TestParseID_RoundTrip 验证生成 ID 后 ParseID 各字段与期望一致
func TestParseID_RoundTrip(t *testing.T) {
	const nodeID = int32(42)
	g, _ := NewGenerator(nodeID)
	id := g.NewID()
	parts := g.ParseID(id)

	if parts.NodeID != int64(nodeID) {
		t.Errorf("NodeID mismatch: want %d, got %d", nodeID, parts.NodeID)
	}
	if parts.Timestamp < 0 {
		t.Errorf("Timestamp should be non-negative, got %d", parts.Timestamp)
	}
	if parts.Sequence < 0 || parts.Sequence > MaxSequence {
		t.Errorf("Sequence out of range: %d", parts.Sequence)
	}
	// 验证时间字段合理（UTC 时间在2025年之后）
	if parts.Time.Year() < 2025 {
		t.Errorf("parsed time %v is before epoch 2025", parts.Time)
	}
}

// TestNewGeneratorWithLayout_CustomLayout 验证自定义布局（公司场景：40+12+11）
func TestNewGeneratorWithLayout_CustomLayout(t *testing.T) {
	companyLayout := Layout{
		TimestampBits: 40,
		NodeIDBits:    12,
		SequenceBits:  11,
		Epoch:         Epoch,
	}

	// nodeID 合法边界：maxNodeID = (1<<12)-1 = 4095
	for _, nodeID := range []int32{0, 1, 4095} {
		g, err := NewGeneratorWithLayout(nodeID, companyLayout)
		if err != nil {
			t.Fatalf("NewGeneratorWithLayout(%d) unexpected error: %v", nodeID, err)
		}

		// 生成 ID 并用实例方法解码，验证字段正确
		id := g.NewID()
		if id <= 0 {
			t.Fatalf("nodeID=%d: ID should be positive, got %d", nodeID, id)
		}
		parts := g.ParseID(id)
		if parts.NodeID != int64(nodeID) {
			t.Errorf("nodeID=%d: NodeID mismatch: want %d, got %d", nodeID, nodeID, parts.NodeID)
		}
		if parts.Timestamp < 0 {
			t.Errorf("nodeID=%d: Timestamp should be non-negative, got %d", nodeID, parts.Timestamp)
		}
		maxSeq := int64((1 << companyLayout.SequenceBits) - 1)
		if parts.Sequence < 0 || parts.Sequence > maxSeq {
			t.Errorf("nodeID=%d: Sequence %d out of range [0, %d]", nodeID, parts.Sequence, maxSeq)
		}
	}

	// nodeID 越界：4096 超出 [0, 4095]
	_, err := NewGeneratorWithLayout(4096, companyLayout)
	if err == nil {
		t.Error("NewGeneratorWithLayout(4096) should return error")
	}
}

// TestNewGeneratorWithLayout_InvalidBitSum 验证位宽之和不等于 63 时返回 error
func TestNewGeneratorWithLayout_InvalidBitSum(t *testing.T) {
	bad := Layout{TimestampBits: 40, NodeIDBits: 12, SequenceBits: 12, Epoch: Epoch} // 64 位，超出
	_, err := NewGeneratorWithLayout(0, bad)
	if err == nil {
		t.Error("bit sum=64 should return error")
	}
}

// TestNewGeneratorWithLayout_ZeroBitField 验证任一位宽为 0 时返回 error
func TestNewGeneratorWithLayout_ZeroBitField(t *testing.T) {
	cases := []Layout{
		{TimestampBits: 0, NodeIDBits: 32, SequenceBits: 31, Epoch: Epoch},
		{TimestampBits: 32, NodeIDBits: 0, SequenceBits: 31, Epoch: Epoch},
		{TimestampBits: 32, NodeIDBits: 31, SequenceBits: 0, Epoch: Epoch},
	}
	for _, layout := range cases {
		_, err := NewGeneratorWithLayout(0, layout)
		if err == nil {
			t.Errorf("zero bit field %+v should return error", layout)
		}
	}
}

// TestNewGeneratorWithLayout_Uniqueness 验证自定义布局下并发生成 ID 全局唯一
func TestNewGeneratorWithLayout_Uniqueness(t *testing.T) {
	const n = 100_000
	g, _ := NewGeneratorWithLayout(1, Layout{
		TimestampBits: 40, NodeIDBits: 12, SequenceBits: 11, Epoch: Epoch,
	})
	ids := make([]int64, n)
	for i := range n {
		ids[i] = g.NewID()
	}
	seen := make(map[int64]struct{}, n)
	for i, id := range ids {
		if _, dup := seen[id]; dup {
			t.Fatalf("自定义布局出现重复 ID at %d: %d", i, id)
		}
		seen[id] = struct{}{}
	}
}

