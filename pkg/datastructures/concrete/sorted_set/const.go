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

import "math"

const (
	SKIPLIST_MAXLEVEL           = 32            // 跳跃表节点的最高高度
	DEFAULT_LEVELUP_PROBABILITY = 0.25          // 提升节点高度的概率
	RAND_MAX                    = math.MaxInt32 // int32的最大值 (0x7fffffff)
)