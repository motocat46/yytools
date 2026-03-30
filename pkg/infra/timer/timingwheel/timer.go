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
// 作者: yangyuan
// 创建日期: 2026-03-28
package timingwheel

import "sync/atomic"

// Timer 是定时任务节点，同时作为 bucket 双向循环链表的节点。
// 调用方通过 *Timer 取消定时器。
type Timer struct {
	expireAt  int64  // 单调到期时间（ms，相对于 TimingWheel.startTime 的偏移）
	interval  int64  // 0=one-shot；>0=repeating 间隔（ms）
	task      func() // 回调，必须非阻塞

	// 双向链表指针（bucket 内链表使用）
	prev, next *Timer

	// bucket 指针：nil 表示在 overflow heap、已触发或已取消
	bucket atomic.Pointer[bucket]
	// 取消标志：bucket==nil 时（overflow heap 中 / Flush 清除指针的窗口期）靠此取消
	cancelled atomic.Bool
}
