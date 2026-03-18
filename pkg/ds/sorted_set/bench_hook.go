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

import "github.com/motocat46/yytools/pkg/common/assert"

// RunBenchCheck 供 internal/bench 调用，校验有序集合内部一致性：
// 验证底层跳表 level-0 链表严格有序，以及 GetByRank 与链表位置一致。
// 注意：此函数与 sortedset_test.go 中的 checkSkiplistOrder 逻辑相同。
// 如修改其中一处，请同步更新另一处。
func RunBenchCheck(ss *SortedSet[int, int]) {
	ss.lengthMustEqual()

	current := ss.sl.Head.Levels[0].Forward
	rank := 1
	for current != nil {
		if current.Levels[0].Forward != nil {
			assert.Assert(current.Data.lessOrder(current.Levels[0].Forward.Data),
				"跳跃表必须是有序的")
		}
		data := ss.GetByRank(rank)
		assert.Assert(data != nil, "rank 实现有问题:", rank)
		assert.Assert(data.equalOrder(current.Data), "rank 实现有问题:", rank)
		rank++
		current = current.Levels[0].Forward
	}
}
