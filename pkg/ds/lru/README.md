# lru

带 TTL 的泛型 LRU 缓存。双向链表 + 哈希表实现，O(1) Put/Get/Delete，`sync.RWMutex` 并发安全。

> **学习目的实现。** 生产环境建议使用 [hashicorp/golang-lru v2](https://github.com/hashicorp/golang-lru)，
> 支持 per-key TTL、onEvict 回调、2Q Cache 等特性。

## 快速上手

```go
import (
    "time"
    "github.com/motocat46/yytools/pkg/ds/lru"
)

// 创建容量 1000、TTL 5 分钟的缓存（作为 Redis 前置缓冲层）
c := lru.New[string, []byte](1000, 5*time.Minute)

// 写入
c.Put("user:123", data)

// 读取（刷新为最近使用）
if val, ok := c.Get("user:123"); ok {
    // 命中，访问延迟 ~50ns（vs Redis ~1ms）
}

// 监控/调试读取（不影响淘汰顺序）
if val, ok := c.Peek("user:123"); ok {
    // 命中，顺序不变
}
```

## API

```go
// New 创建 LRU 缓存。capacity 必须 > 0；ttl=0 表示永不过期。
func New[K comparable, V any](capacity int, ttl time.Duration) *LRUCache[K, V]

// Put 写入或更新 key-value，刷新为最近使用。容量满时淘汰最久未使用 key。
func (c *LRUCache[K, V]) Put(key K, val V)

// Get 读取并刷新为最近使用。key 不存在或已过期返回 (零值, false)。
func (c *LRUCache[K, V]) Get(key K) (V, bool)

// Peek 读取但不更新访问顺序。适合监控/调试，不影响 LRU 淘汰判断。
func (c *LRUCache[K, V]) Peek(key K) (V, bool)

// Delete 删除指定 key，返回是否存在。
func (c *LRUCache[K, V]) Delete(key K) bool

// Contains 判断 key 是否存在且未过期，不更新访问顺序。
func (c *LRUCache[K, V]) Contains(key K) bool

// Len 返回当前缓存元素数（含已过期但未被访问的惰性节点）。
func (c *LRUCache[K, V]) Len() int

// Purge 清空缓存。
func (c *LRUCache[K, V]) Purge()
```

## 注意事项

- **`Len()` 含过期节点**：惰性过期策略下，已过期但未被访问的节点仍计入 `Len()`，直到下次访问时才删除。
- **TTL 为全局统一**：所有 key 共享同一 TTL，`Put` 时设置（或刷新），不支持 per-key TTL。
- **`Get` vs `Peek`**：`Get` 使用写锁并更新访问顺序；`Peek` 正常路径只需读锁，不影响 LRU 淘汰顺序，适合高频监控场景。
- **Put 会复活已过期 key**：若 key 已过期但尚未被访问（惰性删除未触发），再次 `Put` 同一 key 会刷新其值和过期时间（复活），不触发 LRU 淘汰，不消耗额外 capacity 名额。
- **与 hashicorp/golang-lru v2 的差异**：本实现不支持 onEvict 回调、per-key TTL、2Q Cache；构造函数使用 `assert.Assert` 而非返回 error。
