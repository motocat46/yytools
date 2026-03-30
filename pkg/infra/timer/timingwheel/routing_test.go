package timingwheel

import "testing"

// TestAddInternal_RoutesToCorrectLayer 验证不同 delta 路由到正确层级和槽
func TestAddInternal_RoutesToCorrectLayer(t *testing.T) {
	cases := []struct {
		name      string
		expireAt  int64
		wantLayer int // -2=overflow
		wantSlot  int
	}{
		{"L1 slot=1", 1, 0, 1},
		{"L1 slot=255", l1Interval - 1, 0, int(l1Interval - 1)},
		{"L2 slot=1", l1Interval, 1, 1},     // (256 >> 8) & 63 = 1
		{"L2 slot=2", l1Interval * 2, 1, 2}, // (512 >> 8) & 63 = 2
		{"L3 首槽", l2Interval, 2, 1},         // (16384 >> 14) & 63 = 1
		{"L4 首槽", l3Interval, 3, 1},
		{"L5 首槽", l4Interval, 4, 1},
		{"overflow", l5Interval, -2, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tw := New()
			t.Cleanup(func() { tw.taskQueue.Close(); tw.taskQueue.WaitDone() })
			tw.currentTime = 0

			timer := &Timer{expireAt: tc.expireAt}
			tw.addInternal(timer)

			if tc.wantLayer == -2 {
				if tw.overflow.Len() != 1 {
					t.Errorf("expireAt=%d 应路由到 overflow heap，Len=%d", tc.expireAt, tw.overflow.Len())
				}
				return
			}

			b := tw.wheels[tc.wantLayer].buckets[tc.wantSlot]
			found := false
			for cursor := b.root.next; cursor != &b.root; cursor = cursor.next {
				if cursor == timer {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expireAt=%d 应在 layer=%d slot=%d，但未找到",
					tc.expireAt, tc.wantLayer, tc.wantSlot)
			}
		})
	}
}

// TestAdvanceClock_UpdatesCurrentTime 验证时钟推进后 currentTime 更新
func TestAdvanceClock_UpdatesCurrentTime(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close(); tw.taskQueue.WaitDone() })
	tw.currentTime = 0
	tw.advanceClock(100)
	if tw.currentTime != 100 {
		t.Errorf("advanceClock(100) 后 currentTime=%d，预期 100", tw.currentTime)
	}
	// 时间不退后
	tw.advanceClock(50)
	if tw.currentTime != 100 {
		t.Errorf("advanceClock(50) 不应后退 currentTime，got %d", tw.currentTime)
	}
}

// TestAdvanceClock_PromotesOverflowTimer 验证进入调度窗口的 overflow timer 被提升
func TestAdvanceClock_PromotesOverflowTimer(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close(); tw.taskQueue.WaitDone() })
	tw.currentTime = 0

	// expireAt = l5Interval-1：当 currentTime=1，expireAt < 1+l5Interval → 应被提升
	timer := &Timer{expireAt: l5Interval - 1}
	heapPush(&tw.overflow, timer)

	tw.advanceClock(1)

	if tw.overflow.Len() != 0 {
		t.Error("overflow timer 应被提升，heap 应为空")
	}
}

// TestAdvanceClock_DoesNotPromoteFarTimer 验证未进入调度窗口的 timer 留在 overflow
func TestAdvanceClock_DoesNotPromoteFarTimer(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close(); tw.taskQueue.WaitDone() })
	tw.currentTime = 0

	// expireAt = l5Interval：currentTime=1 时，l5Interval < 1+l5Interval → 会被提升
	// 需要更远的 timer：expireAt = 2*l5Interval
	timer := &Timer{expireAt: 2 * l5Interval}
	heapPush(&tw.overflow, timer)

	tw.advanceClock(1)

	if tw.overflow.Len() != 1 {
		t.Error("2*l5Interval 的 timer 在 currentTime=1 时不应提升")
	}
}

// TestCancel_RemovesFromBucket 验证 Cancel 从 bucket 摘除 timer
func TestCancel_RemovesFromBucket(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close(); tw.taskQueue.WaitDone() })
	tw.currentTime = 0

	timer := &Timer{expireAt: 100}
	tw.add(timer)

	timer.Cancel()

	b := tw.wheels[0].buckets[100]
	for cursor := b.root.next; cursor != &b.root; cursor = cursor.next {
		if cursor == timer {
			t.Error("Cancel 后 timer 应从 bucket 摘除")
		}
	}
	if timer.bucket.Load() != nil {
		t.Error("Cancel 后 timer.bucket 应为 nil")
	}
}

// TestCancel_Idempotent 验证重复 Cancel 不 panic
func TestCancel_Idempotent(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close(); tw.taskQueue.WaitDone() })
	tw.currentTime = 0

	timer := &Timer{expireAt: 100}
	tw.add(timer)

	timer.Cancel()
	timer.Cancel() // 不 panic
}

// TestCancel_OverflowTimer 验证 overflow heap 中的 timer 通过 cancelled 标志取消
func TestCancel_OverflowTimer(t *testing.T) {
	timer := &Timer{expireAt: l5Interval + 100}
	// bucket==nil（overflow 中），直接设 cancelled
	timer.Cancel()
	if !timer.cancelled.Load() {
		t.Error("overflow timer Cancel 后 cancelled 应为 true")
	}
}
