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
type NodeData[K comparable, V any] struct {
	Key   K       // 键（用于哈希表查找）
	Score float64 // 分数（跳跃表根据该数值对节点有序排列）
	Val   V       // 卫星数据
	seq   uint64  // 内部插入序列号，用于相同分数时的稳定排序（不对外暴露）
}

func NewNodeData[K comparable, V any](key K, score float64, val V) *NodeData[K, V] {
	return &NodeData[K, V]{Key: key, Score: score, Val: val}
}

// lessOrder 内部跳表排序：使用 seq 实现分数相同时的稳定全序
func (this *NodeData[K, V]) lessOrder(other *NodeData[K, V]) bool {
	return this.Score < other.Score ||
		this.Score == other.Score && this.seq < other.seq
}

// equalOrder 内部跳表查找：seq 唯一确定同一个节点
func (this *NodeData[K, V]) equalOrder(other *NodeData[K, V]) bool {
	return this.seq == other.seq
}

// SortedSet 有序集合（基于跳跃表实现）
type SortedSet[K comparable, V any] struct {
	sl   *SkipList[K, V]
	hash map[K]*NodeData[K, V]
	seq  uint64 // 自增序列号
}

func NewSortedSet[K comparable, V any]() *SortedSet[K, V] {
	return &SortedSet[K, V]{
		sl:   NewSkipList[K, V](),
		hash: make(map[K]*NodeData[K, V]),
	}
}

func (this *SortedSet[K, V]) Get(key K) *NodeData[K, V] {
	return this.hash[key]
}

func (this *SortedSet[K, V]) Insert(data *NodeData[K, V]) bool {
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

func (this *SortedSet[K, V]) Length() int {
	return this.sl.Length
}

func (this *SortedSet[K, V]) lengthMustEqual() {
	assert.Assert(this.sl.Length == len(this.hash),
		"长度不一致 skiplist length:", this.sl.Length, " hash length:", len(this.hash))
}

func (this *SortedSet[K, V]) GetRank(key K) int {
	data, exist := this.hash[key]
	if !exist {
		return 0
	}
	rank := this.sl.GetRank(data)
	assert.Assert(rank != 0, "rank must exist")
	return rank
}

func (this *SortedSet[K, V]) GetByRank(rank int) *NodeData[K, V] {
	assert.Assert(rank > 0, "rank must be positive number")
	node := this.sl.GetNodeByRank(rank)
	if node == nil {
		return nil
	}
	return node.Data
}

func (this *SortedSet[K, V]) GetRangeByRank(start int, end int) []*NodeData[K, V] {
	if start > end {
		start, end = end, start
	}
	return this.sl.GetRangeByRank(start, end)
}

func (this *SortedSet[K, V]) DeleteRangeByRank(start int, end int) []*NodeData[K, V] {
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

func (this *SortedSet[K, V]) UpdateScore(key K, newScore float64) (*NodeData[K, V], bool) {
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

func (this *SortedSet[K, V]) GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V] {
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{MinExclusive: minEx, MaxExclusive: maxEx},
		Min:                min,
		Max:                max,
	}
	return this.sl.GetRangeByScore(r)
}

func (this *SortedSet[K, V]) DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V] {
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
