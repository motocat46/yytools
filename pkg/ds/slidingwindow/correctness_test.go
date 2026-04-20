package slidingwindow_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	sw "github.com/motocat46/yytools/pkg/ds/slidingwindow"
)

// refWindow 是 Window 的参考实现，使用 []int 全量扫描，语义与 Window 完全一致。
type refWindow struct {
	size int
	data []int
}

func newRefWindow(size int) *refWindow { return &refWindow{size: size} }

func (r *refWindow) add(v int) {
	r.data = append(r.data, v)
	if len(r.data) > r.size {
		r.data = r.data[len(r.data)-r.size:]
	}
}

func (r *refWindow) sum() int {
	s := 0
	for _, v := range r.data {
		s += v
	}
	return s
}

func (r *refWindow) max() int {
	m := r.data[0]
	for _, v := range r.data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func (r *refWindow) min() int {
	m := r.data[0]
	for _, v := range r.data[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func (r *refWindow) avg() float64 { return float64(r.sum()) / float64(len(r.data)) }
func (r *refWindow) len() int     { return len(r.data) }
func (r *refWindow) empty() bool  { return len(r.data) == 0 }
func (r *refWindow) full() bool   { return len(r.data) == r.size }

// TestCorrectness_RandomOps 使用随机操作序列对比 Window 与参考模型，验证所有命题。
// 操作次数 100,000，窗口大小随机在 [1,50] 之间。
func TestCorrectness_RandomOps(t *testing.T) {
	const ops = 100_000
	rng := rand.New(rand.NewPCG(42, 0))

	size := int(rng.IntN(50)) + 1
	w := sw.New[int](size)
	ref := newRefWindow(size)

	for i := range ops {
		v := int(rng.IntN(10_000)) - 5_000 // 包含负数

		w.Add(v)
		ref.add(v)

		// 命题1：Len 一致
		if w.Len() != ref.len() {
			t.Fatalf("op %d: Len mismatch: got %d, want %d", i, w.Len(), ref.len())
		}
		// 命题2：Empty/Full 一致
		if w.Empty() != ref.empty() {
			t.Fatalf("op %d: Empty mismatch", i)
		}
		if w.Full() != ref.full() {
			t.Fatalf("op %d: Full mismatch", i)
		}
		// 命题3：Sum 精确
		if int(w.Sum()) != ref.sum() {
			t.Fatalf("op %d: Sum mismatch: got %d, want %d", i, w.Sum(), ref.sum())
		}
		// 命题4：Avg 精确（非空时）
		if !w.Empty() {
			gotAvg := w.Avg()
			wantAvg := ref.avg()
			diff := gotAvg - wantAvg
			if diff < -1e-9 || diff > 1e-9 {
				t.Fatalf("op %d: Avg mismatch: got %v, want %v", i, gotAvg, wantAvg)
			}
		}
		// 命题5：Max 精确（非空时）
		if !w.Empty() {
			if int(w.Max()) != ref.max() {
				t.Fatalf("op %d: Max mismatch: got %d, want %d (window size=%d, data=%v)",
					i, w.Max(), ref.max(), size, ref.data)
			}
		}
		// 命题6：Min 精确（非空时）
		if !w.Empty() {
			if int(w.Min()) != ref.min() {
				t.Fatalf("op %d: Min mismatch: got %d, want %d (window size=%d, data=%v)",
					i, w.Min(), ref.min(), size, ref.data)
			}
		}
	}
}

// TestCorrectness_AllDuplicates 全相同值场景（Max==Min==Sum/N，单调队列退化）。
func TestCorrectness_AllDuplicates(t *testing.T) {
	w := sw.New[int](5)
	for range 20 {
		w.Add(7)
	}
	if w.Max() != 7 {
		t.Errorf("got Max=%d, want 7", w.Max())
	}
	if w.Min() != 7 {
		t.Errorf("got Min=%d, want 7", w.Min())
	}
	if w.Sum() != 35 {
		t.Errorf("got Sum=%d, want 35", w.Sum())
	}
}

// TestCorrectness_StrictlyDecreasing 严格递减序列（Max 始终为最新加入，Min 为最旧）。
func TestCorrectness_StrictlyDecreasing(t *testing.T) {
	const size = 5
	w := sw.New[int](size)
	ref := newRefWindow(size)
	rng := rand.New(rand.NewPCG(99, 0))

	// 生成严格递减序列
	vals := make([]int, 30)
	v := 10000
	for i := range vals {
		v -= int(rng.IntN(10)) + 1
		vals[i] = v
	}

	for i, val := range vals {
		w.Add(val)
		ref.add(val)
		if !w.Empty() {
			if int(w.Max()) != ref.max() {
				t.Fatalf("step %d: Max mismatch: got %d, want %d", i, w.Max(), ref.max())
			}
			if int(w.Min()) != ref.min() {
				t.Fatalf("step %d: Min mismatch: got %d, want %d", i, w.Min(), ref.min())
			}
		}
	}
}

// TestCorrectness_StrictlyIncreasing 严格递增序列。
func TestCorrectness_StrictlyIncreasing(t *testing.T) {
	const size = 5
	w := sw.New[int](size)
	ref := newRefWindow(size)

	for i := range 30 {
		w.Add(i)
		ref.add(i)
		if !w.Empty() {
			if int(w.Max()) != ref.max() {
				t.Fatalf("step %d: Max mismatch: got %d, want %d", i, w.Max(), ref.max())
			}
			if int(w.Min()) != ref.min() {
				t.Fatalf("step %d: Min mismatch: got %d, want %d", i, w.Min(), ref.min())
			}
		}
	}
}

// TestCorrectness_SizeOne 窗口大小为 1 的边界情况。
func TestCorrectness_SizeOne(t *testing.T) {
	w := sw.New[int](1)
	for i := range 100 {
		w.Add(i)
		if w.Len() != 1 {
			t.Fatalf("step %d: Len=%d, want 1", i, w.Len())
		}
		if int(w.Max()) != i || int(w.Min()) != i || int(w.Sum()) != i {
			t.Fatalf("step %d: stats mismatch: Max=%d Min=%d Sum=%d, want all %d",
				i, w.Max(), w.Min(), w.Sum(), i)
		}
	}
}

// TestCorrectness_NegativeValues 含负数的场景。
func TestCorrectness_NegativeValues(t *testing.T) {
	w := sw.New[int](3)
	w.Add(-5)
	w.Add(-1)
	w.Add(-3)
	if int(w.Max()) != -1 {
		t.Errorf("got Max=%d, want -1", w.Max())
	}
	if int(w.Min()) != -5 {
		t.Errorf("got Min=%d, want -5", w.Min())
	}
	if int(w.Sum()) != -9 {
		t.Errorf("got Sum=%d, want -9", w.Sum())
	}
}

// TestCorrectness_FloatWindow 浮点数窗口。
func TestCorrectness_FloatWindow(t *testing.T) {
	w := sw.New[float64](4)
	w.Add(1.5)
	w.Add(2.5)
	w.Add(0.5)
	w.Add(3.0)
	if w.Max() != 3.0 {
		t.Errorf("got Max=%v, want 3.0", w.Max())
	}
	if w.Min() != 0.5 {
		t.Errorf("got Min=%v, want 0.5", w.Min())
	}
}

// TestCorrectness_LargeRandom 大规模随机测试：10万次操作，多种窗口大小。
func TestCorrectness_LargeRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模正确性测试")
	}
	sizes := []int{1, 2, 7, 50, 100, 500}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			const ops = 100_000
			rng := rand.New(rand.NewPCG(uint64(size), 0))
			w := sw.New[int](size)
			ref := newRefWindow(size)
			for i := range ops {
				v := int(rng.IntN(200_000)) - 100_000
				w.Add(v)
				ref.add(v)
				if !w.Empty() {
					if int(w.Max()) != ref.max() {
						t.Fatalf("op %d size %d: Max got %d, want %d", i, size, w.Max(), ref.max())
					}
					if int(w.Min()) != ref.min() {
						t.Fatalf("op %d size %d: Min got %d, want %d", i, size, w.Min(), ref.min())
					}
					if int(w.Sum()) != ref.sum() {
						t.Fatalf("op %d size %d: Sum got %d, want %d", i, size, w.Sum(), ref.sum())
					}
				}
			}
		})
	}
}
