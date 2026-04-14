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
//
// 作者:  yangyuan
package ring_buffer_test

import (
	"fmt"
	"testing"

	rb "github.com/motocat46/yytools/pkg/ds/ring_buffer"
)

var benchSizes = []int{100, 1000, 10000, 100000, 1000000}

// BenchmarkRingBuffer_Enqueue 纯入队基准（预填充到满后持续覆盖，测稳定状态）
func BenchmarkRingBuffer_Enqueue(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			r := rb.NewRingBuffer[int](n)
			for i := range n {
				r.Enqueue(i) // 预填充到满
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				r.Enqueue(1)
			}
		})
	}
}

// BenchmarkRingBuffer_Mixed Enqueue+Dequeue 混合负载（先 Dequeue 再 Enqueue，规模始终为 n）
func BenchmarkRingBuffer_Mixed(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			r := rb.NewRingBuffer[int](n)
			for i := range n {
				r.Enqueue(i) // 预填充到满
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				_ = r.Dequeue() // 先出队，length 降到 n-1
				r.Enqueue(1)    // 再入队，length 恢复到 n
			}
		})
	}
}

// BenchmarkRingBuffer_Range 全量遍历基准（预填充到满后遍历全部元素）
func BenchmarkRingBuffer_Range(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			r := rb.NewRingBuffer[int](n)
			for i := range n {
				r.Enqueue(i) // 预填充到满
			}
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				r.Range(func(int) bool { return true })
			}
		})
	}
}
