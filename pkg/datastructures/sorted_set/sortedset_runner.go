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
// 创建日期:2023/6/29
package sorted_set

import (
	"fmt"
	"time"

	"github.com/stormYuanYang/yytools/pkg/algorithms/mathutils/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
)

func sortedSetMustLegal(ss *SortedSet[int]) {
	ss.lengthMustEqual()

	current := ss.sl.Head.Levels[0].Forward
	rank := 1
	for current != nil {
		if current.Levels[0].Forward != nil {
			assert.Assert(current.Data.lessOrder(current.Levels[0].Forward.Data),
				"跳跃表必须是有序的")

			data := ss.GetByRank(rank)
			assert.Assert(data != nil, "rank实现有问题:", rank)
			assert.Assert(data.equalOrder(current.Data), "rank实现有问题:", rank)
			rank++
		}
		current = current.Levels[0].Forward
	}
}

const (
	testScoreMin = 1
	testScoreMax = 750
)

var globalKey = int(0)

func nextKey() int {
	globalKey++
	return globalKey
}

func sortedSetOpInsert(ss *SortedSet[int], num int) {
	for i := 0; i < num; i++ {
		key := nextKey()
		score := float64(random.RandInt(testScoreMin, testScoreMax))
		data := NewNodeData(key, score, key)
		assert.Assert(ss.Insert(data), "插入不会失败:", data)
	}
}

func sortedSetOpDelete(ss *SortedSet[int], num int) {
	for i := 0; i < num; i++ {
		if ss.Length() > 0 {
			rank := random.RandInt(1, ss.Length())
			data := ss.GetByRank(rank)
			assert.Assert(data != nil, "data 不能为nil, rank:", rank)
			ss.Delete(data.Key)
		}
	}
}

func sortedSetOpUpdateScore(ss *SortedSet[int], num int) {
	for i := 0; i < num; i++ {
		if ss.Length() > 0 {
			rank := random.RandInt(1, ss.Length())
			data := ss.GetByRank(rank)
			assert.Assert(data != nil, "data 不能为nil, rank:", rank)
			newScore := float64(random.RandInt(testScoreMin, testScoreMax))
			_, ok := ss.UpdateScore(data.Key, newScore)
			assert.Assert(ok, "更新分数不能失败")
		}
	}
}

func sortedSetOpGetRank(ss *SortedSet[int], num int) {
	for i := 0; i < num; i++ {
		if ss.Length() > 0 {
			randRank := random.RandInt(1, ss.Length())
			data := ss.GetByRank(randRank)
			assert.Assert(data != nil, "data 不能为nil, rank:", randRank)
			rank := ss.GetRank(data.Key)
			assert.Assert(randRank == rank, "排名不一致, randRank:", randRank, " rank:", rank)
		}
	}
}

func sortedSetOpGetRangeByScore(ss *SortedSet[int], num int) {
	for i := 0; i < num; i++ {
		if ss.Length() == 0 {
			return
		}
		min := float64(random.RandInt(testScoreMin, testScoreMax))
		max := float64(random.RandInt(testScoreMin, testScoreMax))
		if min > max {
			min, max = max, min
		}
		minEx := random.RandInt(0, 1) == 1
		maxEx := random.RandInt(0, 1) == 1
		datas := ss.GetRangeByScore(min, minEx, max, maxEx)
		for j := 0; j < len(datas)-1; j++ {
			assert.Assert(datas[j].Score <= datas[j+1].Score, "返回的元素必须是有序的")
		}
	}
}

func sortedSetOpDeleteRangeByScore(ss *SortedSet[int], num int) {
	for i := 0; i < num; i++ {
		if ss.Length() == 0 {
			return
		}
		min := float64(random.RandInt(testScoreMin, testScoreMax))
		max := float64(random.RandInt(testScoreMin, testScoreMax))
		if min > max {
			min, max = max, min
		}
		ss.DeleteRangeByScore(min, false, max, false)
	}
}

var basicOps = []func(ss *SortedSet[int], num int){
	sortedSetOpInsert,
	sortedSetOpDelete,
	sortedSetOpUpdateScore,
	sortedSetOpGetRank,
}

var rangeOps = []func(ss *SortedSet[int], num int){
	sortedSetOpGetRangeByScore,
	sortedSetOpDeleteRangeByScore,
}

func SortedSetTest(total int) {
	println("有序集合测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	nums := []int{1, 2, 3, 4, 5, 10, 100, 1000, 10000}
	for a := 1; a <= total; a++ {
		fmt.Printf("-------第%d轮测试开始-------\n", a)
		globalKey = 0
		for k, n := range nums {
			ss := NewSortedSet[int]()
			sortedSetOpInsert(ss, n)

			opCnt := 50000
			for i := 0; i < opCnt; i++ {
				op := random.RandInt(0, len(basicOps)-1)
				basicOps[op](ss, 1)
			}
			rangeOpCnt := 100
			for i := 0; i < rangeOpCnt; i++ {
				op := random.RandInt(0, len(rangeOps)-1)
				rangeOps[op](ss, 1)
			}

			sortedSetMustLegal(ss)
			fmt.Printf("测试#%d结束. 初始长度:%d, 当前长度:%d\n", k+1, n, ss.Length())
		}
		fmt.Printf("-------第%d轮测试结束-------\n\n", a)
	}
	println("有序集合测试结束...")
}
