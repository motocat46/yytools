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
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "math/rand"
)

func random() int32 {
	return rand.Int31()
}

// 随机计算跳跃表中某个结点的高度(等级)
// 高度范围在闭区间[1, maxLevel]内
func randomLevel(levelUpProbability float32) int {
	assert.Assert(levelUpProbability >= 0 && levelUpProbability < 1,
		"提升节点高度概率不正确:", levelUpProbability, "正常范围:[0.0,1)")
	
	level := 1
	// 提升等级的概率阈值(将小数形式的概率转换成整数形式的概率)
	// 而且，得到的阈值一定是在[0,RAND_MAX)范围内的
	threshold := int32(levelUpProbability * RAND_MAX)
	// 满足两个条件就可以提升等级:
	// 1.等级小于等于指定最大等级 且
	// 2.满足指定概率
	// 否则退出循环
	for random() < threshold && level <= SKIPLIST_MAXLEVEL {
		level++
	}
	return level
}