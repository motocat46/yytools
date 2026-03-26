package sorted_set

import (
	"fmt"
	"math"
	"math/rand/v2"
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

// ---- 随机混合操作压力测试 ----

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

// ---- Benchmark ----
//
// 运行方式：
//   go test -bench=. -benchmem ./pkg/ds/sorted_set/
//   go test -bench=BenchmarkSortedSet_Mixed/n=10000 -benchtime=5s ./pkg/ds/sorted_set/
//   go test -bench=. -count=3 ./pkg/ds/sorted_set/

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// --- 单操作基准（稳定集合规模，测单次 O(log n) 代价）---
//
// 所有单操作基准均在预填充的稳定集合上运行，集合规模在整个基准过程中保持不变，
// 确保每次迭代的操作代价反映的是"规模为 n 时的成本"，而非建集合的摊销成本。

// BenchmarkSortedSet_Get 测量 O(1) 哈希查找
func BenchmarkSortedSet_Get(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.Get(i%size + 1)
			}
		})
	}
}

// BenchmarkSortedSet_GetRank 测量 O(log n) 排名查询
func BenchmarkSortedSet_GetRank(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.GetRank(i%size + 1)
			}
		})
	}
}

// BenchmarkSortedSet_UpdateScore 测量 O(log n) 分数更新
// UpdateScore 不改变集合大小，天然适合稳定集合测量
func BenchmarkSortedSet_UpdateScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// 使用随机 score 触发真实的重排序路径（删除重插）
				ss.UpdateScore(i%size+1, float64((i*7+3)%size))
			}
		})
	}
}

// BenchmarkSortedSet_Insert 测量 O(log n) 插入
// 每次插入后删除 rank=1 的元素以维持集合规模，防止 n 随 b.N 增长导致后期成本失真
func BenchmarkSortedSet_Insert(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			newKey := size + 1
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				oldest := ss.GetByRank(1)
				ss.Delete(oldest.Key)
				ss.Insert(&NodeData[int, int]{Key: newKey, Score: float64(newKey % size), Val: newKey})
				newKey++
			}
		})
	}
}

// BenchmarkSortedSet_GetRangeByRank 测量 O(log n + k) 排名范围查询（k = n/2）
func BenchmarkSortedSet_GetRangeByRank(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := size / 2
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.GetRangeByRank(1, mid)
			}
		})
	}
}

// BenchmarkSortedSet_GetRangeByScore 测量 O(log n + k) 分数范围查询（k ≈ n/2）
func BenchmarkSortedSet_GetRangeByScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := float64(size / 2)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.GetRangeByScore(1, false, mid, false)
			}
		})
	}
}

// BenchmarkSortedSet_GetMin 测量 O(1) 最小值访问（Head.Levels[0].Forward）
func BenchmarkSortedSet_GetMin(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.GetMin()
			}
		})
	}
}

// BenchmarkSortedSet_GetMax 测量 O(1) 最大值访问（Tail 指针）
func BenchmarkSortedSet_GetMax(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.GetMax()
			}
		})
	}
}

// BenchmarkSortedSet_CountByScore 测量 O(log n + k) 范围计数（k ≈ n/2，无切片分配）
// 与 BenchmarkSortedSet_GetRangeByScore 对比，可量化零分配带来的收益
func BenchmarkSortedSet_CountByScore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			mid := float64(size / 2)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ss.CountByScore(1, false, mid, false)
			}
		})
	}
}

// --- 混合负载基准（模拟真实排行榜场景）---
//
// BenchmarkSortedSet_Mixed 在稳定规模的集合上混合执行多种操作，
// 模拟排行榜典型负载：
//   50% UpdateScore（玩家分数更新，最高频）
//   25% GetRank（查询自己排名）
//   15% GetRangeByRank(1,10)（查看排行榜前 10）
//    5% GetRangeByScore（按积分段查询）
//    5% Insert+Delete（玩家进出场，维持集合规模）
//
// 该基准反映的是整体吞吐量，而非单一操作的延迟。
func BenchmarkSortedSet_Mixed(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			ss := newFilledSet(size)
			rng := rand.New(rand.NewPCG(42, 0))
			nextKey := size + 1
			scoreRange := float64(size * 10)
			
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				key := rng.IntN(size) + 1
				op := rng.IntN(100)
				switch {
				case op < 50: // UpdateScore 50%
					ss.UpdateScore(key, rng.Float64()*scoreRange)
				case op < 75: // GetRank 25%
					ss.GetRank(key)
				case op < 90: // GetRangeByRank(1,10) 15%
					ss.GetRangeByRank(1, 10)
				case op < 95: // GetRangeByScore 5%
					lo := rng.Float64() * scoreRange / 2
					ss.GetRangeByScore(lo, false, lo+scoreRange/4, false)
				default: // Insert+Delete 5%（维持规模）
					oldest := ss.GetByRank(1)
					if oldest != nil {
						ss.Delete(oldest.Key)
					}
					ss.Insert(&NodeData[int, int]{
						Key:   nextKey,
						Score: rng.Float64() * scoreRange,
						Val:   nextKey,
					})
					nextKey++
				}
			}
		})
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
