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

// wheel 是单层时间轮（非线程安全，由 TimingWheel 的顶层锁保护）。
// tickMs、interval 等层级参数为包级常量；currentTime 由 TimingWheel 统一维护。
type wheel struct {
	buckets []*bucket
}

// newWheel 创建 size 个预分配 bucket 的时间轮层。
func newWheel(size int) *wheel {
	buckets := make([]*bucket, size)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &wheel{buckets: buckets}
}
