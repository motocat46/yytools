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
// 创建日期:2026/3/4

// 测试
// # 日常开发 / CI（竞态检测，~7秒）
// go test ./pkg/algorithms/idgen/snowflake/... -v -race -short
//
// # 完整压测（含 30 秒长跑，~45秒）
// # 上线前（+千万级 + 500万单线程，无 race）
// go test ./pkg/algorithms/idgen/snowflake/... -v
// 或者加上测试超时时间
// go test ./pkg/algorithms/idgen/snowflake/... -v -timeout 120s
//
// 核心原则：-race 和大数据量尽量不同时用，-race 本身有 5-20x 开销；-short 用于屏蔽耗时测试，让 CI 保持快速。

// 只有基准测试
// # 只跑 Benchmark
// go test ./pkg/algorithms/idgen/snowflake/... -bench=. -run=^$ -benchmem
//
// # 只跑某个 Benchmark
// go test ./pkg/algorithms/idgen/snowflake/... -bench=BenchmarkGenerator_NewID -run=^$

// 测试+基准测试
// # 运行所有 Benchmark
// go test ./pkg/algorithms/idgen/snowflake/... -bench=.
//
// # 指定具体 Benchmark
// go test ./pkg/algorithms/idgen/snowflake/... -bench=BenchmarkGenerator_NewID
//
// # 加 -benchmem 显示内存分配
// go test ./pkg/algorithms/idgen/snowflake/... -bench=. -benchmem
//
// # 加 -benchtime 控制运行时长（默认 1s，越长结果越稳定）
// go test ./pkg/algorithms/idgen/snowflake/... -bench=. -benchtime=3s
//
// # 加 -count 多次运行取平均（用于统计稳定性）
// go test ./pkg/algorithms/idgen/snowflake/... -bench=. -benchtime=3s -count=3
package snowflake

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestStress_Concurrent_Uniqueness 高并发唯一性压力测试
//
// 100 goroutine 各生成 10000 个 ID，共 100 万个，排序后验证相邻无重复。
// 排序去重比 map 省内存（8MB vs ~150MB），O(n log n) 时间。
func TestStress_Concurrent_Uniqueness(t *testing.T) {
	const (
		goroutines = 100
		perG       = 10_000
		total      = goroutines * perG // 1,000,000
	)
	g, _ := NewGenerator(0)
	ids := make([]int64, total)
	
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		off := i * perG
		go func(off int) {
			defer wg.Done()
			for j := range perG {
				ids[off+j] = g.NewID()
			}
		}(off)
	}
	wg.Wait()
	
	slices.Sort(ids)
	for i := 1; i < total; i++ {
		if ids[i] == ids[i-1] {
			t.Fatalf("重复 ID，sorted[%d]=%d", i, ids[i])
		}
	}
	t.Logf("高并发唯一性: %d goroutine × %d = %d 个 ID，全部唯一", goroutines, perG, total)
}

// TestStress_TenMillion_Uniqueness 千万级唯一性压力测试（-short 跳过）
//
// 设计原则：
//   - 数量级：100 goroutine × 100,000 = 1000 万，是日常并发测试的 10x
//   - 不启用 -race：-race 有 5-20x 开销，千万 + race 会让 CI 超时；
//     竞态已由 TestStress_ConcurrentHighFreq_WithRace 覆盖
//   - 意义：在不依赖人工设置 state 的条件下，自然触发大量序号耗尽，
//     证明生产负载下 CAS + 自旋路径的实际稳定性
//   - 运行方式：go test ./pkg/algorithms/idgen/snowflake/... -v -run TestStress_TenMillion
func TestStress_TenMillion_Uniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过千万级压力测试（-short 模式）")
	}
	const (
		goroutines = 100
		perG       = 100_000
		total      = goroutines * perG // 10,000,000
	)
	g, _ := NewGenerator(0)
	ids := make([]int64, total)
	
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		off := i * perG
		go func(off int) {
			defer wg.Done()
			for j := range perG {
				ids[off+j] = g.NewID()
			}
		}(off)
	}
	wg.Wait()
	
	slices.Sort(ids)
	for i := 1; i < total; i++ {
		if ids[i] == ids[i-1] {
			t.Fatalf("千万级压测发现重复 ID，sorted[%d]=%d", i, ids[i])
		}
	}
	t.Logf("千万级唯一性: %d goroutine × %d = %d 个 ID，全部唯一（自然序号耗尽约 %d 次）",
		goroutines, perG, total, total/4096)
}

// TestStress_SequenceExhaustion 序号耗尽自旋路径测试
//
// 将 state 直接设置为 sequence=MaxSequence（已满），下一次 NewID 必然触发
// 自旋等待逻辑（runtime.Gosched 循环），覆盖核心容错路径。
// 设置 lastMs = currentMillis()+2，保证自旋等待时间 ≤ 2ms。
func TestStress_SequenceExhaustion(t *testing.T) {
	g, _ := NewGenerator(2)
	
	// futureMs 确保 now <= oldMs 成立，第1次调用一定触发序号耗尽分支
	futureMs := g.currentMillis() + 2
	atomic.StoreInt64(&g.state, g.packState(futureMs, MaxSequence))
	
	const n = 20
	ids := make([]int64, n)
	for i := range ids {
		ids[i] = g.NewID()
	}
	
	// 全局无重复
	seen := make(map[int64]struct{}, n)
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			t.Fatalf("序号耗尽边界出现重复 ID: %d", id)
		}
		seen[id] = struct{}{}
	}
	
	// 自旋等到 currentMillis() > futureMs 才返回，所有 ID 的 timestamp > futureMs
	for i, id := range ids {
		if ts := g.ParseID(id).Timestamp; ts <= futureMs {
			t.Errorf("ID[%d] timestamp=%d 应 > futureMs=%d（未正确跨越毫秒边界）", i, ts, futureMs)
		}
	}
	t.Logf("序号耗尽自旋: 成功跨越毫秒边界，续生 %d 个唯一 ID", n)
}

// TestStress_BitFields_AllValid 位字段边界完整性验证
//
// 大量生成 ID，对每个 ID 验证：
//   - 符号位为 0（ID > 0）
//   - NodeID 等于期望值
//   - Timestamp 在 [0, MaxTimestamp]
//   - Sequence 在 [0, MaxSequence]
//   - 解析出的时间在 [2025, 2094] 年内
func TestStress_BitFields_AllValid(t *testing.T) {
	const (
		wantNodeID = int32(512)
		n          = 10_000
	)
	g, _ := NewGenerator(wantNodeID)
	
	for i := range n {
		id := g.NewID()
		
		// 符号位必须为 0
		if id <= 0 {
			t.Fatalf("ID[%d]=%d 不是正 int64（符号位污染）", i, id)
		}
		
		parts := g.ParseID(id)
		if parts.NodeID != int64(wantNodeID) {
			t.Fatalf("ID[%d] NodeID 错误: want %d, got %d", i, wantNodeID, parts.NodeID)
		}
		if parts.Timestamp < 0 || parts.Timestamp > MaxTimestamp {
			t.Fatalf("ID[%d] Timestamp=%d 越界 [0, %d]", i, parts.Timestamp, MaxTimestamp)
		}
		if parts.Sequence < 0 || parts.Sequence > MaxSequence {
			t.Fatalf("ID[%d] Sequence=%d 越界 [0, %d]", i, parts.Sequence, MaxSequence)
		}
		year := parts.Time.Year()
		if year < 2025 || year > 2094 {
			t.Fatalf("ID[%d] 解析时间 %v 超出预期范围 [2025, 2094]", i, parts.Time)
		}
	}
	t.Logf("位字段验证: %d 个 ID 全部合法（nodeID=%d）", n, wantNodeID)
}

// TestStress_ClockRollback_Simulation 时钟回拨模拟测试
//
// 游戏服务器运维中偶有 NTP 校时导致系统时钟短暂回拨。
// 本测试人工将 lastMs 设置为"未来" 500ms，模拟回拨场景：
//   - 验证回拨期间 ID 不重复
//   - 验证回拨期间 ID 大于正常期间 ID（因 lastMs 更大）
//   - 验证解码出的 timestamp 等于人工设置的 rollbackMs
func TestStress_ClockRollback_Simulation(t *testing.T) {
	g, _ := NewGenerator(5)
	
	// 生成一批正常 ID
	const normal = 50
	normalIDs := make([]int64, normal)
	for i := range normalIDs {
		normalIDs[i] = g.NewID()
	}
	maxNormal := normalIDs[normal-1]
	
	// 模拟时钟回拨：将 lastMs 推到"未来" 500ms，sequence 置 0
	// 等价于：系统时钟被调快 500ms 后调回，lastMs 记录了那个"未来"值
	rollbackMs := g.currentMillis() + 500
	atomic.StoreInt64(&g.state, g.packState(rollbackMs, 0))
	
	// 在回拨区间生成 ID（now < rollbackMs，走回拨分支）
	const n = 100
	rollbackIDs := make([]int64, n)
	for i := range rollbackIDs {
		rollbackIDs[i] = g.NewID()
	}
	
	// 回拨区间 ID 无重复
	seen := make(map[int64]struct{}, n)
	for _, id := range rollbackIDs {
		if _, dup := seen[id]; dup {
			t.Fatalf("时钟回拨期间出现重复 ID: %d", id)
		}
		seen[id] = struct{}{}
	}
	
	// 回拨区间 ID 必须大于正常期间最大 ID（rollbackMs > 正常 ms）
	for i, id := range rollbackIDs {
		if id <= maxNormal {
			t.Errorf("回拨 ID[%d]=%d 应 > 正常最大 ID %d", i, id, maxNormal)
		}
	}
	
	// timestamp 均等于 rollbackMs（100个 ID，sequence 1-100，同一毫秒内）
	for i, id := range rollbackIDs {
		if ts := g.ParseID(id).Timestamp; ts != rollbackMs {
			t.Errorf("回拨 ID[%d] timestamp=%d，预期 rollbackMs=%d", i, ts, rollbackMs)
		}
	}
	t.Logf("时钟回拨模拟: rollbackMs=%d，%d 个 ID 全部唯一且时间戳正确", rollbackMs, n)
}

// TestStress_MultiNode_Isolation 多节点 ID 空间隔离测试
//
// 两个 Generator（nodeID=0 和 nodeID=1023，边界值）并发生成 ID，
// 合并后排序验证：跨节点无重复，且每个 ID 的 nodeID 字段正确编码。
func TestStress_MultiNode_Isolation(t *testing.T) {
	const perNode = 50_000
	g0, _ := NewGenerator(0)
	g1, _ := NewGenerator(1023) // 最大合法 nodeID
	
	ids0 := make([]int64, perNode)
	ids1 := make([]int64, perNode)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range ids0 {
			ids0[i] = g0.NewID()
		}
	}()
	go func() {
		defer wg.Done()
		for i := range ids1 {
			ids1[i] = g1.NewID()
		}
	}()
	wg.Wait()
	
	// 跨节点唯一性：合并排序检查
	all := make([]int64, 0, perNode*2)
	all = append(all, ids0...)
	all = append(all, ids1...)
	slices.Sort(all)
	for i := 1; i < len(all); i++ {
		if all[i] == all[i-1] {
			t.Fatalf("跨节点重复 ID: %d", all[i])
		}
	}
	
	// nodeID 字段正确编码
	for i, id := range ids0 {
		if nid := g0.ParseID(id).NodeID; nid != 0 {
			t.Fatalf("g0 ID[%d] nodeID 错误: want 0, got %d", i, nid)
		}
	}
	for i, id := range ids1 {
		if nid := g1.ParseID(id).NodeID; nid != 1023 {
			t.Fatalf("g1023 ID[%d] nodeID 错误: want 1023, got %d", i, nid)
		}
	}
	t.Logf("多节点隔离: nodeID=0 和 nodeID=1023 各生成 %d 个 ID，全局无重复，字段正确", perNode)
}

// TestStress_IDMonotonicity_AcrossMsBoundary 跨毫秒边界单调性验证
//
// 持续生成 ID 直至发生毫秒跳变，验证跨边界前后 ID 仍然严格单调递增。
func TestStress_IDMonotonicity_AcrossMsBoundary(t *testing.T) {
	g, _ := NewGenerator(3)
	
	// 等到新毫秒刚开始时再采样，确保必然覆盖至少2个毫秒边界
	start := g.currentMillis()
	for g.currentMillis() == start {
	}
	
	deadline := time.Now().Add(3 * time.Millisecond)
	var prev int64
	count := 0
	for time.Now().Before(deadline) {
		id := g.NewID()
		if id <= prev {
			t.Fatalf("ID 不单调: prev=%d, cur=%d（第 %d 个）", prev, id, count)
		}
		prev = id
		count++
	}
	t.Logf("跨毫秒边界单调性: 3ms 内生成 %d 个 ID，全部严格单调递增", count)
}

// TestStress_HighFreqSingleThread 单线程极限频率测试（-short 跳过）
//
// 单 goroutine 连续生成 500 万个 ID，验证在序号大量耗尽场景下的正确性。
// 单线程 ~244 ns/op，每毫秒约可到达 4096 次上限，500 万次约触发 ~1220 次
// 自然序号耗尽（不依赖人工设置 state），全面覆盖跨毫秒自旋等待路径。
// 单线程下结果必须严格单调递增（snowflake 对单线程的强保证）。
func TestStress_HighFreqSingleThread(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过单线程极限测试（-short 模式）")
	}
	const n = 5_000_000
	g, _ := NewGenerator(4)
	ids := make([]int64, n)
	for i := range ids {
		ids[i] = g.NewID()
	}
	
	for i := 1; i < n; i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("单线程 ID 不单调: ids[%d]=%d <= ids[%d]=%d", i, ids[i], i-1, ids[i-1])
		}
	}
	t.Logf("单线程极限: %d 个 ID 严格单调递增（自然序号耗尽约 %d 次）", n, n/4096)
}

// TestStress_ConcurrentHighFreq_WithRace 高并发高频唯一性（配合 -race 检测数据竞争）
//
// goroutine 数量少但频率极高，更容易触发竞态窗口。
func TestStress_ConcurrentHighFreq_WithRace(t *testing.T) {
	const (
		goroutines = 8
		perG       = 100_000
		total      = goroutines * perG
	)
	g, _ := NewGenerator(10)
	ids := make([]int64, total)
	var counter int64
	
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range perG {
				idx := atomic.AddInt64(&counter, 1) - 1
				ids[idx] = g.NewID()
			}
		}()
	}
	wg.Wait()
	
	slices.Sort(ids)
	for i := 1; i < total; i++ {
		if ids[i] == ids[i-1] {
			t.Fatalf("高频并发重复 ID: sorted[%d]=%d", i, ids[i])
		}
	}
	t.Logf("高频并发: %d goroutine × %d = %d 个 ID，无重复", goroutines, perG, total)
}

// TestStress_LongRunning 持续运行压力测试（-short 模式下跳过）
//
// 8 个 goroutine 持续生成 2 秒，统计总量和吞吐量，排序验证全局唯一。
// 运行方式：go test ./pkg/algorithms/idgen/snowflake/... -v -race -run TestStress_LongRunning
func TestStress_LongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过长时间压力测试（-short 模式）")
	}
	const (
		duration   = 30 * time.Second
		goroutines = 8
	)
	g, _ := NewGenerator(99)
	
	// 每个 goroutine 本地收集，避免跨 goroutine 锁竞争影响吞吐量测量
	localIDs := make([][]int64, goroutines)
	for i := range localIDs {
		localIDs[i] = make([]int64, 0, 1_000_000)
	}
	
	deadline := time.Now().Add(duration)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			for time.Now().Before(deadline) {
				localIDs[i] = append(localIDs[i], g.NewID())
			}
		}(i)
	}
	wg.Wait()
	
	// 合并所有 ID
	var totalCount int
	for _, s := range localIDs {
		totalCount += len(s)
	}
	all := make([]int64, 0, totalCount)
	for _, s := range localIDs {
		all = append(all, s...)
	}
	
	slices.Sort(all)
	for i := 1; i < len(all); i++ {
		if all[i] == all[i-1] {
			t.Fatalf("长时间压测发现重复 ID: %d", all[i])
		}
	}
	t.Logf("长时间压测（%v，%d goroutine）: %d 个 ID，吞吐量 %.0f 个/秒，自然序号耗尽约 %d 次",
		duration, goroutines, totalCount, float64(totalCount)/duration.Seconds(), totalCount/4096)
}

// BenchmarkGenerator_NewID 单 goroutine 基准测试
func BenchmarkGenerator_NewID(b *testing.B) {
	g, _ := NewGenerator(1)
	for b.Loop() {
		g.NewID()
	}
}

// BenchmarkGenerator_NewID_Parallel 多 goroutine 并发基准测试
func BenchmarkGenerator_NewID_Parallel(b *testing.B) {
	g, _ := NewGenerator(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.NewID()
		}
	})
}

// BenchmarkParseID 解码基准测试
func BenchmarkParseID(b *testing.B) {
	g, _ := NewGenerator(1)
	id := g.NewID()
	for b.Loop() {
		g.ParseID(id)
	}
}