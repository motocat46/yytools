package lru

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New[string, int](10, 0)
	if c == nil {
		t.Fatal("New() 返回了 nil")
	}
	if c.Len() != 0 {
		t.Errorf("新缓存的 Len 应该是 0，实际是 %d", c.Len())
	}
}

func TestNew_Panic_ZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("capacity=0 时应该 panic，但没有")
		}
	}()
	New[string, int](0, 0)
}

// ---- Put ----

func TestPut_Insert(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	if c.Len() != 1 {
		t.Errorf("插入后 Len 期望 1，实际 %d", c.Len())
	}
	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Errorf("Get(\"a\") 期望 (1, true)，实际 (%d, %v)", v, ok)
	}
}

func TestPut_Update_ExistingKey(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	c.Put("a", 99)
	if c.Len() != 1 {
		t.Errorf("重复 Put 后 Len 期望 1，实际 %d", c.Len())
	}
	v, ok := c.Get("a")
	if !ok || v != 99 {
		t.Errorf("Get(\"a\") 期望 (99, true)，实际 (%d, %v)", v, ok)
	}
}

func TestPut_Eviction_LRU(t *testing.T) {
	// capacity=3，Put a/b/c 后 Put d，应淘汰最久未使用的 a
	c := New[string, int](3, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	c.Put("d", 4) // 容量满，淘汰 a（LRU）
	if c.Len() != 3 {
		t.Errorf("淘汰后 Len 期望 3，实际 %d", c.Len())
	}
	if _, ok := c.Get("a"); ok {
		t.Error("a 应该被淘汰，但 Get 仍然返回 true")
	}
	if _, ok := c.Get("d"); !ok {
		t.Error("d 应该存在")
	}
}

func TestPut_Eviction_AfterGet(t *testing.T) {
	// capacity=3，Put a/b/c，Get a（a 变为 MRU），Put d，应淘汰 b（LRU）
	c := New[string, int](3, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	c.Get("a") // a 变为 MRU；顺序变为 a → c → b（b 为 LRU）
	c.Put("d", 4) // 淘汰 b
	if _, ok := c.Get("a"); !ok {
		t.Error("a 应该存在（已被 Get 刷新为 MRU）")
	}
	if _, ok := c.Get("b"); ok {
		t.Error("b 应该被淘汰（LRU），但 Get 仍然返回 true")
	}
}

func TestPut_Capacity1(t *testing.T) {
	c := New[string, int](1, 0)
	c.Put("a", 1)
	c.Put("b", 2) // 淘汰 a
	if c.Len() != 1 {
		t.Errorf("capacity=1 时 Len 期望 1，实际 %d", c.Len())
	}
	if _, ok := c.Get("a"); ok {
		t.Error("a 应该被淘汰")
	}
	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Errorf("Get(\"b\") 期望 (2, true)，实际 (%d, %v)", v, ok)
	}
}

// ---- Get ----

func TestGet_Hit(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("x", 42)
	v, ok := c.Get("x")
	if !ok || v != 42 {
		t.Errorf("Get(\"x\") 期望 (42, true)，实际 (%d, %v)", v, ok)
	}
}

func TestGet_Miss(t *testing.T) {
	c := New[string, int](5, 0)
	v, ok := c.Get("missing")
	if ok || v != 0 {
		t.Errorf("Get(\"missing\") 期望 (0, false)，实际 (%d, %v)", v, ok)
	}
}

func TestGet_Expired(t *testing.T) {
	c := New[string, int](5, 1*time.Millisecond)
	c.Put("x", 42)
	time.Sleep(10 * time.Millisecond) // 等待过期（10x TTL）
	v, ok := c.Get("x")
	if ok || v != 0 {
		t.Errorf("过期后 Get 期望 (0, false)，实际 (%d, %v)", v, ok)
	}
	if c.Len() != 0 {
		t.Errorf("惰性删除后 Len 期望 0，实际 %d", c.Len())
	}
}

func TestGet_UpdatesOrder(t *testing.T) {
	// Get 后顺序更新：被 Get 的 key 变为 MRU，不应被淘汰
	c := New[string, int](2, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a") // a 变为 MRU；b 变为 LRU
	c.Put("c", 3) // 淘汰 b
	if _, ok := c.Get("a"); !ok {
		t.Error("a 应该存在（Get 刷新了顺序）")
	}
	if _, ok := c.Get("b"); ok {
		t.Error("b 应该被淘汰（LRU）")
	}
}

// ---- Peek ----

func TestPeek_Hit(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("k", 7)
	v, ok := c.Peek("k")
	if !ok || v != 7 {
		t.Errorf("Peek(\"k\") 期望 (7, true)，实际 (%d, %v)", v, ok)
	}
}

func TestPeek_Miss(t *testing.T) {
	c := New[string, int](5, 0)
	v, ok := c.Peek("missing")
	if ok || v != 0 {
		t.Errorf("Peek(\"missing\") 期望 (0, false)，实际 (%d, %v)", v, ok)
	}
}

func TestPeek_NoOrderUpdate(t *testing.T) {
	// Peek 不更新顺序，a 仍为 LRU，Put c 后 a 被淘汰
	c := New[string, int](2, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Peek("a") // 不更新顺序，a 仍为 LRU
	c.Put("c", 3) // 淘汰 a
	if _, ok := c.Get("a"); ok {
		t.Error("a 应该被淘汰（Peek 不更新顺序）")
	}
	if _, ok := c.Get("b"); !ok {
		t.Error("b 应该存在")
	}
}

func TestPeek_Expired(t *testing.T) {
	c := New[string, int](5, 1*time.Millisecond)
	c.Put("x", 42)
	time.Sleep(10 * time.Millisecond)
	v, ok := c.Peek("x")
	if ok || v != 0 {
		t.Errorf("过期后 Peek 期望 (0, false)，实际 (%d, %v)", v, ok)
	}
	if c.Len() != 0 {
		t.Errorf("惰性删除后 Len 期望 0，实际 %d", c.Len())
	}
}

// ---- Contains ----

func TestContains_Exists(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	if !c.Contains("a") {
		t.Error("Contains(\"a\") 期望 true")
	}
	if c.Contains("missing") {
		t.Error("Contains(\"missing\") 期望 false")
	}
}

func TestContains_Expired(t *testing.T) {
	c := New[string, int](5, 1*time.Millisecond)
	c.Put("x", 1)
	time.Sleep(10 * time.Millisecond)
	if c.Contains("x") {
		t.Error("过期后 Contains 期望 false")
	}
}

// ---- Delete ----

func TestDelete_Exists(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	if !c.Delete("a") {
		t.Error("Delete(\"a\") 期望返回 true")
	}
	if c.Len() != 0 {
		t.Errorf("删除后 Len 期望 0，实际 %d", c.Len())
	}
	if _, ok := c.Get("a"); ok {
		t.Error("删除后 Get 仍返回 true")
	}
}

func TestDelete_NotExists(t *testing.T) {
	c := New[string, int](5, 0)
	if c.Delete("missing") {
		t.Error("Delete(\"missing\") 期望返回 false")
	}
}

func TestDelete_ThenReinsert(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	c.Delete("a")
	c.Put("a", 2)
	v, ok := c.Get("a")
	if !ok || v != 2 {
		t.Errorf("删除后重插 Get 期望 (2, true)，实际 (%d, %v)", v, ok)
	}
}

// ---- Purge ----

func TestPurge(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Purge()
	if c.Len() != 0 {
		t.Errorf("Purge 后 Len 期望 0，实际 %d", c.Len())
	}
	if _, ok := c.Get("a"); ok {
		t.Error("Purge 后 Get(\"a\") 期望 false")
	}
	// Purge 后再次 Put 应正常
	c.Put("c", 3)
	if v, ok := c.Get("c"); !ok || v != 3 {
		t.Errorf("Purge 后重新插入 Get(\"c\") 期望 (3, true)，实际 (%d, %v)", v, ok)
	}
}

// ---- 边界 ----

func TestEmptyCache_AllOps(t *testing.T) {
	c := New[string, int](5, 0)
	if c.Len() != 0 {
		t.Errorf("空缓存 Len 期望 0，实际 %d", c.Len())
	}
	if _, ok := c.Get("x"); ok {
		t.Error("空缓存 Get 期望 false")
	}
	if _, ok := c.Peek("x"); ok {
		t.Error("空缓存 Peek 期望 false")
	}
	if c.Contains("x") {
		t.Error("空缓存 Contains 期望 false")
	}
	if c.Delete("x") {
		t.Error("空缓存 Delete 期望 false")
	}
	c.Purge() // 不应 panic
}

func TestTTL_Zero_NeverExpires(t *testing.T) {
	c := New[string, int](5, 0)
	c.Put("a", 1)
	time.Sleep(5 * time.Millisecond)
	if _, ok := c.Get("a"); !ok {
		t.Error("ttl=0 时不应该过期")
	}
}

// ---- 基准测试 ----

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// BenchmarkPut 基准：Put（每次迭代触发一次淘汰，规模稳定在 n）
func BenchmarkPut(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			c := New[int, int](n, 0)
			for i := 0; i < n; i++ {
				c.Put(i, i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// key 在 [n, 2n) 循环，始终触发淘汰，map 底层规模稳定在 n
				c.Put(n+i%n, i)
			}
		})
	}
}

// BenchmarkGet 基准：Get 命中（循环访问已有 key）
func BenchmarkGet(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			c := New[int, int](n, 0)
			for i := 0; i < n; i++ {
				c.Put(i, i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				c.Get(i % n)
			}
		})
	}
}

// BenchmarkMixed 混合负载基准：70% Get + 20% Put + 10% Delete
// 模拟典型读多写少的缓存场景（如 Redis 前置缓冲层热点数据访问）
func BenchmarkMixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			c := New[int, int](n, 0)
			for i := 0; i < n; i++ {
				c.Put(i, i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				key := i % n
				switch i % 10 {
				case 0, 1: // 20% Put
					c.Put(key, i)
				case 2: // 10% Delete
					c.Delete(key)
				default: // 70% Get
					c.Get(key)
				}
			}
		})
	}
}

// BenchmarkConcurrent_PeekHeavy 读密集并发：90% Peek（读锁）+ 10% Put（写锁）
// 量化读锁在高并发读场景下的价值：多 Peek 不互斥，p=64 吞吐应接近 p=1
func BenchmarkConcurrent_PeekHeavy(b *testing.B) {
	const n = 10_000
	for _, parallelism := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", parallelism), func(b *testing.B) {
			c := New[int, int](n, 0)
			for i := 0; i < n; i++ {
				c.Put(i, i)
			}
			b.SetParallelism(parallelism)
			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					if i%10 == 0 {
						c.Put(i%n, i) // 10% Put（写锁）
					} else {
						c.Peek(i % n) // 90% Peek（读锁，多调用方不互斥）
					}
					i++
				}
			})
		})
	}
}

// BenchmarkConcurrent_Mixed 并发混合负载：多调用方竞争，暴露锁竞争代价
// 对比 p=1 与 p=64 的 ns/op 差距，量化同步开销
func BenchmarkConcurrent_Mixed(b *testing.B) {
	const n = 10_000
	for _, parallelism := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", parallelism), func(b *testing.B) {
			c := New[int, int](n, 0)
			for i := 0; i < n; i++ {
				c.Put(i, i)
			}
			b.SetParallelism(parallelism)
			b.ResetTimer()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := i % n
					switch i % 10 {
					case 0, 1:
						c.Put(key, i)
					case 2:
						c.Delete(key)
					default:
						c.Get(key)
					}
					i++
				}
			})
		})
	}
}
