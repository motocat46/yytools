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
	"math/rand/v2"

	"github.com/motocat46/yytools/pkg/common/assert"
)

// randInt31 是包级变量，默认使用标准库随机数，测试中可替换为确定性实现。
var randInt31 func() int32 = rand.Int32

// randomLevel 随机计算跳跃表中某个结点的高度，范围在闭区间[1, SKIPLIST_MAXLEVEL]内
func randomLevel(levelUpProbability float32) int {
	assert.Assert(levelUpProbability >= 0 && levelUpProbability < 1,
		"提升节点高度概率不正确:", levelUpProbability, "正常范围:[0.0,1)")
	
	threshold := int32(levelUpProbability * RAND_MAX)
	// 按概率提升等级
	level := 1
	for randInt31() < threshold {
		level++
		// 达到或超过约定上限等级跳出循环（用 >= 而非 ==，防止起始值异常或步长变化导致守卫失效）
		if level >= SKIPLIST_MAXLEVEL {
			break
		}
	}
	assert.Assert(level <= SKIPLIST_MAXLEVEL, level)
	return level
}