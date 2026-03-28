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
// 创建日期:2026/3/28

// Package trie 实现并发安全的前缀树（Trie）。
// map[rune]*node 多叉树，支持 Unicode；sync.RWMutex 读写分离；
// Delete 递归回溯剪枝空节点，无内存泄漏。
package trie

import "sync"

// node 是 Trie 的内部节点。
type node struct {
	children map[rune]*node
	isEnd    bool // 标记从根到此节点的路径构成一个完整词
}

func newNode() *node {
	return &node{children: make(map[rune]*node)}
}

// Trie 是并发安全的前缀树。
// 零值不可用，必须通过 New() 创建。
type Trie struct {
	mu   sync.RWMutex
	root *node
	len  int // 当前存储的词数
}

// New 创建空 Trie。
func New() *Trie {
	return &Trie{root: newNode()}
}

// Len 返回 Trie 中的词数。
func (t *Trie) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.len
}

// findNode 从 n 出发沿 runes 路径走，返回终点节点；路径不存在返回 nil。
// 被 Search、HasPrefix、WithPrefix 共用。
func findNode(n *node, runes []rune) *node {
	for _, r := range runes {
		child := n.children[r]
		if child == nil {
			return nil
		}
		n = child
	}
	return n
}

// Search 精确匹配：词存在返回 true，仅是前缀返回 false。
func (t *Trie) Search(word string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	n := findNode(t.root, []rune(word))
	return n != nil && n.isEnd
}

// collect 从 n 出发 DFS，将所有完整词（isEnd=true 路径）追加到 results。
// path 是从根到 n 的已走路径（rune 切片），用于拼接完整词。
func collect(n *node, path []rune, results *[]string) {
	if n.isEnd {
		*results = append(*results, string(path))
	}
	for r, child := range n.children {
		collect(child, append(path, r), results)
	}
}

// WithPrefix 返回 Trie 中所有以 prefix 开头的词，顺序不保证。
// prefix 为空字符串时返回全部词。无匹配时返回 nil。
func (t *Trie) WithPrefix(prefix string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	n := findNode(t.root, []rune(prefix))
	if n == nil {
		return nil
	}
	var results []string
	collect(n, []rune(prefix), &results)
	return results
}

// HasPrefix 判断 Trie 中是否存在以 prefix 开头的词。
// prefix 为空字符串时，等价于 Len() > 0。
func (t *Trie) HasPrefix(prefix string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if prefix == "" {
		return t.len > 0
	}
	return findNode(t.root, []rune(prefix)) != nil
}

// deleteRec 递归删除 runes[depth:] 对应的词。
// 返回 (prunable, deleted)：
//   - prunable: 当前节点可被父节点从 children 中删除（isEnd=false 且无子节点）
//   - deleted: 词确实存在并被删除
func deleteRec(n *node, runes []rune, depth int) (prunable bool, deleted bool) {
	if depth == len(runes) {
		if !n.isEnd {
			return false, false // 词不存在
		}
		n.isEnd = false
		return len(n.children) == 0, true
	}
	r := runes[depth]
	child := n.children[r]
	if child == nil {
		return false, false // 路径不存在
	}
	prunable, deleted = deleteRec(child, runes, depth+1)
	if prunable {
		delete(n.children, r)
	}
	// 当前节点可剪枝：自身不是词终点且已无子节点
	return !n.isEnd && len(n.children) == 0, deleted
}

// Delete 删除指定词。词存在并成功删除返回 true，词不存在返回 false。
// Delete 会剪枝：若删除后某前缀节点不再有任何词，该节点同步删除。
func (t *Trie) Delete(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, deleted := deleteRec(t.root, []rune(word), 0)
	if deleted {
		t.len--
	}
	return deleted
}

// Insert 插入一个词。词已存在返回 false，新插入返回 true。
// 支持空字符串（空字符串是合法的词）。
func (t *Trie) Insert(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.root
	for _, r := range word {
		if cur.children[r] == nil {
			cur.children[r] = newNode()
		}
		cur = cur.children[r]
	}
	if cur.isEnd {
		return false
	}
	cur.isEnd = true
	t.len++
	return true
}
