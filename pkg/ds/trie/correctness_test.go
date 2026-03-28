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

package trie_test

import (
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/motocat46/yytools/pkg/ds/trie"
)

// refTrie 是参考模型：用 map[string]bool 实现相同语义，正确性显而易见。
type refTrie struct {
	mu    sync.Mutex
	words map[string]bool
}

func newRefTrie() *refTrie { return &refTrie{words: make(map[string]bool)} }

func (r *refTrie) Insert(word string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.words[word] {
		return false
	}
	r.words[word] = true
	return true
}

func (r *refTrie) Search(word string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.words[word]
}

func (r *refTrie) HasPrefix(prefix string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if prefix == "" {
		return len(r.words) > 0
	}
	for k := range r.words {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func (r *refTrie) WithPrefix(prefix string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []string
	for k := range r.words {
		if strings.HasPrefix(k, prefix) {
			result = append(result, k)
		}
	}
	sort.Strings(result)
	return result
}

func (r *refTrie) Delete(word string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.words[word] {
		return false
	}
	delete(r.words, word)
	return true
}

func (r *refTrie) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.words)
}

// genWord 从小写字母生成长度 2-6 的随机词，控制词空间让前缀重叠更多。
func genWord(rng *rand.Rand) string {
	length := 2 + rng.IntN(5)
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = rune('a' + rng.IntN(10)) // 仅用 a-j，增加碰撞
	}
	return string(runes)
}

func strSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestCorrectness_RandomMixed_RefModel 随机混合操作，对比 SUT 与参考模型。
// 100,000 次操作，覆盖 Insert/Search/HasPrefix/WithPrefix/Delete 全部路径。
func TestCorrectness_RandomMixed_RefModel(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过随机混合正确性测试（-short）")
	}
	const ops = 100_000
	rng := rand.New(rand.NewPCG(42, 0))
	tr := trie.New()
	ref := newRefTrie()

	for i := range ops {
		word := genWord(rng)
		op := rng.IntN(5)
		switch op {
		case 0: // Insert
			got := tr.Insert(word)
			want := ref.Insert(word)
			if got != want {
				t.Fatalf("op %d Insert(%q): got %v, want %v", i, word, got, want)
			}
		case 1: // Search
			got := tr.Search(word)
			want := ref.Search(word)
			if got != want {
				t.Fatalf("op %d Search(%q): got %v, want %v", i, word, got, want)
			}
		case 2: // HasPrefix（取词的前 2 个 rune 作为前缀）
			runes := []rune(word)
			prefix := string(runes[:min(2, len(runes))])
			got := tr.HasPrefix(prefix)
			want := ref.HasPrefix(prefix)
			if got != want {
				t.Fatalf("op %d HasPrefix(%q): got %v, want %v", i, prefix, got, want)
			}
		case 3: // WithPrefix
			runes := []rune(word)
			prefix := string(runes[:min(2, len(runes))])
			got := tr.WithPrefix(prefix)
			sort.Strings(got)
			want := ref.WithPrefix(prefix)
			if !strSliceEq(got, want) {
				t.Fatalf("op %d WithPrefix(%q): got %v, want %v", i, prefix, got, want)
			}
		case 4: // Delete
			got := tr.Delete(word)
			want := ref.Delete(word)
			if got != want {
				t.Fatalf("op %d Delete(%q): got %v, want %v", i, word, got, want)
			}
		}
		// 不变量：Len 一致
		if got, want := tr.Len(), ref.Len(); got != want {
			t.Fatalf("op %d 后 Len(): got %d, want %d", i, got, want)
		}
	}
}

// TestCorrectness_Concurrent 多 goroutine 并发读写，验证无 data race 且最终 Len 一致。
func TestCorrectness_Concurrent(t *testing.T) {
	const (
		goroutines = 20
		perG       = 5_000
	)
	tr := trie.New()

	// 预填充 500 个词，在并发开始前精确记录实际词数（genWord 词空间有碰撞，实际词数 <= 500）
	rng0 := rand.New(rand.NewPCG(0, 0))
	for range 500 {
		tr.Insert(genWord(rng0))
	}
	initialLen := tr.Len()

	var wg sync.WaitGroup
	var insertCount, deleteCount atomic.Int64
	for g := range goroutines {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(seed, 0))
			for range perG {
				word := genWord(rng)
				switch rng.IntN(3) {
				case 0:
					if tr.Insert(word) {
						insertCount.Add(1)
					}
				case 1:
					tr.Search(word)
					runes := []rune(word)
					tr.HasPrefix(string(runes[:1]))
				case 2:
					if tr.Delete(word) {
						deleteCount.Add(1)
					}
				}
			}
		}(uint64(g))
	}
	wg.Wait()

	// 精确验证不变量：最终词数 == 初始词数 + 成功插入 - 成功删除
	want := int64(initialLen) + insertCount.Load() - deleteCount.Load()
	if got := int64(tr.Len()); got != want {
		t.Errorf("并发结束后 Len(): got %d, want %d (initial=%d +inserts=%d -deletes=%d)",
			got, want, initialLen, insertCount.Load(), deleteCount.Load())
	}
}
