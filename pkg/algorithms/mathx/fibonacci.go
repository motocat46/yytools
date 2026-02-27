// Package mathx.

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
package mathx

import (
	"slices"

	"github.com/stormYuanYang/yytools/pkg/common/assert"
	"github.com/stormYuanYang/yytools/pkg/common/base"
)

// TODO 可以考虑给出初始的两个数值(这里默认是0，1;可以继续扩展设定初始数字)
type Fibonacci[T base.Integer] struct {
	mem []T
}

func NewFibMem[T base.Integer]() *Fibonacci[T] {
	// 默认构造2个斐波那契数
	return &Fibonacci[T]{mem: []T{0, 1}}
}

// 计算斐波那契数（并将斐波那契数列保存到备忘录中）
// 该方法需要传入初始为空的备忘录
func (a *Fibonacci[T]) Calculate(n T) T {
	assert.Assert(n >= 1)
	// 如果已经计算过的斐波那契数就从备忘录中获取
	target := int(n - 1)
	if target < len(a.mem) {
		return a.mem[target]
	}

	// 预扩容：避免在热点路径上频繁realloc
	need := (target + 1) - len(a.mem)
	if need > 0 {
		a.mem = slices.Grow(a.mem, need)
	}

	for i := len(a.mem); i <= target; i++ {
		f2 := a.mem[i-2]
		f1 := a.mem[i-1]
		a.mem = append(a.mem, f2+f1)
	}
	return a.mem[target]
}
