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

package lru_test

import (
	"fmt"
	"time"

	"github.com/motocat46/yytools/pkg/ds/lru"
)

// ExampleLRUCache 展示基本 Put/Get 操作及 LRU 淘汰行为。
func ExampleLRUCache() {
	c := lru.New[string, int](3, 0) // capacity=3，无 TTL

	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)

	v, ok := c.Get("a") // a 变为 MRU；顺序：a → c → b
	fmt.Printf("Get(\"a\") = %d, %v\n", v, ok)

	// 容量满，Put d 淘汰最久未使用的 b
	c.Put("d", 4)

	_, ok = c.Get("b")
	fmt.Printf("Get(\"b\") after eviction = %v\n", ok)

	// Output:
	// Get("a") = 1, true
	// Get("b") after eviction = false
}

// ExampleLRUCache_ttl 展示 TTL 惰性过期。
func ExampleLRUCache_ttl() {
	c := lru.New[string, string](10, 50*time.Millisecond)

	c.Put("session:123", "user-alice")

	v, ok := c.Get("session:123")
	fmt.Printf("before TTL: %q, %v\n", v, ok)

	time.Sleep(100 * time.Millisecond) // 等待过期

	v, ok = c.Get("session:123")
	fmt.Printf("after TTL:  %q, %v\n", v, ok)

	// Output:
	// before TTL: "user-alice", true
	// after TTL:  "", false
}

// ExampleLRUCache_Peek 展示 Peek 不影响 LRU 淘汰顺序。
func ExampleLRUCache_Peek() {
	c := lru.New[string, int](2, 0)
	c.Put("a", 1)
	c.Put("b", 2)

	// Peek a，不刷新顺序（a 仍为 LRU）
	v, ok := c.Peek("a")
	fmt.Printf("Peek(\"a\") = %d, %v\n", v, ok)

	// Put c 触发淘汰，a 被淘汰（而不是 b）
	c.Put("c", 3)
	_, ok = c.Get("a")
	fmt.Printf("Get(\"a\") after eviction = %v\n", ok)

	// Output:
	// Peek("a") = 1, true
	// Get("a") after eviction = false
}
