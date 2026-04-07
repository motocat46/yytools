package timingwheel

import "testing"

// TestBucket_AddFlush_DeliverAll 验证 Add 后 Flush 把所有 timer 交给 fn
func TestBucket_AddFlush_DeliverAll(t *testing.T) {
	b := newBucket()
	timers := make([]*Timer, 5)
	for i := range timers {
		timers[i] = &Timer{}
		b.Add(timers[i])
	}

	b.expireAt.Store(100)

	var got []*Timer
	b.Flush(func(t *Timer) { got = append(got, t) })

	if len(got) != 5 {
		t.Fatalf("Flush 交付 %d 个 timer，预期 5", len(got))
	}
	if b.expireAt.Load() != -1 {
		t.Errorf("Flush 后 expireAt=%d，预期 -1", b.expireAt.Load())
	}
	for i, timer := range timers {
		if timer.bucket.Load() != nil {
			t.Errorf("timer[%d].bucket 非 nil，Flush 应清零", i)
		}
	}
	if b.root.next != &b.root {
		t.Error("Flush 后链表未清空")
	}
}

// TestBucket_Add_IsFirst 验证第一个 timer 返回 isFirst=true
func TestBucket_Add_IsFirst(t *testing.T) {
	b := newBucket()
	t1, t2 := &Timer{}, &Timer{}

	if isFirst := b.Add(t1); !isFirst {
		t.Error("第一个 timer 加入空 bucket，isFirst 应为 true")
	}
	if isFirst := b.Add(t2); isFirst {
		t.Error("第二个 timer，isFirst 应为 false")
	}
}

// TestBucket_Flush_BucketNilBeforeFn 验证 Flush 在调用 fn 前已清除 timer.bucket
func TestBucket_Flush_BucketNilBeforeFn(t *testing.T) {
	b := newBucket()
	timer := &Timer{}
	b.Add(timer)
	b.expireAt.Store(100)

	var bucketNotNil bool
	b.Flush(func(ti *Timer) {
		if ti.bucket.Load() != nil {
			bucketNotNil = true
		}
	})
	if bucketNotNil {
		t.Error("Flush 调用 fn 时 timer.bucket 应已为 nil（Flush 应在 mu 内先清除 bucket 指针再释放锁调用 fn）")
	}
}

// TestBucket_Flush_Empty 验证空 bucket Flush 不调用 fn，expireAt 重置为 -1
func TestBucket_Flush_Empty(t *testing.T) {
	b := newBucket()
	b.expireAt.Store(100)

	var called bool
	b.Flush(func(_ *Timer) { called = true })

	if called {
		t.Error("空 bucket Flush 不应调用 fn")
	}
	if b.expireAt.Load() != -1 {
		t.Errorf("空 bucket Flush 后 expireAt=%d，预期 -1", b.expireAt.Load())
	}
}
