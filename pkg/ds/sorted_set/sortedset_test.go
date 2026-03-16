package sorted_set

import (
	"fmt"
	"testing"
)

// ---- 辅助函数 ----

// newFilledSet 创建并填充 n 个元素（key=i, score=float64(i)）的有序集合
func newFilledSet(n int) *SortedSet[int, int] {
	ss := NewSortedSet[int, int]()
	for i := 1; i <= n; i++ {
		ss.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}
	return ss
}

// ---- 构造函数 ----

func TestNewSortedSet(t *testing.T) {
	ss := NewSortedSet[int, int]()
	if ss == nil {
		t.Fatal("NewSortedSet() 返回了 nil")
	}
	if ss.Length() != 0 {
		t.Errorf("新有序集合的长度应该是 0，实际是 %d", ss.Length())
	}
}

func TestNewNodeData(t *testing.T) {
	nd := NewNodeData(1, 3.14, 42)
	if nd == nil {
		t.Fatal("NewNodeData() 返回了 nil")
	}
	if nd.Key != 1 || nd.Score != 3.14 || nd.Val != 42 {
		t.Errorf("字段不正确: Key=%d Score=%f Val=%d", nd.Key, nd.Score, nd.Val)
	}
}

// ---- Insert ----

func TestSortedSet_Insert(t *testing.T) {
	ss := NewSortedSet[int, int]()

	for i, item := range []*NodeData[int, int]{
		{Key: 1, Score: 3.0, Val: 1},
		{Key: 2, Score: 1.0, Val: 2},
		{Key: 3, Score: 2.0, Val: 3},
	} {
		if !ss.Insert(item) {
			t.Errorf("第 %d 个元素插入失败", i+1)
		}
		if ss.Length() != i+1 {
			t.Errorf("插入后长度期望 %d，实际 %d", i+1, ss.Length())
		}
	}

	// 重复 Key 应拒绝（不覆盖）
	if ss.Insert(&NodeData[int, int]{Key: 1, Score: 99.0, Val: 99}) {
		t.Error("重复 Key 插入应返回 false")
	}
	if ss.Get(1).Score != 3.0 {
		t.Error("重复插入不应修改原有分数")
	}
}

// TestSortedSet_ReinsertAfterDelete 删除后同 Key 应能重新插入
func TestSortedSet_ReinsertAfterDelete(t *testing.T) {
	ss := newFilledSet(3)

	ss.Delete(2)
	if ss.Length() != 2 {
		t.Fatalf("删除后长度期望 2，实际 %d", ss.Length())
	}

	// 用新 score 重新插入
	if !ss.Insert(&NodeData[int, int]{Key: 2, Score: 99.0, Val: 2}) {
		t.Fatal("删除后重插相同 Key 应成功")
	}
	if ss.Length() != 3 {
		t.Errorf("重插后长度期望 3，实际 %d", ss.Length())
	}
	if ss.GetRank(2) != 3 {
		t.Errorf("score=99 应排最后（rank=3），实际 rank=%d", ss.GetRank(2))
	}
}

// ---- Get ----

func TestSortedSet_Get(t *testing.T) {
	ss := NewSortedSet[string, string]()
	items := []*NodeData[string, string]{
		{Key: "a", Score: 1.0, Val: "va"},
		{Key: "b", Score: 2.0, Val: "vb"},
	}
	for _, item := range items {
		ss.Insert(item)
	}

	for _, item := range items {
		if got := ss.Get(item.Key); got == nil || got.Key != item.Key {
			t.Errorf("Get(%s) 期望找到，实际 %v", item.Key, got)
		}
	}
	if ss.Get("nonexistent") != nil {
		t.Error("不存在的 Key 应返回 nil")
	}
}

// ---- Delete ----

func TestSortedSet_Delete(t *testing.T) {
	ss := newFilledSet(3)

	d, ok := ss.Delete(2)
	if !ok || d.Key != 2 {
		t.Errorf("删除 key=2 应成功，got ok=%v key=%v", ok, d)
	}
	if ss.Length() != 2 {
		t.Errorf("删除后长度期望 2，实际 %d", ss.Length())
	}
	if ss.Get(2) != nil {
		t.Error("删除后 Get 应返回 nil")
	}

	if _, ok := ss.Delete(999); ok {
		t.Error("不存在的 Key 删除应返回 false")
	}
}

// TestSortedSet_DeleteAll 删光所有元素后集合应完全为空
func TestSortedSet_DeleteAll(t *testing.T) {
	const n = 5
	ss := newFilledSet(n)

	for i := 1; i <= n; i++ {
		if _, ok := ss.Delete(i); !ok {
			t.Errorf("删除 key=%d 失败", i)
		}
	}
	if ss.Length() != 0 {
		t.Errorf("删完后长度期望 0，实际 %d", ss.Length())
	}
	// 再插入应正常工作
	if !ss.Insert(&NodeData[int, int]{Key: 1, Score: 1.0, Val: 1}) {
		t.Error("清空后再插入应成功")
	}
}

// ---- GetRank ----

func TestSortedSet_GetRank(t *testing.T) {
	ss := NewSortedSet[int, int]()
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 3.0, Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 2, Score: 1.0, Val: 2})
	ss.Insert(&NodeData[int, int]{Key: 3, Score: 2.0, Val: 3})

	for key, want := range map[int]int{2: 1, 3: 2, 1: 3} {
		if got := ss.GetRank(key); got != want {
			t.Errorf("GetRank(%d) 期望 %d，实际 %d", key, want, got)
		}
	}
	if ss.GetRank(999) != 0 {
		t.Error("不存在的 Key 应返回 0")
	}
}

// ---- GetByRank ----

func TestSortedSet_GetByRank(t *testing.T) {
	ss := NewSortedSet[string, string]()
	ss.Insert(&NodeData[string, string]{Key: "a", Score: 3.0, Val: "a"})
	ss.Insert(&NodeData[string, string]{Key: "b", Score: 1.0, Val: "b"})
	ss.Insert(&NodeData[string, string]{Key: "c", Score: 2.0, Val: "c"})

	for i, want := range []string{"b", "c", "a"} {
		rank := i + 1
		got := ss.GetByRank(rank)
		if got == nil || got.Key != want {
			t.Errorf("GetByRank(%d) 期望 %s，实际 %v", rank, want, got)
		}
	}
	if ss.GetByRank(999) != nil {
		t.Error("超出范围应返回 nil")
	}
}

// ---- GetRangeByRank ----

func TestSortedSet_GetRangeByRank(t *testing.T) {
	ss := newFilledSet(10)

	// 正常范围 [2,5] → 4 个元素，升序
	result := ss.GetRangeByRank(2, 5)
	if len(result) != 4 {
		t.Fatalf("[2,5] 期望 4 个，实际 %d", len(result))
	}
	for i := 0; i < len(result)-1; i++ {
		if result[i].Score > result[i+1].Score {
			t.Errorf("结果应升序，位置 %d: %f > %f", i, result[i].Score, result[i+1].Score)
		}
	}

	// start > end 自动交换，结果与 [2,5] 相同
	result2 := ss.GetRangeByRank(5, 2)
	if len(result2) != len(result) {
		t.Errorf("start>end 自动交换：期望 %d 个，实际 %d", len(result), len(result2))
	}
	for i := range result2 {
		if result2[i].Key != result[i].Key {
			t.Errorf("start>end 结果不一致，位置 %d", i)
		}
	}
}

// TestSortedSet_GetRangeByRank_BeyondLength end 超出总长度应截断返回，不 panic
func TestSortedSet_GetRangeByRank_BeyondLength(t *testing.T) {
	ss := newFilledSet(5)

	result := ss.GetRangeByRank(3, 100) // 只有 5 个元素
	if len(result) != 3 {               // rank 3,4,5
		t.Errorf("end 超界应截断，期望 3 个，实际 %d", len(result))
	}
	if result[0].Key != 3 || result[2].Key != 5 {
		t.Errorf("截断后元素不正确：%v", result)
	}
}

// ---- UpdateScore ----

func TestSortedSet_UpdateScore(t *testing.T) {
	ss := newFilledSet(3) // key=1 score=1, key=2 score=2, key=3 score=3

	// 将 key=1 更新到最大分，应移到末尾
	if updated, ok := ss.UpdateScore(1, 10.0); !ok || updated.Score != 10.0 {
		t.Fatalf("UpdateScore 失败: ok=%v score=%v", ok, updated)
	}
	if ss.GetRank(1) != 3 {
		t.Errorf("更新到最大分后应 rank=3，实际 %d", ss.GetRank(1))
	}
	if ss.GetRank(2) != 1 {
		t.Errorf("key=2 应升为 rank=1，实际 %d", ss.GetRank(2))
	}

	if _, ok := ss.UpdateScore(999, 0); ok {
		t.Error("不存在的 Key 应返回 false")
	}
}

// TestSortedSet_UpdateScore_SameValue 更新为相同分数应视为无操作，排名不变
func TestSortedSet_UpdateScore_SameValue(t *testing.T) {
	ss := newFilledSet(3)

	rankBefore := ss.GetRank(2)
	if _, ok := ss.UpdateScore(2, 2.0); !ok {
		t.Fatal("更新为相同分数应成功")
	}
	if ss.GetRank(2) != rankBefore {
		t.Errorf("更新相同分数后 rank 不应变化：before=%d after=%d", rankBefore, ss.GetRank(2))
	}
	if ss.Length() != 3 {
		t.Errorf("更新相同分数后长度应不变，实际 %d", ss.Length())
	}
}

// TestSortedSet_UpdateScore_EqualNeighbor 更新后分数与相邻节点相同，应进入删除重插路径
func TestSortedSet_UpdateScore_EqualNeighbor(t *testing.T) {
	ss := newFilledSet(3) // key=1 score=1, key=2 score=2, key=3 score=3

	// key=2 分数改为与前驱 key=1 相同 → 触发删除重插
	if _, ok := ss.UpdateScore(2, 1.0); !ok {
		t.Fatal("更新为前驱相同分数应成功")
	}
	if ss.Length() != 3 {
		t.Errorf("更新后长度应不变，实际 %d", ss.Length())
	}
	// key=1 先插入（seq 更小），key=2 后插入（seq 更大），相同分数下 key=1 排前
	if ss.GetRank(1) != 1 || ss.GetRank(2) != 2 {
		t.Errorf("相同分数时应按 seq 排序：rank(1)=%d rank(2)=%d", ss.GetRank(1), ss.GetRank(2))
	}

	// key=2 分数改为与后继 key=3 相同 → 触发删除重插
	ss = newFilledSet(3)
	if _, ok := ss.UpdateScore(2, 3.0); !ok {
		t.Fatal("更新为后继相同分数应成功")
	}
	// key=2 seq 小于 key=3，相同分数下排前
	if ss.GetRank(2) != 2 || ss.GetRank(3) != 3 {
		t.Errorf("相同分数时应按 seq 排序：rank(2)=%d rank(3)=%d", ss.GetRank(2), ss.GetRank(3))
	}
}

// ---- GetRangeByScore ----

func TestSortedSet_GetRangeByScore(t *testing.T) {
	ss := newFilledSet(10)

	result := ss.GetRangeByScore(3.0, false, 7.0, false)
	if len(result) != 5 {
		t.Errorf("[3,7] 期望 5 个，实际 %d", len(result))
	}
	for _, item := range result {
		if item.Score < 3.0 || item.Score > 7.0 {
			t.Errorf("元素 score=%f 不在 [3,7]", item.Score)
		}
	}
}

func TestSortedSet_GetRangeByScore_Exclusive(t *testing.T) {
	ss := newFilledSet(10)

	cases := []struct {
		minEx, maxEx bool
		want         int
		desc         string
	}{
		{false, false, 5, "[3,7]"},
		{true, false, 4, "(3,7]"},
		{false, true, 4, "[3,7)"},
		{true, true, 3, "(3,7)"},
	}
	for _, c := range cases {
		result := ss.GetRangeByScore(3.0, c.minEx, 7.0, c.maxEx)
		if len(result) != c.want {
			t.Errorf("%s 期望 %d 个，实际 %d", c.desc, c.want, len(result))
		}
	}
}

// TestSortedSet_GetRangeByScore_EmptyResult 范围内无元素应返回空切片，不 panic
func TestSortedSet_GetRangeByScore_EmptyResult(t *testing.T) {
	ss := newFilledSet(10) // score 1~10

	result := ss.GetRangeByScore(50.0, false, 100.0, false)
	if result == nil || len(result) != 0 {
		t.Errorf("范围外应返回空切片，实际 %v", result)
	}
}

// TestSortedSet_GetRangeByScore_SinglePoint 单点查询 [x,x]
func TestSortedSet_GetRangeByScore_SinglePoint(t *testing.T) {
	ss := newFilledSet(10)

	// [5,5] 含两端 → 1 个元素
	result := ss.GetRangeByScore(5.0, false, 5.0, false)
	if len(result) != 1 || result[0].Score != 5.0 {
		t.Errorf("[5,5] 期望 1 个 score=5，实际 %v", result)
	}

	// (5,5) 两端排他 → 空（isInRange 直接拒绝）
	result = ss.GetRangeByScore(5.0, true, 5.0, true)
	if len(result) != 0 {
		t.Errorf("(5,5) 期望空，实际 %v", result)
	}

	// [5,5) 右排他 → 空
	result = ss.GetRangeByScore(5.0, false, 5.0, true)
	if len(result) != 0 {
		t.Errorf("[5,5) 期望空，实际 %v", result)
	}
}

// ---- DeleteRangeByScore ----

func TestSortedSet_DeleteRangeByScore(t *testing.T) {
	cases := []struct {
		minEx, maxEx bool
		wantDel      int
		desc         string
	}{
		{false, false, 5, "[3,7]"},
		{true, false, 4, "(3,7]"},
		{false, true, 4, "[3,7)"},
	}
	for _, c := range cases {
		ss := newFilledSet(10)
		deleted := ss.DeleteRangeByScore(3.0, c.minEx, 7.0, c.maxEx)
		if len(deleted) != c.wantDel {
			t.Errorf("%s 期望删除 %d 个，实际 %d", c.desc, c.wantDel, len(deleted))
		}
		if ss.Length() != 10-c.wantDel {
			t.Errorf("%s 删除后长度期望 %d，实际 %d", c.desc, 10-c.wantDel, ss.Length())
		}
		for _, d := range deleted {
			if ss.Get(d.Key) != nil {
				t.Errorf("%s 已删除的 key=%d 不应再存在", c.desc, d.Key)
			}
		}
	}
}

// ---- DeleteRangeByRank ----

func TestSortedSet_DeleteRangeByRank(t *testing.T) {
	ss := newFilledSet(10)
	deleted := ss.DeleteRangeByRank(3, 6)
	if len(deleted) != 4 {
		t.Errorf("期望删除 4 个，实际 %d", len(deleted))
	}
	if ss.Length() != 6 {
		t.Errorf("删除后长度期望 6，实际 %d", ss.Length())
	}
	for _, d := range deleted {
		if ss.Get(d.Key) != nil {
			t.Errorf("已删除的 key=%d 不应再存在", d.Key)
		}
	}

	// start > end 自动交换
	ss2 := newFilledSet(5)
	if deleted2 := ss2.DeleteRangeByRank(4, 2); len(deleted2) != 3 {
		t.Errorf("start>end 自动交换：期望删除 3 个，实际 %d", len(deleted2))
	}
}

// ---- 负数 Score ----

// TestSortedSet_NegativeScore 负数分数应排在正数前面（rank=1 为最小值）
func TestSortedSet_NegativeScore(t *testing.T) {
	ss := NewSortedSet[int, int]()
	ss.Insert(&NodeData[int, int]{Key: 1, Score: -100.0, Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 2, Score: 0.0, Val: 2})
	ss.Insert(&NodeData[int, int]{Key: 3, Score: 100.0, Val: 3})

	if ss.GetRank(1) != 1 {
		t.Errorf("score=-100 应 rank=1，实际 %d", ss.GetRank(1))
	}
	if ss.GetRank(2) != 2 {
		t.Errorf("score=0 应 rank=2，实际 %d", ss.GetRank(2))
	}
	if ss.GetRank(3) != 3 {
		t.Errorf("score=100 应 rank=3，实际 %d", ss.GetRank(3))
	}
	if ss.GetByRank(1).Key != 1 {
		t.Errorf("rank=1 应为 key=1（score=-100），实际 %d", ss.GetByRank(1).Key)
	}
}

// ---- 空集合边界 ----

// TestSortedSet_EmptySet_Operations 空集合上所有查询应返回空/零值，不 panic
func TestSortedSet_EmptySet_Operations(t *testing.T) {
	ss := NewSortedSet[int, int]()

	if ss.Get(1) != nil {
		t.Error("空集合 Get 应返回 nil")
	}
	if _, ok := ss.Delete(1); ok {
		t.Error("空集合 Delete 应返回 false")
	}
	if ss.GetRank(1) != 0 {
		t.Error("空集合 GetRank 应返回 0")
	}
	if ss.GetByRank(1) != nil {
		t.Error("空集合 GetByRank 应返回 nil")
	}
	if result := ss.GetRangeByRank(1, 10); len(result) != 0 {
		t.Errorf("空集合 GetRangeByRank 应返回空切片，实际 %v", result)
	}
	if result := ss.GetRangeByScore(0, false, 100, false); len(result) != 0 {
		t.Errorf("空集合 GetRangeByScore 应返回空切片，实际 %v", result)
	}
	if result := ss.DeleteRangeByRank(1, 1); len(result) != 0 {
		t.Errorf("空集合 DeleteRangeByRank 应返回空切片，实际 %v", result)
	}
	if result := ss.DeleteRangeByScore(0, false, 100, false); len(result) != 0 {
		t.Errorf("空集合 DeleteRangeByScore 应返回空切片，实际 %v", result)
	}
	if _, ok := ss.UpdateScore(1, 1.0); ok {
		t.Error("空集合 UpdateScore 应返回 false")
	}
}

// ---- 同分数稳定排序 ----

func TestSortedSet_SameScore_StableOrder(t *testing.T) {
	ss := NewSortedSet[int, int]()
	for _, key := range []int{10, 20, 30, 40, 50} {
		ss.Insert(&NodeData[int, int]{Key: key, Score: 1.0, Val: key})
	}
	result := ss.GetRangeByRank(1, 5)
	for i := 0; i < len(result)-1; i++ {
		if result[i].seq >= result[i+1].seq {
			t.Errorf("相同分数应按插入顺序稳定排序，seq[%d]=%d >= seq[%d]=%d",
				i, result[i].seq, i+1, result[i+1].seq)
		}
	}
}

// ---- Key/Val 类型分离 ----

func TestSortedSet_KeyValDifferentTypes(t *testing.T) {
	type Player struct{ Name string }

	ss := NewSortedSet[string, Player]()
	ss.Insert(NewNodeData("alice", 1500.0, Player{"Alice"}))
	ss.Insert(NewNodeData("bob", 1200.0, Player{"Bob"}))

	got := ss.Get("bob")
	if got == nil || got.Val.Name != "Bob" {
		t.Errorf("Val 类型独立于 Key：期望 Bob，实际 %v", got)
	}
	if ss.GetRank("bob") != 1 {
		t.Errorf("score=1200 应 rank=1，实际 %d", ss.GetRank("bob"))
	}
}

// ---- Benchmark ----
//
// 使用 b.Run 按规模分组，测试者可通过以下方式控制：
//   - go test -bench=. -benchtime=3s         // 每个子测试运行 3 秒
//   - go test -bench=BenchmarkSortedSet_Get/n=10000  // 只跑特定规模
//   - go test -bench=. -count=3              // 每个基准重复 3 次

var benchSizes = []int{100, 1000, 10000}

func BenchmarkSortedSet_Insert(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss := NewSortedSet[int, int]()
				for j := 0; j < size; j++ {
					ss.Insert(&NodeData[int, int]{Key: j, Score: float64(j), Val: j})
				}
			}
		})
	}
}

func BenchmarkSortedSet_Get(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss.Get(i % size)
			}
		})
	}
}

func BenchmarkSortedSet_GetRank(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss.GetRank(i%size + 1)
			}
		})
	}
}

func BenchmarkSortedSet_UpdateScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss.UpdateScore(i%size+1, float64(i))
			}
		})
	}
}

func BenchmarkSortedSet_GetRangeByRank(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := size / 2
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss.GetRangeByRank(1, mid)
			}
		})
	}
}

func BenchmarkSortedSet_GetRangeByScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := float64(size / 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss.GetRangeByScore(1, false, mid, false)
			}
		})
	}
}

func BenchmarkSortedSet_Delete(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ss := newFilledSet(size)
				b.StartTimer()
				for j := 1; j <= size; j++ {
					ss.Delete(j)
				}
			}
		})
	}
}
