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
package sorted_set

// BasicOps 定义基础键值操作：O(1) 查找、O(log n) 增删改、长度查询，以及最值快捷方法。
//
// 使用建议：
//   - 只需按 Key 查找或更新分数的代码，依赖此接口而非 *SortedSet。
//   - 实现方：*SortedSet[K, V] 满足此接口。
type BasicOps[K comparable, V any] interface {
	// Get 按 Key 查找元素，O(1)。不存在时返回 nil。
	Get(key K) *NodeData[K, V]

	// Insert 插入元素，O(log n)。Key 已存在时返回 false，原有元素不受影响。
	Insert(data *NodeData[K, V]) bool

	// Delete 按 Key 删除元素，O(log n)。Key 不存在时返回 (nil, false)。
	Delete(key K) (*NodeData[K, V], bool)

	// Length 返回集合元素总数，O(1)。
	Length() int

	// UpdateScore 更新 key 的分数并触发重新排序，O(log n)。Key 不存在时返回 (nil, false)。
	UpdateScore(key K, newScore float64) (*NodeData[K, V], bool)

	// GetMin 返回排序最靠前的元素，O(1)。Score 最小；Score 相同取 seq 最小（最先插入）。集合为空时返回 nil。
	GetMin() *NodeData[K, V]

	// GetMax 返回排序最靠后的元素，O(1)。Score 最大；Score 相同取 seq 最大（最后插入）。集合为空时返回 nil。
	GetMax() *NodeData[K, V]
}

// AscRankOps 定义升序排名操作（rank=1 对应 Score 最小的元素）。
//
// 使用建议：
//   - 需要按正序排名取值的场景（时间竞速、延迟排序等），依赖此接口。
//   - 实现方：*SortedSet[K, V] 满足此接口。
type AscRankOps[K comparable, V any] interface {
	// GetRank 查询 key 的升序排名，O(log n)。Key 不存在时返回 0。
	GetRank(key K) int

	// GetByRank 按升序排名取元素，O(log n)。rank 超出范围返回 nil；rank ≤ 0 触发断言 panic。
	GetByRank(rank int) *NodeData[K, V]

	// GetRangeByRank 返回升序排名范围 [start, end] 内的元素，O(log n + k)。结果升序排列。
	// start > end 时自动交换；start ≤ 0 截断为 1；end 超出总长度时截断，不 panic。
	GetRangeByRank(start, end int) []*NodeData[K, V]

	// DeleteRangeByRank 删除升序排名范围 [start, end] 内的元素，O(log n + k)。返回被删除的元素列表（升序）。
	// start > end 时自动交换；start ≤ 0 截断为 1；end 超出总长度时截断，不 panic。
	DeleteRangeByRank(start, end int) []*NodeData[K, V]
}

// DescRankOps 定义降序排名操作（rank=1 对应 Score 最大的元素）。
//
// 使用建议：
//   - 排行榜、积分榜等高分靠前的场景，依赖此接口。
//   - 实现方：*SortedSet[K, V] 满足此接口。
type DescRankOps[K comparable, V any] interface {
	// GetRankDesc 查询 key 的降序排名，O(log n)。Key 不存在时返回 0。
	GetRankDesc(key K) int

	// GetByRankDesc 按降序排名取元素，O(log n)。rank 超出范围返回 nil；rank ≤ 0 触发断言 panic。
	GetByRankDesc(rank int) *NodeData[K, V]

	// GetRangeByRankDesc 返回降序排名范围 [start, end] 内的元素，O(log n + k)。结果降序排列。
	// start > end 时自动交换；end 超出总长度时截断，不 panic。
	GetRangeByRankDesc(start, end int) []*NodeData[K, V]

	// DeleteRangeByRankDesc 删除降序排名范围 [start, end] 内的元素，O(log n + k)。返回被删除的元素列表（降序）。
	// start > end 时自动交换；end 超出总长度时截断，不 panic。
	DeleteRangeByRankDesc(start, end int) []*NodeData[K, V]
}

// ScoreOps 定义分数范围操作。
//
// 使用建议：
//   - 按积分段查询、清理过期数据等场景，依赖此接口。
//   - 实现方：*SortedSet[K, V] 满足此接口。
type ScoreOps[K comparable, V any] interface {
	// GetRangeByScore 返回 Score 在 [min, max] 范围内的元素，O(log n + k)。结果升序排列。
	// minEx=true 排除 min 端点，maxEx=true 排除 max 端点。
	GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V]

	// DeleteRangeByScore 删除 Score 在指定范围内的元素，O(log n + k)。返回被删除的元素列表（升序）。
	DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V]

	// CountByScore 统计 Score 在指定范围内的元素数量，O(log n + k)。只需 1 次结构体分配，避免结果切片开销，适合高频计数场景。
	CountByScore(min float64, minEx bool, max float64, maxEx bool) int
}

// SortedSetOps 是四个分组接口的组合，涵盖 *SortedSet 的全部公开方法。
//
// 使用建议：
//   - 大多数情况下，直接使用 *SortedSet[K, V] 即可，无需通过接口。
//   - 需要 mock 整个有序集合（如单元测试中替换为内存实现）时，用此接口作为参数类型。
//   - 只需要部分能力时，优先使用更小的子接口（BasicOps/AscRankOps/DescRankOps/ScoreOps），
//     明确表达函数对调用方的真实依赖，降低耦合。
type SortedSetOps[K comparable, V any] interface {
	BasicOps[K, V]
	AscRankOps[K, V]
	DescRankOps[K, V]
	ScoreOps[K, V]
}

// 编译期检查：*SortedSet[K, V] 必须满足 SortedSetOps。
// 若新增公开方法未加入任何子接口，此行会报错提醒。
var _ SortedSetOps[int, int] = (*SortedSet[int, int])(nil)
