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
	parts := ParseID(id)

	if parts.NodeID != int64(nodeID) {
		t.Errorf("NodeID mismatch: want %d, got %d", nodeID, parts.NodeID)
	}
	if parts.Timestamp < 0 {
		t.Errorf("Timestamp should be non-negative, got %d", parts.Timestamp)
	}
	if parts.Sequence < 0 || parts.Sequence > MaxSequence {
		t.Errorf("Sequence out of range: %d", parts.Sequence)
	}
	// 验证重组后等于原始 ID
	reconstructed := (parts.Timestamp << TimestampShift) | (parts.NodeID << NodeIDShift) | parts.Sequence
	if reconstructed != id {
		t.Errorf("round-trip failed: original=%d, reconstructed=%d", id, reconstructed)
	}
	// 验证时间字段合理（UTC 时间在2025年之后）
	if parts.Time.Year() < 2025 {
		t.Errorf("parsed time %v is before epoch 2025", parts.Time)
	}
}

// TestInit_InvalidNodeID_Panics 验证 Init 传非法 nodeID 时 panic
func TestInit_InvalidNodeID_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Init with invalid nodeID should panic, but did not")
		}
	}()
	Init(9999)
}

// TestNewID_BeforeInit_Panics 验证未调用 Init 直接调用包级 NewID 时 panic
func TestNewID_BeforeInit_Panics(t *testing.T) {
	// 保存并重置 defaultGen，确保测试隔离
	orig := defaultGen
	defaultGen = nil
	defer func() {
		defaultGen = orig
		if r := recover(); r == nil {
			t.Error("NewID before Init should panic, but did not")
		}
	}()
	NewID()
}
