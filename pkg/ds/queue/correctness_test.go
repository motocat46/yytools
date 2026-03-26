// 正确性命题测试 — Queue
// 验证 Queue 的核心不变量：长度精确性、FIFO 顺序、ring buffer 环绕完整性、扩缩容安全性。
// 非并发结构，无并发命题；重点在随机混合操作 + 参考模型对比。

package queue_test

import (
	"math/rand/v2"
	"testing"

	"github.com/motocat46/yytools/pkg/ds/queue"
)

// ─── 命题 1：Length 精确性 ────────────────────────────────────────────────────
//
// 不变量：每次 Enqueue/Dequeue 后，Len() 严格等于当前元素数量，Empty() 严格等于 (Len==0)。

func TestCorrectness_Queue_LengthPrecision(t *testing.T) {
	q := queue.NewQueue[int]()
	count := 0
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 10_000 {
		if count == 0 || rng.IntN(2) == 0 {
			q.Enqueue(rng.IntN(10_000))
			count++
		} else {
			q.Dequeue()
			count--
		}
		if got := q.Len(); got != count {
			t.Fatalf("第 %d 次操作后：Len() = %d，期望 %d", i+1, got, count)
		}
		if got, want := q.Empty(), count == 0; got != want {
			t.Fatalf("第 %d 次操作后：Empty() = %v，期望 %v", i+1, got, want)
		}
	}
}

// ─── 命题 2：FIFO 顺序 ────────────────────────────────────────────────────────
//
// 不变量：Dequeue 的顺序严格等于 Enqueue 的顺序。

func TestCorrectness_Queue_FIFOOrder(t *testing.T) {
	const n = 10_000
	q := queue.NewQueue[int]()
	rng := rand.New(rand.NewPCG(42, 0))

	enqueued := make([]int, n)
	for i := range n {
		v := rng.IntN(1_000_000)
		enqueued[i] = v
		q.Enqueue(v)
	}

	for i := range n {
		got := q.Dequeue()
		if got != enqueued[i] {
			t.Fatalf("Dequeue 第 %d 次：got %d，want %d（FIFO 顺序被破坏）", i+1, got, enqueued[i])
		}
	}

	if !q.Empty() {
		t.Fatal("Dequeue 完所有元素后 Empty() = false")
	}
}

// ─── 命题 3：环绕完整性 ───────────────────────────────────────────────────────
//
// 不变量：多次 Enqueue/Dequeue 交替使 ring buffer 的 Head 前移，触发 Tail 环绕后，
// FIFO 顺序仍然正确，无元素丢失或错位。

func TestCorrectness_Queue_WrapAroundIntegrity(t *testing.T) {
	// 从小容量开始，强制触发环绕和扩容
	q := queue.NewQueueWithSize[int](8)
	var ref []int
	rng := rand.New(rand.NewPCG(42, 0))

	for round := range 100 {
		// 入队：超过当前容量，触发扩容和环绕
		n := 10 + rng.IntN(50)
		for range n {
			v := round*1000 + rng.IntN(1000)
			q.Enqueue(v)
			ref = append(ref, v)
		}

		// 出队部分元素：制造 Head > 0，下轮入队时 Tail 将环绕
		m := rng.IntN(len(ref) + 1)
		for range m {
			got := q.Dequeue()
			if got != ref[0] {
				t.Fatalf("第 %d 轮 Dequeue：got %d，want %d（环绕后顺序错误）",
					round+1, got, ref[0])
			}
			ref = ref[1:]
		}

		if got := q.Len(); got != len(ref) {
			t.Fatalf("第 %d 轮后：Len() = %d，期望 %d", round+1, got, len(ref))
		}
	}

	// 清空剩余，验证顺序完整
	for len(ref) > 0 {
		got := q.Dequeue()
		if got != ref[0] {
			t.Fatalf("清空阶段 Dequeue：got %d，want %d", got, ref[0])
		}
		ref = ref[1:]
	}

	if !q.Empty() {
		t.Fatal("清空后 Empty() = false")
	}
}

// ─── 命题 4：扩缩容完整性 ─────────────────────────────────────────────────────
//
// 不变量：大量 Enqueue 触发多次扩容、大量 Dequeue 触发多次缩容后，
// 剩余元素仍可按 FIFO 顺序完整取出，无丢失无错乱。

func TestCorrectness_Queue_ExpandShrinkIntegrity(t *testing.T) {
	const (
		enqN   = 10_000
		deqN   = 9_750 // 触发多次缩容
		remain = enqN - deqN
	)
	q := queue.NewQueue[int]()
	rng := rand.New(rand.NewPCG(42, 0))

	enqueued := make([]int, enqN)
	for i := range enqN {
		v := rng.IntN(1_000_000)
		enqueued[i] = v
		q.Enqueue(v)
	}

	for range deqN {
		q.Dequeue()
	}

	if got := q.Len(); got != remain {
		t.Fatalf("缩容后 Len() = %d，期望 %d", got, remain)
	}

	for i := deqN; i < enqN; i++ {
		got := q.Dequeue()
		if got != enqueued[i] {
			t.Fatalf("缩容后第 %d 次 Dequeue：got %d，want %d（扩缩容破坏了元素完整性）",
				i-deqN+1, got, enqueued[i])
		}
	}

	if !q.Empty() {
		t.Fatal("所有元素取出后 Empty() = false")
	}
}

// ─── 命题 5：随机混合操作参考模型对比 ─────────────────────────────────────────
//
// 不变量：任意 Enqueue/Dequeue/Peek 操作序列下，Queue 行为与参考模型（标准 Go slice）完全一致。
// 数据量：100,000 次操作。

func TestCorrectness_Queue_RandomMixed_ReferenceModel(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机混合测试")
	}

	const ops = 100_000
	q := queue.NewQueue[int]()
	var ref []int
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range ops {
		op := rng.IntN(3)
		switch {
		case op == 0 || len(ref) == 0:
			// Enqueue
			v := rng.IntN(1_000_000)
			q.Enqueue(v)
			ref = append(ref, v)
		case op == 1:
			// Dequeue
			got := q.Dequeue()
			want := ref[0]
			ref = ref[1:]
			if got != want {
				t.Fatalf("第 %d 次操作（Dequeue）：got %d，want %d", i+1, got, want)
			}
		default:
			// Peek
			got := q.Peek()
			want := ref[0]
			if got != want {
				t.Fatalf("第 %d 次操作（Peek）：got %d，want %d", i+1, got, want)
			}
		}

		if got, want := q.Len(), len(ref); got != want {
			t.Fatalf("第 %d 次操作后：Len() = %d，期望 %d", i+1, got, want)
		}
	}
}
