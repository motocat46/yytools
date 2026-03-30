// Package timingwheel 实现分层时间轮定时器，O(1) add/cancel，毫秒精度。
//
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

const (
	tickMs = int64(1)

	// L1：8 位，256 槽，覆盖 256ms（最热层，高位宽减少每槽平均节点数）
	l1Bits = 8
	l1Size = 1 << l1Bits // 256
	l1Mask = l1Size - 1

	// L2-L5：各 6 位，64 槽；8+6*4=32 覆盖完整 32 位时间戳
	l2Bits = 6
	l2Size = 1 << l2Bits // 64
	l2Mask = l2Size - 1

	// levelShift：expireAt 对应层的起始位
	l1Shift = 0
	l2Shift = l1Bits
	l3Shift = l1Bits + l2Bits
	l4Shift = l1Bits + l2Bits*2
	l5Shift = l1Bits + l2Bits*3

	// 各层覆盖时间范围（ms）
	l1Interval = int64(l1Size)              // 256ms
	l2Interval = l1Interval * int64(l2Size) // 16384ms  (~16s)
	l3Interval = l2Interval * int64(l2Size) // 1048576ms (~17.5min)
	l4Interval = l3Interval * int64(l2Size) // 67108864ms (~18.7h)
	l5Interval = l4Interval * int64(l2Size) // 4294967296ms (~49.7天)
)
