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

type SkipList[T comparable] struct {
	Head        *Node[T]
	Tail        *Node[T]
	Length      int
	Level       int
	LevelUpProb float32
}

type SkipListLevel[T comparable] struct {
	Forward *Node[T]
	Span    int
}

type Node[T comparable] struct {
	Levels   []*SkipListLevel[T]
	Backward *Node[T]
	Data     *NodeData[T]
}

type RangeSpecifiedBase struct {
	MinExclusive bool
	MaxExclusive bool
}

type RangeSpecified struct {
	RangeSpecifiedBase
	Min float64
	Max float64
}

func CreateNode[T comparable](level int, data *NodeData[T]) *Node[T] {
	levelArr := make([]*SkipListLevel[T], level)
	for i := 0; i < level; i++ {
		levelArr[i] = &SkipListLevel[T]{Forward: nil, Span: 0}
	}
	return &Node[T]{Levels: levelArr, Backward: nil, Data: data}
}

func (this *Node[T]) High() int {
	return len(this.Levels)
}

func NewSkipList[T comparable]() *SkipList[T] {
	return NewSkipListByParams[T](DEFAULT_LEVELUP_PROBABILITY)
}

func NewSkipListByParams[T comparable](nodeLevelUpProb float32) *SkipList[T] {
	assert.Assert(nodeLevelUpProb >= 0 && nodeLevelUpProb < 1,
		"提升节点高度概率不正确:", nodeLevelUpProb, "正常范围:[0.0,1)")

	var zero T
	skipList := &SkipList[T]{
		Head: CreateNode[T](SKIPLIST_MAXLEVEL, &NodeData[T]{
			Score: 0,
			Val:   zero,
		}),
		Tail:        nil,
		Length:      0,
		Level:       0,
		LevelUpProb: nodeLevelUpProb,
	}
	return skipList
}

func (this *SkipList[T]) Get(score float64, data *NodeData[T]) (*Node[T], bool) {
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

func (this *SkipList[T]) Insert(data *NodeData[T]) (*Node[T], bool) {
	assert.Assert(data != nil, "data must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[T]{}
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

	newNode := CreateNode[T](level, data)
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

func (this *SkipList[T]) findNode(data *NodeData[T], prevNodes *[SKIPLIST_MAXLEVEL]*Node[T]) (*Node[T], bool) {
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

func (this *SkipList[T]) deleteNode(current *Node[T], prevNodes *[SKIPLIST_MAXLEVEL]*Node[T]) *Node[T] {
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

func (this *SkipList[T]) Delete(data *NodeData[T]) (*Node[T], bool) {
	assert.Assert(data != nil, "val must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[T]{}
	current, ok := this.findNode(data, &prevNodes)
	if !ok {
		return current, ok
	}
	return this.deleteNode(current, &prevNodes), true
}

func (this *SkipList[T]) GetRank(data *NodeData[T]) int {
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

func (this *SkipList[T]) GetNodeByRank(rank int) *Node[T] {
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

func (this *SkipList[T]) GetRangeByRank(start int, end int) []*NodeData[T] {
	assert.Assert(start > 0 && end > 0 && start <= end, "rank范围不合法, start:", start, " end:", end)

	current := this.GetNodeByRank(start)
	traversed := start
	datas := make([]*NodeData[T], 0, 4)
	for current != nil && traversed <= end {
		next := current.Levels[0].Forward
		datas = append(datas, current.Data)
		traversed++
		current = next
	}
	return datas
}

func (this *SkipList[T]) DeleteRangeByRank(start int, end int) []*NodeData[T] {
	assert.Assert(start > 0 && end > 0 && start <= end, "rank范围不合法, start:", start, " end:", end)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[T]{}
	traversed := 0
	prev := this.Head
	var current *Node[T] = nil
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

	deleted := make([]*NodeData[T], 0, 4)
	for current != nil && traversed <= end {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		traversed++
		current = next
	}
	return deleted
}

func (this *SkipList[T]) UpdateScore(data *NodeData[T], newScore float64) (*Node[T], bool) {
	assert.Assert(data != nil, "data must not be nil")
	assert.Assert(!math.IsNaN(data.Score), "oldScore is not a number:", data.Score)
	assert.Assert(!math.IsNaN(newScore), "newScore is not a number:", newScore)

	prevNodes := [SKIPLIST_MAXLEVEL]*Node[T]{}
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

func scoreGreaterThanMin(score float64, r *RangeSpecified) bool {
	if r.MinExclusive {
		return score > r.Min
	}
	return score >= r.Min
}

func scoreLessThanMax(score float64, r *RangeSpecified) bool {
	if r.MaxExclusive {
		return score < r.Max
	}
	return score <= r.Max
}

func (this *SkipList[T]) isInRange(r *RangeSpecified) bool {
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

func (this *SkipList[T]) FirstInRange(r *RangeSpecified) *Node[T] {
	assert.Assert(r != nil, "r range cannot be nil")
	if !this.isInRange(r) {
		return nil
	}

	prev := this.Head
	var current *Node[T] = nil
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

func (this *SkipList[T]) GetRangeByScore(r *RangeSpecified) []*NodeData[T] {
	current := this.FirstInRange(r)
	datas := make([]*NodeData[T], 0, 4)
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		next := current.Levels[0].Forward
		datas = append(datas, current.Data)
		current = next
	}
	return datas
}

func (this *SkipList[T]) DeleteRangeByScore(r *RangeSpecified) []*NodeData[T] {
	prevNodes := [SKIPLIST_MAXLEVEL]*Node[T]{}
	prev := this.Head
	var current *Node[T] = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		for current != nil && !scoreGreaterThanMin(current.Data.Score, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}

	deleted := make([]*NodeData[T], 0, 4)
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		current = next
	}
	return deleted
}
