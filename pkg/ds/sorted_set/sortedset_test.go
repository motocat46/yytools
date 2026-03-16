package sorted_set

import (
	"testing"
)

func TestNewSortedSet(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()
	if sortedSet == nil {
		t.Fatal("NewSortedSet() 返回了 nil")
	}
	if sortedSet.Length() != 0 {
		t.Errorf("新有序集合的长度应该是 0，实际是 %d", sortedSet.Length())
	}
}

func TestNewNodeData(t *testing.T) {
	key := 1
	score := 3.14
	val := 42

	nodeData := NewNodeData(key, score, val)
	if nodeData == nil {
		t.Fatal("NewNodeData() 返回了 nil")
	}
	if nodeData.Key != key {
		t.Errorf("期望键 %d，实际是 %d", key, nodeData.Key)
	}
	if nodeData.Score != score {
		t.Errorf("期望分数 %f，实际是 %f", score, nodeData.Score)
	}
	if nodeData.Val != val {
		t.Errorf("期望值 %d，实际是 %d", val, nodeData.Val)
	}
}

func TestSortedSet_Insert(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()

	items := []*NodeData[int, int]{
		{Key: 1, Score: 3.0, Val: 1},
		{Key: 2, Score: 1.0, Val: 2},
		{Key: 3, Score: 2.0, Val: 3},
	}

	for i, item := range items {
		success := sortedSet.Insert(item)
		if !success {
			t.Errorf("插入第 %d 个元素失败", i)
		}
		if sortedSet.Length() != i+1 {
			t.Errorf("插入后长度应该是 %d，实际是 %d", i+1, sortedSet.Length())
		}
	}

	// 重复插入应该失败
	duplicateItem := &NodeData[int, int]{Key: 1, Score: 4.0, Val: 1}
	success := sortedSet.Insert(duplicateItem)
	if success {
		t.Error("重复插入应该失败")
	}
}

func TestSortedSet_Get(t *testing.T) {
	sortedSet := NewSortedSet[string, string]()

	items := []*NodeData[string, string]{
		{Key: "key1", Score: 1.0, Val: "val1"},
		{Key: "key2", Score: 2.0, Val: "val2"},
		{Key: "key3", Score: 3.0, Val: "val3"},
	}

	for _, item := range items {
		sortedSet.Insert(item)
	}

	for _, item := range items {
		retrieved := sortedSet.Get(item.Key)
		if retrieved == nil {
			t.Errorf("键 %s 应该存在", item.Key)
		}
		if retrieved.Key != item.Key {
			t.Errorf("期望键 %s，实际是 %s", item.Key, retrieved.Key)
		}
	}

	notFound := sortedSet.Get("nonexistent")
	if notFound != nil {
		t.Error("不存在的键应该返回 nil")
	}
}

func TestSortedSet_Delete(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()

	items := []*NodeData[int, int]{
		{Key: 1, Score: 1.0, Val: 1},
		{Key: 2, Score: 2.0, Val: 2},
		{Key: 3, Score: 3.0, Val: 3},
	}

	for _, item := range items {
		sortedSet.Insert(item)
	}

	deleted, success := sortedSet.Delete(2)
	if !success {
		t.Error("删除存在的键应该成功")
	}
	if deleted.Key != 2 {
		t.Errorf("期望删除键 2，实际删除键 %d", deleted.Key)
	}
	if sortedSet.Length() != 2 {
		t.Errorf("删除后长度应该是 2，实际是 %d", sortedSet.Length())
	}

	_, success = sortedSet.Delete(999)
	if success {
		t.Error("删除不存在的键应该失败")
	}
}

func TestSortedSet_GetRank(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()

	items := []*NodeData[int, int]{
		{Key: 1, Score: 3.0, Val: 1}, // 排名 3
		{Key: 2, Score: 1.0, Val: 2}, // 排名 1
		{Key: 3, Score: 2.0, Val: 3}, // 排名 2
	}

	for _, item := range items {
		sortedSet.Insert(item)
	}

	expectedRanks := map[int]int{
		1: 3,
		2: 1,
		3: 2,
	}

	for key, expectedRank := range expectedRanks {
		rank := sortedSet.GetRank(key)
		if rank != expectedRank {
			t.Errorf("键 %d 期望排名 %d，实际排名 %d", key, expectedRank, rank)
		}
	}

	rank := sortedSet.GetRank(999)
	if rank != 0 {
		t.Errorf("不存在的键应该返回排名 0，实际返回 %d", rank)
	}
}

func TestSortedSet_GetByRank(t *testing.T) {
	sortedSet := NewSortedSet[string, string]()

	items := []*NodeData[string, string]{
		{Key: "key1", Score: 3.0, Val: "val1"},
		{Key: "key2", Score: 1.0, Val: "val2"},
		{Key: "key3", Score: 2.0, Val: "val3"},
	}

	for _, item := range items {
		sortedSet.Insert(item)
	}

	expectedKeys := []string{"key2", "key3", "key1"}

	for i, expectedKey := range expectedKeys {
		rank := i + 1
		nodeData := sortedSet.GetByRank(rank)
		if nodeData == nil {
			t.Errorf("排名 %d 应该存在数据", rank)
		}
		if nodeData.Key != expectedKey {
			t.Errorf("排名 %d 期望键 %s，实际是 %s", rank, expectedKey, nodeData.Key)
		}
	}

	nodeData := sortedSet.GetByRank(999)
	if nodeData != nil {
		t.Error("超出范围的排名应该返回 nil")
	}
}

func TestSortedSet_GetRangeByRank(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()
	for i := 1; i <= 10; i++ {
		sortedSet.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}

	// 正常范围
	result := sortedSet.GetRangeByRank(2, 5)
	if len(result) != 4 {
		t.Errorf("期望获取 4 个元素，实际获取 %d 个", len(result))
	}
	for i := 0; i < len(result)-1; i++ {
		if result[i].Score > result[i+1].Score {
			t.Errorf("结果应按分数升序，但 %f > %f", result[i].Score, result[i+1].Score)
		}
	}

	// start > end 自动交换，结果应与 (2,5) 相同
	result2 := sortedSet.GetRangeByRank(5, 2)
	if len(result2) != 4 {
		t.Errorf("start>end 自动交换：期望 4 个元素，实际 %d", len(result2))
	}
	for i, item := range result2 {
		if item.Key != result[i].Key {
			t.Errorf("start>end 结果与正常顺序不一致，位置 %d: %d != %d", i, item.Key, result[i].Key)
		}
	}
}

func TestSortedSet_UpdateScore(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()

	sortedSet.Insert(&NodeData[int, int]{Key: 1, Score: 1.0, Val: 1}) // 初始 rank=1
	sortedSet.Insert(&NodeData[int, int]{Key: 2, Score: 2.0, Val: 2}) // 初始 rank=2
	sortedSet.Insert(&NodeData[int, int]{Key: 3, Score: 3.0, Val: 3}) // 初始 rank=3

	// 将 key=1 的分数更新到最大，应排到末尾
	updated, success := sortedSet.UpdateScore(1, 10.0)
	if !success {
		t.Error("更新存在的键应该成功")
	}
	if updated.Score != 10.0 {
		t.Errorf("期望分数 10.0，实际是 %f", updated.Score)
	}
	if rank := sortedSet.GetRank(1); rank != 3 {
		t.Errorf("更新到最大分数后应排名第 3，实际是 %d", rank)
	}
	if rank := sortedSet.GetRank(2); rank != 1 {
		t.Errorf("key=2 应升为排名第 1，实际是 %d", rank)
	}

	_, success = sortedSet.UpdateScore(999, 10.0)
	if success {
		t.Error("更新不存在的键应该失败")
	}
}

func TestSortedSet_GetRangeByScore(t *testing.T) {
	sortedSet := NewSortedSet[int, int]()
	for i := 1; i <= 10; i++ {
		sortedSet.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}

	result := sortedSet.GetRangeByScore(3.0, false, 7.0, false)
	if len(result) != 5 {
		t.Errorf("期望获取 5 个元素，实际获取 %d 个", len(result))
	}

	for _, item := range result {
		if item.Score < 3.0 || item.Score > 7.0 {
			t.Errorf("元素分数 %f 不在范围 [3.0, 7.0] 内", item.Score)
		}
	}
}

func TestSortedSet_DeleteRangeByScore(t *testing.T) {
	newSet := func() *SortedSet[int, int] {
		ss := NewSortedSet[int, int]()
		for i := 1; i <= 10; i++ {
			ss.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
		}
		return ss
	}

	// [3, 7] 含两端：score=3,4,5,6,7 共 5 个
	ss := newSet()
	deleted := ss.DeleteRangeByScore(3.0, false, 7.0, false)
	if len(deleted) != 5 {
		t.Errorf("[3,7] 期望删除 5 个，实际 %d", len(deleted))
	}
	if ss.Length() != 5 {
		t.Errorf("[3,7] 删除后长度期望 5，实际 %d", ss.Length())
	}

	// (3, 7] 不含左端：score=4,5,6,7 共 4 个
	ss = newSet()
	deleted = ss.DeleteRangeByScore(3.0, true, 7.0, false)
	if len(deleted) != 4 {
		t.Errorf("(3,7] 期望删除 4 个，实际 %d", len(deleted))
	}

	// [3, 7) 不含右端：score=3,4,5,6 共 4 个
	ss = newSet()
	deleted = ss.DeleteRangeByScore(3.0, false, 7.0, true)
	if len(deleted) != 4 {
		t.Errorf("[3,7) 期望删除 4 个，实际 %d", len(deleted))
	}

	// 验证已删元素确实从集合中移除
	for _, d := range deleted {
		if ss.Get(d.Key) != nil {
			t.Errorf("已删除的 key=%d 不应再存在", d.Key)
		}
	}
}

func TestSortedSet_WithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	// Key 和 Val 使用不同类型：Key=Person（查找用），Val=string（业务数据）
	sortedSet := NewSortedSet[Person, string]()

	persons := []*NodeData[Person, string]{
		{Key: Person{"Alice", 25}, Score: 3.0, Val: "alice_data"},
		{Key: Person{"Bob", 30}, Score: 1.0, Val: "bob_data"},
		{Key: Person{"Charlie", 35}, Score: 2.0, Val: "charlie_data"},
	}

	for _, person := range persons {
		sortedSet.Insert(person)
	}

	retrieved := sortedSet.Get(Person{"Bob", 30})
	if retrieved == nil {
		t.Error("应该能找到 Bob")
	}
	if retrieved.Key.Name != "Bob" {
		t.Errorf("期望找到 Bob，实际找到 %s", retrieved.Key.Name)
	}
	if retrieved.Val != "bob_data" {
		t.Errorf("期望 Val=bob_data，实际是 %s", retrieved.Val)
	}
}

func TestSortedSet_DeleteRangeByRank(t *testing.T) {
	ss := NewSortedSet[int, int]()
	for i := 1; i <= 10; i++ {
		ss.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}

	// 删除排名 3~6
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
	ss2 := NewSortedSet[int, int]()
	for i := 1; i <= 5; i++ {
		ss2.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}
	deleted2 := ss2.DeleteRangeByRank(4, 2)
	if len(deleted2) != 3 {
		t.Errorf("start>end 自动交换：期望删除 3 个，实际 %d", len(deleted2))
	}
}

func TestSortedSet_GetRangeByScore_Exclusive(t *testing.T) {
	ss := NewSortedSet[int, int]()
	for i := 1; i <= 10; i++ {
		ss.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}

	// 含左含右 [3, 7]
	result := ss.GetRangeByScore(3.0, false, 7.0, false)
	if len(result) != 5 {
		t.Errorf("[3,7] 期望 5 个，实际 %d", len(result))
	}

	// 不含左 (3, 7]
	result = ss.GetRangeByScore(3.0, true, 7.0, false)
	if len(result) != 4 {
		t.Errorf("(3,7] 期望 4 个，实际 %d", len(result))
	}

	// 不含右 [3, 7)
	result = ss.GetRangeByScore(3.0, false, 7.0, true)
	if len(result) != 4 {
		t.Errorf("[3,7) 期望 4 个，实际 %d", len(result))
	}

	// 两端均不含 (3, 7)
	result = ss.GetRangeByScore(3.0, true, 7.0, true)
	if len(result) != 3 {
		t.Errorf("(3,7) 期望 3 个，实际 %d", len(result))
	}
}

func TestSortedSet_SameScore_StableOrder(t *testing.T) {
	ss := NewSortedSet[int, int]()
	for _, key := range []int{10, 20, 30, 40, 50} {
		ss.Insert(&NodeData[int, int]{Key: key, Score: 1.0, Val: key})
	}
	result := ss.GetRangeByRank(1, 5)
	for i := 0; i < len(result)-1; i++ {
		if result[i].seq >= result[i+1].seq {
			t.Errorf("相同分数应按插入顺序稳定排序，但 seq[%d]=%d >= seq[%d]=%d",
				i, result[i].seq, i+1, result[i+1].seq)
		}
	}
}

func BenchmarkSortedSet_Insert(b *testing.B) {
	sortedSet := NewSortedSet[int, int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sortedSet.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}
}

func BenchmarkSortedSet_Get(b *testing.B) {
	sortedSet := NewSortedSet[int, int]()
	for i := 0; i < 1000; i++ {
		sortedSet.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedSet.Get(i % 1000)
	}
}

func BenchmarkSortedSet_Delete(b *testing.B) {
	sortedSet := NewSortedSet[int, int]()
	for i := 0; i < b.N; i++ {
		sortedSet.Insert(&NodeData[int, int]{Key: i, Score: float64(i), Val: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortedSet.Delete(i)
	}
}
