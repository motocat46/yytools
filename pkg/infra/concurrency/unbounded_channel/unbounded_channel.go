// Package unbounded_channel.

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
// 创建日期: 2025/6/12
package unbounded_channel

// UnboundedChannel 是无界通道的公开类型，当前实现为 V6。
//
// 提供有序（FIFO）、无界（带背压）的消息传递：
//   - 快速路径：无 buffer 积压时直接投入底层 channel，接近 native channel 性能
//   - 慢路径：channel 满时暂存 buffer，由 worker 事件驱动搬运，不阻塞生产者
//   - 背压：buffer 超过 limit 时生产者阻塞，避免内存无限增长
//
// 基本用法：
//
//	uc := NewUnboundedChannel[int](1024, 100_000)
//	defer uc.Close()
//
//	// 生产者
//	uc.Send(42)
//
//	// 消费者（阻塞接收）
//	val, ok := uc.Receive()
//
//	// 消费者（select 多路复用）
//	for v := range uc.Out() { ... }
//
// 注意：必须调用 Close()，否则内部 worker goroutine 永久泄漏。
type UnboundedChannel[T any] = UnboundedChannelV6[T]

// NewUnboundedChannel 创建一个无界通道。
//
//   - chanSize：底层 channel 的缓冲大小，建议设为预期并发量的 2-4 倍
//   - limit：buffer 中允许积压的消息上限，超过后生产者会阻塞（背压）
func NewUnboundedChannel[T any](chanSize int, limit int) *UnboundedChannel[T] {
	return NewUnboundedChannelV6[T](chanSize, limit)
}
