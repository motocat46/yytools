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

// 作者:  yangyuan
// 创建日期:2026/3/25

package workerpool

import "sync"

// submitLocker 抽象 Submit 与 Close 的加锁行为。
//
// Submit 调用 lockSubmit/unlockSubmit（可并发持锁）。
// Close  调用 lockClose/unlockClose（独占）。
//
// Mutex 实现：两端都用排他锁。
// RWMutex 实现：Submit 用读锁（允许并发），Close 用写锁（独占）。
type submitLocker interface {
	lockSubmit()
	unlockSubmit()
	lockClose()
	unlockClose()
}

// mutexLocker 使用 sync.Mutex，Submit 与 Close 均排他。
type mutexLocker struct{ mu sync.Mutex }

func (l *mutexLocker) lockSubmit()   { l.mu.Lock() }
func (l *mutexLocker) unlockSubmit() { l.mu.Unlock() }
func (l *mutexLocker) lockClose()    { l.mu.Lock() }
func (l *mutexLocker) unlockClose()  { l.mu.Unlock() }

// rwMutexLocker 使用 sync.RWMutex，Submit 持读锁，Close 持写锁。
// 多个 Submit 可并发通过，Close 等待所有当前 Submit 完成后独占推进。
type rwMutexLocker struct{ mu sync.RWMutex }

func (l *rwMutexLocker) lockSubmit()   { l.mu.RLock() }
func (l *rwMutexLocker) unlockSubmit() { l.mu.RUnlock() }
func (l *rwMutexLocker) lockClose()    { l.mu.Lock() }
func (l *rwMutexLocker) unlockClose()  { l.mu.Unlock() }
