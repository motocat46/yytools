# delayqueue

并发安全的延迟队列，基于 min-heap，元素按到期时间升序出队。

## 适用场景

定时触发场景：持有一批带"到期时间"的元素，阻塞等待最近一个到期，到期即消费。
TimingWheel 用此模块驱动时钟推进（零空转）。

## 快速上手

```go
// clock 提供当前时间（ms），与 Item.ExpireAt() 同一时间域
start := time.Now()
clock := func() int64 { return time.Since(start).Milliseconds() }

q := delayqueue.New[*MyItem](clock)

// 生产者（并发安全）
q.Offer(&MyItem{expireAt: clock() + 1000}) // 1秒后到期

// 消费者（阻塞，直到有元素到期或 ctx 取消）
item, ok := q.Poll(ctx)

// 非阻塞尝试
item, ok = q.TryPoll()
```

## API

```go
type Item interface { ExpireAt() int64 }

func New[T Item](clock func() int64) *DelayQueue[T]
func (q *DelayQueue[T]) Offer(item T)
func (q *DelayQueue[T]) Poll(ctx context.Context) (T, bool)
func (q *DelayQueue[T]) TryPoll() (T, bool)
```

| 方法 | 复杂度 | 说明 |
|------|--------|------|
| Offer | O(log n) | 并发安全；加入更早元素时唤醒 Poll |
| Poll | O(log n) | 阻塞；ctx 取消返回 false |
| TryPoll | O(log n) | 非阻塞；未到期返回 (zero, false) |
