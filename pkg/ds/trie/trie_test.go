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

package trie

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"sync/atomic"
	"testing"
)

func TestInsert_NewWord(t *testing.T) {
	tr := New()
	if tr.Len() != 0 {
		t.Errorf("空 Trie Len(): got %d, want 0", tr.Len())
	}
	if ok := tr.Insert("hello"); !ok {
		t.Error("Insert 新词: got false, want true")
	}
	if tr.Len() != 1 {
		t.Errorf("Insert 后 Len(): got %d, want 1", tr.Len())
	}
}

func TestInsert_Duplicate(t *testing.T) {
	tr := New()
	tr.Insert("hello")
	if ok := tr.Insert("hello"); ok {
		t.Error("Insert 重复词: got true, want false")
	}
	if tr.Len() != 1 {
		t.Errorf("重复 Insert 后 Len(): got %d, want 1", tr.Len())
	}
}

// slicesEqual 比较两个字符串切片是否相等（nil 和空切片视为相等）。
func slicesEqual(a, b []string) bool {
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

func TestSearch(t *testing.T) {
	tr := New()
	tr.Insert("apple")
	tr.Insert("app")
	tr.Insert("应用") // Unicode

	cases := []struct {
		word string
		want bool
	}{
		{"apple", true},
		{"app", true},
		{"应用", true},
		{"ap", false},     // 只是前缀
		{"apples", false}, // 未插入
		{"", false},       // 空字符串未插入
		{"banana", false},
	}
	for _, tc := range cases {
		t.Run(tc.word, func(t *testing.T) {
			got := tr.Search(tc.word)
			if got != tc.want {
				t.Errorf("Search(%q): got %v, want %v", tc.word, got, tc.want)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tr := New()
	tr.Insert("apple")
	tr.Insert("application")
	tr.Insert("应用程序")

	cases := []struct {
		prefix string
		want   bool
	}{
		{"app", true},
		{"apple", true},        // 精确匹配也是合法前缀
		{"apples", false},      // 超出任意词的长度
		{"应", true},
		{"应用", true},
		{"应用程序", true},
		{"应用程序x", false},
		{"b", false},
	}
	for _, tc := range cases {
		t.Run(tc.prefix, func(t *testing.T) {
			got := tr.HasPrefix(tc.prefix)
			if got != tc.want {
				t.Errorf("HasPrefix(%q): got %v, want %v", tc.prefix, got, tc.want)
			}
		})
	}
}

func TestHasPrefix_EmptyPrefix(t *testing.T) {
	emptyTrie := New()
	if emptyTrie.HasPrefix("") {
		t.Error(`空 Trie HasPrefix(""): got true, want false`)
	}
	tr := New()
	tr.Insert("hello")
	if !tr.HasPrefix("") {
		t.Error(`非空 Trie HasPrefix(""): got false, want true`)
	}
}

func TestWithPrefix(t *testing.T) {
	tr := New()
	for _, w := range []string{"apple", "application", "apt", "app", "banana"} {
		tr.Insert(w)
	}

	cases := []struct {
		prefix string
		want   []string // 排序后
	}{
		{"app", []string{"app", "apple", "application"}},
		{"appl", []string{"apple", "application"}},
		{"apple", []string{"apple"}},
		{"b", []string{"banana"}},
		{"c", nil},
	}
	for _, tc := range cases {
		t.Run(tc.prefix, func(t *testing.T) {
			got := tr.WithPrefix(tc.prefix)
			sort.Strings(got)
			if !slicesEqual(got, tc.want) {
				t.Errorf("WithPrefix(%q): got %v, want %v", tc.prefix, got, tc.want)
			}
		})
	}
}

func TestWithPrefix_EmptyPrefix(t *testing.T) {
	tr := New()
	words := []string{"apple", "banana", "cherry"}
	for _, w := range words {
		tr.Insert(w)
	}
	got := tr.WithPrefix("")
	sort.Strings(got)
	want := []string{"apple", "banana", "cherry"}
	if !slicesEqual(got, want) {
		t.Errorf(`WithPrefix(""): got %v, want %v`, got, want)
	}
}

func TestWithPrefix_EmptyTrie(t *testing.T) {
	tr := New()
	got := tr.WithPrefix("app")
	if len(got) != 0 {
		t.Errorf(`空 Trie WithPrefix("app"): got %v, want []`, got)
	}
}

func TestDelete_Existing(t *testing.T) {
	tr := New()
	tr.Insert("apple")
	tr.Insert("app")

	if ok := tr.Delete("apple"); !ok {
		t.Error("Delete 存在的词: got false, want true")
	}
	if tr.Len() != 1 {
		t.Errorf("Delete 后 Len(): got %d, want 1", tr.Len())
	}
	if tr.Search("apple") {
		t.Error("Delete 后 Search(apple): got true, want false")
	}
	// app 仍应存在，Delete apple 时不能把共享前缀节点 app 删掉
	if !tr.Search("app") {
		t.Error("Delete apple 后 Search(app): got false, want true")
	}
}

func TestDelete_NonExistent(t *testing.T) {
	tr := New()
	tr.Insert("hello")
	if ok := tr.Delete("world"); ok {
		t.Error("Delete 不存在的词: got true, want false")
	}
	if tr.Len() != 1 {
		t.Errorf("Delete 不存在词后 Len(): got %d, want 1", tr.Len())
	}
}

func TestDelete_Prune(t *testing.T) {
	// 验证 Delete 后，HasPrefix 对已删除词的前缀返回正确结果
	tr := New()
	tr.Insert("abc")
	tr.Delete("abc")
	// abc 是 Trie 中唯一的词，删后 HasPrefix("ab") 应为 false
	if tr.HasPrefix("ab") {
		t.Error("Delete abc 后 HasPrefix(ab): got true, want false（剪枝未生效）")
	}
	if tr.Len() != 0 {
		t.Errorf("Delete 全部词后 Len(): got %d, want 0", tr.Len())
	}
}

func TestDelete_PrefixWordRemains(t *testing.T) {
	// 删除较长词时，作为其前缀的较短词应保留
	tr := New()
	tr.Insert("app")
	tr.Insert("apple")
	tr.Delete("apple")
	if !tr.Search("app") {
		t.Error("Delete apple 后 Search(app): got false, want true")
	}
	if tr.HasPrefix("apple") {
		t.Error("Delete apple 后 HasPrefix(apple): got true, want false")
	}
}

func TestBoundary_EmptyTrie(t *testing.T) {
	tr := New()
	if tr.Search("") {
		t.Error(`空 Trie Search(""): got true, want false`)
	}
	if tr.HasPrefix("") {
		t.Error(`空 Trie HasPrefix(""): got true, want false`)
	}
	if got := tr.WithPrefix(""); len(got) != 0 {
		t.Errorf(`空 Trie WithPrefix(""): got %v, want []`, got)
	}
	if tr.Delete("x") {
		t.Error("空 Trie Delete(x): got true, want false")
	}
}

func TestBoundary_SingleElement(t *testing.T) {
	tr := New()
	tr.Insert("x")
	if !tr.Search("x") {
		t.Error("Search(x): got false, want true")
	}
	if !tr.HasPrefix("x") {
		t.Error("HasPrefix(x): got false, want true")
	}
	tr.Delete("x")
	if tr.Search("x") {
		t.Error("Delete 后 Search(x): got true, want false")
	}
	if tr.Len() != 0 {
		t.Errorf("Delete 后 Len(): got %d, want 0", tr.Len())
	}
}

func TestBoundary_Unicode(t *testing.T) {
	tr := New()
	words := []string{"你好", "你好世界", "你好吗", "hello", "héllo"}
	for _, w := range words {
		tr.Insert(w)
	}
	if !tr.HasPrefix("你好") {
		t.Error("HasPrefix(你好): got false, want true")
	}
	got := tr.WithPrefix("你好")
	sort.Strings(got)
	want := []string{"你好", "你好世界", "你好吗"}
	if !slicesEqual(got, want) {
		t.Errorf("WithPrefix(你好): got %v, want %v", got, want)
	}
}

func TestBoundary_ReInsertAfterDelete(t *testing.T) {
	tr := New()
	tr.Insert("hello")
	tr.Delete("hello")
	if ok := tr.Insert("hello"); !ok {
		t.Error("Delete 后重新 Insert: got false, want true")
	}
	if !tr.Search("hello") {
		t.Error("重新 Insert 后 Search: got false, want true")
	}
	if tr.Len() != 1 {
		t.Errorf("重新 Insert 后 Len(): got %d, want 1", tr.Len())
	}
}

func TestBoundary_DeleteOnlySharedPrefix(t *testing.T) {
	// 删除作为其他词前缀的词（如 app 是 apple 的前缀）
	tr := New()
	tr.Insert("app")
	tr.Insert("apple")
	tr.Delete("app")
	if tr.Search("app") {
		t.Error("Delete app 后 Search(app): got true, want false")
	}
	if !tr.Search("apple") {
		t.Error("Delete app 后 Search(apple): got false, want true")
	}
	// apple 存在，HasPrefix("app") 仍应为 true
	if !tr.HasPrefix("app") {
		t.Error("Delete app 后 HasPrefix(app): got false, want true（apple 仍存在）")
	}
}

func TestSearch_EmptyStringInserted(t *testing.T) {
	tr := New()
	tr.Insert("")
	if !tr.Search("") {
		t.Error(`插入空字符串后 Search(""): got false, want true`)
	}
	if tr.Len() != 1 {
		t.Errorf("Len(): got %d, want 1", tr.Len())
	}
}

var benchSizes = []int{100, 1_000, 10_000, 100_000, 1_000_000}

// makeBenchWords 生成 n 个随机词（固定种子，结果可复现）。
func makeBenchWords(n int) []string {
	rng := rand.New(rand.NewPCG(42, 0))
	words := make([]string, n)
	for i := range words {
		length := 3 + rng.IntN(8)
		runes := make([]rune, length)
		for j := range runes {
			runes[j] = rune('a' + rng.IntN(26))
		}
		words[i] = string(runes)
	}
	return words
}

// BenchmarkSearch n 个词的 Trie 上随机精确查找。
func BenchmarkSearch(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			words := makeBenchWords(n)
			tr := New()
			for _, w := range words {
				tr.Insert(w)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := range b.N {
				tr.Search(words[i%n])
			}
		})
	}
}

// BenchmarkInsert 维持集合规模稳定：删一个插一个。
func BenchmarkInsert(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			words := makeBenchWords(n * 2) // [0,n) 预填充，[n,2n) 替换池
			tr := New()
			for i := range n {
				tr.Insert(words[i])
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := range b.N {
				tr.Delete(words[i%n])
				tr.Insert(words[n+i%n])
			}
		})
	}
}

// BenchmarkHasPrefix n 个词的 Trie 上前缀查询。
func BenchmarkHasPrefix(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			words := makeBenchWords(n)
			tr := New()
			for _, w := range words {
				tr.Insert(w)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := range b.N {
				w := words[i%n]
				runes := []rune(w)
				prefix := string(runes[:max(1, len(runes)/2)])
				tr.HasPrefix(prefix)
			}
		})
	}
}

// BenchmarkMixed 模拟真实负载：60% Search，20% Insert/Delete，20% HasPrefix/WithPrefix。
func BenchmarkMixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			words := makeBenchWords(n * 2)
			tr := New()
			for i := range n {
				tr.Insert(words[i])
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := range b.N {
				switch i % 10 {
				case 0, 1: // 20% Insert/Delete
					tr.Delete(words[i%n])
					tr.Insert(words[n+i%n])
				case 2, 3: // 20% HasPrefix/WithPrefix
					w := words[i%n]
					runes := []rune(w)
					prefix := string(runes[:max(1, len(runes)/2)])
					if i%2 == 0 {
						tr.HasPrefix(prefix)
					} else {
						tr.WithPrefix(prefix)
					}
				default: // 60% Search
					tr.Search(words[i%n])
				}
			}
		})
	}
}

// BenchmarkConcurrent_ReadHeavy 并发读（Search+HasPrefix）为主，少量写。
// 观察读写锁在不同并发度下的竞争代价。
func BenchmarkConcurrent_ReadHeavy(b *testing.B) {
	const n = 10_000
	words := makeBenchWords(n)
	for _, p := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("p=%d", p), func(b *testing.B) {
			tr := New()
			for _, w := range words {
				tr.Insert(w)
			}
			b.SetParallelism(p)
			b.ResetTimer()
			b.ReportAllocs()
			var i atomic.Int64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					idx := int(i.Add(1)) % n
					if idx%10 == 0 { // 10% 写
						tr.Insert(words[idx])
					} else { // 90% 读
						tr.Search(words[idx])
					}
				}
			})
		})
	}
}
