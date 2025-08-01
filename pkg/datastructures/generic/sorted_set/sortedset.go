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

// NodeData 节点数据
type NodeData[T comparable] struct {
	Key   T       // 键
	Score float64 // 分数(跳跃表根据该数值来对节点进行有序排列)
	Val   T       // 卫星数据
}

func NewNodeData[T comparable](key T, score float64, val T) *NodeData[T] {
	return &NodeData[T]{
		Key:   key,
		Score: score,
		Val:   val,
	}
}

func (this *NodeData[T]) LessThan(other *NodeData[T]) bool {
	return this.Score < other.Score ||
		this.Score == other.Score && this.Val != other.Val // 简化比较
}

func (this *NodeData[T]) EqualTo(other *NodeData[T]) bool {
	return this.Score == other.Score && this.Val == other.Val
}

// SortedSet 有序集合
type SortedSet[T comparable] struct {
	Items []*NodeData[T] // 简化为数组实现，实际应该使用跳跃表
	Hash  map[T]*NodeData[T]
}

func NewSortedSet[T comparable]() *SortedSet[T] {
	return &SortedSet[T]{
		Items: make([]*NodeData[T], 0),
		Hash:  make(map[T]*NodeData[T]),
	}
}

/*
	基本操作
*/

func (this *SortedSet[T]) Get(key T) *NodeData[T] {
	return this.Hash[key]
}

func (this *SortedSet[T]) Insert(data *NodeData[T]) bool {
	assert.Assert(data != nil, "data == nil")

	if _, has := this.Hash[data.Key]; has {
		// 不能重复插入
		return false
	}

	// 插入到有序位置
	this.insertOrdered(data)
	this.Hash[data.Key] = data
	return true
}

func (this *SortedSet[T]) insertOrdered(data *NodeData[T]) {
	// 找到插入位置
	insertIndex := 0
	for i, item := range this.Items {
		if item.LessThan(data) {
			insertIndex = i + 1
		} else {
			break
		}
	}

	// 插入到指定位置
	this.Items = append(this.Items, nil)
	copy(this.Items[insertIndex+1:], this.Items[insertIndex:])
	this.Items[insertIndex] = data
}

func (this *SortedSet[T]) Delete(key T) (*NodeData[T], bool) {
	data, exist := this.Hash[key]
	if !exist {
		return nil, false
	}

	// 从有序数组中删除
	for i, item := range this.Items {
		if item == data {
			this.Items = append(this.Items[:i], this.Items[i+1:]...)
			break
		}
	}

	// 同步删除哈希表中的元素
	delete(this.Hash, key)
	return data, true
}

func (this *SortedSet[T]) Length() int {
	return len(this.Items)
}

/*
	排名相关操作
*/

// 获取排名
func (this *SortedSet[T]) GetRank(key T) int {
	data, exist := this.Hash[key]
	if !exist {
		return 0
	}

	for i, item := range this.Items {
		if item == data {
			return i + 1
		}
	}
	return 0
}

// 通过指定排名获得数据
func (this *SortedSet[T]) GetByRank(rank int) *NodeData[T] {
	assert.Assert(rank > 0, "rank must be positive number")

	if rank > len(this.Items) {
		return nil
	}
	return this.Items[rank-1]
}

// 获得指定排名范围的数据
func (this *SortedSet[T]) GetRangeByRank(start int, end int) []*NodeData[T] {
	if start > end {
		start, end = end, start
	}

	if start < 1 {
		start = 1
	}
	if end > len(this.Items) {
		end = len(this.Items)
	}

	result := make([]*NodeData[T], 0, end-start+1)
	for i := start - 1; i < end; i++ {
		result = append(result, this.Items[i])
	}
	return result
}

// 删除指定排名范围的数据
func (this *SortedSet[T]) DeleteRangeByRank(start int, end int) []*NodeData[T] {
	if start > end {
		start, end = end, start
	}

	if start < 1 {
		start = 1
	}
	if end > len(this.Items) {
		end = len(this.Items)
	}

	deleted := make([]*NodeData[T], 0, end-start+1)
	for i := start - 1; i < end; i++ {
		deleted = append(deleted, this.Items[i])
		delete(this.Hash, this.Items[i].Key)
	}

	// 从数组中删除
	this.Items = append(this.Items[:start-1], this.Items[end:]...)

	return deleted
}

/*
	分数相关操作
*/

// 更新分数
func (this *SortedSet[T]) UpdateScore(key T, newScore float64) (*NodeData[T], bool) {
	data, exist := this.Hash[key]
	if !exist {
		return nil, false
	}

	// 从有序数组中删除
	for i, item := range this.Items {
		if item == data {
			this.Items = append(this.Items[:i], this.Items[i+1:]...)
			break
		}
	}

	// 更新分数
	data.Score = newScore

	// 重新插入到有序位置
	this.insertOrdered(data)

	return data, true
}

// 通过分数范围(开闭区间由调用者指定)得到若干数据
func (this *SortedSet[T]) GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[T] {
	result := make([]*NodeData[T], 0)

	for _, item := range this.Items {
		score := item.Score

		// 检查最小值
		if minEx {
			if score <= min {
				continue
			}
		} else {
			if score < min {
				continue
			}
		}

		// 检查最大值
		if maxEx {
			if score >= max {
				continue
			}
		} else {
			if score > max {
				continue
			}
		}

		result = append(result, item)
	}

	return result
}

// 通过分数范围(开闭区间由调用者指定)删除若干数据
func (this *SortedSet[T]) DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[T] {
	deleted := make([]*NodeData[T], 0)
	newItems := make([]*NodeData[T], 0)

	for _, item := range this.Items {
		score := item.Score

		// 检查是否在范围内
		inRange := true

		// 检查最小值
		if minEx {
			if score <= min {
				inRange = false
			}
		} else {
			if score < min {
				inRange = false
			}
		}

		// 检查最大值
		if maxEx {
			if score >= max {
				inRange = false
			}
		} else {
			if score > max {
				inRange = false
			}
		}

		if inRange {
			deleted = append(deleted, item)
			delete(this.Hash, item.Key)
		} else {
			newItems = append(newItems, item)
		}
	}

	this.Items = newItems
	return deleted
}
