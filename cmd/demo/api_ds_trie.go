// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/4/15
package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	trie_pkg "github.com/motocat46/yytools/pkg/ds/trie"
)

const (
	trieOpsPerMeasure = 1000
	trieCharset       = "abcdefghijklmnopqrstuvwxyz"
)

// randString 生成长度为 length 的随机小写字母字符串。
func randString(rng *rand.Rand, length int) string {
	var b strings.Builder
	b.Grow(length)
	for range length {
		b.WriteByte(trieCharset[rng.IntN(len(trieCharset))])
	}
	return b.String()
}

// fillTrie 预填充 n 个长度为 keyLen 的词到 Trie，使用固定种子。
// 返回 Trie 及插入的所有词（用于后续 Search 计时）。
func fillTrie(n, keyLen int) (*trie_pkg.Trie, []string) {
	rng := rand.New(rand.NewPCG(42, 0))
	t := trie_pkg.New()
	words := make([]string, n)
	for i := range n {
		w := randString(rng, keyLen)
		t.Insert(w)
		words[i] = w
	}
	return t, words
}

// ---- Chart 1: 耗时 vs key 长度（固定词典 10 万） ----

const trieFixedDictSize = 100_000

func handleDsTrie(w http.ResponseWriter, _ *http.Request) {
	// Chart 1：耗时 vs key 长度
	keyLengths := []int{4, 8, 16, 32, 64}
	xLabels1 := make([]string, len(keyLengths))
	insertNs := make([]int64, len(keyLengths))
	searchNs := make([]int64, len(keyLengths))
	hasPrefixNs := make([]int64, len(keyLengths))

	for i, kl := range keyLengths {
		xLabels1[i] = fmt.Sprintf("%d", kl)

		t, words := fillTrie(trieFixedDictSize, kl)
		rng := rand.New(rand.NewPCG(99, 0))

		// Insert：向已有 trieFixedDictSize 词的 Trie 插入新词
		start := time.Now()
		for j := range trieOpsPerMeasure {
			t.Insert(randString(rng, kl) + fmt.Sprintf("%d", j)) // 加序号保证不重复
		}
		insertNs[i] = time.Since(start).Nanoseconds() / trieOpsPerMeasure

		// Search：全命中，随机选已有词
		start = time.Now()
		for range trieOpsPerMeasure {
			t.Search(words[rng.IntN(len(words))])
		}
		searchNs[i] = time.Since(start).Nanoseconds() / trieOpsPerMeasure

		// HasPrefix：取词的前半段作为前缀
		prefixLen := kl / 2
		if prefixLen == 0 {
			prefixLen = 1
		}
		start = time.Now()
		for range trieOpsPerMeasure {
			prefix := words[rng.IntN(len(words))][:prefixLen]
			t.HasPrefix(prefix)
		}
		hasPrefixNs[i] = time.Since(start).Nanoseconds() / trieOpsPerMeasure
	}

	// Chart 2：Trie Search vs map[string]bool（vs 词典规模，固定 key 长 16）
	const fixedKeyLen = 16
	dictSizes := []int{10000, 50000, 100000, 200000, 500000}
	xLabels2 := make([]string, len(dictSizes))
	trieSearchNs := make([]int64, len(dictSizes))
	mapSearchNs := make([]int64, len(dictSizes))

	for i, n := range dictSizes {
		xLabels2[i] = fmt.Sprintf("%d万", n/10000)

		t2, words2 := fillTrie(n, fixedKeyLen)
		rng2 := rand.New(rand.NewPCG(77, 0))

		// 同时构建等价 map
		m := make(map[string]bool, n)
		for _, word := range words2 {
			m[word] = true
		}

		// Trie Search
		start := time.Now()
		for range trieOpsPerMeasure {
			t2.Search(words2[rng2.IntN(len(words2))])
		}
		trieSearchNs[i] = time.Since(start).Nanoseconds() / trieOpsPerMeasure

		// map lookup（重置 rng 使用相同 key 序列）
		rng2 = rand.New(rand.NewPCG(77, 0))
		start = time.Now()
		for range trieOpsPerMeasure {
			_ = m[words2[rng2.IntN(len(words2))]]
		}
		mapSearchNs[i] = time.Since(start).Nanoseconds() / trieOpsPerMeasure
	}

	json.NewEncoder(w).Encode(pageData{ //nolint:errcheck
		Title: "Trie 操作耗时",
		Charts: []chartData{
			{
				Type:      "line",
				Title:     fmt.Sprintf("Insert / Search / HasPrefix 均摊耗时 vs key 长度（词典 %d 个，每规模 %d 次）", trieFixedDictSize, trieOpsPerMeasure),
				XAxis:     xLabels1,
				XAxisName: "key 长度（字符数）",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Insert", Data: insertNs},
					{Name: "Search（全命中）", Data: searchNs},
					{Name: "HasPrefix（前缀=key/2）", Data: hasPrefixNs},
				},
			},
			{
				Type:      "line",
				Title:     fmt.Sprintf("Trie vs map[string]bool — Search 均摊耗时 vs 词典规模（key 长 %d，每规模 %d 次）", fixedKeyLen, trieOpsPerMeasure),
				XAxis:     xLabels2,
				XAxisName: "词典规模",
				YAxisName: "ns/op",
				Series: []chartSeries{
					{Name: "Trie Search（O(L)，与规模无关）", Data: trieSearchNs},
					{Name: "map[string]bool lookup", Data: mapSearchNs},
				},
			},
		},
	})
}

func init() {
	Register(VisEntry{
		Pkg: "pkg/ds", SubPkg: "trie/", Title: "Trie 操作耗时",
		Desc: "Insert/Search/HasPrefix 耗时 vs key 长度（O(L)特性）；Trie vs map Search 耗时 vs 词典规模",
		Path: "/api/ds/trie", DataHandler: handleDsTrie,
	})
}
