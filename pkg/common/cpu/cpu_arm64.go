// Package cpu.

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

//go:build arm64 || arm64be

package cpu

// CacheLinePadSize 与 Go internal/cpu 保持一致。
// 取 128 而非实际的 64，是因为 Apple Silicon（M1 及后续）的 cache line 为 128 字节，
// 多 padding 64 字节不影响正确性，且对未来硬件更具兼容性。
// 参考：go/src/internal/cpu/cpu_arm64.go
const CacheLinePadSize = 128
