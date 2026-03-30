package timingwheel

import "testing"

// TestConstants_LayerCoverage 验证常量层级覆盖范围正确
func TestConstants_LayerCoverage(t *testing.T) {
	if l1Interval != 256 {
		t.Errorf("l1Interval=%d, want 256", l1Interval)
	}
	if l2Interval != 16384 {
		t.Errorf("l2Interval=%d, want 16384", l2Interval)
	}
	wantL5 := int64(256) * 64 * 64 * 64 * 64
	if l5Interval != wantL5 {
		t.Errorf("l5Interval=%d, want %d", l5Interval, wantL5)
	}
	if l5Shift != l1Bits+l2Bits*3 {
		t.Errorf("l5Shift=%d, want %d", l5Shift, l1Bits+l2Bits*3)
	}
}

// TestBucketInitialState 验证 newBucket 初始状态
func TestBucketInitialState(t *testing.T) {
	b := newBucket()
	if b.expireAt.Load() != -1 {
		t.Errorf("newBucket expireAt=%d, want -1", b.expireAt.Load())
	}
	if b.root.next != &b.root || b.root.prev != &b.root {
		t.Error("newBucket 双向链表未正确初始化为循环空链表")
	}
}
