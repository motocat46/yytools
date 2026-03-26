// 正确性命题测试 — Heap / MaxHeap / PriorityQueue
// 验证核心不变量：MinHeap 始终弹出最小值，MaxHeap 始终弹出最大值，
// PriorityQueue 始终弹出最高优先级，UpdatePriority 不破坏堆序。
// 非并发结构，无并发命题；重点在随机混合操作 + 参考模型对比。

package heap_test

import (
	"math/rand/v2"
	"testing"

	"github.com/motocat46/yytools/pkg/ds/heap"
)

// ── 参考模型辅助函数 ──────────────────────────────────────────────────────────

// refPopMin 在参考切片中找到权重最小的元素，将其从切片中移除并返回。O(n)。
func refPopMin(ref *[]*heap.Item[int]) *heap.Item[int] {
	minIdx := 0
	for i, item := range *ref {
		if item.Weight < (*ref)[minIdx].Weight {
			minIdx = i
		}
	}
	item := (*ref)[minIdx]
	last := len(*ref) - 1
	(*ref)[minIdx] = (*ref)[last]
	(*ref)[last] = nil
	*ref = (*ref)[:last]
	return item
}

// refPopMax 在参考切片中找到权重最大的元素，将其从切片中移除并返回。O(n)。
func refPopMax(ref *[]*heap.Item[int]) *heap.Item[int] {
	maxIdx := 0
	for i, item := range *ref {
		if item.Weight > (*ref)[maxIdx].Weight {
			maxIdx = i
		}
	}
	item := (*ref)[maxIdx]
	last := len(*ref) - 1
	(*ref)[maxIdx] = (*ref)[last]
	(*ref)[last] = nil
	*ref = (*ref)[:last]
	return item
}

// refPopMaxPriority 在参考切片中找到优先级最高的元素，移除并返回。O(n)。
func refPopMaxPriority(ref *[]*heap.PriorityItem[int]) *heap.PriorityItem[int] {
	maxIdx := 0
	for i, item := range *ref {
		if item.Priority > (*ref)[maxIdx].Priority {
			maxIdx = i
		}
	}
	item := (*ref)[maxIdx]
	last := len(*ref) - 1
	(*ref)[maxIdx] = (*ref)[last]
	(*ref)[last] = nil
	*ref = (*ref)[:last]
	return item
}

// ─── 命题 1：MinHeap 最小值语义 ──────────────────────────────────────────────
//
// 不变量：PopItem 始终返回当前权重最小的元素（与参考模型每步对比）。
// 数据量：100,000 次操作；堆大小上限 500，避免参考模型 O(n) 扫描使测试过慢。

func TestCorrectness_Heap_MinSemantics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机混合测试")
	}

	const (
		ops     = 100_000
		maxSize = 500
	)

	h := heap.NewHeap[int]()
	var ref []*heap.Item[int]
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range ops {
		if len(ref) == 0 || (len(ref) < maxSize && rng.IntN(2) == 0) {
			item := &heap.Item[int]{Data: i, Weight: rng.IntN(1_000_000)}
			h.PushItem(item)
			ref = append(ref, item)
		} else {
			got := h.PopItem()
			want := refPopMin(&ref)
			if got.Weight != want.Weight {
				t.Fatalf("第 %d 次操作（Pop）：got.Weight=%d，want.Weight=%d（最小值语义被破坏）",
					i+1, got.Weight, want.Weight)
			}
		}

		if got, want := h.Length(), len(ref); got != want {
			t.Fatalf("第 %d 次操作后：Length() = %d，期望 %d", i+1, got, want)
		}
	}
}

// ─── 命题 2：MaxHeap 最大值语义 ──────────────────────────────────────────────
//
// 不变量：PopItem 始终返回当前权重最大的元素。

func TestCorrectness_Heap_MaxSemantics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机混合测试")
	}

	const (
		ops     = 100_000
		maxSize = 500
	)

	h := heap.NewMaxHeap[int]()
	var ref []*heap.Item[int]
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range ops {
		if len(ref) == 0 || (len(ref) < maxSize && rng.IntN(2) == 0) {
			item := &heap.Item[int]{Data: i, Weight: rng.IntN(1_000_000)}
			h.PushItem(item)
			ref = append(ref, item)
		} else {
			got := h.PopItem()
			want := refPopMax(&ref)
			if got.Weight != want.Weight {
				t.Fatalf("第 %d 次操作（Pop）：got.Weight=%d，want.Weight=%d（最大值语义被破坏）",
					i+1, got.Weight, want.Weight)
			}
		}

		if got, want := h.Length(), len(ref); got != want {
			t.Fatalf("第 %d 次操作后：Length() = %d，期望 %d", i+1, got, want)
		}
	}
}

// ─── 命题 3：PriorityQueue 最高优先级语义 ────────────────────────────────────
//
// 不变量：PopItem 始终返回当前优先级最高（Priority 数值最大）的元素。

func TestCorrectness_PriorityQueue_MaxPrioritySemantics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机混合测试")
	}

	const (
		ops     = 100_000
		maxSize = 500
	)

	pq := heap.NewPriorityQueue[int]()
	var ref []*heap.PriorityItem[int]
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range ops {
		if len(ref) == 0 || (len(ref) < maxSize && rng.IntN(2) == 0) {
			item := &heap.PriorityItem[int]{Data: i, Priority: rng.IntN(1_000_000)}
			pq.PushItem(item)
			ref = append(ref, item)
		} else {
			got := pq.PopItem()
			want := refPopMaxPriority(&ref)
			if got.Priority != want.Priority {
				t.Fatalf("第 %d 次操作（Pop）：got.Priority=%d，want.Priority=%d（最高优先级语义被破坏）",
					i+1, got.Priority, want.Priority)
			}
		}

		if got, want := pq.Length(), len(ref); got != want {
			t.Fatalf("第 %d 次操作后：Length() = %d，期望 %d", i+1, got, want)
		}
	}
}

// ─── 命题 4：UpdatePriority 不破坏堆序 ───────────────────────────────────────
//
// 不变量：随机更新队列中任意元素的优先级后，Pop 序列仍严格按优先级降序排列。

func TestCorrectness_PriorityQueue_UpdatePriority(t *testing.T) {
	const n = 5_000
	pq := heap.NewPriorityQueue[int]()
	rng := rand.New(rand.NewPCG(42, 0))

	items := make([]*heap.PriorityItem[int], n)
	for i := range n {
		items[i] = &heap.PriorityItem[int]{Data: i, Priority: rng.IntN(100_000)}
		pq.PushItem(items[i])
	}

	// 随机更新 2,000 次优先级，所有元素均在队列中
	for range 2_000 {
		item := items[rng.IntN(n)]
		pq.UpdatePriority(item, rng.IntN(100_000))
	}

	// Pop 所有元素，验证优先级降序
	prev := pq.PopItem()
	if prev == nil {
		t.Fatal("队列为空，无法测试")
	}
	for pq.Length() > 0 {
		curr := pq.PopItem()
		if curr.Priority > prev.Priority {
			t.Fatalf("UpdatePriority 后堆序被破坏：弹出 Priority=%d 后又弹出 Priority=%d（应 ≤ 前者）",
				prev.Priority, curr.Priority)
		}
		prev = curr
	}
}
