// Package math.

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
// 创建日期:2023/10/19
package mathutils

import (
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
	"github.com/stormYuanYang/yytools/pkg/common/concrete/base"
)

// 计算斐波那契数（并将斐波那契数列保存到备忘录中）
// 该方法需要传入备忘录
// 例如：
// var mem *[]int
// Fibonacci(2, mem)
// Fibonacci(5, mem)
// Fibonacci(10, mem)
func Fibonacci[T base.Integer](n T, mem *[]T) T {
	assert.Assert(n >= 1)
	if len(*mem) == 0 {
		// 初始化备忘录
		mem = &[]T{0, 1}
	}
	// 如果已经计算过的斐波那契数就从备忘录中获取
	if int(n-1) < len(*mem) {
		return (*mem)[n-1]
	}
	
	// 将新的斐波那契数放入备忘录中
	for i := len(*mem); i <= int(n); i++ {
		f2 := (*mem)[i-2]
		f1 := (*mem)[i-1]
		sum := f2 + f1
		(*mem) = append((*mem), sum)
	}
	return (*mem)[n-1]
}