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
//
// 作者: yangyuan
// 创建日期: 2026-03-28
package timingwheel

import (
	"container/heap"
	"slices"
)

// timerHeap 是 *Timer 的 min-heap（按 expireAt 升序），实现 container/heap.Interface。
// 由 TimingWheel.overflowMu 保护并发写；advanceClock（writeLock 下）读不需要 overflowMu。
type timerHeap []*Timer

func (h timerHeap) Len() int           { return len(h) }
func (h timerHeap) Less(i, j int) bool { return h[i].expireAt < h[j].expireAt }
func (h timerHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *timerHeap) Push(x any)        { *h = append(*h, x.(*Timer)) }
func (h *timerHeap) Pop() any {
	old := *h
	n := len(old)
	t := old[n-1]
	*h = slices.Delete(old, n-1, n)
	return t
}

// heapPeek 返回堆顶 timer（expireAt 最小），不弹出；堆为空时返回 nil。
func heapPeek(h *timerHeap) *Timer {
	if len(*h) == 0 {
		return nil
	}
	return (*h)[0]
}

// heapPop 弹出堆顶 timer；调用前需确认堆非空。
func heapPop(h *timerHeap) *Timer {
	return heap.Pop(h).(*Timer)
}

// heapPush 将 timer 推入堆。
func heapPush(h *timerHeap, t *Timer) {
	heap.Push(h, t)
}
