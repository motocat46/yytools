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
