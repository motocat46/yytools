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

// 作者:  yangyuan
// 创建日期:2023/6/17
package sorted_set

// NodeData 有序集合中的节点数据，包含唯一键、排序分数和业务数据。
//
// Key 用于哈希表 O(1) 查找，Score 决定跳表中的排列顺序（升序），
// Val 是与 Key 关联的业务数据，类型完全独立于 Key。
// seq 由 SortedSet 在插入时自动分配，外部不可见，用于相同 Score 时的稳定排序。
//
// 注意：Score 不支持 NaN。NaN 的比较语义（NaN != NaN、NaN < NaN == false）
// 会破坏跳表所依赖的全序关系，Insert 和 UpdateScore 会对 NaN 触发 assert panic。
// ±Inf 是合法的 Score 值。
type NodeData[K comparable, V any] struct {
	Key   K       // 键（用于哈希表查找）
	Score float64 // 分数（跳跃表根据该数值对节点有序排列）
	Val   V       // 卫星数据
	seq   uint64  // 内部插入序列号，用于相同分数时的稳定排序（不对外暴露）
}

// NewNodeData 创建一个节点数据，seq 由 SortedSet.Insert 负责赋值。
func NewNodeData[K comparable, V any](key K, score float64, val V) *NodeData[K, V] {
	return &NodeData[K, V]{Key: key, Score: score, Val: val}
}

// lessOrder 跳表内部排序比较：按 (Score, seq) 二元组升序。
// Score 相同时 seq 更小的排在前面，保证稳定全序。
func (this *NodeData[K, V]) lessOrder(other *NodeData[K, V]) bool {
	return this.Score < other.Score ||
		this.Score == other.Score && this.seq < other.seq
}

// equalOrder 跳表内部节点定位：seq 全局唯一，可精确区分同 Score 的不同节点。
func (this *NodeData[K, V]) equalOrder(other *NodeData[K, V]) bool {
	return this.seq == other.seq
}