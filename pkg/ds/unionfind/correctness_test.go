package unionfind_test

import (
	"math/rand/v2"
	"testing"

	uf "github.com/motocat46/yytools/pkg/ds/unionfind"
)

// refUF 是并查集的参考模型，用 map[int]int 追踪每个元素所属的组 ID。
// 语义与 UnionFind 完全相同，实现最朴素（O(N) union），用于随机测试对比。
type refUF struct {
	group map[int]int // group[x] = x 所在的组 ID（代表元）
}

func newRefUF() *refUF { return &refUF{group: make(map[int]int)} }

func (r *refUF) register(x int) {
	if _, ok := r.group[x]; !ok {
		r.group[x] = x
	}
}

func (r *refUF) find(x int) int {
	r.register(x)
	return r.group[x]
}

// union 合并 a 和 b 的组，返回是否发生了实际合并。
func (r *refUF) union(a, b int) bool {
	ra, rb := r.find(a), r.find(b)
	if ra == rb {
		return false
	}
	// 将所有 group==rb 的元素改为 ra
	for k, v := range r.group {
		if v == rb {
			r.group[k] = ra
		}
	}
	return true
}

func (r *refUF) connected(a, b int) bool { return r.find(a) == r.find(b) }

func (r *refUF) count() int {
	roots := make(map[int]struct{})
	for _, g := range r.group {
		roots[g] = struct{}{}
	}
	return len(roots)
}

func (r *refUF) size(a int) int {
	root := r.find(a)
	cnt := 0
	for _, g := range r.group {
		if g == root {
			cnt++
		}
	}
	return cnt
}

// TestCorrectness_RandomOps 在 universe=[0,99] 上执行 100k 随机操作，
// 每步将 UnionFind 与参考模型对比。
func TestCorrectness_RandomOps(t *testing.T) {
	const (
		universe = 100
		ops      = 100_000
	)
	rng := rand.New(rand.NewPCG(42, 0))
	u := uf.New[int]()
	ref := newRefUF()

	for i := range ops {
		a, b := rng.IntN(universe), rng.IntN(universe)
		switch rng.IntN(5) {
		case 0, 1: // Union（较高频率）
			gotMerged := u.Union(a, b)
			wantMerged := ref.union(a, b)
			if gotMerged != wantMerged {
				t.Fatalf("op %d Union(%d,%d): got merged=%v, want %v", i, a, b, gotMerged, wantMerged)
			}
		case 2: // Connected
			got := u.Connected(a, b)
			want := ref.connected(a, b)
			if got != want {
				t.Fatalf("op %d Connected(%d,%d): got %v, want %v", i, a, b, got, want)
			}
		case 3: // Count
			got := u.Count()
			want := ref.count()
			if got != want {
				t.Fatalf("op %d Count: got %d, want %d", i, got, want)
			}
		case 4: // Size
			got := u.Size(a)
			want := ref.size(a)
			if got != want {
				t.Fatalf("op %d Size(%d): got %d, want %d", i, a, got, want)
			}
		}
	}
}

// TestCorrectness_ChainUnion 验证链式合并后全联通、Count=1、Size=N。
func TestCorrectness_ChainUnion(t *testing.T) {
	const N = 1000
	u := uf.New[int]()
	for i := range N - 1 {
		if !u.Union(i, i+1) {
			t.Fatalf("Union(%d,%d) 返回 false，期望 true", i, i+1)
		}
	}
	if got := u.Count(); got != 1 {
		t.Errorf("Count: got %d, want 1", got)
	}
	for i := range N {
		if got := u.Size(i); got != N {
			t.Errorf("Size(%d): got %d, want %d", i, got, N)
			break
		}
	}
	for i := range N {
		for j := i + 1; j < N; j += 100 { // 抽样验证传递连通
			if !u.Connected(i, j) {
				t.Errorf("Connected(%d,%d): got false, want true", i, j)
			}
		}
	}
}

// TestCorrectness_StarUnion 验证星形合并：所有节点合并到中心节点后全连通。
func TestCorrectness_StarUnion(t *testing.T) {
	const (
		N      = 500
		center = 0
	)
	u := uf.New[int]()
	for i := 1; i < N; i++ {
		u.Union(center, i)
	}
	if got := u.Count(); got != 1 {
		t.Errorf("Count: got %d, want 1", got)
	}
	if got := u.Size(center); got != N {
		t.Errorf("Size(center): got %d, want %d", got, N)
	}
	for i := 1; i < N; i++ {
		if !u.Connected(center, i) {
			t.Errorf("Connected(%d,%d): got false", center, i)
		}
	}
}

// TestCorrectness_LargeRandom 在不同规模的 universe 下执行 100k 操作，验证不变量。
func TestCorrectness_LargeRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模正确性测试")
	}
	universes := []int{10, 50, 200, 1000}
	for _, univ := range universes {
		univ := univ
		t.Run("univ="+itoa(univ), func(t *testing.T) {
			const ops = 100_000
			rng := rand.New(rand.NewPCG(uint64(univ), 0))
			u := uf.New[int]()
			ref := newRefUF()
			for i := range ops {
				a, b := rng.IntN(univ), rng.IntN(univ)
				u.Union(a, b)
				ref.union(a, b)
				if i%1000 == 999 {
					// 每 1000 步对账一次
					if got, want := u.Count(), ref.count(); got != want {
						t.Fatalf("op %d Count: got %d, want %d (univ=%d)", i, got, want, univ)
					}
					for k := range univ {
						if got, want := u.Size(k), ref.size(k); got != want {
							t.Fatalf("op %d Size(%d): got %d, want %d (univ=%d)", i, k, got, want, univ)
						}
					}
				}
			}
		})
	}
}

// itoa 将 int 转为字符串，避免引入 strconv 依赖。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
