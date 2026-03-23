// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sorted_set_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/ds/sorted_set"
)

// ExampleSortedSet_leaderboard 展示游戏积分排行榜的典型用法。
func ExampleSortedSet_leaderboard() {
	ss := sorted_set.NewSortedSet[string, int]()

	// 插入玩家积分（Score 越大降序排名越靠前）
	ss.Insert(sorted_set.NewNodeData("alice", 1500.0, 1500))
	ss.Insert(sorted_set.NewNodeData("bob", 1200.0, 1200))
	ss.Insert(sorted_set.NewNodeData("carol", 1800.0, 1800))

	// 查询单个玩家排名（降序，1 = 第一名）
	fmt.Println(ss.GetRankDesc("carol")) // 1
	fmt.Println(ss.GetRankDesc("alice")) // 2
	fmt.Println(ss.GetRankDesc("bob"))   // 3

	// 获取前 3 名
	top3 := ss.GetRangeByRankDesc(1, 3)
	for _, node := range top3 {
		fmt.Printf("%s: %.0f\n", node.Key, node.Score)
	}
	// Output:
	// 1
	// 2
	// 3
	// carol: 1800
	// alice: 1500
	// bob: 1200
}

// ExampleSortedSet_updateScore 展示积分更新后排名自动调整。
func ExampleSortedSet_updateScore() {
	ss := sorted_set.NewSortedSet[string, int]()

	ss.Insert(sorted_set.NewNodeData("alice", 1500.0, 1500))
	ss.Insert(sorted_set.NewNodeData("bob", 1200.0, 1200))

	// bob 积分超越 alice
	ss.UpdateScore("bob", 2000.0)

	fmt.Println(ss.GetRankDesc("bob"))   // 1
	fmt.Println(ss.GetRankDesc("alice")) // 2
	// Output:
	// 1
	// 2
}
