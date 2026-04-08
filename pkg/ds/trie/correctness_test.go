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
	"fmt"
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

// TestCorrectness_Trie_ConcurrentReadWriteSearch 命题3：并发读写时 Search 值正确性。
// 先插入 1000 个已知词（fixedWords），之后 writer goroutine 仅操作独立的随机词池，
// reader goroutine 持续 Search 已知词，验证命中时返回 true（已知词不被删除）。
func TestCorrectness_Trie_ConcurrentReadWriteSearch(t *testing.T) {
	const fixedCount = 1000
	const poolSize = 100  // writer 操作的词池，与 fixedWords 不重叠
	const goroutines = 10 // 10 writers + 10 readers
	const opsPerGoroutine = 5000

	tr := trie.New()

	// 顺序插入 fixedWords
	fixedWords := make([]string, fixedCount)
	for i := range fixedCount {
		fixedWords[i] = fmt.Sprintf("fixed%d", i)
		tr.Insert(fixedWords[i])
	}

	// writer 操作的词池：以 "rw" 开头，与 "fixed" 开头的词不重叠
	pool := make([]string, poolSize)
	for i := range poolSize {
		pool[i] = fmt.Sprintf("rw%d", i)
	}

	var wg sync.WaitGroup
	errCount := atomic.Int64{}

	// 10 个 writer：只操作 pool，不碰 fixedWords
	for g := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewPCG(uint64(id), 0))
			for range opsPerGoroutine {
				w := pool[r.IntN(poolSize)]
				if r.IntN(2) == 0 {
					tr.Insert(w)
				} else {
					tr.Delete(w)
				}
			}
		}(g)
	}

	// 10 个 reader：只搜索 fixedWords，这些词不会被 writer 删除
	for g := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewPCG(uint64(id+goroutines), 0))
			for range opsPerGoroutine {
				w := fixedWords[r.IntN(fixedCount)]
				if !tr.Search(w) {
					// fixedWords 没有被删除，Search 必须返回 true
					errCount.Add(1)
				}
			}
		}(g)
	}

	wg.Wait()

	if errCount.Load() > 0 {
		t.Errorf("命题3失败：%d 次 Search(fixedWord) 返回 false（fixedWord 不应被删除）", errCount.Load())
	}
}

// TestCorrectness_Trie_ConcurrentPanic 命题4：并发 Insert/Delete/Search 不触发 panic。
// goroutine 随机执行三种操作，验证在任意交叉下不产生 panic 或数据竞争。
func TestCorrectness_Trie_ConcurrentPanic(t *testing.T) {
	const rounds = 50_000
	const goroutines = 6

	tr := trie.New()

	pool := make([]string, 200)
	for i := range pool {
		pool[i] = fmt.Sprintf("w%d", i)
	}

	var wg sync.WaitGroup
	panicked := atomic.Bool{}

	for g := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					panicked.Store(true)
				}
			}()
			defer wg.Done()
			r := rand.New(rand.NewPCG(uint64(id), 0))
			for range rounds {
				w := pool[r.IntN(len(pool))]
				switch r.IntN(3) {
				case 0:
					tr.Insert(w)
				case 1:
					tr.Delete(w)
				case 2:
					tr.Search(w)
				}
			}
		}(g)
	}
	wg.Wait()

	if panicked.Load() {
		t.Error("命题4失败：并发 Insert/Delete/Search 触发了 panic")
	}
}
