// 正确性命题测试 — Stack
// 验证 Stack 的核心不变量：长度精确性、LIFO 顺序、缩容安全性。
// 非并发结构，无并发命题；重点在随机混合操作 + 参考模型对比。

package stack_test

import (
	"math/rand/v2"
	"testing"

	"github.com/motocat46/yytools/pkg/ds/stack"
)

// ─── 命题 1：Length 精确性 ────────────────────────────────────────────────────
//
// 不变量：每次 Push/Pop 后，Length() 严格等于当前元素数量，Empty() 严格等于 (Length==0)。

func TestCorrectness_Stack_LengthPrecision(t *testing.T) {
	s := stack.NewStack[int]()
	count := 0
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range 10_000 {
		if count == 0 || rng.IntN(2) == 0 {
			s.Push(rng.IntN(10_000))
			count++
		} else {
			s.Pop()
			count--
		}
		if got := s.Length(); got != count {
			t.Fatalf("第 %d 次操作后：Length() = %d，期望 %d", i+1, got, count)
		}
		if got, want := s.Empty(), count == 0; got != want {
			t.Fatalf("第 %d 次操作后：Empty() = %v，期望 %v", i+1, got, want)
		}
	}
}

// ─── 命题 2：LIFO 顺序 ────────────────────────────────────────────────────────
//
// 不变量：Pop 的顺序严格等于 Push 顺序的逆序。

func TestCorrectness_Stack_LIFOOrder(t *testing.T) {
	const n = 10_000
	s := stack.NewStack[int]()
	rng := rand.New(rand.NewPCG(42, 0))

	pushed := make([]int, n)
	for i := range n {
		v := rng.IntN(1_000_000)
		pushed[i] = v
		s.Push(v)
	}

	for i := n - 1; i >= 0; i-- {
		got := s.Pop()
		if got != pushed[i] {
			t.Fatalf("Pop 第 %d 次：got %d，want %d（LIFO 顺序被破坏）", n-i, got, pushed[i])
		}
	}

	if !s.Empty() {
		t.Fatal("Pop 完所有元素后 Empty() = false")
	}
}

// ─── 命题 3：缩容安全性 ───────────────────────────────────────────────────────
//
// 不变量：大量 Pop 触发多次缩容后，剩余元素仍可按 LIFO 顺序完整取出，无丢失无错乱。

func TestCorrectness_Stack_ShrinkSafety(t *testing.T) {
	const (
		pushN  = 10_000
		popN   = 9_750 // 触发多次缩容
		remain = pushN - popN
	)
	s := stack.NewStack[int]()
	rng := rand.New(rand.NewPCG(42, 0))

	pushed := make([]int, pushN)
	for i := range pushN {
		v := rng.IntN(1_000_000)
		pushed[i] = v
		s.Push(v)
	}

	for range popN {
		s.Pop()
	}

	if got := s.Length(); got != remain {
		t.Fatalf("缩容后 Length() = %d，期望 %d", got, remain)
	}

	// 已 pop 了 pushN 个中的后 popN 个（栈顶部分），剩余是 pushed[0..remain-1]，
	// 栈顶为 pushed[remain-1]，按 LIFO 顺序从 remain-1 降至 0。
	for i := remain - 1; i >= 0; i-- {
		got := s.Pop()
		if got != pushed[i] {
			t.Fatalf("缩容后第 %d 次 Pop：got %d，want %d（缩容破坏了元素完整性）",
				remain-i, got, pushed[i])
		}
	}

	if !s.Empty() {
		t.Fatal("所有元素取出后 Empty() = false")
	}
}

// ─── 命题 4：随机混合操作参考模型对比 ─────────────────────────────────────────
//
// 不变量：任意 Push/Pop/Top 操作序列下，Stack 行为与参考模型（标准 Go slice）完全一致。
// 数据量：100,000 次操作。

func TestCorrectness_Stack_RandomMixed_ReferenceModel(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机混合测试")
	}

	const ops = 100_000
	s := stack.NewStack[int]()
	var ref []int
	rng := rand.New(rand.NewPCG(42, 0))

	for i := range ops {
		op := rng.IntN(3)
		switch {
		case op == 0 || len(ref) == 0:
			// Push
			v := rng.IntN(1_000_000)
			s.Push(v)
			ref = append(ref, v)
		case op == 1:
			// Pop
			got := s.Pop()
			want := ref[len(ref)-1]
			ref = ref[:len(ref)-1]
			if got != want {
				t.Fatalf("第 %d 次操作（Pop）：got %d，want %d", i+1, got, want)
			}
		default:
			// Top
			got := s.Top()
			want := ref[len(ref)-1]
			if got != want {
				t.Fatalf("第 %d 次操作（Top）：got %d，want %d", i+1, got, want)
			}
		}

		if got, want := s.Length(), len(ref); got != want {
			t.Fatalf("第 %d 次操作后：Length() = %d，期望 %d", i+1, got, want)
		}
	}
}
