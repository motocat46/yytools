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
	"sort"

	"github.com/motocat46/yytools/pkg/ds/trie"
)

func ExampleTrie() {
	tr := trie.New()
	tr.Insert("apple")
	tr.Insert("application")
	tr.Insert("app")
	tr.Insert("banana")

	fmt.Println(tr.Search("apple"))     // true
	fmt.Println(tr.Search("ap"))        // false（仅前缀）
	fmt.Println(tr.HasPrefix("app"))    // true
	fmt.Println(tr.HasPrefix("cherry")) // false

	// WithPrefix 无序，排序后输出
	words := tr.WithPrefix("app")
	sort.Strings(words)
	fmt.Println(words)

	fmt.Println(tr.Len()) // 4
	// Output:
	// true
	// false
	// true
	// false
	// [app apple application]
	// 4
}

func ExampleTrie_Delete() {
	tr := trie.New()
	tr.Insert("app")
	tr.Insert("apple")

	fmt.Println(tr.Delete("apple"))  // true
	fmt.Println(tr.Delete("apple"))  // false（已删除）
	fmt.Println(tr.Search("app"))    // true（共享前缀未受影响）
	fmt.Println(tr.HasPrefix("app")) // true
	// Output:
	// true
	// false
	// true
	// true
}
