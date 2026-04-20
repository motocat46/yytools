package slidingwindow_test

import (
	"testing"

	sw "github.com/motocat46/yytools/pkg/ds/slidingwindow"
)

func TestNew_InitialState(t *testing.T) {
	w := sw.New[int](5)
	if w.Len() != 0 {
		t.Errorf("got Len=%d, want 0", w.Len())
	}
	if !w.Empty() {
		t.Error("新窗口应为空")
	}
	if w.Full() {
		t.Error("新窗口不应为满")
	}
}

func TestNew_ZeroSizePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("size=0 应 panic")
		}
	}()
	sw.New[int](0)
}

func TestAdd_Sum_UnderCapacity(t *testing.T) {
	w := sw.New[int](3)
	w.Add(10)
	w.Add(20)
	if w.Len() != 2 {
		t.Errorf("got Len=%d, want 2", w.Len())
	}
	if w.Sum() != 30 {
		t.Errorf("got Sum=%d, want 30", w.Sum())
	}
}

func TestAdd_Sum_ExactCapacity(t *testing.T) {
	w := sw.New[int](3)
	w.Add(1)
	w.Add(2)
	w.Add(3)
	if !w.Full() {
		t.Error("should be full")
	}
	if w.Sum() != 6 {
		t.Errorf("got Sum=%d, want 6", w.Sum())
	}
}

func TestAdd_Sum_OverCapacity_Evicts(t *testing.T) {
	w := sw.New[int](3)
	w.Add(1)
	w.Add(2)
	w.Add(3)
	w.Add(4) // evicts 1，窗口变为 [2,3,4]
	if w.Len() != 3 {
		t.Errorf("got Len=%d, want 3", w.Len())
	}
	if w.Sum() != 9 {
		t.Errorf("got Sum=%d, want 9 (2+3+4)", w.Sum())
	}
}

func TestAvg_Basic(t *testing.T) {
	w := sw.New[int](4)
	w.Add(10)
	w.Add(20)
	w.Add(30)
	got := w.Avg()
	want := 20.0
	if got != want {
		t.Errorf("got Avg=%v, want %v", got, want)
	}
}

func TestAvg_AfterEviction(t *testing.T) {
	w := sw.New[int](2)
	w.Add(10)
	w.Add(20)
	w.Add(30) // 窗口=[20,30]
	got := w.Avg()
	want := 25.0
	if got != want {
		t.Errorf("got Avg=%v, want %v", got, want)
	}
}

func TestAvg_EmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("空窗口调用 Avg() 应 panic")
		}
	}()
	w := sw.New[float64](3)
	w.Avg()
}

func TestMax_Basic(t *testing.T) {
	w := sw.New[int](3)
	w.Add(3)
	w.Add(1)
	w.Add(4)
	if w.Max() != 4 {
		t.Errorf("got Max=%d, want 4", w.Max())
	}
}

func TestMax_AfterEviction(t *testing.T) {
	// 最大值被淘汰后，Max() 应返回窗口内新的最大值
	w := sw.New[int](3)
	w.Add(10)
	w.Add(3)
	w.Add(5) // 窗口=[10,3,5]，Max=10
	w.Add(7) // 窗口=[3,5,7]，10 被淘汰，Max=7
	if w.Max() != 7 {
		t.Errorf("got Max=%d, want 7", w.Max())
	}
}

func TestMax_EmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("空窗口调用 Max() 应 panic")
		}
	}()
	sw.New[int](3).Max()
}

func TestMin_Basic(t *testing.T) {
	w := sw.New[int](3)
	w.Add(3)
	w.Add(1)
	w.Add(4)
	if w.Min() != 1 {
		t.Errorf("got Min=%d, want 1", w.Min())
	}
}

func TestMin_AfterEviction(t *testing.T) {
	// 最小值被淘汰后，Min() 应返回窗口内新的最小值
	w := sw.New[int](3)
	w.Add(1)
	w.Add(5)
	w.Add(8) // 窗口=[1,5,8]，Min=1
	w.Add(3) // 窗口=[5,8,3]，1 被淘汰，Min=3
	if w.Min() != 3 {
		t.Errorf("got Min=%d, want 3", w.Min())
	}
}

func TestMin_EmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("空窗口调用 Min() 应 panic")
		}
	}()
	sw.New[int](3).Min()
}
