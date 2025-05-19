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
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
)

type SortedSet struct {
	Sl   *SkipList
	Hash map[interface{}]*NodeData
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		Sl:   NewSkipList(),
		Hash: map[interface{}]*NodeData{},
	}
}

/*
	基本操作
*/

func (this *SortedSet) Get(key interface{}) *NodeData {
	assert.Assert(key != nil, "key == nil")
	return this.Hash[key]
}

func (this *SortedSet) Insert(data *NodeData) bool {
	assert.Assert(data != nil, "data == nil")

	if _, has := this.Hash[data.Key]; has {
		// 不能重复插入
		return false
	}
	
	_, ok := this.Sl.Insert(data)
	assert.Assert(ok, "insert must success, data.Key:", data.Key)
	if ok {
		this.Hash[data.Key] = data
	}
	this.lengthMustEqual()
	return ok
}

func (this *SortedSet) Delete(key interface{}) (*NodeData, bool) {
	assert.Assert(key != nil, "key == nil")

	data, exist := this.Hash[key]
	if !exist {
		return nil, false
	}
	
	if node, ok := this.Sl.Delete(data); ok {
		// 同步删除哈希表中的元素
		delete(this.Hash, key)
		this.lengthMustEqual()
		return node.Data, ok
	} else {
		return nil, ok
	}
}

func (this *SortedSet) Length() int {
	return this.Sl.Length
}

func (this *SortedSet) lengthMustEqual() {
	assert.Assert(this.Sl.Length == len(this.Hash),
		"长度不一致 skiplist length:", this.Sl.Length, " hash length:", this.Hash)
}

/*
	排名相关操作
*/

// 获取排名
func (this *SortedSet) GetRank(key interface{}) int {
	assert.Assert(key != nil, "key == nil")
	
	data, exist := this.Hash[key]
	if !exist {
		return 0
	}
	rank := this.Sl.GetRank(data)
	// 一定能找到排名(哈希表保证了元素一定存在)
	assert.Assert(rank != 0, "rank must exist")
	return rank
}

// 通过指定排名获得数据
func (this *SortedSet) GetByRank(rank int) *NodeData {
	assert.Assert(rank > 0, "rank must be positive number")

	node := this.Sl.GetNodeByRank(rank)
	if node == nil {
		return nil
	}
	return node.Data
}

// 获得指定排名范围的数据
func (this *SortedSet) GetRangeByRank(start int, end int) []*NodeData {
	if start > end {
		start, end = end, start
	}
	return this.Sl.GetRangeByRank(start, end)
}

// 删除指定排名范围的数据
func (this *SortedSet) DeleteRangeByRank(start int, end int) []*NodeData {
	if start > end {
		start, end = end, start
	}
	deleted := this.Sl.DeleteRangeByRank(start, end)
	// 同步删除哈希表中映射的数据
	for _, one := range deleted {
		delete(this.Hash, one.Key)
	}
	this.lengthMustEqual()
	return deleted
}

/*
	分数相关操作
*/

// 更新分数
func (this *SortedSet) UpdateScore(key interface{}, newScore float64) (*NodeData, bool) {
	data, exist := this.Hash[key]
	if !exist {
		return nil, false
	}
	node, ok := this.Sl.UpdateScore(data, newScore)
	if !ok {
		return nil, ok
	}
	this.lengthMustEqual()
	return node.Data, ok
}

// 通过分数范围(开闭区间由调用者指定)得到若干数据
func (this *SortedSet) GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData {
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{
			MinExclusive: minEx,
			MaxExclusive: maxEx,
		},
		Min: min,
		Max: max,
	}
	return this.Sl.GetRangeByScore(r)
}

// 通过分数范围(开闭区间由调用者指定)删除若干数据
func (this *SortedSet) DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData {
	r := &RangeSpecified{
		RangeSpecifiedBase: RangeSpecifiedBase{
			MinExclusive: minEx,
			MaxExclusive: maxEx,
		},
		Min: min,
		Max: max,
	}
	deleted := this.Sl.DeleteRangeByScore(r)
	// 同步删除哈希表中映射的数据
	for _, one := range deleted {
		delete(this.Hash, one.Key)
	}
	this.lengthMustEqual()
	return deleted
}