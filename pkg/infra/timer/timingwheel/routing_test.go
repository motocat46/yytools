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
			t.Cleanup(func() { tw.taskQueue.Close() })
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
	t.Cleanup(func() { tw.taskQueue.Close() })
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

// TestAdvanceClock_PromotesOverflowTimer 验证进入调度窗口的 overflow timer 被提升到时间轮 bucket
func TestAdvanceClock_PromotesOverflowTimer(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close() })
	tw.currentTime = 0

	// expireAt = l5Interval-1：当 currentTime=1，expireAt < 1+l5Interval → 应被提升
	timer := &Timer{expireAt: l5Interval - 1}
	heapPush(&tw.overflow, timer)

	tw.advanceClock(1)

	if tw.overflow.Len() != 0 {
		t.Error("overflow timer 应被提升，heap 应为空")
	}
	// 提升后 timer 应路由到时间轮某个 bucket，而非直接入 taskQueue 或被丢弃
	if timer.bucket.Load() == nil {
		t.Error("promoted overflow timer 应在时间轮某个 bucket 中（bucket 指针不应为 nil）")
	}
}

// TestAdvanceClock_DoesNotPromoteFarTimer 验证未进入调度窗口的 timer 留在 overflow
func TestAdvanceClock_DoesNotPromoteFarTimer(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close() })
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
	t.Cleanup(func() { tw.taskQueue.Close() })
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
	t.Cleanup(func() { tw.taskQueue.Close() })
	tw.currentTime = 0

	timer := &Timer{expireAt: 100}
	tw.add(timer)

	timer.Cancel()
	timer.Cancel() // 不 panic
}

// TestCancel_OverflowTimer 验证实际在 overflow heap 中的 timer 被 Cancel 后，
// advanceClock 提升时跳过该 timer（不路由到任何 bucket）。
func TestCancel_OverflowTimer(t *testing.T) {
	tw := New()
	t.Cleanup(func() { tw.taskQueue.Close() })
	tw.currentTime = 0

	// 将 timer 实际插入 overflow heap
	timer := &Timer{expireAt: l5Interval + 100}
	tw.addInternal(timer)
	if tw.overflow.Len() != 1 {
		t.Fatalf("expireAt=l5Interval+100 应路由到 overflow heap，Len=%d", tw.overflow.Len())
	}

	// Cancel：bucket==nil（overflow 中），设置 cancelled 标志
	timer.Cancel()
	if !timer.cancelled.Load() {
		t.Error("overflow timer Cancel 后 cancelled 应为 true")
	}

	// 推进时钟使 timer 进入调度窗口，advanceClock 应弹出 timer 后跳过（因为 cancelled）
	// currentTime=l5Interval+1, expireAt=l5Interval+100 < currentTime+l5Interval → 触发提升逻辑
	tw.advanceClock(l5Interval + 1)

	if tw.overflow.Len() != 0 {
		t.Error("已弹出的 overflow timer 不应留在 heap 中")
	}
	// cancelled timer 提升时应被丢弃（continue 跳过 addInternal），不入任何 bucket
	if timer.bucket.Load() != nil {
		t.Error("cancelled overflow timer 提升后不应入 bucket（应被 advanceClock 丢弃）")
	}
}
