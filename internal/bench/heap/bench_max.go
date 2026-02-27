// Package benchheap.

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
// 创建日期:2023/7/11
package benchheap

import (
	"fmt"
	"time"

	"github.com/stormYuanYang/yytools/pkg/algorithms/mathx/random"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/ds/heap"
)

var muniq = 1000000
var mmax = muniq

func maxHeapOpPushItem(h heap.InterfaceHeap[any], num int) interface{} {
	for i := 0; i < num; i++ {
		one := &heap.Item[any]{
			Data:   nil,
			Weight: muniq,
		}
		h.PushItem(one)
		muniq--
		assert.Assert(h.Length() > 0)
	}
	return nil
}

func maxHeapOpPopItem(h heap.InterfaceHeap[any], num int) interface{} {
	res := make([]*heap.Item[any], 0, num)
	for i := 0; i < num; i++ {
		if h.Length() > 0 {
			tmp := h.PeekItem()
			item := h.PopItem()
			assert.Assert(item.Weight == mmax)
			assert.Assert(item == tmp)
			mmax--
			res = append(res, item)
		}
	}
	// 必须是从大到小的
	for i := 0; i < len(res)-1; i++ {
		assert.Assert(res[i].Weight > res[i+1].Weight)
	}
	return res
}

func maxHeapOpPeekItem(h heap.InterfaceHeap[any], num int) interface{} {
	for i := 0; i < num; i++ {
		oldLength := h.Length()
		if oldLength > 0 {
			item := h.PeekItem()
			assert.Assert(item.Weight == mmax)
			assert.Assert(oldLength == h.Length())
		}
	}
	return nil
}

var maxHeapHandlers = []func(h heap.InterfaceHeap[any], num int) interface{}{
	maxHeapOpPushItem,
	maxHeapOpPopItem,
	maxHeapOpPeekItem,
}

func maxHeapMustBeLegal(h heap.InterfaceHeap[any], deleted []*heap.Item[any]) {
	items := maxHeapOpPopItem(h, h.Length()).([]*heap.Item[any])
	assert.Assert(h.Length() == 0)
	deleted = append(deleted, items...)
	for i := 0; i < len(deleted)-1; i++ {
		assert.Assert(deleted[i].Weight > deleted[i+1].Weight)
	}
}

func MaxHeapTest(num int) {
	println("最大堆测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	scale := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100, 1000, 10000, 100000, 1000000}
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		for k, s := range scale {
			var h heap.InterfaceHeap[any] = heap.NewMaxHeap[any]()
			deleted := []*heap.Item[any]{}
			muniq = 1000000
			mmax = muniq
			maxHeapOpPushItem(h, s)

			opCnt := 100000
			for j := 0; j < opCnt; j++ {
				r := random.RandInt(0, len(maxHeapHandlers)-1)
				res := maxHeapHandlers[r](h, 1)
				if res != nil {
					deleted = append(deleted, res.([]*heap.Item[any])...)
				}
			}
			maxHeapMustBeLegal(h, deleted)
			fmt.Printf("测试#%d. 起始长度:%d, 当前长度:%d\n", k, s, h.Length())
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("最大堆测试完毕...")
}
