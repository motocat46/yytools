// Package sorted_set.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 有序集合

// 作者:  yangyuan
// 创建日期:2023/7/3
package sorted_set

import (
	"math"
	
	"github.com/motocat46/yytools/pkg/common/assert"
)

// SortedSet 有序集合，基于跳表 + 哈希表实现。
//
// 跳表负责按 Score 维护有序结构，支持 O(log n) 的排名查询和范围操作；
// 哈希表负责 Key → NodeData 的 O(1) 映射，支持按 Key 直接查找和删除。
// 非并发安全，多 goroutine 访问需自行加锁。
type SortedSet[K comparable, V any] struct {
	sl   *SkipList[K, V]
	hash map[K]*NodeData[K, V]
	seq  uint64 // 自增序列号，每次 Insert 递增，赋给新节点的 seq 字段
}

// NewSortedSet 创建一个空的有序集合。
func NewSortedSet[K comparable, V any]() *SortedSet[K, V] {
	return &SortedSet[K, V]{
		sl:   NewSkipList[K, V](),
		hash: make(map[K]*NodeData[K, V]),
	}
}

// Get 按 Key 查找元素，O(1)。不存在时返回 nil。
func (this *SortedSet[K, V]) Get(key K) *NodeData[K, V] {
	return this.hash[key]
}

// Insert 插入元素，O(log n)。
// Key 已存在时返回 false，原有元素不受影响。
// 插入时自动分配 seq，相同 Score 的元素按插入顺序稳定排列。
// 注意：不要通过 Delete + Insert 代替 UpdateScore，
// 重新插入会分配新的 seq，破坏相同 Score 下原有的稳定顺序。
func (this *SortedSet[K, V]) Insert(data *NodeData[K, V]) bool {
	assert.Assert(data != nil, "data == nil")
	assert.Assert(!math.IsNaN(data.Score), "Score 不能为 NaN：NaN 的比较语义（NaN != NaN）会破坏跳表全序")
	
	if _, has := this.hash[data.Key]; has {
		return false
	}
	
	this.seq++
	data.seq = this.seq
	_, ok := this.sl.Insert(data)
	assert.Assert(ok, "insert must succeed, data.Key:", data.Key)
	if ok {
		this.hash[data.Key] = data
	}
	this.lengthMustEqual()
	return ok
}

// Delete 按 Key 删除元素，O(log n)。
// 删除成功时返回被删除的节点数据和 true；Key 不存在时返回 (nil, false)。
func (this *SortedSet[K, V]) Delete(key K) (*NodeData[K, V], bool) {
	data, exist := this.hash[key]
	if !exist {
		return nil, false
	}
	
	if node, ok := this.sl.Delete(data); ok {
		delete(this.hash, key)
		this.lengthMustEqual()
		return node.Data, ok
	}
	return nil, false
}

// Length 返回集合中的元素总数，O(1)。
func (this *SortedSet[K, V]) Length() int {
	return this.sl.Length
}

// lengthMustEqual 内部一致性断言：跳表长度必须与哈希表长度相等。
func (this *SortedSet[K, V]) lengthMustEqual() {
	assert.Assert(this.sl.Length == len(this.hash),
		"长度不一致 skiplist length:", this.sl.Length, " hash length:", len(this.hash))
}

// GetMin 返回排序最靠前的元素，O(1)。集合为空时返回 nil。
// Score 最小；Score 相同时取 seq 最小（最先插入）的元素。
// 直接读跳表 Head.Levels[0].Forward，无需遍历索引层。
func (this *SortedSet[K, V]) GetMin() *NodeData[K, V] {
	first := this.sl.Head.Levels[0].Forward
	if first == nil {
		return nil
	}
	return first.Data
}

// GetMax 返回排序最靠后的元素，O(1)。集合为空时返回 nil。
// Score 最大；Score 相同时取 seq 最大（最后插入）的元素。
// 直接读跳表 Tail 指针，无需遍历索引层。
func (this *SortedSet[K, V]) GetMax() *NodeData[K, V] {
	if this.sl.Tail == nil {
		return nil
	}
	return this.sl.Tail.Data
}

// GetRank 查询 key 的升序排名，O(log n)。
// 排名从 1 开始，rank=1 对应 Score 最小的元素。
// Key 不存在时返回 0。
func (this *SortedSet[K, V]) GetRank(key K) int {
	data, exist := this.hash[key]
	if !exist {
		return 0
	}
	rank := this.sl.GetRank(data)
	assert.Assert(rank != 0, "rank must exist")
	return rank
}

// GetByRank 按升序排名取元素，O(log n)。
// rank=1 对应 Score 最小的元素。
// rank 超出 [1, Length] 范围时返回 nil；rank <= 0 触发断言 panic。
func (this *SortedSet[K, V]) GetByRank(rank int) *NodeData[K, V] {
	assert.Assert(rank > 0, "rank must be positive number")
	node := this.sl.GetNodeByRank(rank)
	if node == nil {
		return nil
	}
	return node.Data
}

// GetRankDesc 查询 key 的降序排名，O(log n)。
// rank=1 对应 Score 最大的元素，适合排行榜场景（高分靠前）。
// Key 不存在时返回 0。
// 与 GetRank 的关系：GetRankDesc(key) + GetRank(key) == Length + 1。
func (this *SortedSet[K, V]) GetRankDesc(key K) int {
	r := this.GetRank(key)
	if r == 0 {
		return 0
	}
	return this.Length() - r + 1
}

// GetByRankDesc 按降序排名取元素，O(log n)。
// rank=1 对应 Score 最大的元素。
// rank 超出 [1, Length] 范围时返回 nil；rank <= 0 触发断言 panic。
func (this *SortedSet[K, V]) GetByRankDesc(rank int) *NodeData[K, V] {
	assert.Assert(rank > 0, "rank must be positive number")
	ascRank := this.Length() - rank + 1
	if ascRank <= 0 {
		return nil
	}
	return this.GetByRank(ascRank)
}

// GetRangeByRankDesc 返回降序排名范围 [start, end] 内的元素，O(log n + k)，k 为返回元素数。
// 结果按 Score 从大到小排列，rank=1 为 Score 最大的元素。
// start > end 时自动交换；end 超出总长度时截断返回，不 panic。
func (this *SortedSet[K, V]) GetRangeByRankDesc(start int, end int) []*NodeData[K, V] {
	if start > end {
		start, end = end, start
	}
	n := this.Length()
	if n == 0 {
		return []*NodeData[K, V]{}
	}
	ascStart := n - end + 1
	ascEnd := n - start + 1
	if ascStart < 1 {
		ascStart = 1
	}
	if ascEnd < 1 {
		return []*NodeData[K, V]{}
	}
	result := this.GetRangeByRank(ascStart, ascEnd)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// GetRangeByRank 返回升序排名范围 [start, end] 内的元素，O(log n + k)，k 为返回元素数。
// 结果按 Score 从小到大排列。
// start > end 时自动交换；start ≤ 0 时截断为 1；end 超出总长度时截断返回，不 panic。
func (this *SortedSet[K, V]) GetRangeByRank(start int, end int) []*NodeData[K, V] {
	if start > end {
		start, end = end, start
	}
	if start < 1 {
		start = 1
	}
	return this.sl.GetRangeByRank(start, end)
}

// DeleteRangeByRank 删除升序排名范围 [start, end] 内的所有元素，O(log n + k)，k 为删除元素数。
// 返回被删除的元素列表。start > end 时自动交换；start ≤ 0 时截断为 1。
func (this *SortedSet[K, V]) DeleteRangeByRank(start int, end int) []*NodeData[K, V] {
	if start > end {
		start, end = end, start
	}
	if start < 1 {
		start = 1
	}
	deleted := this.sl.DeleteRangeByRank(start, end)
	for _, one := range deleted {
		delete(this.hash, one.Key)
	}
	this.lengthMustEqual()
	return deleted
}

// DeleteRangeByRankDesc 删除降序排名范围 [start, end] 内的所有元素，O(log n + k)，k 为删除元素数。
// 返回被删除的元素列表，按 Score 从大到小排列。
// start > end 时自动交换；end 超出总长度时截断，不 panic。
func (this *SortedSet[K, V]) DeleteRangeByRankDesc(start int, end int) []*NodeData[K, V] {
	if start > end {
		start, end = end, start
	}
	n := this.Length()
	if n == 0 {
		return []*NodeData[K, V]{}
	}
	ascStart := n - end + 1
	ascEnd := n - start + 1
	if ascStart < 1 {
		ascStart = 1
	}
	if ascEnd < 1 {
		return []*NodeData[K, V]{}
	}
	deleted := this.DeleteRangeByRank(ascStart, ascEnd)
	for i, j := 0, len(deleted)-1; i < j; i, j = i+1, j-1 {
		deleted[i], deleted[j] = deleted[j], deleted[i]
	}
	return deleted
}

// UpdateScore 更新 key 的分数并触发重新排序，O(log n)。
// 返回更新后的节点数据和 true；Key 不存在时返回 (nil, false)。
// 相比 Delete + Insert，UpdateScore 保留原有的 seq，
// 相同 Score 下的稳定排序不受影响。
// 若新分数未改变节点的相对位置（前驱 < newScore < 后继），则原地更新，无需删除重插。
func (this *SortedSet[K, V]) UpdateScore(key K, newScore float64) (*NodeData[K, V], bool) {
	assert.Assert(!math.IsNaN(newScore), "newScore 不能为 NaN：NaN 的比较语义（NaN != NaN）会破坏跳表全序")
	data, exist := this.hash[key]
	if !exist {
		return nil, false
	}
	node, ok := this.sl.UpdateScore(data, newScore)
	if !ok {
		return nil, ok
	}
	this.lengthMustEqual()
	return node.Data, ok
}

// GetRangeByScore 返回 Score 在 [min, max] 范围内的所有元素，O(log n + k)，k 为返回元素数。
// 结果按 Score 从小到大排列。
// minEx=true 表示排除 min 端点（即 (min, ...]），maxEx=true 表示排除 max 端点（即 [..., max)）。
// 范围内无元素时返回空切片，不 panic。
func (this *SortedSet[K, V]) GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V] {
	assert.Assert(!math.IsNaN(min), "min 不能为 NaN：NaN 的比较语义会导致范围查询结果未定义")
	assert.Assert(!math.IsNaN(max), "max 不能为 NaN：NaN 的比较语义会导致范围查询结果未定义")
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{MinExclusive: minEx, MaxExclusive: maxEx},
		Min:                min,
		Max:                max,
	}
	return this.sl.GetRangeByScore(r)
}

// CountByScore 统计 Score 在 [min, max] 范围内的元素数量，O(log n + k)，k 为命中元素数。
// minEx/maxEx 语义与 GetRangeByScore 相同。
// 与 len(GetRangeByScore(...)) 等价，但只需 1 次结构体分配，避免结果切片的多次分配，适合高频计数场景。
func (this *SortedSet[K, V]) CountByScore(min float64, minEx bool, max float64, maxEx bool) int {
	assert.Assert(!math.IsNaN(min), "min 不能为 NaN：NaN 的比较语义会导致范围计数结果未定义")
	assert.Assert(!math.IsNaN(max), "max 不能为 NaN：NaN 的比较语义会导致范围计数结果未定义")
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{MinExclusive: minEx, MaxExclusive: maxEx},
		Min:                min,
		Max:                max,
	}
	return this.sl.CountByScore(r)
}

// DeleteRangeByScore 删除 Score 在 [min, max] 范围内的所有元素，O(log n + k)，k 为删除元素数。
// 返回被删除的元素列表。
// minEx/maxEx 语义与 GetRangeByScore 相同。
func (this *SortedSet[K, V]) DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V] {
	assert.Assert(!math.IsNaN(min), "min 不能为 NaN：NaN 的比较语义会导致范围删除结果未定义")
	assert.Assert(!math.IsNaN(max), "max 不能为 NaN：NaN 的比较语义会导致范围删除结果未定义")
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{MinExclusive: minEx, MaxExclusive: maxEx},
		Min:                min,
		Max:                max,
	}
	deleted := this.sl.DeleteRangeByScore(r)
	for _, one := range deleted {
		delete(this.hash, one.Key)
	}
	this.lengthMustEqual()
	return deleted
}