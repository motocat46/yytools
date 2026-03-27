// 正确性命题测试 — LRUCache
// 验证核心不变量：LRU 淘汰顺序、TTL 惰性过期、容量不变量、Peek 不影响顺序、
// 随机混合操作与参考模型一致（ttl=0，10 万次）、并发安全（Len ≤ capacity）。
//
// 运行命令：
//
//	go test -race -run TestCorrectness -v ./pkg/ds/lru/
package lru_test

import (
	"container/list"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/motocat46/yytools/pkg/ds/lru"
)

// ── 参考模型 ─────────────────────────────────────────────────────────────────
//
// refLRU 用 container/list + map 实现最简正确的 LRU（ttl=0）。
// 作为随机混合测试的 ground truth，与被测对象（SUT）逐步对比。

type refLRU struct {
	cap   int
	items map[int]int
	order *list.List          // front = MRU，back = LRU
	elems map[int]*list.Element
}

func newRefLRU(cap int) *refLRU {
	return &refLRU{
		cap:   cap,
		items: make(map[int]int),
		order: list.New(),
		elems: make(map[int]*list.Element),
	}
}

func (r *refLRU) put(key, val int) {
	if e, ok := r.elems[key]; ok {
		r.order.MoveToFront(e)
		r.items[key] = val
		return
	}
	if len(r.items) >= r.cap {
		back := r.order.Back()
		backKey := back.Value.(int)
		r.order.Remove(back)
		delete(r.items, backKey)
		delete(r.elems, backKey)
	}
	r.items[key] = val
	e := r.order.PushFront(key)
	r.elems[key] = e
}

func (r *refLRU) get(key int) (int, bool) {
	v, ok := r.items[key]
	if !ok {
		return 0, false
	}
	r.order.MoveToFront(r.elems[key])
	return v, true
}

func (r *refLRU) delete(key int) bool {
	if _, ok := r.items[key]; !ok {
		return false
	}
	r.order.Remove(r.elems[key])
	delete(r.items, key)
	delete(r.elems, key)
	return true
}

// ─── 命题 1：LRU 淘汰顺序 ──────────────────────────────────────────────────
//
// 不变量：满容量时被淘汰的始终是最久未 Get/Put 的 key。
// 用确定序列验证：capacity=3，Put a/b/c，Get a，Put d → b 被淘汰。

func TestCorrectness_LRU_EvictionOrder(t *testing.T) {
	c := lru.New[string, int](3, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	// 顺序（MRU→LRU）：c → b → a

	c.Get("a") // a 变为 MRU；顺序变为：a → c → b

	c.Put("d", 4) // 容量满，淘汰 b（LRU）

	cases := []struct {
		key    string
		wantOK bool
	}{
		{"a", true},
		{"b", false}, // 被淘汰
		{"c", true},
		{"d", true},
	}
	for _, tc := range cases {
		_, ok := c.Get(tc.key)
		if ok != tc.wantOK {
			t.Errorf("淘汰顺序错误：Get(%q) ok=%v，期望 %v", tc.key, ok, tc.wantOK)
		}
	}
}

// ─── 命题 2：TTL 惰性过期 ──────────────────────────────────────────────────
//
// 不变量：过期后 Get/Peek/Contains 均返回未命中；惰性删除后容量槽被释放。

func TestCorrectness_TTL_LazyExpiry(t *testing.T) {
	const ttl = 10 * time.Millisecond
	c := lru.New[string, int](3, ttl)

	c.Put("a", 1)
	c.Put("b", 2)

	time.Sleep(5 * ttl) // 等待过期（5x TTL）

	// 所有访问路径均应返回未命中
	if _, ok := c.Get("a"); ok {
		t.Error("过期后 Get(\"a\") 期望 false")
	}
	if _, ok := c.Peek("b"); ok {
		t.Error("过期后 Peek(\"b\") 期望 false")
	}
	if c.Contains("a") {
		t.Error("过期后 Contains(\"a\") 期望 false")
	}

	// 惰性删除后，capacity 槽应被释放
	// 重新插入 3 个新元素（与 capacity 相等），不应触发 LRU 淘汰
	c.Put("x", 10)
	c.Put("y", 20)
	c.Put("z", 30)
	for _, k := range []string{"x", "y", "z"} {
		if _, ok := c.Get(k); !ok {
			t.Errorf("新插入 %q 应该存在，但 Get 返回 false（可能被错误淘汰）", k)
		}
	}
}

// ─── 命题 3：容量不变量 ──────────────────────────────────────────────────────
//
// 不变量：任意操作序列后 Len() ≤ capacity。
// 数据量：10,000 次随机 Put/Get/Delete。

func TestCorrectness_CapacityInvariant(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模容量不变量测试")
	}
	const (
		capacity = 50
		ops      = 100_000
	)
	c := lru.New[int, int](capacity, 0)
	rng := rand.New(rand.NewPCG(42, 0))

	for i := 0; i < ops; i++ {
		key := rng.IntN(100)
		switch rng.IntN(3) {
		case 0:
			c.Put(key, i)
		case 1:
			c.Get(key)
		case 2:
			c.Delete(key)
		}
		if got := c.Len(); got > capacity {
			t.Fatalf("第 %d 次操作后 Len=%d > capacity=%d（违反容量不变量）", i+1, got, capacity)
		}
	}
}

// ─── 命题 4：Peek 不影响淘汰顺序 ──────────────────────────────────────────
//
// 不变量：Peek 后淘汰顺序与 Peek 前完全一致（a 仍为 LRU，Put d 后 a 被淘汰）。

func TestCorrectness_Peek_NoOrderEffect(t *testing.T) {
	c := lru.New[string, int](3, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	// 顺序（MRU→LRU）：c → b → a

	c.Peek("a") // 不更新顺序，a 仍为 LRU

	c.Put("d", 4) // 应淘汰 a（LRU 未变）
	if _, ok := c.Get("a"); ok {
		t.Error("Peek 后 a 仍为 LRU，应该被淘汰，但 Get 返回 true")
	}
	if _, ok := c.Get("d"); !ok {
		t.Error("d 应该存在")
	}
}

// ─── 命题 5：随机混合操作 + 参考模型对比 ──────────────────────────────────
//
// 不变量：10 万次随机 Put/Get/Delete，每步与 refLRU 对比返回值。
// 使用 ttl=0 规避时序不确定性；keyRange > capacity 确保触发淘汰。

func TestCorrectness_RandomMixed_RefModel(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模随机混合测试")
	}
	const (
		ops      = 100_000
		capacity = 200
		keyRange = 300 // > capacity，确保触发淘汰
	)
	c := lru.New[int, int](capacity, 0)
	ref := newRefLRU(capacity)
	rng := rand.New(rand.NewPCG(99, 0))

	for i := 0; i < ops; i++ {
		key := rng.IntN(keyRange)
		switch rng.IntN(3) {
		case 0: // Put
			val := rng.IntN(10_000)
			c.Put(key, val)
			ref.put(key, val)
		case 1: // Get
			gotVal, gotOK := c.Get(key)
			wantVal, wantOK := ref.get(key)
			if gotOK != wantOK {
				t.Fatalf("第 %d 次操作 Get(%d)：SUT ok=%v，ref ok=%v", i+1, key, gotOK, wantOK)
			}
			if gotOK && gotVal != wantVal {
				t.Fatalf("第 %d 次操作 Get(%d)：SUT val=%d，ref val=%d", i+1, key, gotVal, wantVal)
			}
		case 2: // Delete
			gotOK := c.Delete(key)
			wantOK := ref.delete(key)
			if gotOK != wantOK {
				t.Fatalf("第 %d 次操作 Delete(%d)：SUT ok=%v，ref ok=%v", i+1, key, gotOK, wantOK)
			}
		}
		if got := c.Len(); got > capacity {
			t.Fatalf("第 %d 次操作后 Len=%d > capacity=%d", i+1, got, capacity)
		}
	}
}

// ─── 命题 6：并发安全 + Len ≤ capacity ──────────────────────────────────────
//
// 不变量：多 goroutine 并发 Put/Get/Delete，Len() ≤ capacity 始终成立，无数据竞争。
// 配合 -race 运行：go test -race -run TestCorrectness_Concurrent ./pkg/ds/lru/

func TestCorrectness_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发压力测试")
	}
	const (
		capacity   = 100
		goroutines = 20
		opsEach    = 5_000
		keyRange   = 200
	)
	c := lru.New[int, int](capacity, 0)
	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(uint64(id), 0))
			for i := 0; i < opsEach; i++ {
				key := rng.IntN(keyRange)
				switch rng.IntN(3) {
				case 0:
					c.Put(key, i)
				case 1:
					c.Get(key)
				case 2:
					c.Delete(key)
				}
				if got := c.Len(); got > capacity {
					t.Errorf("goroutine %d 第 %d 次操作后 Len=%d > capacity=%d", id, i+1, got, capacity)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}
