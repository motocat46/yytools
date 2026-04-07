// Copyright 2026 yangyuan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package random_split

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
)

// ────────────────────────────────────────────────────────
// State.Validate
// ────────────────────────────────────────────────────────

func TestState_Validate(t *testing.T) {
	cases := []struct {
		name    string
		state   State
		wantErr bool
	}{
		{"合法最小值", State{1, 1, 1}, false},
		{"合法 amount==count*min", State{10, 5, 2}, false},
		{"合法 amount>count*min", State{100, 10, 1}, false},
		{"非法 count=0", State{10, 0, 1}, true},
		{"非法 min=0", State{10, 5, 0}, true},
		{"非法 amount<count*min", State{9, 5, 2}, true},
		{"非法 count负数", State{10, -1, 1}, true},
		{"非法 min负数", State{10, 5, -1}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.state.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Validate()=nil，期望 ErrInvalidState，state=%+v", tc.state)
				}
				if !errors.Is(err, ErrInvalidState) {
					t.Fatalf("Validate()=%v，期望 errors.Is(err,ErrInvalidState)=true，state=%+v", err, tc.state)
				}
			} else {
				if err != nil {
					t.Fatalf("Validate()=%v，期望 nil，state=%+v", err, tc.state)
				}
			}
		})
	}
}

// ────────────────────────────────────────────────────────
// New / Done / Remaining
// ────────────────────────────────────────────────────────

func TestNew_ValidState(t *testing.T) {
	fn := Fixed()
	d, err := New(State{100, 10, 1}, fn, nil)
	if err != nil {
		t.Fatalf("New()=%v，期望 nil", err)
	}
	if d == nil {
		t.Fatal("New() 返回 nil Distributor")
	}
	if d.Done() {
		t.Fatal("新建 Distributor Done() 应为 false")
	}
	rem := d.Remaining()
	if rem.RemainAmount != 100 || rem.RemainCount != 10 || rem.MinPerPart != 1 {
		t.Fatalf("Remaining()=%+v，期望 {100,10,1}", rem)
	}
}

func TestNew_InvalidState(t *testing.T) {
	fn := Fixed()
	_, err := New(State{0, 0, 0}, fn, nil)
	if err == nil {
		t.Fatal("New() 期望返回错误，但返回 nil")
	}
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("错误类型期望 ErrInvalidState，got: %v", err)
	}
}

func TestNew_NilSampleFunc(t *testing.T) {
	_, err := New(State{100, 10, 1}, nil, nil)
	if err == nil {
		t.Fatal("SampleFunc 为 nil 时 New() 应返回错误")
	}
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("错误类型期望 ErrInvalidParam，got: %v", err)
	}
}

func TestDone_AfterAllAllocated(t *testing.T) {
	fn := Fixed()
	d, _ := New(State{10, 2, 1}, fn, nil)
	if d.Done() {
		t.Fatal("初始 Done() 应为 false")
	}
	d.Next() //nolint
	if d.Done() {
		t.Fatal("分配 1/2 份后 Done() 仍应为 false")
	}
	d.Next() //nolint
	if !d.Done() {
		t.Fatal("分配完所有份后 Done() 应为 true")
	}
}

// ────────────────────────────────────────────────────────
// Next
// ────────────────────────────────────────────────────────

func TestNext_NormalSequence(t *testing.T) {
	// Fixed策略：S=10,N=3,min=1 → 预期序列 3,3,4
	fn := Fixed()
	rng := rand.New(rand.NewPCG(42, 0))
	d, _ := New(State{10, 3, 1}, fn, rng)

	v1, err := d.Next()
	if err != nil {
		t.Fatalf("第1次 Next() 错误: %v", err)
	}
	if v1 != 3 {
		t.Errorf("第1次 Next()=%d，期望 3（10/3=3）", v1)
	}

	v2, err := d.Next()
	if err != nil {
		t.Fatalf("第2次 Next() 错误: %v", err)
	}
	if v2 != 3 {
		t.Errorf("第2次 Next()=%d，期望 3（7/2=3）", v2)
	}

	v3, err := d.Next()
	if err != nil {
		t.Fatalf("第3次 Next() 错误: %v", err)
	}
	if v3 != 4 {
		t.Errorf("第3次（最后一份）Next()=%d，期望 4（剩余全部）", v3)
	}

	if !d.Done() {
		t.Error("全部分配后 Done() 应为 true")
	}
	if v1+v2+v3 != 10 {
		t.Errorf("sum=%d，期望 10（守恒性）", v1+v2+v3)
	}
}

func TestNext_AfterDone_ReturnsErrDone(t *testing.T) {
	fn := Fixed()
	d, _ := New(State{5, 1, 1}, fn, nil)
	d.Next() //nolint
	if !d.Done() {
		t.Fatal("应已 Done")
	}
	v, err := d.Next()
	if !errors.Is(err, ErrDone) {
		t.Fatalf("Done后 Next() 应返回 ErrDone，got v=%d err=%v", v, err)
	}
	if v != 0 {
		t.Errorf("Done后 Next() 应返回 0，got %d", v)
	}
}

func TestNext_SampleFuncError_StateUnchanged(t *testing.T) {
	callCount := 0
	fn := SampleFunc(func(state State, rng *rand.Rand) (int64, error) {
		callCount++
		if callCount == 1 {
			return 0, fmt.Errorf("模拟外部资源失败")
		}
		return state.MinPerPart, nil
	})

	rng := rand.New(rand.NewPCG(1, 0))
	d, _ := New(State{10, 2, 1}, fn, rng)
	remBefore := d.Remaining()

	_, err := d.Next()
	if err == nil {
		t.Fatal("期望 SampleFunc 返回 error 时 Next() 透传错误")
	}
	remAfter := d.Remaining()
	if remAfter != remBefore {
		t.Errorf("SampleFunc 返回 error 后状态改变了：before=%+v after=%+v", remBefore, remAfter)
	}

	v, err := d.Next()
	if err != nil {
		t.Fatalf("重试时 Next() 错误: %v", err)
	}
	if v < 1 {
		t.Errorf("重试时 Next()=%d，期望 >= min=1", v)
	}
}

// ────────────────────────────────────────────────────────
// Allocate
// ────────────────────────────────────────────────────────

func TestAllocate_LengthAndConservation(t *testing.T) {
	fn := Fixed()
	rng := rand.New(rand.NewPCG(42, 0))
	state := State{100, 10, 1}
	d, _ := New(state, fn, rng)

	result, err := d.Allocate()
	if err != nil {
		t.Fatalf("Allocate() 错误: %v", err)
	}
	if len(result) != 10 {
		t.Errorf("len(result)=%d，期望 10（RemainCount）", len(result))
	}
	var sum int64
	for i, v := range result {
		if v < 1 {
			t.Errorf("result[%d]=%d < min=1（合法性违反）", i, v)
		}
		sum += v
	}
	if sum != 100 {
		t.Errorf("sum=%d，期望 100（守恒性违反）", sum)
	}
}

func TestAllocate_AfterNext_ReturnsErrAlreadyStarted(t *testing.T) {
	fn := Fixed()
	d, _ := New(State{10, 3, 1}, fn, nil)
	d.Next() //nolint
	_, err := d.Allocate()
	if !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("调用过 Next() 后 Allocate() 应返回 ErrAlreadyStarted，got: %v", err)
	}
}

func TestAllocate_Twice_ReturnsErrAlreadyStarted(t *testing.T) {
	fn := Fixed()
	d, _ := New(State{6, 3, 1}, fn, nil)
	_, err := d.Allocate()
	if err != nil {
		t.Fatalf("Allocate() 首次调用应成功，got: %v", err)
	}
	_, err = d.Allocate()
	if !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("Allocate() 第二次调用应返回 ErrAlreadyStarted，got: %v", err)
	}
}

// ────────────────────────────────────────────────────────
// Uniform 策略
// ────────────────────────────────────────────────────────

func TestUniform_AllValuesInRange(t *testing.T) {
	fn := Uniform()
	state := State{RemainAmount: 100, RemainCount: 5, MinPerPart: 1}
	safeUpper := state.RemainAmount - (state.RemainCount-1)*state.MinPerPart
	rng := rand.New(rand.NewPCG(99, 0))

	for i := 0; i < 10000; i++ {
		v, err := fn(state, rng)
		if err != nil {
			t.Fatalf("Uniform() 错误: %v", err)
		}
		if v < state.MinPerPart || v > safeUpper {
			t.Fatalf("第%d次 Uniform()=%d 超出范围 [%d,%d]", i, v, state.MinPerPart, safeUpper)
		}
	}
}

func TestUniform_ConservationViaAllocate(t *testing.T) {
	fn := Uniform()
	rng := rand.New(rand.NewPCG(7, 0))
	d, _ := New(State{1000, 20, 3}, fn, rng)
	result, err := d.Allocate()
	if err != nil {
		t.Fatalf("Allocate() 错误: %v", err)
	}
	var sum int64
	for i, v := range result {
		if v < 3 {
			t.Errorf("result[%d]=%d < min=3（合法性违反）", i, v)
		}
		sum += v
	}
	if sum != 1000 {
		t.Errorf("sum=%d，期望 1000（守恒）", sum)
	}
}

// ────────────────────────────────────────────────────────
// MeanBounded 策略
// ────────────────────────────────────────────────────────

func TestMeanBounded_MultiplierLessThan1_ReturnsError(t *testing.T) {
	cases := []float64{0.0, 0.5, 0.99, -1.0}
	for _, m := range cases {
		fn, err := MeanBounded(m)
		if fn != nil {
			t.Errorf("MeanBounded(%.2f) 应返回 nil SampleFunc", m)
		}
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("MeanBounded(%.2f) 期望 ErrInvalidParam，got: %v", m, err)
		}
	}
}

func TestMeanBounded_Multiplier1_IsLegal(t *testing.T) {
	fn, err := MeanBounded(1.0)
	if err != nil {
		t.Fatalf("MeanBounded(1.0) 应合法，got: %v", err)
	}
	if fn == nil {
		t.Fatal("MeanBounded(1.0) 返回 nil SampleFunc")
	}
}

func TestMeanBounded_2_EquivalentToDoubleMean(t *testing.T) {
	state := State{100, 5, 1}
	rng1 := rand.New(rand.NewPCG(42, 0))
	rng2 := rand.New(rand.NewPCG(42, 0))

	fn1, _ := MeanBounded(2.0)
	fn2 := DoubleMean()

	for i := 0; i < 100; i++ {
		v1, _ := fn1(state, rng1)
		v2, _ := fn2(state, rng2)
		if v1 != v2 {
			t.Fatalf("第%d次：MeanBounded(2.0)=%d vs DoubleMean()=%d，期望相等", i, v1, v2)
		}
	}
}

func TestMeanBounded_BoundsRespected(t *testing.T) {
	state := State{100, 10, 1}
	fn := DoubleMean()
	rng := rand.New(rand.NewPCG(13, 0))

	for i := 0; i < 50000; i++ {
		avg := state.RemainAmount / state.RemainCount
		safeUpper := state.RemainAmount - (state.RemainCount-1)*state.MinPerPart
		upper := min(safeUpper, int64(math.Floor(2.0*float64(avg))))

		v, err := fn(state, rng)
		if err != nil {
			t.Fatalf("DoubleMean() 错误: %v", err)
		}
		if v < state.MinPerPart {
			t.Fatalf("v=%d < min=%d（合法性违反）", v, state.MinPerPart)
		}
		if v > upper {
			t.Fatalf("v=%d > upper=%d（有界性违反，avg=%d，safeUpper=%d）", v, upper, avg, safeUpper)
		}
	}
}

// ────────────────────────────────────────────────────────
// Fixed 策略
// ────────────────────────────────────────────────────────

func TestFixed_ReturnsTruncatedMean(t *testing.T) {
	fn := Fixed()
	rng := rand.New(rand.NewPCG(42, 0))

	v1, err := fn(State{10, 3, 1}, rng)
	if err != nil {
		t.Fatalf("Fixed() 错误: %v", err)
	}
	if v1 != 3 {
		t.Errorf("Fixed({10,3,1})=%d，期望 3", v1)
	}

	v2, err := fn(State{7, 2, 1}, rng)
	if err != nil {
		t.Fatalf("Fixed() 错误: %v", err)
	}
	if v2 != 3 {
		t.Errorf("Fixed({7,2,1})=%d，期望 3", v2)
	}
}

func TestFixed_AllocateSumConservation(t *testing.T) {
	fn := Fixed()
	rng := rand.New(rand.NewPCG(0, 0))
	d, _ := New(State{7, 3, 1}, fn, rng)
	result, err := d.Allocate()
	if err != nil {
		t.Fatalf("Allocate() 错误: %v", err)
	}
	var sum int64
	for _, v := range result {
		sum += v
	}
	if sum != 7 {
		t.Errorf("sum=%d，期望 7（守恒）", sum)
	}
}

// ────────────────────────────────────────────────────────
// error 类型识别（%w 包装有效）
// ────────────────────────────────────────────────────────

func TestErrorsIs(t *testing.T) {
	t.Run("ErrInvalidState via New", func(t *testing.T) {
		_, err := New(State{0, 0, 0}, Fixed(), nil)
		if !errors.Is(err, ErrInvalidState) {
			t.Errorf("errors.Is(err, ErrInvalidState) 应为 true，got: %v", err)
		}
	})

	t.Run("ErrInvalidParam via MeanBounded", func(t *testing.T) {
		_, err := MeanBounded(0.5)
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("errors.Is(err, ErrInvalidParam) 应为 true，got: %v", err)
		}
	})

	t.Run("ErrInvalidParam via New nil fn", func(t *testing.T) {
		_, err := New(State{10, 3, 1}, nil, nil)
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("errors.Is(err, ErrInvalidParam) 应为 true，got: %v", err)
		}
	})

	t.Run("ErrDone via Next after Done", func(t *testing.T) {
		d, _ := New(State{1, 1, 1}, Fixed(), nil)
		d.Next() //nolint
		_, err := d.Next()
		if !errors.Is(err, ErrDone) {
			t.Errorf("errors.Is(err, ErrDone) 应为 true，got: %v", err)
		}
	})

	t.Run("ErrAlreadyStarted via Allocate after Next", func(t *testing.T) {
		d, _ := New(State{10, 3, 1}, Fixed(), nil)
		d.Next() //nolint
		_, err := d.Allocate()
		if !errors.Is(err, ErrAlreadyStarted) {
			t.Errorf("errors.Is(err, ErrAlreadyStarted) 应为 true，got: %v", err)
		}
	})
}

// ────────────────────────────────────────────────────────
// 随机混合测试
// 1000 组参数 × 3 种策略 × 100 轮 = 300,000 次完整分配
// ────────────────────────────────────────────────────────

func TestRandom_Mixed(t *testing.T) {
	const (
		paramSets = 1000
		rounds    = 100
	)
	rng := rand.New(rand.NewPCG(12345, 0))

	strategies := map[string]SampleFunc{
		"DoubleMean": DoubleMean(),
		"Uniform":    Uniform(),
		"Fixed":      Fixed(),
	}

	for i := 0; i < paramSets; i++ {
		amount := int64(rng.IntN(9991) + 10)
		// count 不能超过 amount（否则 count*1 > amount，State 非法）
		maxCount := amount
		if maxCount > 50 {
			maxCount = 50
		}
		count := int64(rng.IntN(int(maxCount-1)) + 2) // [2, maxCount]
		// maxMin = amount/count >= 1（amount >= count 保证）
		maxMin := amount / count
		minPer := int64(rng.IntN(int(maxMin)) + 1) // [1, maxMin]，确保 count*minPer <= amount
		state := State{RemainAmount: amount, RemainCount: count, MinPerPart: minPer}

		for name, fn := range strategies {
			for r := 0; r < rounds; r++ {
				d, err := New(state, fn, rng)
				if err != nil {
					t.Fatalf("paramSet=%d strategy=%s round=%d New() 错误: %v", i, name, r, err)
				}
				alloc, err := d.Allocate()
				if err != nil {
					t.Fatalf("paramSet=%d strategy=%s round=%d Allocate() 错误: %v", i, name, r, err)
				}
				if int64(len(alloc)) != count {
					t.Errorf("paramSet=%d strategy=%s round=%d len=%d，期望 %d（state=%+v）",
						i, name, r, len(alloc), count, state)
				}
				var sum int64
				for j, v := range alloc {
					if v < minPer {
						t.Errorf("paramSet=%d strategy=%s round=%d alloc[%d]=%d < min=%d（state=%+v）",
							i, name, r, j, v, minPer, state)
					}
					sum += v
				}
				if sum != amount {
					t.Errorf("paramSet=%d strategy=%s round=%d sum=%d，期望 %d（state=%+v）",
						i, name, r, sum, amount, state)
				}
			}
		}
	}
}

// ────────────────────────────────────────────────────────
// 压力测试（-short 跳过）
// ────────────────────────────────────────────────────────

func TestStress_LargeScale(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模压力测试")
	}

	state := State{RemainAmount: 1_000_000, RemainCount: 1000, MinPerPart: 1}
	const stressRounds = 1000

	strategies := map[string]SampleFunc{
		"DoubleMean": DoubleMean(),
		"Uniform":    Uniform(),
		"Fixed":      Fixed(),
	}

	for name, fn := range strategies {
		t.Run(name, func(t *testing.T) {
			result, err := Simulate(state, fn, stressRounds, 42)
			if err != nil {
				t.Fatalf("Simulate() 失败: %v", err)
			}
			if err := result.CheckConservation(); err != nil {
				t.Errorf("守恒性违反: %v", err)
			}
			if err := result.CheckLegality(); err != nil {
				t.Errorf("合法性违反: %v", err)
			}
			if len(result.Positions) != 1000 {
				t.Errorf("完整性违反: len(Positions)=%d，期望 1000", len(result.Positions))
			}
		})
	}
}

// ────────────────────────────────────────────────────────
// 基准测试
// ────────────────────────────────────────────────────────

var benchNSizes = []int{10, 100, 1000}

func BenchmarkAllocate(b *testing.B) {
	strategies := map[string]SampleFunc{
		"Fixed":      Fixed(),
		"Uniform":    Uniform(),
		"DoubleMean": DoubleMean(),
	}
	for stratName, fn := range strategies {
		for _, n := range benchNSizes {
			fn := fn
			n := n
			b.Run(fmt.Sprintf("strategy=%s/n=%d", stratName, n), func(b *testing.B) {
				state := State{
					RemainAmount: int64(n) * 100,
					RemainCount:  int64(n),
					MinPerPart:   1,
				}
				rng := rand.New(rand.NewPCG(42, 0))
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					d, _ := New(state, fn, rng)
					d.Allocate() //nolint
				}
			})
		}
	}
}

func BenchmarkSimulate(b *testing.B) {
	b.Run("rounds=100000", func(b *testing.B) {
		state := State{100, 10, 1}
		fn := DoubleMean()
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Simulate(state, fn, 100_000, 42) //nolint
		}
	})
}
