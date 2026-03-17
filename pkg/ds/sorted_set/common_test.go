// Package sorted_set.

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
// 创建日期:2023/6/1
package sorted_set

import (
	"math/rand"
	"testing"
)

// --- 方案一：Mock 注入（确定性，必然触发上界）---

// TestRandomLevel_AlwaysPromote 注入"永远晋升"的假随机函数，
// 确保 randomLevel 在必然触顶时返回值恰好等于 SKIPLIST_MAXLEVEL，不越界。
// 这是上界 bug 的确定性回归测试——修复前此测试必然失败（level 会到 33）。
func TestRandomLevel_AlwaysPromote(t *testing.T) {
	orig := randInt31
	randInt31 = func() int32 { return 0 } // 0 < 任意正 threshold，始终晋升
	defer func() { randInt31 = orig }()
	
	level := randomLevel(0.25)
	if level != SKIPLIST_MAXLEVEL {
		t.Errorf("永远晋升时期望 level=%d，实际 level=%d", SKIPLIST_MAXLEVEL, level)
	}
}

// TestRandomLevel_NeverPromote 注入"永不晋升"的假随机函数，
// 确保 randomLevel 返回最小值 1。
func TestRandomLevel_NeverPromote(t *testing.T) {
	orig := randInt31
	randInt31 = func() int32 { return 0x7fff_ffff } // MaxInt32，始终 >= threshold
	defer func() { randInt31 = orig }()
	
	level := randomLevel(0.25)
	if level != 1 {
		t.Errorf("永不晋升时期望 level=1，实际 level=%d", level)
	}
}

// --- 方案二：高概率采样（统计覆盖，补充随机路径）---

// TestRandomLevel_Range_HighProb 使用接近 1 的概率跑 100 万次，
// 验证 level 始终在 [1, SKIPLIST_MAXLEVEL] 内，不依赖特定随机路径。
func TestRandomLevel_Range_HighProb(t *testing.T) {
	const iterations = 1_000_000
	for i := range iterations {
		level := randomLevel(0.999)
		if level < 1 || level > SKIPLIST_MAXLEVEL {
			t.Fatalf("第 %d 次: level=%d 越界 [1, %d]", i, level, SKIPLIST_MAXLEVEL)
		}
	}
}

// TestRandomLevel_Range_NormalProb 使用正常概率（0.25）跑 100 万次，验证值域正确。
func TestRandomLevel_Range_NormalProb(t *testing.T) {
	const iterations = 1_000_000
	rng := rand.New(rand.NewSource(42)) // 固定种子，失败可复现
	
	orig := randInt31
	randInt31 = func() int32 { return rng.Int31() }
	defer func() { randInt31 = orig }()
	
	for i := range iterations {
		level := randomLevel(0.25)
		if level < 1 || level > SKIPLIST_MAXLEVEL {
			t.Fatalf("第 %d 次: level=%d 越界 [1, %d]", i, level, SKIPLIST_MAXLEVEL)
		}
	}
}

// --- 方案三：Fuzz（探索未知的概率参数边界）---

// FuzzRandomLevel 对 randomLevel 做模糊测试，探索 prob 参数空间下的未知边界。
// 运行方式：go test -fuzz=FuzzRandomLevel -fuzztime=30s ./pkg/ds/sorted_set/
// Fuzz 引擎发现的失败语料会保存到 testdata/fuzz/FuzzRandomLevel/，后续作为回归用例自动执行。
func FuzzRandomLevel(f *testing.F) {
	// 初始语料：覆盖低概率、常用概率、接近上界的概率
	f.Add(float32(0.0))
	f.Add(float32(0.25))
	f.Add(float32(0.5))
	f.Add(float32(0.75))
	f.Add(float32(0.999))
	
	f.Fuzz(func(t *testing.T, prob float32) {
		// assert 会对无效概率 panic，跳过无效输入
		if prob < 0 || prob >= 1 {
			return
		}
		level := randomLevel(prob)
		if level < 1 || level > SKIPLIST_MAXLEVEL {
			t.Errorf("prob=%.6f → level=%d 越界 [1, %d]", prob, level, SKIPLIST_MAXLEVEL)
		}
	})
}