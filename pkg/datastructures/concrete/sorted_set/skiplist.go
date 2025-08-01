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

// 参考自redis的跳跃表实现

// 作者:  yangyuan
// 创建日期:2023/6/2
package sorted_set

import (
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
	"math"
)

type SkipList struct {
	Head        *Node   // 头结点(哨兵结点)
	Tail        *Node   // 尾结点
	Length      int     // 结点总数(不包含头结点)
	Level       int     // 链表中当前结点的最大高度(除开头结点的其他结点中的最高的高度)
	LevelUpProb float32 // 提升结点高度的概率
}

type SkipListLevel struct {
	Forward *Node// 同一高度下，指向的下一个结点
	Span    int // 同一高度下, 结点之间的跨度(方便取结点的排名) 跨度是基于1的
}

type Value interface {
	LessThan(other Value) bool
	EqualTo(other Value) bool
}

// 不小于也不等于,那么就是大于
func ValueGreaterThan(one Value, other Value) bool {
	return !one.LessThan(other) && !one.EqualTo(other)
}

type NodeData struct {
	Key   interface{}
	Score float64 // 分数(跳跃表根据该数值来对节点进行有序排列)
	Val   Value   // 卫星数据
}

func NewNodeData(key interface{}, score float64, val Value) *NodeData {
	return &NodeData{
		Key:   key,
		Score: score,
		Val:   val,
	}
}

func (this *NodeData) LessThan(other *NodeData) bool {
	return this.Score < other.Score ||
		this.Score == other.Score && this.Val.LessThan(other.Val)
}

func (this *NodeData) EqualTo(other *NodeData) bool {
	return this.Score == other.Score && this.Val.EqualTo(other.Val)
}

type Node struct {
	Levels   []*SkipListLevel // 向前的(每个高度的下一个)结点数组
	Backward *Node            // 上一个结点(这样最下层就是双向链表，方便向后的遍历)
	Data     *NodeData        // 结点携带的数据(包含分数)
}

type RangeSpecifiedBase struct {
	MinExclusive bool // true:不包含边界值Min
	MaxExclusive bool // true：不包含边界值Max
}

// 比较分数时，指定分数范围和边界(开闭区间)
type RangeSpecified struct {
	RangeSpecifiedBase
	Min float64
	Max float64
}

// 比较值时，指定值范围和边界(开闭区间)
type ValueRangeSpecified struct {
	RangeSpecifiedBase
	Min Value
	Max Value
}

/*
	method of Node
*/
func CreateNode(level int, data *NodeData) *Node {
	levelArr := make([]*SkipListLevel, level)
	for i := 0; i < level; i++ {
		levelArr[i] = &SkipListLevel{
			Forward: nil,
			Span:    0,
		}
	}
	node := &Node{
		Levels:   levelArr,
		Backward: nil,
		Data:     data,
	}
	return node
}

func (this *Node) High() int {
	return len(this.Levels)
}

/*
	method of SkipList
*/
func NewSkipList() *SkipList {
	return NewSkipListByParams(DEFAULT_LEVELUP_PROBABILITY)
}

func NewSkipListByParams(nodeLevelUpProb float32) *SkipList {
	assert.Assert(nodeLevelUpProb >= 0 && nodeLevelUpProb < 1,
		"提升节点高度概率不正确:", nodeLevelUpProb, "正常范围:[0.0,1)")
	
	skipList := &SkipList{
		Head: CreateNode(SKIPLIST_MAXLEVEL, &NodeData{
			Score: 0,
			Val:   nil,
		}),
		Tail:        nil,
		Length:      0,
		Level:       0,
		LevelUpProb: nodeLevelUpProb,
	}
	return skipList
}

/*
 跳跃表基本操作（增删改查）
*/

// 根据分数和数据查找结点
// 时间复杂度为O(logn)
// 空间复杂度为O(1)
func (this *SkipList) Get(score float64, data *NodeData) (*Node, bool) {
	// 断言(判断传入的分数值)
	assert.Assert(!math.IsNaN(score), "score is not a number:", score)
	// 断言(不允许传入nil)
	assert.Assert(data != nil, "data must not be nil, score:", score)
	
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		// 满足两个条件，查找就继续在当前高度向前继续查找：
		// 1.指定分数大于当前结点的分数，说明要查找的节点一定在当前结点的前方;
		// 2.当前节点不能是(哨兵)尾结点
		current := prev.Levels[i].Forward
		for current != nil && current.Data.LessThan(data) {
			// 双指针继续向前移动
			prev = current
			current = prev.Levels[i].Forward
		}
		if current != nil && current.Data.EqualTo(data) {
			// 找到了
			return current, true
		}
		// 否则1.score分数比当前结点的分数更大,或者
		// 2.当前(高度),这个结点是nil
		// 所以再次进入循环，到更低高度去查找
	}
	return nil, false
}

// 根据分数和数据插入新结点到跳跃表中
// 需要由调用者保证不插入重复的结点,如果结点已存在则会插入失败
// 时间复杂度为O(logn)
// 空间复杂度为O(1)
func (this *SkipList) Insert(data *NodeData) (*Node, bool) {
	// 断言(不允许传入nil)
	assert.Assert(data != nil, "data must not be nil")
	// 断言(判断传入的分数值)
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)
	
	//注意这里，使用数组而不是切片，避免不必要的堆内存分配(插入操作可能会很频繁)
	// 当前这种情况，(只要该函数不返回数组)数组就是分配在栈上的
	prevNodes := [SKIPLIST_MAXLEVEL]*Node{}
	rank := [SKIPLIST_MAXLEVEL]int{}
	
	//	跨度是基于1的(即跨度单位是1)
	//	跳跃表结点记录的是，当前结点到其前置结点的跨度，是一个相对值
	//	也正因为如此，就可以通过累计这些相对值，从而取得指定结点的排名
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		// 第一个(即rank[this.Levels-1])肯定是0
		if i != this.Level-1 {
			// 随着高度的下降，重新赋值当前高度下的排名初始值
			rank[i] = rank[i+1]
		}
		
		current := prev.Levels[i].Forward
		for current != nil && current.Data.LessThan(data) {
			// 指针向前行进说明，要插入的结点在当前结点的前方
			// 那么加上当前结点的跨度
			rank[i] += prev.Levels[i].Span
			
			// 当前层链表，前置指针前进一个节点
			prev = current
			// 当前层链表，当前指针也前进一个节点
			current = prev.Levels[i].Forward
		}
		if current != nil && current.Data.EqualTo(data) {
			// 如果逻辑走到这里，意味着将插入重复元素(不允许插入重复元素)
			// 返回已存在的元素
			// BTW,如果调用者能够保证不会插入重复的元素，那么这里的判断就是不必要的
			return current, false
		}
		
		// 需要记录每一层的前置结点
		prevNodes[i] = prev
		// 继续进入下一高度，直到进入最低的高度后退出循环
	}
	
	level := randomLevel(this.LevelUpProb)
	if level > this.Level {
		for i := this.Level; i < level; i++ {
			// 比跳跃表原有的结点高度还高,则需要将头结点作为高高度的前置结点
			prevNodes[i] = this.Head
			// 当前高度，结点(头结点)到下一结点(nil)的跨度就是整个跳跃表的长度
			prevNodes[i].Levels[i].Span = this.Length
		}
		// 更新最大高度
		this.Level = level
	}
	
	//	1.创建新的结点,设置相关数据;2.并插入指定位置,并且更新和维护结点每一层的索引关系
	newNode := CreateNode(level, data)
	for i := 0; i < level; i++ {
		// 将每一层向前(方向)的链表都重新链接起来
		newNode.Levels[i].Forward = prevNodes[i].Levels[i].Forward
		prevNodes[i].Levels[i].Forward = newNode
		
		// 一定注意，每一高度的前置结点可能是不同的结点
		// 当新结点在对应高度插入,则更新跨度
		// oldSpan表示最低高度的前置结点排名(rank[0])减去
		// 在当前高度前置结点的排名rank[i]
		// 也就是 (当前高度)前置结点到(最低高度)前置结点的之间的跨度[这两个结点很可能不是同一个结点]
		oldSpan := rank[0] - rank[i]
		// 新结点到(当前高度)下一结点的跨度就等于
		// (当前高度的)前置结点到(当前高度)下一结点的跨度 减去oldSpan
		newNode.Levels[i].Span = prevNodes[i].Levels[i].Span - oldSpan
		// (当前高度的)前置结点的下一个结点指向的就是新结点
		// 因为新结点的插入，那么(当前高度的)前置结点到下一结点的跨度必然要加一
		prevNodes[i].Levels[i].Span = oldSpan + 1
	}
	// (如果有)更高高度结点的跨度需要加一
	for i := level; i < this.Level; i++ {
		prevNodes[i].Levels[i].Span++
	}
	
	// 最下面一层实际是个双向链表
	// 将向后(方向)的链表链接起来
	if prevNodes[0] == this.Head {
		// backward不指向头结点，而是指向nil
		// 方便统一的nil判断
		newNode.Backward = nil
	} else {
		newNode.Backward = prevNodes[0]
	}
	
	if newNode.Levels[0].Forward != nil {
		newNode.Levels[0].Forward.Backward = newNode
	} else {
		this.Tail = newNode
	}
	// 更新跳跃表中的结点总数
	this.Length++
	
	// 至此，结点正确插入到跳跃表中
	// 并且结点间的索引关系也得到正确维护
	return newNode, true
}

// 通过分数和值查找指定结点，并更新前置结点数组
// 注意和Get方法区分：Get方法找到结点就返回，这个方法还要更新前置结点数组，更消耗一些
func (this *SkipList) findNode(data *NodeData, prevNodes *[SKIPLIST_MAXLEVEL]*Node) (*Node, bool) {
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		for current != nil && current.Data.LessThan(data) {
			prev = current
			current = prev.Levels[i].Forward
		}
		// 更新每一高度的前置结点
		prevNodes[i] = prev
		// 继续循环，到下一高度查找和处理
	}
	current := prev.Levels[0].Forward
	if current != nil && current.Data.EqualTo(data) {
		return current, true
	}
	// 找不到指定的结点(必须要分数和数据都等，才算是同一个结点)
	return nil, false
}

// 注意，go和C/C++甚至Java不同的地方
// 在go中，数组是值传递——传递给一个函数时，是拷贝原数组而不是传递的指针(引用)
// 为了避免拷贝这里传递数组的指针
func (this *SkipList) deleteNode(current *Node, prevNodes *[SKIPLIST_MAXLEVEL]*Node) *Node{
	//	1.移除结点
	//	2.处理结点每一层的索引关系
	for i := 0; i < this.Level; i++ {
		if prevNodes[i].Levels[i].Forward == current {
			// 更新当前高度的前置结点到下一结点的跨度
			// 减一是因为删除一个结点，跨度当然就会减一
			prevNodes[i].Levels[i].Span += current.Levels[i].Span - 1
			// 更新前置结点指向的结点
			prevNodes[i].Levels[i].Forward = current.Levels[i].Forward
		} else {
			// 此时prevNodes[i].Levels[i].Forward应该是nil
			// 那么跨度减一即可
			prevNodes[i].Levels[i].Span--
		}
	}
	if current.Levels[0].Forward != nil {
		current.Levels[0].Forward.Backward = current.Backward
	} else {
		// 被删除的结点是尾结点，那么其前置结点成为新的尾结点
		this.Tail = current.Backward
	}
	
	//重新设置高度
	for this.Level > 1 &&
		this.Head.Levels[this.Level-1].Forward == nil {
		this.Level--
	}
	// 结点总数减一
	this.Length--
	// 至此，指定某个结点已被删除
	// 并正确维护了剩余结点间的关系
	return current
}

// 根据分数和值，删除指定结点
// 时间复杂度为O(logn)
// 空间复杂度为O(1)
func (this *SkipList) Delete(data *NodeData) (*Node, bool) {
	// 断言(不允许传入nil)
	assert.Assert(data != nil, "val must not be nil")
	// 断言(判断传入的分数值)
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)
	
	// 注意这里，使用数组而不是切片，避免不必要的堆内存分配
	// 当前这种情况，(只要该函数不返回数组)数组就是分配在栈上的
	prevNodes := [SKIPLIST_MAXLEVEL]*Node{}
	current, ok := this.findNode(data, &prevNodes)
	if !ok {
		return current, ok
	}
	return this.deleteNode(current, &prevNodes), true
}

/*
 rank相关操作
*/

// 这个方法的实现和Get()几乎一模一样
func (this *SkipList) GetRank(data *NodeData) int {
	// 断言(不允许传入nil)
	assert.Assert(data != nil, "val must not be nil")
	// 断言(判断传入的分数值)
	assert.Assert(!math.IsNaN(data.Score), "score is not a number:", data.Score)
	
	rank := 0
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		// 满足两个条件，查找就继续在当前高度向前继续查找：
		// 1.指定分数大于当前结点的分数，说明要查找的节点一定在当前结点的前方;
		// 2.当前节点不能是(哨兵)尾结点
		current := prev.Levels[i].Forward
		for current != nil && (current.Data.LessThan(data) || current.Data.EqualTo(data)) {
			// 累计跨度
			rank += prev.Levels[i].Span
			
			// 双指针继续向前移动
			prev = current
			current = prev.Levels[i].Forward
		}
		if prev != nil && prev.Data.EqualTo(data) {
			// 找到了
			return rank
		}
		// 否则1.score分数比当前结点的分数更大,或者
		// 2.当前(高度),这个结点是nil
		// 所以再次进入循环，到更低高度去查找
	}
	// 没有找到
	return 0
}

func (this *SkipList) GetNodeByRank(rank int) *Node {
	assert.Assert(rank > 0, "rank must >= 0,rank:", rank)
	
	traversed := 0
	prev := this.Head
	for i := this.Level - 1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		// 当在当前高度，累计的跨度小于等于指定排名时继续向右查找
		for current != nil &&
			(traversed+prev.Levels[i].Span) <= rank {
			// 累加跨度
			traversed += prev.Levels[i].Span
			
			// 指针前进
			prev = current
			current = prev.Levels[i].Forward
		}
		if traversed == rank {
			return prev
		}
		// else 当前高度未找到结点，继续循环进入更低一层查找
	}
	return nil
}

func (this *SkipList) GetRangeByRank(start int, end int) []*NodeData {
	assert.Assert(start > 0 && end > 0 && start <= end, "rank范围不合法, start:", start, " end:", end)
	
	// 找到在指定范围中最小的结点(如果没有就是nil)
	current := this.GetNodeByRank(start)
	traversed := start
	datas := make([]*NodeData, 0, 4)
	// 在给定的排名范围内，依次遍历结点
	for current != nil && traversed <= end {
		next := current.Levels[0].Forward
		datas = append(datas, current.Data)
		traversed++
		current = next
	}
	return datas
}

func (this *SkipList) DeleteRangeByRank(start int, end int) []*NodeData {
	assert.Assert(start > 0 && end > 0 && start <= end, "rank范围不合法, start:", start, " end:", end)
	
	// 注意这里，使用数组而不是切片，避免不必要的堆内存分配
	// 当前这种情况，(只要该函数不返回数组)数组就是分配在栈上的
	prevNodes := [SKIPLIST_MAXLEVEL]*Node{}
	traversed := 0
	prev := this.Head
	var current *Node = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		// 走过的跨度小于指定的起始位置时继续在当前高度向右前进
		for current != nil &&
			(traversed+prev.Levels[i].Span) < start {
			// 累加跨度
			traversed += prev.Levels[i].Span
			
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}
	// 循环结束，找到在指定范围中最小的结点(如果没有就是nil)
	
	// 前面的for循环累加跨度确定了前置结点的排名，这里加一得到当前结点的排名
	traversed++
	assert.Assert(traversed == start,
		"traversed must equal to start. traversed:", traversed, " start:", start)
	// 在给定的排名范围内，依次删除结点
	deleted := make([]*NodeData, 0, 4)
	for current != nil && traversed <= end {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		traversed++
		current = next
	}
	return deleted
}

/*
 Score相关操作
 */

func (this *SkipList) UpdateScore(data *NodeData, newScore float64) (*Node, bool) {
	// 断言(不允许传入nil)
	assert.Assert(data != nil, "data must not be nil")
	// 断言(判断传入的分数值)
	assert.Assert(!math.IsNaN(data.Score), "oldScore is not a number:", data.Score)
	assert.Assert(!math.IsNaN(newScore), "newScore is not a number:", newScore)
	
	// 注意这里，使用数组而不是切片，避免不必要的堆内存分配
	// 当前这种情况，(只要该函数不返回数组)数组就是分配在栈上的
	prevNodes := [SKIPLIST_MAXLEVEL]*Node{}
	current, ok := this.findNode(data, &prevNodes)
	if !ok {
		// 找不到指定结点，就返回
		return current, ok
	}
	
	// 先看能不能复用之前的结点对象
	// 如果新分数和旧的分数的位置一样不会变化的话就可以复用之前的旧结点
	// 那么就只需要更新结点的分数即可
	if (current.Backward == nil || current.Backward.Data.Score < newScore) &&
		(current.Levels[0].Forward == nil || current.Levels[0].Forward.Data.Score > newScore) {
		current.Data.Score = newScore
		return current, true
	}
	
	// 不能复用的话，那就需要删除结点
	this.deleteNode(current, &prevNodes)
	// 然后插入新的结点 (前文的逻辑已经保证了这里肯定能插入成功)
	data.Score = newScore
	return this.Insert(data)
}

func scoreGeaterThanMin(score float64, r *RangeSpecified) bool{
	if r.MinExclusive {
		return score > r.Min
	} else {
		return score >= r.Min
	}
}

func scoreLessThanMax(score float64, r *RangeSpecified) bool{
	if r.MaxExclusive {
		return score < r.Max
	} else {
		return score <= r.Max
	}
}

func (this *SkipList) isInRange(r *RangeSpecified) bool{
	if r.Min > r.Max ||
		(r.Min == r.Max &&
			(r.MinExclusive || r.MaxExclusive)) {
		// 指定范围不合法
		return false
	}
	
	// 判断最右边值边界是否合法
	// 即尾结点的分数要比范围的最小值大才合法
	last := this.Tail
	if last == nil ||
		!scoreGeaterThanMin(last.Data.Score, r) {
		return false
	}
	
	// 再判断最左边值是否合法
	// 即第一个结点的分数要比范围的最大值小才合法
	first := this.Head.Levels[0].Forward
	if first == nil ||
		!scoreLessThanMax(first.Data.Score, r) {
		return false
	}
	
	return true
}

func (this *SkipList) FirstInRange(r *RangeSpecified) *Node{
	assert.Assert(r != nil, "r range cannot be nil")
	if !this.isInRange(r) {
		return nil
	}
	
	prev := this.Head
	var current *Node = nil
	for i := this.Level-1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		// 如果当前结点的分数小于指定范围的最小分数
		// 则继续在当前高度向右查找
		for current != nil &&
			!scoreGeaterThanMin(current.Data.Score, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		// 进入下一层继续查找
	}
	// 因为函数第一步就判断了给定的范围是否在跳跃表中
	// 所以这里一定不能为空
	assert.Assert(current != nil, "current != nil,range r:", r)
	
	// 再判断一下找到的结点的分数
	// 一定要比指定范围的最大值小才行
	if !scoreLessThanMax(current.Data.Score, r) {
		return nil
	}
	return current
}

func (this *SkipList) LastInRange(r *RangeSpecified) *Node{
	assert.Assert(r != nil, "r range cannot be nil")
	if !this.isInRange(r) {
		return nil
	}
	
	prev := this.Head
	for i := this.Level-1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		// 如果当前结点的分数小于指定范围的最大分数
		// 则继续在当前高度向右查找
		for current != nil &&
			scoreLessThanMax(current.Data.Score, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		// 进入下一层继续查找
	}
	// 因为函数第一步就判断了给定的范围是否在跳跃表中
	// 所以这里一定不能为空
	assert.Assert(prev != nil, "prev != nil,range r:", r)
	
	// 再判断一下找到的结点的分数
	// 一定要比指定范围的最小值大才行
	if !scoreGeaterThanMin(prev.Data.Score, r) {
		return nil
	}
	return prev
}

func (this *SkipList) GetRangeByScore(r *RangeSpecified) []*NodeData {
	current := this.FirstInRange(r)
	datas := make([]*NodeData, 0, 4)
	// 从范围中最小的结点开始向右遍历，依次遍历结点
	// 直到范围结束
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		next := current.Levels[0].Forward
		datas = append(datas, current.Data)
		current = next
	}
	return datas
}

func (this *SkipList) DeleteRangeByScore(r *RangeSpecified) []*NodeData {
	// 注意这里，使用数组而不是切片，避免不必要的堆内存分配
	// 当前这种情况，(只要该函数不返回数组)数组就是分配在栈上的
	prevNodes := [SKIPLIST_MAXLEVEL]*Node{}
	prev := this.Head
	var current *Node = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		// 当前结点的分数小于指定的范围的最小分数就继续在当前高度向右查找
		// !scoreGeaterThanMin == ValueLessThanMin
		for current != nil &&
			!scoreGeaterThanMin(current.Data.Score, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}
	// 循环结束，找到在指定范围中最小的结点(如果没有就是nil)
	
	deleted := make([]*NodeData, 0, 4)
	// 从范围中最小的结点开始向右遍历，依次删除结点
	// 直到范围结束
	for current != nil && scoreLessThanMax(current.Data.Score, r) {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		current = next
	}
	return deleted
}

/*
 Value相关操作
 */

func valueGeaterThanMin(val Value, r *ValueRangeSpecified) bool {
	if r.MinExclusive {
		return ValueGreaterThan(val, r.Max)
	} else {
		// 不小于，则认为就是大于等于
		return !val.LessThan(r.Min)
	}
}

func valueLessThanMax(val Value, r *ValueRangeSpecified) bool{
	if r.MaxExclusive {
		return val.LessThan(r.Max)
	} else {
		// 小于或者等于
		return val.LessThan(r.Max) || val.EqualTo(r.Max)
	}
}

func (this *SkipList) isInValueRange(r *ValueRangeSpecified) bool{
	if ValueGreaterThan(r.Min, r.Max) ||
		(r.Min.EqualTo(r.Max) && (r.MinExclusive || r.MaxExclusive)) {
		// 指定范围不合法
		return false
	}
	
	// 判断最右边值边界是否合法
	// 即尾结点的值要比范围的最小值大才合法
	last := this.Tail
	if last == nil || !valueGeaterThanMin(last.Data.Val, r) {
		return false
	}
	
	// 再判断最左边值是否合法
	// 即第一个结点的值要比范围的最大值小才合法
	first := this.Head.Levels[0].Forward
	if first == nil || !valueLessThanMax(first.Data.Val, r) {
		return false
	}
	
	return true
}

func (this *SkipList) FirstInValueRange(r *ValueRangeSpecified) *Node{
	assert.Assert(r != nil, "r range cannot be nil")
	if !this.isInValueRange(r) {
		return nil
	}
	
	prev := this.Head
	var current *Node = nil
	for i := this.Level-1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		// 如果当前结点的值小于指定范围的最小值
		// 则继续在当前高度向右查找
		for current != nil &&
			!valueGeaterThanMin(current.Data.Val, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		// 进入下一层继续查找
	}
	// 因为函数第一步就判断了给定的范围是否在跳跃表中
	// 所以这里一定不能为空
	assert.Assert(current != nil, "current != nil,range r:", r)
	
	// 再判断一下找到的结点的值
	// 一定要比指定范围的最大值小才行
	if !valueLessThanMax(current.Data.Val, r) {
		return nil
	}
	return current
}

func (this *SkipList) LastInValueRange(r *ValueRangeSpecified) *Node{
	assert.Assert(r != nil, "r range cannot be nil")
	if !this.isInValueRange(r) {
		return nil
	}
	
	prev := this.Head
	for i := this.Level-1; i >= 0; i-- {
		current := prev.Levels[i].Forward
		// 如果当前结点的值小于指定范围的最大值
		// 则继续在当前高度向右查找
		for current != nil &&
			valueLessThanMax(current.Data.Val, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		// 进入下一层继续查找
	}
	// 因为函数第一步就判断了给定的范围是否在跳跃表中
	// 所以这里一定不能为空
	assert.Assert(prev != nil, "prev != nil,range r:", r)
	
	// 再判断一下找到的结点的值
	// 一定要比指定范围的最小值大才行
	if !valueGeaterThanMin(prev.Data.Val, r) {
		return nil
	}
	return prev
}

func (this *SkipList) DeleteRangeByValue(r *ValueRangeSpecified) []*NodeData {
	// 注意这里，使用数组而不是切片，避免不必要的堆内存分配
	// 当前这种情况，(只要该函数不返回数组)数组就是分配在栈上的
	prevNodes := [SKIPLIST_MAXLEVEL]*Node{}
	prev := this.Head
	var current *Node = nil
	for i := this.Level - 1; i >= 0; i-- {
		current = prev.Levels[i].Forward
		// 当前结点的值小于指定的范围的最小值就继续在当前高度向右查找
		// !valueGeaterThanMin == ValueLessThanMin
		for current != nil &&
			!valueGeaterThanMin(current.Data.Val, r) {
			prev = current
			current = prev.Levels[i].Forward
		}
		prevNodes[i] = prev
	}
	// 循环结束，找到在指定范围中最小的结点(如果没有就是nil)
	
	deleted := make([]*NodeData, 0, 4)
	// 从范围中最小的结点开始向右遍历，依次删除结点
	// 直到范围结束
	for current != nil && valueLessThanMax(current.Data.Val, r) {
		next := current.Levels[0].Forward
		this.deleteNode(current, &prevNodes)
		deleted = append(deleted, current.Data)
		current = next
	}
	return deleted
}