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

package sampling_test

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/motocat46/yytools/pkg/algorithms/mathx/sampling"
)

// ExampleSampleKDistinctFloyd 展示 Floyd 采样——从 [lo..hi] 均匀采 k 个不重复的值，O(k) 时间。
func ExampleSampleKDistinctFloyd() {
	r := rand.New(rand.NewPCG(42, 0))
	result := sampling.SampleKDistinctFloyd(1, 100, 5, r)

	fmt.Println(len(result)) // 始终返回 k 个值

	// 验证无重复：排序后无相邻相等元素
	sorted := slices.Clone(result)
	slices.Sort(sorted)
	hasDup := false
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			hasDup = true
		}
	}
	fmt.Println(hasDup) // false：无重复
	// Output:
	// 5
	// false
}

// ExampleSampleWithMinGap 展示带间距约束的采样——从 [L..R] 采 k 个值，相邻值间隔 ≥ gap。
// 返回有序切片，空间压缩算法保证 O(k log k) 确定性终止，无重试。
func ExampleSampleWithMinGap() {
	r := rand.New(rand.NewPCG(42, 0))
	result := sampling.SampleWithMinGap(1, 100, 4, 10, r)

	fmt.Println(len(result)) // 返回 k=4 个值

	// 结果已排序，验证最小间隔 ≥ 10
	minGapOk := true
	for i := 1; i < len(result); i++ {
		if result[i]-result[i-1] < 10 {
			minGapOk = false
		}
	}
	fmt.Println(minGapOk)
	// Output:
	// 4
	// true
}
