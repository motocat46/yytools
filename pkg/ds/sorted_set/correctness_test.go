package sorted_set

import (
	"math/rand/v2"
	"testing"
)

// ---- 参考模型 ----

// refModel 是 SortedSet 的参考实现，使用 map + 有序切片，语义与 SortedSet 完全一致。
// 用于与 SortedSet 的返回值做逐步对比，发现偏差。
type refModel struct {
	scores map[int]float64 // key -> score
}

func newRefModel() *refModel {
	return &refModel{scores: make(map[int]float64)}
}

func (r *refModel) insert(key int, score float64) bool {
	if _, exists := r.scores[key]; exists {
		return false
	}
	r.scores[key] = score
	return true
}

func (r *refModel) delete(key int) bool {
	if _, exists := r.scores[key]; !exists {
		return false
	}
	delete(r.scores, key)
	return true
}

func (r *refModel) updateScore(key int, score float64) bool {
	if _, exists := r.scores[key]; !exists {
		return false
	}
	r.scores[key] = score
	return true
}

func (r *refModel) keys() []int {
	keys := make([]int, 0, len(r.scores))
	for k := range r.scores {
		keys = append(keys, k)
	}
	return keys
}

// ---- 不变量验证 ----

// checkSkiplistOrder 白盒检查：直接遍历跳表底层链表，
// 验证相邻节点满足 lessOrder（score+seq 全序），
// 以及 GetByRank(rank) 与链表位置严格对应。
//
// 注意：此函数与 bench_hook.go 中的 RunBenchCheck 逻辑相同。
// 如修改其中一处，请同步更新另一处。
func checkSkiplistOrder(t *testing.T, ss *SortedSet[int, int]) {
	t.Helper()
	current := ss.sl.Head.Levels[0].Forward
	rank := 1
	for current != nil {
		if current.Levels[0].Forward != nil {
			if !current.Data.lessOrder(current.Levels[0].Forward.Data) {
				t.Errorf("跳表底层链表顺序破坏：rank=%d score=%f seq=%d >= rank=%d score=%f seq=%d",
					rank, current.Data.Score, current.Data.seq,
					rank+1, current.Levels[0].Forward.Data.Score, current.Levels[0].Forward.Data.seq)
			}
		}
		data := ss.GetByRank(rank)
		if data == nil || !data.equalOrder(current.Data) {
			t.Errorf("GetByRank(%d) 与链表位置不一致", rank)
		}
		current = current.Levels[0].Forward
		rank++
	}
}

// checkInvariants 在每次操作后验证 SortedSet 的全量不变量。
// 不变量：
//  1. Length == ref 中的元素数量
//  2. ref 中每个 key 都可通过 Get 找到，且 score 与 ref 一致
//  3. GetByRank 遍历结果单调不降（score 升序）
//  4. GetRank(GetByRank(r).Key) == r（排名双向一致）
//  5. GetByRank 覆盖的 key 集合 == ref 中的 key 集合（无多无少）
//  6. GetRankDesc(key) + GetRank(key) == Length+1（升降序对称）
//  7. GetMin/GetMax 与 GetByRank(1)/GetByRank(n) 一致
//  8. 跳表底层链表 lessOrder 全序（白盒，通过 checkSkiplistOrder 验证）
func checkInvariants(t *testing.T, ss *SortedSet[int, int], ref *refModel) {
	t.Helper()
	n := ss.Length()

	// 1. 长度一致
	if n != len(ref.scores) {
		t.Errorf("Length 不一致: ss=%d ref=%d", n, len(ref.scores))
	}

	// 2. ref 中每个 key 在 ss 中都可查到，score 正确
	for key, wantScore := range ref.scores {
		node := ss.Get(key)
		if node == nil {
			t.Errorf("Get(%d) 返回 nil，但 ref 中存在 score=%f", key, wantScore)
			continue
		}
		if node.Score != wantScore {
			t.Errorf("Get(%d).Score=%f，ref 期望 %f", key, node.Score, wantScore)
		}
	}

	if n == 0 {
		return
	}

	// 3. GetByRank 遍历：score 单调不降
	prev := ss.GetByRank(1)
	if prev == nil {
		t.Errorf("Length=%d 但 GetByRank(1) 返回 nil", n)
		return
	}
	for r := 2; r <= n; r++ {
		cur := ss.GetByRank(r)
		if cur == nil {
			t.Errorf("Length=%d 但 GetByRank(%d) 返回 nil", n, r)
			break
		}
		if cur.Score < prev.Score {
			t.Errorf("排序破坏：rank=%d score=%f < rank=%d score=%f", r, cur.Score, r-1, prev.Score)
		}
		prev = cur
	}

	// 4. GetRank 与 GetByRank 双向一致
	for r := 1; r <= n; r++ {
		node := ss.GetByRank(r)
		if node == nil {
			continue
		}
		if gotRank := ss.GetRank(node.Key); gotRank != r {
			t.Errorf("GetRank(GetByRank(%d).Key) = %d，期望 %d", r, gotRank, r)
		}
	}

	// 5. GetByRank 遍历的 key 集合 == ref key 集合
	ssKeys := make(map[int]struct{}, n)
	for r := 1; r <= n; r++ {
		node := ss.GetByRank(r)
		if node != nil {
			ssKeys[node.Key] = struct{}{}
		}
	}
	for key := range ref.scores {
		if _, ok := ssKeys[key]; !ok {
			t.Errorf("ref key=%d 无法通过 GetByRank 遍历到", key)
		}
	}
	for key := range ssKeys {
		if _, ok := ref.scores[key]; !ok {
			t.Errorf("ss 中存在 key=%d，但 ref 中没有", key)
		}
	}

	// 6. 升降序 rank 对称：GetRankDesc + GetRank == Length+1
	for r := 1; r <= n; r++ {
		node := ss.GetByRank(r)
		if node == nil {
			continue
		}
		desc := ss.GetRankDesc(node.Key)
		if desc+r != n+1 {
			t.Errorf("rank 对称性破坏：key=%d asc=%d desc=%d，期望和为 %d", node.Key, r, desc, n+1)
		}
	}

	// 7. GetMin/GetMax 与 GetByRank(1)/GetByRank(n) 一致
	minNode := ss.GetMin()
	if minNode == nil {
		t.Errorf("Length=%d 但 GetMin() 返回 nil", n)
	} else if minNode.Key != ss.GetByRank(1).Key {
		t.Errorf("GetMin().Key=%d，期望与 GetByRank(1).Key=%d 一致", minNode.Key, ss.GetByRank(1).Key)
	}
	maxNode := ss.GetMax()
	if maxNode == nil {
		t.Errorf("Length=%d 但 GetMax() 返回 nil", n)
	} else if maxNode.Key != ss.GetByRank(n).Key {
		t.Errorf("GetMax().Key=%d，期望与 GetByRank(%d).Key=%d 一致", maxNode.Key, n, ss.GetByRank(n).Key)
	}

	// 8. 白盒：跳表底层链表全序
	checkSkiplistOrder(t, ss)
}

// ---- 随机混合操作正确性测试 ----

// TestSortedSet_RandomOps 通过随机混合操作序列验证 SortedSet 的整体正确性。
//
// 策略：
//   - 维护与 SortedSet 语义完全相同的参考模型（ref），每次操作后对比结果
//   - 每批操作后调用 checkInvariants 验证全量不变量
//   - 固定随机种子保证可复现
func TestSortedSet_RandomOps(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机测试")
	}
	const (
		rounds      = 20    // 轮数
		opsPerRound = 5_000 // 每轮操作数，总计 10 万次操作
		scoreRange  = 1000  // score 范围 [-500, 500)
		maxKeys     = 1_000 // key 池大小，控制 insert/delete/update 比例
	)

	rng := rand.New(rand.NewPCG(42, 0))
	ss := NewSortedSet[int, int]()
	ref := newRefModel()
	nextKey := 1

	randScore := func() float64 {
		return float64(rng.IntN(scoreRange)) - float64(scoreRange/2)
	}

	for round := range rounds {
		for range opsPerRound {
			existingKeys := ref.keys()
			hasElements := len(existingKeys) > 0

			// 操作权重（满员时停止 insert）：
			//   insert=35%  delete=20%  update=20%
			//   DeleteRangeByScore=15%  DeleteRangeByRank/Desc=10%
			var op int
			if !hasElements || len(existingKeys) < maxKeys/4 {
				op = 0 // 强制 insert
			} else if len(existingKeys) >= maxKeys {
				op = 1 + rng.IntN(4) // 只删除 / 更新
			} else {
				op = rng.IntN(20)
			}

			switch {
			case op <= 6: // insert
				key := nextKey
				nextKey++
				score := randScore()
				ssOk := ss.Insert(&NodeData[int, int]{Key: key, Score: score, Val: key})
				refOk := ref.insert(key, score)
				if ssOk != refOk {
					t.Fatalf("round=%d Insert(key=%d) ss=%v ref=%v 不一致", round, key, ssOk, refOk)
				}

			case op <= 10: // delete single
				if !hasElements {
					continue
				}
				key := existingKeys[rng.IntN(len(existingKeys))]
				_, ssOk := ss.Delete(key)
				refOk := ref.delete(key)
				if ssOk != refOk {
					t.Fatalf("round=%d Delete(key=%d) ss=%v ref=%v 不一致", round, key, ssOk, refOk)
				}
				if ss.Get(key) != nil {
					t.Errorf("round=%d Delete(key=%d) 后 Get 应返回 nil", round, key)
				}

			case op <= 14: // update
				if !hasElements {
					continue
				}
				key := existingKeys[rng.IntN(len(existingKeys))]
				newScore := randScore()
				_, ssOk := ss.UpdateScore(key, newScore)
				refOk := ref.updateScore(key, newScore)
				if ssOk != refOk {
					t.Fatalf("round=%d UpdateScore(key=%d) ss=%v ref=%v 不一致", round, key, ssOk, refOk)
				}

			case op <= 17: // DeleteRangeByScore
				if !hasElements {
					continue
				}
				lo, hi := randScore(), randScore()
				if lo > hi {
					lo, hi = hi, lo
				}
				deleted := ss.DeleteRangeByScore(lo, false, hi, false)
				// 将 ss 删除的 key 同步到 ref；单元测试已验证删除元素的正确性，
				// 此处关注的是大量混合操作后内部结构的一致性。
				for _, d := range deleted {
					ref.delete(d.Key)
				}

			default: // DeleteRangeByRank 或 DeleteRangeByRankDesc（各 op 约 5%）
				if !hasElements {
					continue
				}
				n := ss.Length()
				// 每次最多删除 5 个，避免集合过快缩小
				start := rng.IntN(n) + 1
				end := min(start+rng.IntN(5), n)
				var deleted []*NodeData[int, int]
				if rng.IntN(2) == 0 {
					deleted = ss.DeleteRangeByRank(start, end)
				} else {
					deleted = ss.DeleteRangeByRankDesc(start, end)
				}
				for _, d := range deleted {
					ref.delete(d.Key)
				}
			}
		}

		// 每轮结束验证全量不变量
		checkInvariants(t, ss, ref)
		if t.Failed() {
			t.Fatalf("round=%d 不变量检查失败，终止", round)
		}

		// 额外验证：GetRangeByScore 结果有序且范围正确
		if ss.Length() > 0 {
			lo, hi := randScore(), randScore()
			if lo > hi {
				lo, hi = hi, lo
			}
			result := ss.GetRangeByScore(lo, false, hi, false)
			for j, node := range result {
				if node.Score < lo || node.Score > hi {
					t.Errorf("round=%d GetRangeByScore[%f,%f] 第%d个元素 score=%f 超出范围",
						round, lo, hi, j, node.Score)
				}
				if j > 0 && result[j-1].Score > node.Score {
					t.Errorf("round=%d GetRangeByScore 结果未升序，位置 %d", round, j)
				}
			}
			// 验证结果数量与 ref 一致
			refCount := 0
			for _, score := range ref.scores {
				if score >= lo && score <= hi {
					refCount++
				}
			}
			if len(result) != refCount {
				t.Errorf("round=%d GetRangeByScore[%f,%f] ss返回%d个，ref期望%d个",
					round, lo, hi, len(result), refCount)
			}
		}
	}
}

// TestSortedSet_StressOps 百万级压力测试：验证大规模混合操作后不变量仍然成立。
//
// 与 TestSortedSet_RandomOps 的差异：
//   - 总操作数 100 万（10 倍），充分压测 O(log n) 路径
//   - maxKeys=10,000，集合规模更大，暴露大集合下的跳表边界 bug
//   - 使用 keyPool 切片实现 O(1) 随机键选取，避免 O(n) map 遍历成为瓶颈
//   - 不对每步操作做 ref 结果对比（依靠 checkInvariants 验证每轮后全量不变量）
func TestSortedSet_StressOps(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过百万级压力测试")
	}
	const (
		rounds      = 10
		opsPerRound = 100_000 // 总计 100 万次操作
		scoreRange  = 100_000
		maxKeys     = 10_000 // 较大集合，充分压测 O(log n)
	)

	rng := rand.New(rand.NewPCG(99, 0))
	ss := NewSortedSet[int, int]()
	ref := newRefModel()
	// keyPool 与 ref 同步维护，swap-remove 保持 O(1) 随机选取
	keyPool := make([]int, 0, maxKeys)
	nextKey := 1

	randScore := func() float64 {
		return float64(rng.IntN(scoreRange)) - float64(scoreRange/2)
	}

	for round := range rounds {
		for range opsPerRound {
			poolSize := len(keyPool)
			var op int
			if poolSize == 0 || poolSize < maxKeys/4 {
				op = 0 // 强制 insert
			} else if poolSize >= maxKeys {
				op = 1 + rng.IntN(2) // 只 delete 或 update
			} else {
				op = rng.IntN(4)
			}

			switch {
			case op == 0: // insert
				key := nextKey
				nextKey++
				score := randScore()
				ss.Insert(NewNodeData[int, int](key, score, key))
				ref.insert(key, score)
				keyPool = append(keyPool, key)

			case op == 1: // delete
				idx := rng.IntN(poolSize)
				key := keyPool[idx]
				ss.Delete(key)
				ref.delete(key)
				// swap-remove：O(1) 删除，不保留 keyPool 顺序（无需有序）
				keyPool[idx] = keyPool[poolSize-1]
				keyPool = keyPool[:poolSize-1]

			default: // update
				key := keyPool[rng.IntN(poolSize)]
				newScore := randScore()
				ss.UpdateScore(key, newScore)
				ref.updateScore(key, newScore)
			}
		}

		// 每轮结束验证全量不变量
		checkInvariants(t, ss, ref)
		if t.Failed() {
			t.Fatalf("round=%d 不变量检查失败，终止", round)
		}
	}
}
