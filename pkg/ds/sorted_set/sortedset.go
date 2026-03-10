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
	"github.com/motocat46/yytools/pkg/common/assert"
)

// NodeData 节点数据
type NodeData[T comparable] struct {
	Key   T       // 键（用于哈希表查找）
	Score float64 // 分数（跳跃表根据该数值对节点有序排列）
	Val   T       // 卫星数据
	seq   uint64  // 内部插入序列号，用于相同分数时的稳定排序（不对外暴露）
}

func NewNodeData[T comparable](key T, score float64, val T) *NodeData[T] {
	return &NodeData[T]{Key: key, Score: score, Val: val}
}

// LessThan 公开比较：以分数为准（分数越小越靠前）
func (this *NodeData[T]) LessThan(other *NodeData[T]) bool {
	return this.Score < other.Score
}

// EqualTo 公开比较：分数相同且值相同视为相等
func (this *NodeData[T]) EqualTo(other *NodeData[T]) bool {
	return this.Score == other.Score && this.Val == other.Val
}

// lessOrder 内部跳表排序：使用 seq 实现分数相同时的稳定全序
func (this *NodeData[T]) lessOrder(other *NodeData[T]) bool {
	return this.Score < other.Score ||
		this.Score == other.Score && this.seq < other.seq
}

// equalOrder 内部跳表查找：seq 唯一确定同一个节点
func (this *NodeData[T]) equalOrder(other *NodeData[T]) bool {
	return this.seq == other.seq
}

// SortedSet 有序集合（基于跳跃表实现）
type SortedSet[T comparable] struct {
	sl   *SkipList[T]
	hash map[T]*NodeData[T]
	seq  uint64 // 自增序列号
}

func NewSortedSet[T comparable]() *SortedSet[T] {
	return &SortedSet[T]{
		sl:   NewSkipList[T](),
		hash: make(map[T]*NodeData[T]),
	}
}

func (this *SortedSet[T]) Get(key T) *NodeData[T] {
	return this.hash[key]
}

func (this *SortedSet[T]) Insert(data *NodeData[T]) bool {
	assert.Assert(data != nil, "data == nil")

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

func (this *SortedSet[T]) Delete(key T) (*NodeData[T], bool) {
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

func (this *SortedSet[T]) Length() int {
	return this.sl.Length
}

func (this *SortedSet[T]) lengthMustEqual() {
	assert.Assert(this.sl.Length == len(this.hash),
		"长度不一致 skiplist length:", this.sl.Length, " hash length:", len(this.hash))
}

func (this *SortedSet[T]) GetRank(key T) int {
	data, exist := this.hash[key]
	if !exist {
		return 0
	}
	rank := this.sl.GetRank(data)
	assert.Assert(rank != 0, "rank must exist")
	return rank
}

func (this *SortedSet[T]) GetByRank(rank int) *NodeData[T] {
	assert.Assert(rank > 0, "rank must be positive number")
	node := this.sl.GetNodeByRank(rank)
	if node == nil {
		return nil
	}
	return node.Data
}

func (this *SortedSet[T]) GetRangeByRank(start int, end int) []*NodeData[T] {
	if start > end {
		start, end = end, start
	}
	return this.sl.GetRangeByRank(start, end)
}

func (this *SortedSet[T]) DeleteRangeByRank(start int, end int) []*NodeData[T] {
	if start > end {
		start, end = end, start
	}
	deleted := this.sl.DeleteRangeByRank(start, end)
	for _, one := range deleted {
		delete(this.hash, one.Key)
	}
	this.lengthMustEqual()
	return deleted
}

func (this *SortedSet[T]) UpdateScore(key T, newScore float64) (*NodeData[T], bool) {
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

func (this *SortedSet[T]) GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[T] {
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{MinExclusive: minEx, MaxExclusive: maxEx},
		Min:                min,
		Max:                max,
	}
	return this.sl.GetRangeByScore(r)
}

func (this *SortedSet[T]) DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[T] {
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
