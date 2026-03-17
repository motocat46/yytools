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

// 参考自redis的跳跃表实现

// 作者:  yangyuan
// 创建日期:2023/6/2
package sorted_set

import (
	"math"

	"github.com/motocat46/yytools/pkg/common/assert"
)

// SkipList 跳表，按 NodeData 的 (Score, seq) 二元组升序维护节点。
//
// 结构：Head 是哨兵节点（不存储业务数据），其后连接所有真实节点。
// Head 拥有 SKIPLIST_MAXLEVEL 层，每层维护一个 Forward 指针和 Span。
// Level 记录当前跳表实际使用的最大层数，随插入/删除动态调整。
// Tail 指向底层（level 0）最后一个节点，用于 O(1) 检查最大值。
type SkipList[K comparable, V any] struct {
	Head        *Node[K, V]
	Tail        *Node[K, V]
	Length      int
	Level       int
	LevelUpProb float32
}

// SkipListLevel 跳表节点某一层的索引信息。
//
// Forward 是该层的前向指针，指向同层下一个节点。
// Span 是该层 Forward 指针跨越的底层（level 0）节点数，包含 Forward 指向的节点本身。
// 例如：Head.Levels[2].Span = 5 表示从 Head 沿第 2 层跳一步可跨过 5 个底层节点。
// Span 用于在 O(log n) 内计算任意节点的排名（累加沿途的 Span 即得 rank）。
type SkipListLevel[K comparable, V any] struct {
	Forward *Node[K, V]
	Span    int
}

// Node 跳表节点。
//
// Levels 存储该节点在每一层的索引（Forward + Span），长度即节点高度。
// Backward 是底层（level 0）的反向指针，指向前一个节点，用于 UpdateScore 时检查前驱分数。
// Data 指向存储业务数据的 NodeData。
type Node[K comparable, V any] struct {
	Levels   []*SkipListLevel[K, V]
	Backward *Node[K, V]
	Data     *NodeData[K, V]
}

// RangeSpecifiedBase 分数范围查询的端点排他性配置。
// MinExclusive=true 表示排除 Min 端点（开区间），false 表示包含（闭区间）；MaxExclusive 同理。
type RangeSpecifiedBase struct {
	MinExclusive bool
	MaxExclusive bool
}

// RangeSpecified 完整的分数范围查询参数，包含端点值和排他性配置。
type RangeSpecified struct {
	RangeSpecifiedBase
	Min float64
	Max float64
}

// CreateNode 分配一个指定高度的跳表节点。
// level 为节点层数，data 为关联的业务数据（Head 哨兵节点传零值）。
func CreateNode[K comparable, V any](level int, data *NodeData[K, V]) *Node[K, V] {
	levelArr := make([]*SkipListLevel[K, V], level)
	for i := 0; i < level; i++ {
		levelArr[i] = &SkipListLevel[K, V]{Forward: nil, Span: 0}
	}
	return &Node[K, V]{Levels: levelArr, Backward: nil, Data: data}
}

// High 返回节点的层数（高度）。
func (this *Node[K, V]) High() int {
	return len(this.Levels)
}

// NewSkipList 使用默认晋升概率（0.25）创建跳表。
func NewSkipList[K comparable, V any]() *SkipList[K, V] {
	return NewSkipListByParams[K, V](DEFAULT_LEVELUP_PROBABILITY)
}

// NewSkipListByParams 使用指定晋升概率创建跳表。
// nodeLevelUpProb 为节点层数晋升概率，必须在 [0, 1) 范围内。
// 概率越高，平均节点层数越高，查询更快但内存占用更多；
// 默认值 0.25 是 Redis 的选择，平均节点高度约 1.33 层，空间与性能均衡。
func NewSkipListByParams[K comparable, V any](nodeLevelUpProb float32) *SkipList[K, V] {
	assert.Assert(nodeLevelUpProb >= 0 && nodeLevelUpProb < 1,
		"提升节点高度概率不正确:", nodeLevelUpProb, "正常范围:[0.0,1)")

	var zeroK K
	var zeroV V
	skipList := &SkipList[K, V]{
		Head: CreateNode[K, V](SKIPLIST_MAXLEVEL, &NodeData[K, V]{
			Score: 0,
			Key:   zeroK,
			Val:   zeroV,
		}),
		Tail:        nil,
		Length:      0,
		Level:       0,
		LevelUpProb: nodeLevelUpProb,
	}
	return skipList
}

// Get 按 (score, data) 精确查找节点，O(log n)。
// 内部使用 equalOrder（seq 比较）定位节点，score 用于快速跳过无关节点。
// 节点存在时返回 (*Node, true)，不存在时返回 (nil, false)。
func (this *SkipList[K, V]) Get(score float64, data *NodeData[K, V]) (*Node[K, V], bool) {
	assert.Assert(!math.IsNaN(score), "score is not a number:", score)
	assert.Assert(data != nil, "data must not be nil, score:", score)

	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		for current != nil && current.Data.lessOrder(data) {
			prev = current
			current = prev.Levels[i].Forward
		}
		if current != nil && current.Data.equalOrder(data) {
			return current, true
		}
	}
	return nil, false
}

// Insert 将 data 插入跳表，O(log n)。
//
// 算法分三阶段：
//  1. 从最高层向下遍历，记录每层插入点的前驱节点（prevNodes[i]）和
//     从 Head 到该前驱经过的底层节点数（rank[i]）；
//  2. 随机生成新节点高度，若超过当前最大层数则将 Head 作为新层的前驱，
//     初始 Span 设为 Length（新层跨越全部现有节点）；
//  3. 在每层链接新节点，并根据 rank 数组重新计算 Span：
//     - prevNodes[i].Span 更新为 rank[0]-rank[i]+1（到新节点的距离）
//     - newNode.Levels[i].Span 为原 prevNodes[i] Span 减去上述值（新节点到原后继的距离）
//     超出新节点高度的层只需将 prevNodes[i].Span 加 1（多了一个底层节点）。
//
// 节点已存在（equalOrder）时返回 (existingNode, false)，插入成功返回 (newNode, true)。
func (this *SkipList[K, V]) Insert(data *NodeData[K, V]) (*Node[K, V], bool) {
	assert.Assert(data != nil, "data must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[K, V]{}
	rank := [SKIPLIST_MAXLEVEL]int{}

	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		if i != this.Level-1 {
			rank[i] = rank[i+1]
		}
		current := prev.Levels[i].Forward
		for current != nil && current.Data.lessOrder(data) {
			rank[i] += prev.Levels[i].Span
			prev = current
			current = prev.Levels[i].Forward
		}
		if current != nil && current.Data.equalOrder(data) {
			return current, false
		}
		prevNodes[i] = prev
	}

	level := randomLevel(this.LevelUpProb)
	if level > this.Level {
		for i := this.Level; i < level; i++ {
			prevNodes[i] = this.Head
			prevNodes[i].Levels[i].Span = this.Length
		}
		this.Level = level
	}

	newNode := CreateNode[K, V](level, data)
	for i := 0; i < level; i++ {
		newNode.Levels[i].Forward = prevNodes[i].Levels[i].Forward
		prevNodes[i].Levels[i].Forward = newNode

		oldSpan := rank[0] - rank[i]
		newNode.Levels[i].Span = prevNodes[i].Levels[i].Span - oldSpan
		prevNodes[i].Levels[i].Span = oldSpan + 1
	}
	for i := level; i < this.Level; i++ {
		prevNodes[i].Levels[i].Span++
	}

	if prevNodes[0] == this.Head {
		newNode.Backward = nil
	} else {
		newNode.Backward = prevNodes[0]
	}

	if newNode.Levels[0].Forward != nil {
		newNode.Levels[0].Forward.Backward = newNode
	} else {
		this.Tail = newNode
	}
	this.Length++

	return newNode, true
}

// findNode 内部辅助：定位 data 对应的节点，同时记录每层的前驱节点到 prevNodes。
// 找到时返回 (node, true)，未找到返回 (nil, false)。
// prevNodes 由调用方传入，供后续 deleteNode 使用。
func (this *SkipList[K, V]) findNode(data *NodeData[K, V], prevNodes *[SKIPLIST_MAXLEVEL]*Node[K, V]) (*Node[K, V], bool) {
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		for current != nil && current.Data.lessOrder(data) {
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}
	current := prev.Levels[0].Forward
	if current != nil && current.Data.equalOrder(data) {
		return current, true
	}
	return nil, false
}

// deleteNode 内部辅助：从跳表中摘除 current 节点并维护 Span 和双向链表。
// prevNodes 必须是 findNode 或等效遍历记录的每层前驱。
// 每层处理：
//   - 若 prevNodes[i] 直接指向 current，则合并 Span（吸收 current 的 Span - 1）并跳过 current；
//   - 否则该层不经过 current，只将 Span 减 1（少了一个底层节点）。
//
// 删除后收缩 Level（去掉顶部空层），更新 Backward 和 Tail。
func (this *SkipList[K, V]) deleteNode(current *Node[K, V], prevNodes *[SKIPLIST_MAXLEVEL]*Node[K, V]) *Node[K, V] {
	for i := 0; i < this.Level; i++ {
		if prevNodes[i].Levels[i].Forward == current {
			prevNodes[i].Levels[i].Span += current.Levels[i].Span - 1
			prevNodes[i].Levels[i].Forward = current.Levels[i].Forward
		} else {
			prevNodes[i].Levels[i].Span--
		}
	}
	if current.Levels[0].Forward != nil {
		current.Levels[0].Forward.Backward = current.Backward
	} else {
		this.Tail = current.Backward
	}

	for this.Level > 1 && this.Head.Levels[this.Level-1].Forward == nil {
		this.Level--
	}
	this.Length--
	return current
}

// Delete 按 data 删除对应节点，O(log n)。
// 节点存在时返回 (deletedNode, true)，不存在时返回 (nil, false)。
func (this *SkipList[K, V]) Delete(data *NodeData[K, V]) (*Node[K, V], bool) {
	assert.Assert(data != nil, "val must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[K, V]{}
	current, ok := this.findNode(data, &prevNodes)
	if !ok {
		return current, ok
	}
	return this.deleteNode(current, &prevNodes), true
}

// GetRank 计算 data 在跳表中的升序排名（从 1 开始），O(log n)。
// 原理：从最高层向下遍历，凡推进一步则累加该层的 Span；
// 当找到目标节点（equalOrder）时，累计的 Span 之和即为其排名。
// data 不存在时返回 0。
func (this *SkipList[K, V]) GetRank(data *NodeData[K, V]) int {
	assert.Assert(data != nil, "val must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)

	rank := 0
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		for current != nil && (current.Data.lessOrder(data) || current.Data.equalOrder(data)) {
			rank += prev.Levels[i].Span
			prev = current
			current = prev.Levels[i].Forward
		}
		if prev != nil && prev.Data.equalOrder(data) {
			return rank
		}
	}
	return 0
}

// GetNodeByRank 按升序排名取节点，O(log n)。
// 原理：维护已遍历的底层节点计数（traversed），在每层尽可能向前推进，
// 直到再推一步会超过目标 rank；当 traversed == rank 时即到达目标节点。
// rank 超出 [1, Length] 范围时返回 nil；rank <= 0 触发断言 panic。
func (this *SkipList[K, V]) GetNodeByRank(rank int) *Node[K, V] {
	assert.Assert(rank > 0, "rank must >= 1, rank:", rank)

	traversed := 0
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		for current != nil && (traversed+prev.Levels[i].Span) <= rank {
			traversed += prev.Levels[i].Span
			prev = current
			current = prev.Levels[i].Forward
		}
		if traversed == rank {
			return prev
		}
	}
	return nil
}

// GetRangeByRank 返回升序排名范围 [start, end] 内的所有节点数据，O(log n + k)，k 为返回元素数。
// 先用 GetNodeByRank 定位到 start，再沿底层 Forward 顺序遍历至 end。
// start、end 必须满足 start > 0 && end > 0 && start <= end，否则触发断言 panic。
// end 超出 Length 时自然截断（current 变为 nil），不 panic。
func (this *SkipList[K, V]) GetRangeByRank(start int, end int) []*NodeData[K, V] {
	assert.Assert(start > 0 && end > 0 && start <= end, "rank范围不合法, start:", start, " end:", end)

	current := this.GetNodeByRank(start)
	traversed := start
	datas := make([]*NodeData[K, V], 0, 4)
	for current != nil && traversed <= end {
		next := current.Levels[0].Forward
		datas = append(datas, current.Data)
		traversed++
		current = next
	}
	return datas
}

// DeleteRangeByRank 删除升序排名范围 [start, end] 内的所有节点，O(log n + k)，k 为删除元素数。
// 先从顶层向下定位到 start 的前驱（条件为 < start，不含 start 本身），记录 prevNodes；
// 再沿底层顺序逐一删除 start 到 end 的节点，复用已记录的 prevNodes 避免重复定位。
// start、end 约束同 GetRangeByRank。
func (this *SkipList[K, V]) DeleteRangeByRank(start int, end int) []*NodeData[K, V] {
	assert.Assert(start > 0 && end > 0 && start <= end, "rank范围不合法, start:", start, " end:", end)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[K, V]{}
	traversed := 0
	prev := this.Head
	var current *Node[K, V] = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		for current != nil && (traversed+prev.Levels[i].Span) < start {
			traversed += prev.Levels[i].Span
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}
	traversed++

	deleted := make([]*NodeData[K, V], 0, 4)
	for current != nil && traversed <= end {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		traversed++
		current = next
	}
	return deleted
}

// UpdateScore 将 data 的分数更新为 newScore，O(log n)。
//
// 优先走原地更新路径：若新分数不改变节点相对位置
//
//	（前驱节点的 Score < newScore 且后继节点的 Score > newScore，或不存在前驱/后继），
//
// 则直接修改 Score 字段，无需删除重插。
// 否则删除节点后以 newScore 重新插入，seq 保持不变，相同分数下的稳定顺序不受影响。
// data 不存在时返回 (nil, false)。
func (this *SkipList[K, V]) UpdateScore(data *NodeData[K, V], newScore float64) (*Node[K, V], bool) {
	assert.Assert(data != nil, "data must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "oldScore is not a number:", data.Score)
	assert.Assert(!math.IsNaN(newScore), "newScore is not a number:", newScore)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[K, V]{}
	current, ok := this.findNode(data, &prevNodes)
	if !ok {
		return current, ok
	}

	// 如果新分数不改变结点位置，直接更新
	if (current.Backward == nil || current.Backward.Data.Score < newScore) &&
		(current.Levels[0].Forward == nil || current.Levels[0].Forward.Data.Score > newScore) {
		current.Data.Score = newScore
		return current, true
	}

	// 否则删除后重新插入
	this.deleteNode(current, &prevNodes)
	data.Score = newScore
	return this.Insert(data)
}

// scoreGreaterThanMin 判断 score 是否满足范围的 min 条件。
// MinExclusive=true 时要求严格大于 Min，否则大于等于。
func scoreGreaterThanMin(score float64, r *RangeSpecified) bool {
	if r.MinExclusive {
		return score > r.Min
	}
	return score >= r.Min
}

// scoreLessThanMax 判断 score 是否满足范围的 max 条件。
// MaxExclusive=true 时要求严格小于 Max，否则小于等于。
func scoreLessThanMax(score float64, r *RangeSpecified) bool {
	if r.MaxExclusive {
		return score < r.Max
	}
	return score <= r.Max
}

// isInRange 快速判断跳表中是否存在至少一个节点落在范围 r 内，O(1)。
// 用于 FirstInRange / GetRangeByScore 的前置剪枝，避免无谓的遍历。
// 判断逻辑：
//  1. 范围本身合法性：Min > Max，或 Min == Max 且任一端排他，则不可能有交集；
//  2. 跳表最大值（Tail）< Min，所有节点都在范围左侧；
//  3. 跳表最小值（Head 后第一个节点）> Max，所有节点都在范围右侧。
func (this *SkipList[K, V]) isInRange(r *RangeSpecified) bool {
	if r.Min > r.Max || (r.Min == r.Max && (r.MinExclusive || r.MaxExclusive)) {
		return false
	}
	last := this.Tail
	if last == nil || !scoreGreaterThanMin(last.Data.Score, r) {
		return false
	}
	first := this.Head.Levels[0].Forward
	if first == nil || !scoreLessThanMax(first.Data.Score, r) {
		return false
	}
	return true
}

// FirstInRange 返回分数范围 r 内的第一个节点（Score 最小），O(log n)。
// 先用 isInRange 做 O(1) 剪枝；再从最高层向下跳过所有不满足 min 条件的节点，
// 底层停止后检查当前节点是否也满足 max 条件。
// 范围内无节点时返回 nil。
func (this *SkipList[K, V]) FirstInRange(r *RangeSpecified) *Node[K, V] {
	assert.Assert(r != nil, "r range cannot be nil")
	if !this.isInRange(r) {
		return nil
	}

	prev := this.Head
	var current *Node[K, V] = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		for current != nil && !scoreGreaterThanMin(current.Data.Score, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
	}
	assert.Assert(current != nil, "current != nil, range r:", r)
	if !scoreLessThanMax(current.Data.Score, r) {
		return nil
	}
	return current
}

// GetRangeByScore 返回分数在范围 r 内的所有节点数据，O(log n + k)，k 为返回元素数。
// 用 FirstInRange 定位起点，再沿底层 Forward 顺序遍历直到超出 max。
func (this *SkipList[K, V]) GetRangeByScore(r *RangeSpecified) []*NodeData[K, V] {
	current := this.FirstInRange(r)
	datas := make([]*NodeData[K, V], 0, 4)
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		next := current.Levels[0].Forward
		datas = append(datas, current.Data)
		current = next
	}
	return datas
}

// CountByScore 统计分数在范围 r 内的节点数量，O(log n + k)，k 为命中节点数。
// 与 GetRangeByScore 等价，但不构建切片，无堆分配。
func (this *SkipList[K, V]) CountByScore(r *RangeSpecified) int {
	current := this.FirstInRange(r)
	count := 0
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		count++
		current = current.Levels[0].Forward
	}
	return count
}

// DeleteRangeByScore 删除分数在范围 r 内的所有节点，O(log n + k)，k 为删除元素数。
// 先从顶层向下定位到满足 min 条件的前驱（记录 prevNodes），
// 再沿底层顺序逐一删除满足 max 条件的节点。
func (this *SkipList[K, V]) DeleteRangeByScore(r *RangeSpecified) []*NodeData[K, V] {
	prevNodes := [SKIPLIST_MAXLEVEL]*Node[K, V]{}
	prev := this.Head
	var current *Node[K, V] = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		for current != nil && !scoreGreaterThanMin(current.Data.Score, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}

	deleted := make([]*NodeData[K, V], 0, 4)
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		current = next
	}
	return deleted
}
