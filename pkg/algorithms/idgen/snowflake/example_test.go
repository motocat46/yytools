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

package snowflake_test

import (
	"fmt"

	"github.com/motocat46/yytools/pkg/algorithms/idgen/snowflake"
)

// ExampleNewGenerator 展示标准布局（41+10+12 位）的基本用法。
func ExampleNewGenerator() {
	g, err := snowflake.NewGenerator(1) // nodeID=1，范围 0–1023
	if err != nil {
		panic(err)
	}

	id := g.NewID()
	parts := g.ParseID(id)

	fmt.Println(parts.NodeID)            // 1
	fmt.Println(parts.Sequence >= 0)     // true
	fmt.Println(parts.Timestamp > 0)     // true
	// Output:
	// 1
	// true
	// true
}

// ExampleNewGeneratorWithLayout 展示自定义位布局：
// 扩展 nodeID 到 12 位（支持 4096 个节点），压缩 sequence 到 11 位。
func ExampleNewGeneratorWithLayout() {
	layout := snowflake.Layout{
		TimestampBits: 40,
		NodeIDBits:    12, // 最多 4096 个节点
		SequenceBits:  11, // 每毫秒每节点最多 2048 个 ID
		Epoch:         snowflake.Epoch,
	}
	g, err := snowflake.NewGeneratorWithLayout(100, layout)
	if err != nil {
		panic(err)
	}

	parts := g.ParseID(g.NewID())
	fmt.Println(parts.NodeID) // 100
	// Output:
	// 100
}
