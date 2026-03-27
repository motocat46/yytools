// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/3/27

// Package lru 实现带 TTL 的泛型 LRU 缓存。
// 双向链表 + 哈希表，O(1) Put/Get/Delete，惰性过期，sync.RWMutex 并发安全。
// 学习目的实现；生产环境建议使用 hashicorp/golang-lru v2。
package lru

import (
	"sync"
	"time"

	"github.com/motocat46/yytools/pkg/common/assert"
)

// entry 是双向链表的节点，同时是 hashmap 的值。
type entry[K comparable, V any] struct {
	key      K
	val      V
	expireAt time.Time // 零值表示永不过期
	prev     *entry[K, V]
	next     *entry[K, V]
}

// isExpired 返回节点是否已过期。零值（ttl=0）永不过期。
func (e *entry[K, V]) isExpired() bool {
	return !e.expireAt.IsZero() && time.Now().After(e.expireAt)
}

// LRUCache 是带 TTL 的泛型 LRU 缓存。
// 并发安全，使用 sync.RWMutex 保护内部状态。
// 链表布局：head ↔ [MRU] ↔ ... ↔ [LRU] ↔ tail
// ttl == 0 表示永不过期，只靠 LRU 淘汰。
// 惰性过期：过期节点在被访问时才删除，不运行后台 goroutine。
type LRUCache[K comparable, V any] struct {
	mu       sync.RWMutex
	items    map[K]*entry[K, V]
	head     *entry[K, V] // 哨兵头节点，head.next 是最近使用的节点（MRU）
	tail     *entry[K, V] // 哨兵尾节点，tail.prev 是最久未使用的节点（LRU）
	capacity int
	ttl      time.Duration // 0 表示永不过期
}

// New 创建 LRU 缓存。
// capacity 必须 > 0，否则 assert panic。
// ttl == 0 表示永不过期，只靠 LRU 淘汰。
func New[K comparable, V any](capacity int, ttl time.Duration) *LRUCache[K, V] {
	assert.Assert(capacity > 0, "LRUCache: capacity 必须大于 0")
	head := &entry[K, V]{}
	tail := &entry[K, V]{}
	head.next = tail
	tail.prev = head
	return &LRUCache[K, V]{
		items:    make(map[K]*entry[K, V], capacity),
		head:     head,
		tail:     tail,
		capacity: capacity,
		ttl:      ttl,
	}
}

// Len 返回当前缓存的元素数量。
// 注意：惰性过期策略下，已过期但未被访问的节点仍计入 Len。
func (c *LRUCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// detach 将节点从链表中摘除（不修改 map）。调用方持有写锁。
func (c *LRUCache[K, V]) detach(e *entry[K, V]) {
	e.prev.next = e.next
	e.next.prev = e.prev
}

// attachFront 将节点插入链表头（head.next，即 MRU 位置）。调用方持有写锁。
func (c *LRUCache[K, V]) attachFront(e *entry[K, V]) {
	e.next = c.head.next
	e.prev = c.head
	c.head.next.prev = e
	c.head.next = e
}

// removeEntry 从链表和 map 中同时删除节点。调用方持有写锁。
func (c *LRUCache[K, V]) removeEntry(e *entry[K, V]) {
	c.detach(e)
	delete(c.items, e.key)
}

// Put 存入 key-value。
// 若 key 已存在则更新值并移到链表头（刷新为最近使用）。
// 若容量已满则淘汰链表尾节点（最久未使用，tail.prev），再插入。
func (c *LRUCache[K, V]) Put(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[key]; ok {
		// 已存在：更新值，刷新 TTL，移到链表头
		e.val = val
		if c.ttl > 0 {
			e.expireAt = time.Now().Add(c.ttl)
		}
		c.detach(e)
		c.attachFront(e)
		return
	}

	// 容量满：淘汰最久未使用节点（tail.prev）
	if len(c.items) >= c.capacity {
		c.removeEntry(c.tail.prev)
	}

	// 创建新节点，插入链表头
	e := &entry[K, V]{key: key, val: val}
	if c.ttl > 0 {
		e.expireAt = time.Now().Add(c.ttl)
	}
	c.attachFront(e)
	c.items[key] = e
}

// Delete 删除指定 key，返回是否存在。
func (c *LRUCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		return false
	}
	c.removeEntry(e)
	return true
}

// Purge 清空缓存。
func (c *LRUCache[K, V]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*entry[K, V], c.capacity)
	c.head.next = c.tail
	c.tail.prev = c.head
}

// Peek 获取值但不更新访问顺序。
// 适合监控/调试场景，不影响 LRU 淘汰判断。
// key 不存在或已过期返回 (零值, false)；过期时惰性删除节点（double-check 模式）。
func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	if !ok {
		c.mu.RUnlock()
		var zero V
		return zero, false
	}
	if !e.isExpired() {
		val := e.val
		c.mu.RUnlock()
		return val, true
	}
	// 发现过期：释放读锁，升级为写锁，double-check 后删除
	// double-check 必要：另一个 goroutine 可能在读锁释放和写锁加锁之间已删除该节点
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if e2, still := c.items[key]; still && e2.isExpired() {
		c.removeEntry(e2)
	}
	var zero V
	return zero, false
}

// Contains 判断 key 是否存在且未过期，不更新访问顺序。
func (c *LRUCache[K, V]) Contains(key K) bool {
	_, ok := c.Peek(key)
	return ok
}

// Get 获取值并将该 key 移到链表头（标记为最近使用）。
// key 不存在或已过期返回 (零值, false)；过期时惰性删除节点。
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	if e.isExpired() {
		c.removeEntry(e)
		var zero V
		return zero, false
	}
	c.detach(e)
	c.attachFront(e)
	return e.val, true
}
