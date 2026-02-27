// Package mathx.

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
// 创建日期:2026/2/27
package mathx

import (
	"testing"
)

// 用于验证派生类型能正确派发
type myInt8 int8

func TestFibN(t *testing.T) {
	// 基础语义验证：多类型共享相同的下标
	t.Run("int8/F(0)=0", func(t *testing.T) {
		if got := FibN[int8](0); got != 0 {
			t.Errorf("FibN[int8](0) = %d, 期望 0", got)
		}
	})
	t.Run("int8/F(1)=1", func(t *testing.T) {
		if got := FibN[int8](1); got != 1 {
			t.Errorf("FibN[int8](1) = %d, 期望 1", got)
		}
	})
	t.Run("int8/F(2)=1", func(t *testing.T) {
		if got := FibN[int8](2); got != 1 {
			t.Errorf("FibN[int8](2) = %d, 期望 1", got)
		}
	})
	t.Run("int8/F(7)=13", func(t *testing.T) {
		if got := FibN[int8](7); got != 13 {
			t.Errorf("FibN[int8](7) = %d, 期望 13", got)
		}
	})
	t.Run("uint8/F(0)=0", func(t *testing.T) {
		if got := FibN[uint8](0); got != 0 {
			t.Errorf("FibN[uint8](0) = %d, 期望 0", got)
		}
	})
	t.Run("uint8/F(7)=13", func(t *testing.T) {
		if got := FibN[uint8](7); got != 13 {
			t.Errorf("FibN[uint8](7) = %d, 期望 13", got)
		}
	})
	t.Run("int64/F(1)=1", func(t *testing.T) {
		if got := FibN[int64](1); got != 1 {
			t.Errorf("FibN[int64](1) = %d, 期望 1", got)
		}
	})
	t.Run("int64/F(7)=13", func(t *testing.T) {
		if got := FibN[int64](7); got != 13 {
			t.Errorf("FibN[int64](7) = %d, 期望 13", got)
		}
	})

	// 各类型最大有效下标的正确值
	t.Run("int8/F(11)=89", func(t *testing.T) {
		if got := FibN[int8](11); got != 89 {
			t.Errorf("FibN[int8](11) = %d, 期望 89", got)
		}
	})
	t.Run("int16/F(23)=28657", func(t *testing.T) {
		if got := FibN[int16](23); got != 28657 {
			t.Errorf("FibN[int16](23) = %d, 期望 28657", got)
		}
	})
	t.Run("int32/F(46)=1836311903", func(t *testing.T) {
		if got := FibN[int32](46); got != 1836311903 {
			t.Errorf("FibN[int32](46) = %d, 期望 1836311903", got)
		}
	})
	t.Run("int64/F(92)=7540113804746346429", func(t *testing.T) {
		if got := FibN[int64](92); got != 7540113804746346429 {
			t.Errorf("FibN[int64](92) = %d, 期望 7540113804746346429", got)
		}
	})
	t.Run("uint8/F(13)=233", func(t *testing.T) {
		if got := FibN[uint8](13); got != 233 {
			t.Errorf("FibN[uint8](13) = %d, 期望 233", got)
		}
	})
	t.Run("uint16/F(24)=46368", func(t *testing.T) {
		if got := FibN[uint16](24); got != 46368 {
			t.Errorf("FibN[uint16](24) = %d, 期望 46368", got)
		}
	})
	t.Run("uint32/F(47)=2971215073", func(t *testing.T) {
		if got := FibN[uint32](47); got != 2971215073 {
			t.Errorf("FibN[uint32](47) = %d, 期望 2971215073", got)
		}
	})
	t.Run("uint64/F(93)=12200160415121876738", func(t *testing.T) {
		if got := FibN[uint64](93); got != 12200160415121876738 {
			t.Errorf("FibN[uint64](93) = %d, 期望 12200160415121876738", got)
		}
	})

	// 派生类型能正确派发
	t.Run("myInt8/F(0)=0", func(t *testing.T) {
		if got := FibN[myInt8](0); got != 0 {
			t.Errorf("FibN[myInt8](0) = %d, 期望 0", got)
		}
	})
	t.Run("myInt8/F(7)=13", func(t *testing.T) {
		if got := FibN[myInt8](7); got != 13 {
			t.Errorf("FibN[myInt8](7) = %d, 期望 13", got)
		}
	})
	t.Run("myInt8/F(11)=89", func(t *testing.T) {
		if got := FibN[myInt8](11); got != 89 {
			t.Errorf("FibN[myInt8](11) = %d, 期望 89", got)
		}
	})
}

func TestFibNMax(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"int8", FibNMax[int8](), 11},
		{"int16", FibNMax[int16](), 23},
		{"int32", FibNMax[int32](), 46},
		{"int64", FibNMax[int64](), 92},
		{"uint8", FibNMax[uint8](), 13},
		{"uint16", FibNMax[uint16](), 24},
		{"uint32", FibNMax[uint32](), 47},
		{"uint64", FibNMax[uint64](), 93},
		{"myInt8", FibNMax[myInt8](), 11},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("FibNMax[%s]() = %d, 期望 %d", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestFibN_OutOfBound(t *testing.T) {
	t.Run("int8超出上限", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		_ = FibN[int8](12) // int8 最大有效下标为 11
	})

	t.Run("int8负数下标", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		_ = FibN[int8](-1)
	})

	t.Run("uint8超出上限", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		_ = FibN[uint8](14) // uint8 最大有效下标为 13
	})

	t.Run("int64超出上限", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望 panic，但未发生")
			}
		}()
		_ = FibN[int64](93) // int64 最大有效下标为 92
	})
}

var benchSinkInt64 int64

func BenchmarkFibNTable(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkInt64 = int64(FibN[int64](92))
	}
}

func BenchmarkFibCalculate(b *testing.B) {
	fib := NewFibMem[int64]()
	fib.Calculate(int64(93)) // 预热备忘录至 F(92)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkInt64 = int64(fib.Calculate(int64(93)))
	}
}
