package sorted_set

import (
	"math"
	"testing"
)

// ---- 辅助函数 ----

// newFilledSet 创建并填充 n 个元素（key=i, score=float64(i)）的有序集合
func newFilledSet(n int) *SortedSet[int, int] {
	ss := NewSortedSet[int, int]()
	for i := 1; i <= n; i++ {
		ss.Insert(NewNodeData[int, int](i, float64(i), i))
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

// ---- GetRankDesc / GetByRankDesc / GetRangeByRankDesc ----

func TestSortedSet_GetRankDesc(t *testing.T) {
	ss := NewSortedSet[int, int]()
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 3.0, Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 2, Score: 1.0, Val: 2})
	ss.Insert(&NodeData[int, int]{Key: 3, Score: 2.0, Val: 3})
	// 升序：key2(1.0) key3(2.0) key1(3.0)
	// 降序：key1(3.0)=rank1  key3(2.0)=rank2  key2(1.0)=rank3
	for key, want := range map[int]int{1: 1, 3: 2, 2: 3} {
		if got := ss.GetRankDesc(key); got != want {
			t.Errorf("GetRankDesc(%d) 期望 %d，实际 %d", key, want, got)
		}
	}
	if ss.GetRankDesc(999) != 0 {
		t.Error("不存在的 Key 应返回 0")
	}
}

func TestSortedSet_GetByRankDesc(t *testing.T) {
	ss := NewSortedSet[string, string]()
	ss.Insert(&NodeData[string, string]{Key: "a", Score: 3.0, Val: "a"})
	ss.Insert(&NodeData[string, string]{Key: "b", Score: 1.0, Val: "b"})
	ss.Insert(&NodeData[string, string]{Key: "c", Score: 2.0, Val: "c"})
	// 降序：a(3.0) c(2.0) b(1.0)
	for i, want := range []string{"a", "c", "b"} {
		rank := i + 1
		got := ss.GetByRankDesc(rank)
		if got == nil || got.Key != want {
			t.Errorf("GetByRankDesc(%d) 期望 %s，实际 %v", rank, want, got)
		}
	}
	if ss.GetByRankDesc(999) != nil {
		t.Error("超出范围应返回 nil")
	}
}

func TestSortedSet_GetRangeByRankDesc(t *testing.T) {
	ss := newFilledSet(10) // key=i, score=i，升序 1..10

	// 降序前 3：score=10,9,8
	result := ss.GetRangeByRankDesc(1, 3)
	if len(result) != 3 {
		t.Fatalf("降序[1,3] 期望 3 个，实际 %d", len(result))
	}
	for i := 0; i < len(result)-1; i++ {
		if result[i].Score < result[i+1].Score {
			t.Errorf("结果应降序，位置 %d: %f < %f", i, result[i].Score, result[i+1].Score)
		}
	}
	if result[0].Key != 10 || result[2].Key != 8 {
		t.Errorf("降序前 3 应为 key=10,9,8，实际 %v", result)
	}

	// start > end 自动交换
	result2 := ss.GetRangeByRankDesc(3, 1)
	if len(result2) != len(result) || result2[0].Key != result[0].Key {
		t.Error("start>end 自动交换后结果应与 [1,3] 相同")
	}
}

// TestSortedSet_GetRangeByRankDesc_BeyondLength end 超出总长度应截断返回，不 panic
func TestSortedSet_GetRangeByRankDesc_BeyondLength(t *testing.T) {
	ss := newFilledSet(5)

	result := ss.GetRangeByRankDesc(3, 100)
	if len(result) != 3 { // desc rank 3,4,5 → score=3,2,1
		t.Errorf("end 超界应截断，期望 3 个，实际 %d", len(result))
	}
	if result[0].Key != 3 || result[2].Key != 1 {
		t.Errorf("截断后元素不正确：%v", result)
	}
}

// TestSortedSet_GetRangeByRankDesc_BothBeyondLength start 和 end 均超出总长度，期望返回空切片
func TestSortedSet_GetRangeByRankDesc_BothBeyondLength(t *testing.T) {
	ss := newFilledSet(5)

	// start=10, end=20，两端均超出（n=5）
	result := ss.GetRangeByRankDesc(10, 20)
	if len(result) != 0 {
		t.Errorf("两端均超界应返回空切片，实际 %d 个", len(result))
	}

	// DeleteRangeByRankDesc 同样场景
	deleted := ss.DeleteRangeByRankDesc(10, 20)
	if len(deleted) != 0 {
		t.Errorf("两端均超界删除应返回空切片，实际 %d 个", len(deleted))
	}
	if ss.Length() != 5 {
		t.Errorf("无效范围删除不应改变集合长度，实际 %d", ss.Length())
	}
}

// TestSortedSet_DescAscRankSymmetry 逆序 rank 与正序 rank 之和应等于 Length+1
func TestSortedSet_DescAscRankSymmetry(t *testing.T) {
	ss := newFilledSet(10)
	n := ss.Length()
	for key := 1; key <= n; key++ {
		asc := ss.GetRank(key)
		desc := ss.GetRankDesc(key)
		if asc+desc != n+1 {
			t.Errorf("key=%d asc=%d desc=%d，期望和为 %d", key, asc, desc, n+1)
		}
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
		{true, true, 3, "(3,7)"},
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

// ---- DeleteRangeByRankDesc ----

func TestSortedSet_DeleteRangeByRankDesc(t *testing.T) {
	// 删除降序前 3（score 最大的 3 个）
	ss := newFilledSet(10)
	deleted := ss.DeleteRangeByRankDesc(1, 3)
	if len(deleted) != 3 {
		t.Fatalf("期望删除 3 个，实际 %d", len(deleted))
	}
	// 返回结果应降序
	for i := 0; i < len(deleted)-1; i++ {
		if deleted[i].Score < deleted[i+1].Score {
			t.Errorf("结果应降序，位置 %d: %f < %f", i, deleted[i].Score, deleted[i+1].Score)
		}
	}
	// score=10,9,8 被删除
	if deleted[0].Key != 10 || deleted[2].Key != 8 {
		t.Errorf("应删除 key=10,9,8，实际 %v", deleted)
	}
	if ss.Length() != 7 {
		t.Errorf("删除后长度期望 7，实际 %d", ss.Length())
	}
	for _, d := range deleted {
		if ss.Get(d.Key) != nil {
			t.Errorf("已删除的 key=%d 不应再存在", d.Key)
		}
	}

	// start > end 自动交换
	ss2 := newFilledSet(5)
	if d2 := ss2.DeleteRangeByRankDesc(3, 1); len(d2) != 3 {
		t.Errorf("start>end 自动交换：期望删除 3 个，实际 %d", len(d2))
	}
}

// TestSortedSet_DeleteRangeByRankDesc_BeyondLength end 超出总长度应截断返回，不 panic
func TestSortedSet_DeleteRangeByRankDesc_BeyondLength(t *testing.T) {
	ss := newFilledSet(5)
	deleted := ss.DeleteRangeByRankDesc(3, 100)
	if len(deleted) != 3 {
		t.Errorf("end 超界应截断，期望 3 个，实际 %d", len(deleted))
	}
	if ss.Length() != 2 {
		t.Errorf("删除后长度期望 2，实际 %d", ss.Length())
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

// TestSortedSet_DeleteRangeByRank_BeyondLength end 超出总长度应截断返回，不 panic
func TestSortedSet_DeleteRangeByRank_BeyondLength(t *testing.T) {
	ss := newFilledSet(5)
	deleted := ss.DeleteRangeByRank(3, 100)
	if len(deleted) != 3 { // rank 3,4,5
		t.Errorf("end 超界应截断，期望 3 个，实际 %d", len(deleted))
	}
	if ss.Length() != 2 {
		t.Errorf("删除后长度期望 2，实际 %d", ss.Length())
	}
	for _, d := range deleted {
		if ss.Get(d.Key) != nil {
			t.Errorf("已删除的 key=%d 不应再存在", d.Key)
		}
	}
}

// TestSortedSet_NaN_Rejected NaN Score 应触发 assert panic，不允许写入集合
func TestSortedSet_NaN_Rejected(t *testing.T) {
	ss := NewSortedSet[int, int]()

	// Insert NaN 应 panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Insert NaN Score 应触发 assert panic")
			}
		}()
		ss.Insert(&NodeData[int, int]{Key: 1, Score: math.NaN(), Val: 1})
	}()

	// 集合应仍为空（NaN 被拒绝）
	if ss.Length() != 0 {
		t.Errorf("NaN 被拒绝后集合应为空，实际 Length=%d", ss.Length())
	}

	// UpdateScore NaN 应 panic
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 1.0, Val: 1})
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("UpdateScore NaN 应触发 assert panic")
			}
		}()
		ss.UpdateScore(1, math.NaN())
	}()

	// 原有数据应不受影响
	if got := ss.Get(1); got == nil || got.Score != 1.0 {
		t.Errorf("UpdateScore NaN panic 后原数据应不变，实际 %v", got)
	}

	// GetRangeByScore NaN min 应 panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("GetRangeByScore NaN min 应触发 assert panic")
			}
		}()
		ss.GetRangeByScore(math.NaN(), false, 10.0, false)
	}()

	// DeleteRangeByScore NaN max 应 panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("DeleteRangeByScore NaN max 应触发 assert panic")
			}
		}()
		ss.DeleteRangeByScore(0, false, math.NaN(), false)
	}()
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
	// Desc API
	if ss.GetRankDesc(1) != 0 {
		t.Error("空集合 GetRankDesc 应返回 0")
	}
	if ss.GetByRankDesc(1) != nil {
		t.Error("空集合 GetByRankDesc 应返回 nil")
	}
	if result := ss.GetRangeByRankDesc(1, 10); len(result) != 0 {
		t.Errorf("空集合 GetRangeByRankDesc 应返回空切片，实际 %v", result)
	}
	if result := ss.DeleteRangeByRankDesc(1, 10); len(result) != 0 {
		t.Errorf("空集合 DeleteRangeByRankDesc 应返回空切片，实际 %v", result)
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

// ---- GetMin / GetMax ----

func TestSortedSet_GetMin(t *testing.T) {
	ss := NewSortedSet[int, int]()
	if ss.GetMin() != nil {
		t.Error("空集合 GetMin 应返回 nil")
	}

	ss.Insert(&NodeData[int, int]{Key: 3, Score: 30.0, Val: 3})
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 10.0, Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 2, Score: 20.0, Val: 2})

	got := ss.GetMin()
	if got == nil || got.Key != 1 || got.Score != 10.0 {
		t.Errorf("GetMin 应返回 score=10 的元素，实际 %v", got)
	}

	// 单元素
	ss2 := NewSortedSet[int, int]()
	ss2.Insert(&NodeData[int, int]{Key: 7, Score: 5.0, Val: 7})
	if got2 := ss2.GetMin(); got2 == nil || got2.Key != 7 {
		t.Errorf("单元素 GetMin 应返回该元素，实际 %v", got2)
	}
}

func TestSortedSet_GetMax(t *testing.T) {
	ss := NewSortedSet[int, int]()
	if ss.GetMax() != nil {
		t.Error("空集合 GetMax 应返回 nil")
	}

	ss.Insert(&NodeData[int, int]{Key: 3, Score: 30.0, Val: 3})
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 10.0, Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 2, Score: 20.0, Val: 2})

	got := ss.GetMax()
	if got == nil || got.Key != 3 || got.Score != 30.0 {
		t.Errorf("GetMax 应返回 score=30 的元素，实际 %v", got)
	}

	// 单元素
	ss2 := NewSortedSet[int, int]()
	ss2.Insert(&NodeData[int, int]{Key: 7, Score: 5.0, Val: 7})
	if got2 := ss2.GetMax(); got2 == nil || got2.Key != 7 {
		t.Errorf("单元素 GetMax 应返回该元素，实际 %v", got2)
	}
}

// TestSortedSet_GetMin_SameScore Score 相同时 GetMin 应返回最先插入（seq 最小）的元素
func TestSortedSet_GetMin_SameScore(t *testing.T) {
	ss := NewSortedSet[int, int]()
	// 按顺序插入，score 均为 1.0，seq 依次递增
	ss.Insert(&NodeData[int, int]{Key: 10, Score: 1.0, Val: 10}) // seq 最小，应被 GetMin 返回
	ss.Insert(&NodeData[int, int]{Key: 20, Score: 1.0, Val: 20})
	ss.Insert(&NodeData[int, int]{Key: 30, Score: 1.0, Val: 30})
	// 混入更大 score，确保 GetMin 取的是 score=1.0 中 seq 最小的
	ss.Insert(&NodeData[int, int]{Key: 99, Score: 2.0, Val: 99})

	got := ss.GetMin()
	if got == nil || got.Key != 10 {
		t.Errorf("同分数时 GetMin 应返回最先插入的 key=10，实际 %v", got)
	}
}

// TestSortedSet_GetMax_SameScore Score 相同时 GetMax 应返回最后插入（seq 最大）的元素
func TestSortedSet_GetMax_SameScore(t *testing.T) {
	ss := NewSortedSet[int, int]()
	// 混入更小 score，确保 GetMax 取的是 score=5.0 中 seq 最大的
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 1.0, Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 10, Score: 5.0, Val: 10})
	ss.Insert(&NodeData[int, int]{Key: 20, Score: 5.0, Val: 20})
	ss.Insert(&NodeData[int, int]{Key: 30, Score: 5.0, Val: 30}) // seq 最大，应被 GetMax 返回

	got := ss.GetMax()
	if got == nil || got.Key != 30 {
		t.Errorf("同分数时 GetMax 应返回最后插入的 key=30，实际 %v", got)
	}
}

// ---- CountByScore ----

func TestSortedSet_CountByScore(t *testing.T) {
	ss := newFilledSet(10) // score=1..10

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
		got := ss.CountByScore(3.0, c.minEx, 7.0, c.maxEx)
		if got != c.want {
			t.Errorf("%s CountByScore 期望 %d，实际 %d", c.desc, c.want, got)
		}
		// 与 len(GetRangeByScore) 保持一致
		wantFromGet := len(ss.GetRangeByScore(3.0, c.minEx, 7.0, c.maxEx))
		if got != wantFromGet {
			t.Errorf("%s CountByScore=%d 与 len(GetRangeByScore)=%d 不一致", c.desc, got, wantFromGet)
		}
	}

	// 空集合
	empty := NewSortedSet[int, int]()
	if empty.CountByScore(0, false, 100, false) != 0 {
		t.Error("空集合 CountByScore 应返回 0")
	}

	// 范围外无元素
	if ss.CountByScore(50, false, 100, false) != 0 {
		t.Error("范围外 CountByScore 应返回 0")
	}
}

// ---- 单元素 ----

// TestSortedSet_SingleElement 只含 1 个元素时，所有操作应正常返回，不 panic
func TestSortedSet_SingleElement(t *testing.T) {
	ss := NewSortedSet[int, int]()
	ss.Insert(&NodeData[int, int]{Key: 1, Score: 5.0, Val: 1})

	if got := ss.Get(1); got == nil || got.Score != 5.0 {
		t.Error("单元素 Get 应返回元素")
	}
	if ss.GetRank(1) != 1 {
		t.Errorf("单元素 GetRank 应为 1，实际 %d", ss.GetRank(1))
	}
	if got := ss.GetByRank(1); got == nil || got.Key != 1 {
		t.Error("单元素 GetByRank(1) 应返回元素")
	}
	if ss.GetByRank(2) != nil {
		t.Error("单元素 GetByRank(2) 应返回 nil")
	}
	if result := ss.GetRangeByRank(1, 1); len(result) != 1 {
		t.Errorf("单元素 GetRangeByRank(1,1) 应返回 1 个，实际 %d", len(result))
	}
	if result := ss.GetRangeByScore(5.0, false, 5.0, false); len(result) != 1 {
		t.Errorf("单元素 GetRangeByScore[5,5] 应返回 1 个，实际 %d", len(result))
	}
	if _, ok := ss.UpdateScore(1, 10.0); !ok {
		t.Error("单元素 UpdateScore 应成功")
	}
	if ss.GetRank(1) != 1 {
		t.Error("单元素更新分数后 rank 仍应为 1")
	}
	// Desc API：单元素时降序 rank == 升序 rank == 1
	if ss.GetRankDesc(1) != 1 {
		t.Errorf("单元素 GetRankDesc 应为 1，实际 %d", ss.GetRankDesc(1))
	}
	if got := ss.GetByRankDesc(1); got == nil || got.Key != 1 {
		t.Error("单元素 GetByRankDesc(1) 应返回元素")
	}
	if ss.GetByRankDesc(2) != nil {
		t.Error("单元素 GetByRankDesc(2) 应返回 nil")
	}
	if result := ss.GetRangeByRankDesc(1, 1); len(result) != 1 || result[0].Key != 1 {
		t.Errorf("单元素 GetRangeByRankDesc(1,1) 应返回 1 个元素，实际 %v", result)
	}
	if _, ok := ss.Delete(1); !ok {
		t.Error("单元素 Delete 应成功")
	}
	if ss.Length() != 0 {
		t.Errorf("删除后 Length 应为 0，实际 %d", ss.Length())
	}
}

// ---- 越界负数 rank ----

// TestSortedSet_GetByRankDesc_NegativeAndZero rank 为 0 或负数是合约违反，assert 开启时应 panic
func TestSortedSet_GetByRankDesc_NegativeAndZero(t *testing.T) {
	ss := newFilledSet(5)

	for _, rank := range []int{0, -1, -100} {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("GetByRankDesc(%d) 应因 assert 失败而 panic", rank)
				}
			}()
			ss.GetByRankDesc(rank)
		}()
	}
}

// TestSortedSet_GetByRank_NegativeAndZero rank 为 0 或负数是合约违反，assert 开启时应 panic
func TestSortedSet_GetByRank_NegativeAndZero(t *testing.T) {
	ss := newFilledSet(5)

	for _, rank := range []int{0, -1, -100} {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("GetByRank(%d) 应因 assert 失败而 panic", rank)
				}
			}()
			ss.GetByRank(rank)
		}()
	}
}

// ---- 特殊浮点数 Score ----

// TestSortedSet_InfScore ±Inf 作为 score 应正常插入、排序、查询，不 panic
func TestSortedSet_InfScore(t *testing.T) {
	ss := NewSortedSet[int, int]()
	ss.Insert(&NodeData[int, int]{Key: 1, Score: math.Inf(-1), Val: 1})
	ss.Insert(&NodeData[int, int]{Key: 2, Score: 0.0, Val: 2})
	ss.Insert(&NodeData[int, int]{Key: 3, Score: math.Inf(1), Val: 3})

	if ss.Length() != 3 {
		t.Fatalf("插入 ±Inf score 后 Length 应为 3，实际 %d", ss.Length())
	}
	// -Inf 最小，rank=1
	if ss.GetRank(1) != 1 {
		t.Errorf("score=-Inf 应 rank=1，实际 %d", ss.GetRank(1))
	}
	// +Inf 最大，rank=3
	if ss.GetRank(3) != 3 {
		t.Errorf("score=+Inf 应 rank=3，实际 %d", ss.GetRank(3))
	}
	if got := ss.GetByRank(1); got == nil || got.Key != 1 {
		t.Errorf("rank=1 应为 key=1(score=-Inf)，实际 %v", got)
	}
	if got := ss.GetByRank(3); got == nil || got.Key != 3 {
		t.Errorf("rank=3 应为 key=3(score=+Inf)，实际 %v", got)
	}
	// GetRangeByScore 含 -Inf 端点
	result := ss.GetRangeByScore(math.Inf(-1), false, 0.0, false)
	if len(result) != 2 {
		t.Errorf("[-Inf,0] 应返回 2 个，实际 %d", len(result))
	}
}

// benchSizes 及所有 BenchmarkSortedSet_* 已迁移至 bench_test.go。

// TestSortedSet_StressOps 已迁移至 correctness_test.go。
