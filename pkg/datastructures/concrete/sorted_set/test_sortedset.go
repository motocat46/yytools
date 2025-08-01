// Package test.

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
    "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/probability_distribution"
    random2 "github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

var uniq = int64(0)

type Val struct {
	ID   int64 //唯一id
	Name string
	Meta string
}

func NewVal() *Val {
	uniq++
	return &Val{
		ID:   uniq,
		Meta: "",
	}
}

func (this *Val) LessThan(data Value) bool {
	x := data.(*Val)
	return this.ID < x.ID
}

func (this *Val) EqualTo(data Value) bool {
	x := data.(*Val)
	return this.ID == x.ID
}

const (
	TEST_SORTED_SET_SCORE_MIN = 1
	TEST_SORTED_SET_SCORE_MAX = 750
)

func SortedSetMustLegal(ss *SortedSet) {
	ss.lengthMustEqual()
	
	current := ss.Sl.Head.Levels[0].Forward
	rank := 1
	for current != nil {
		//print(int(current.Data.Score), "->")
		if current.Levels[0].Forward != nil {
			assert.Assert(current.Data.LessThan(current.Levels[0].Forward.Data),
				"跳跃表表必须是有序的:", fmt.Sprintf("current:%+v, forward:%+v", current, current.Levels[0].Forward))

			data := ss.GetByRank(rank)
			assert.Assert(data != nil, "rank实现有问题:", rank)
			assert.Assert(data.EqualTo(current.Data), "rank实现有问题", rank)
			rank++
		}
		current = current.Levels[0].Forward
	}
}

// 插入
func SortedSetOp_Insert(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		n := random2.RandInt(TEST_SORTED_SET_SCORE_MIN, TEST_SORTED_SET_SCORE_MAX)
		val := NewVal()
		data := NewNodeData(val.ID, float64(n), val)
		assert.Assert(ss.Insert(data), "插入不会失败:", data)
	}
}

// 删除
func SortedSetOp_Delete(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		if ss.Length() > 0 {
			randomRank := random2.RandInt(1, ss.Length())
			data := ss.GetByRank(randomRank)
			assert.Assert(data != nil, "data 不能为nil, rank:", randomRank)
			ss.Delete(data.Key)
		}
	}
}

// 更新分数
func SortedSetOp_UpdateScore(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		if ss.Length() > 0 {
			randomRank := random2.RandInt(1, ss.Length())
			data := ss.GetByRank(randomRank)
			assert.Assert(data != nil, "data 不能为nil, rank:", randomRank)
			
			n := random2.RandInt(TEST_SORTED_SET_SCORE_MIN, TEST_SORTED_SET_SCORE_MAX)
			_, ok := ss.UpdateScore(data.Key, float64(n))
			assert.Assert(ok, "更新分数不能失败, data:", data, " newScore:", n)
		}
	}
}

// 通过分数范围获得多个元素
func SortedSetOp_GetRangeByScore(ss *SortedSet, num int) {
	if ss.Length() > 0 {
		for i := 0; i < num; i++ {
			min := float64(random2.RandInt(TEST_SORTED_SET_SCORE_MIN, TEST_SORTED_SET_SCORE_MAX))
			max := float64(random2.RandInt(TEST_SORTED_SET_SCORE_MIN, TEST_SORTED_SET_SCORE_MAX))
			if min > max {
				min, max = max, min
			}
			
			minEx := false
			if random2.RandInt(0, 1) == 1 {
				minEx = true
			}
			
			maxEx := false
			if random2.RandInt(0, 1) == 1 {
				maxEx = true
			}

			datas := ss.GetRangeByScore(min, minEx, max, maxEx)

			// 判断是否有序
			for j := 0; j < len(datas); j++ {
				if j+1 < len(datas) {
					assert.Assert(datas[j].LessThan(datas[j+1]), "返回的元素必须是有序的")
				}
			}

			// 判断返回的元素是否在指定范围内
			for _, one := range datas {
				if minEx {
					assert.Assert(one.Score > min, "分数要在范围内,", "score:", one.Score, " max:", max)
				} else {
					assert.Assert(one.Score >= min, "分数要在范围内,", "score:", one.Score, " max:", max)
				}
				if maxEx {
					assert.Assert(one.Score < max, "分数要在范围内,", "score:", one.Score, " max:", max)
				} else {
					assert.Assert(one.Score <= max, "分数要在范围内,", "score:", one.Score, " max:", max)
				}
			}
		}
	}
}

func SortedSetOp_DeleteRangeByScore(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		if ss.Length() == 0 {
			break
		}
		min := float64(random2.RandInt(TEST_SORTED_SET_SCORE_MIN, TEST_SORTED_SET_SCORE_MAX))
		max := float64(random2.RandInt(TEST_SORTED_SET_SCORE_MIN, TEST_SORTED_SET_SCORE_MAX))
		if min > max {
			min, max = max, min
		}
		
		minEx := false
		if random2.RandInt(0, 1) == 1 {
			minEx = true
		}
		
		maxEx := false
		if random2.RandInt(0, 1) == 1 {
			maxEx = true
		}
		
		datas := ss.DeleteRangeByScore(min, minEx, max, maxEx)
		
		// 判断是否有序
		for j := 0; j < len(datas); j++ {
			if j+1 < len(datas) {
				assert.Assert(datas[j].LessThan(datas[j+1]), "返回的元素必须是有序的")
			}
		}
		
		// 判断返回的元素是否在指定范围内
		for _, one := range datas {
			if minEx {
				assert.Assert(one.Score > min, "分数要在范围内,", "score:", one.Score, " max:", max)
			} else {
				assert.Assert(one.Score >= min, "分数要在范围内,", "score:", one.Score, " max:", max)
			}
			if maxEx {
				assert.Assert(one.Score < max, "分数要在范围内,", "score:", one.Score, " max:", max)
			} else {
				assert.Assert(one.Score <= max, "分数要在范围内,", "score:", one.Score, " max:", max)
			}
		}
	}
}

// 获取排名 和 通过排名获取元素 互相验证
func SortedSetOp_GetRank(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		if ss.Length() > 0 {
			randomRank := random2.RandInt(1, ss.Length())
			data := ss.GetByRank(randomRank)
			assert.Assert(data != nil, "data 不能为nil, rank:", randomRank)
			
			rank := ss.GetRank(data.Key)
			assert.Assert(randomRank == rank, "排名不一致, randomRank:", randomRank, " rank:", rank, " key:", data.Key)
		}
	}
}

// 通过排名范围获取元素
func SortedSetOp_GetRangeByRank(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		length := ss.Length()
		if length == 0 {
			return
		}
		start := random2.RandInt(1, length)
		end := random2.RandInt(1, length)
		if start > end {
			start, end = end, start
		}
		datas := ss.GetRangeByRank(start, end)
		for j, one := range datas {
			rank := ss.GetRank(one.Key)
			assert.Assert(rank == start+j, "排名不正确", rank, " ", start+j)
			
			if j < len(datas)-1 {
				assert.Assert(datas[j].LessThan(datas[j+1]), "返回的元素必须是有序的")
			}
		}
	}
}

func SortedSetOp_DeleteRangeByRank(ss *SortedSet, num int) {
	for i := 0; i < num; i++ {
		length := ss.Length()
		if length > 0 {
			start := random2.RandInt(1, length)
			end := random2.RandInt(1, length)
			if start > end {
				start, end = end, start
			}
			datas := ss.DeleteRangeByRank(start, end)
			for j, one := range datas {
				rank := ss.GetRank(one.Key)
				assert.Assert(rank == 0, "排名不正确:", rank)
				
				if j < len(datas)-1 {
					assert.Assert(datas[j].LessThan(datas[j+1]), "返回的元素必须是有序的")
				}
			}
		}
	}
}

var SortedSetOp_Handlers = []func(ss *SortedSet, num int){
	SortedSetOp_Insert,
	SortedSetOp_Delete,
	SortedSetOp_UpdateScore,
	SortedSetOp_GetRank,
}

var SortedSetOp_RangeHandlers = []func(ss *SortedSet, num int){
	SortedSetOp_GetRangeByScore,
	SortedSetOp_DeleteRangeByScore,
	SortedSetOp_GetRangeByRank,
	SortedSetOp_DeleteRangeByRank,
}

func SortedSetTest(total int) {
	println("有序集合测试开始...")
	random2.RandSeed(time.Now().UnixMilli())
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 1e4, 1e5, 1e6}
	for a := 1; a <= total; a++ {
		fmt.Printf("-------第%d轮测试开始-------\n", a)
		for k, n := range nums {
			ss := NewSortedSet()
			// 插入指定数量的元素
			SortedSetOp_Insert(ss, n)
			
			// 基本操作的测试次数
			opCnt := random2.RandInt(1e5, 1e5)
			// range相关操作都很耗时，减少测试的量级
			rangeOpCnt := random2.RandInt(10, 10)
			opWeights := []int{opCnt, rangeOpCnt}
			
			realCnt := []int{0, 0}
			// 根据操作次数得到对应的执行概率
			aliasMethod := probability_distribution.ProbFactory(probability_distribution.VoseAlias, opWeights)
			for i := 0; i < opCnt+rangeOpCnt; i++ {
				index := aliasMethod.Generate()
				if index == 0 {
					op := random2.RandInt(0, len(SortedSetOp_Handlers)-1)
					fn := SortedSetOp_Handlers[op]
					fn(ss, 1)
				} else if index == 1 {
					rangeOp := random2.RandInt(0, len(SortedSetOp_RangeHandlers)-1)
					fn := SortedSetOp_RangeHandlers[rangeOp]
					fn(ss, 1)
				} else {
					assert.Assert(false, "不应该执行到这里", opCnt, " ", rangeOpCnt)
				}
				realCnt[index]++
			}
			SortedSetMustLegal(ss)
			fmt.Printf("测试#%d结束. 初始长度:%d, 当前长度:%d, 执行基本操作:%d次(理论:%d)，执行range操作:%d次(理论:%d)\n",
				k+1, n, ss.Length(), realCnt[0], opCnt, realCnt[1], rangeOpCnt)
		}
		fmt.Printf("-------第%d轮测试结束-------\n\n", a)
	}
	println("有序集合测试结束...")
}