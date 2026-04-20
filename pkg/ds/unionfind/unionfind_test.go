package unionfind_test

import (
	"testing"

	uf "github.com/motocat46/yytools/pkg/ds/unionfind"
)

func TestNew_Empty(t *testing.T) {
	u := uf.New[int]()
	if u.Count() != 0 {
		t.Errorf("got Count=%d, want 0", u.Count())
	}
}

func TestAutoRegister_NewElement(t *testing.T) {
	u := uf.New[string]()
	root := u.Find("alice")
	if root != "alice" {
		t.Errorf("got Find=%q, want %q", root, "alice")
	}
	if u.Count() != 1 {
		t.Errorf("got Count=%d, want 1", u.Count())
	}
	if u.Size("alice") != 1 {
		t.Errorf("got Size=%d, want 1", u.Size("alice"))
	}
}

func TestFind_ReturnsSelf_WhenAlone(t *testing.T) {
	u := uf.New[int]()
	if u.Find(42) != 42 {
		t.Errorf("单独元素 Find 应返回自身")
	}
}

func TestFind_SameRoot_AfterUnion(t *testing.T) {
	u := uf.New[int]()
	u.Union(1, 2)
	u.Union(2, 3)
	r1, r2, r3 := u.Find(1), u.Find(2), u.Find(3)
	if r1 != r2 || r2 != r3 {
		t.Errorf("同组元素 Find 应返回相同根：got %d %d %d", r1, r2, r3)
	}
}

func TestFind_PathCompression(t *testing.T) {
	u := uf.New[int]()
	u.Union(1, 2)
	u.Union(2, 3)
	u.Union(3, 4)
	u.Union(4, 5)
	root := u.Find(1)
	if u.Find(1) != root {
		t.Errorf("路径压缩后 Find 结果不稳定")
	}
}

func TestUnion_ReturnsTrueOnFirstMerge(t *testing.T) {
	u := uf.New[string]()
	if !u.Union("a", "b") {
		t.Error("首次合并应返回 true")
	}
}

func TestUnion_ReturnsFalseWhenAlreadyConnected(t *testing.T) {
	u := uf.New[string]()
	u.Union("a", "b")
	if u.Union("a", "b") {
		t.Error("已连通再次 Union 应返回 false")
	}
}

func TestUnion_ReducesCount(t *testing.T) {
	u := uf.New[int]()
	u.Find(1)
	u.Find(2)
	u.Find(3)
	if u.Count() != 3 {
		t.Fatalf("got Count=%d, want 3", u.Count())
	}
	u.Union(1, 2)
	if u.Count() != 2 {
		t.Errorf("Union 后 got Count=%d, want 2", u.Count())
	}
	u.Union(2, 3)
	if u.Count() != 1 {
		t.Errorf("全连通后 got Count=%d, want 1", u.Count())
	}
}

func TestUnion_SelfUnion(t *testing.T) {
	u := uf.New[int]()
	if u.Union(1, 1) {
		t.Error("自身 Union 应返回 false")
	}
	if u.Count() != 1 {
		t.Errorf("自身 Union 后 Count 应为 1，got %d", u.Count())
	}
}

func TestConnected_FalseForNewElements(t *testing.T) {
	u := uf.New[int]()
	if u.Connected(1, 2) {
		t.Error("未合并的两个元素不应连通")
	}
}

func TestConnected_TrueAfterUnion(t *testing.T) {
	u := uf.New[int]()
	u.Union(1, 2)
	if !u.Connected(1, 2) {
		t.Error("Union 后应连通")
	}
}

func TestConnected_Transitive(t *testing.T) {
	u := uf.New[int]()
	u.Union(1, 2)
	u.Union(2, 3)
	if !u.Connected(1, 3) {
		t.Error("传递连通：1-2、2-3 Union 后 1-3 应连通")
	}
}

func TestSize_SingleElement(t *testing.T) {
	u := uf.New[string]()
	if u.Size("x") != 1 {
		t.Errorf("单独元素 Size 应为 1，got %d", u.Size("x"))
	}
}

func TestSize_AfterUnion(t *testing.T) {
	u := uf.New[int]()
	u.Union(1, 2)
	u.Union(1, 3)
	if u.Size(1) != 3 {
		t.Errorf("got Size(1)=%d, want 3", u.Size(1))
	}
	if u.Size(2) != 3 {
		t.Errorf("got Size(2)=%d, want 3", u.Size(2))
	}
}
