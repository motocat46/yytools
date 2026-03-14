// Package cpu 提供 CPU 相关的平台级常量与工具类型。
//
// 命名与数值直接对齐 Go 标准库 internal/cpu，但以 build tag 分文件的方式暴露给用户代码使用：
//   - CacheLinePadSize：当前平台的 cache line 字节数（编译期常量）
//   - CacheLinePad：用于结构体字段 padding、避免伪共享（false sharing）的零业务字段类型

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
// 创建日期:2025/6/12
package cpu

// CacheLinePad 用于在结构体中插入 padding，将相邻字段隔离到不同的 cache line，
// 从而避免多核并发写时的伪共享（false sharing）问题。
//
// 用法示例：
//
//	type HotField struct {
//	    val atomic.Int64
//	    _   cpu.CacheLinePad  // 将 val 与后续字段隔离到不同 cache line
//	}
//
// 与 Go 标准库 internal/cpu.CacheLinePad 的定义完全一致。
type CacheLinePad struct{ _ [CacheLinePadSize]byte }
