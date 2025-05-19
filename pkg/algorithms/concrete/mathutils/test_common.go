// Package mathutils.

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
// 创建日期:2023/7/11
package mathutils

import (
    "fmt"
	"github.com/stormYuanYang/yytools/pkg/algorithms/concrete/mathutils/random"
    "github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
    "time"
)

func GcdTest(num int) {
	for i := 0; i < num; i++ {
		x := random.RandInt(0, 10000)
		y := random.RandInt(0, 10000)
		d1 := GcdR(x, y)
		d2 := GcdI(x, y)
		d3 := Gcd(x, y)
		assert.Assert(d1 == d2 && d1 == d3, "x:", x, " y:", y)
	}
}

var MathCommon_handlers = []func(num int){
    GcdTest,
}

func MathCommonTest(num int) {
	println("mathutils.common测试开始...")
	random.RandSeed(time.Now().UnixMilli())
	for i := 1; i <= num; i++ {
		fmt.Printf("第%d轮测试开始\n", i)
		handlerLength := len(MathCommon_handlers)
		for k := 0; k < handlerLength; k++ {
			// 十万次
			for j := 0; j < 100000; j++ {
				handler := MathCommon_handlers[k]
				handler(1)
			}
		}
		fmt.Printf("第%d轮测试结束\n\n", i)
	}
	println("mathutils.common测试完毕...")
}