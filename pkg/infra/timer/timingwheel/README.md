# timingwheel

分层时间轮定时器，O(1) add/cancel，毫秒精度，覆盖约 50 天，适用于百万级并发定时器场景（技能冷却、心跳超时、断线保护等）。

## 快速上手

```go
tw := timingwheel.New()
tw.Start()
defer tw.Stop()

// one-shot：1 秒后执行
timer, err := tw.AfterFunc(time.Second, func() {
    fmt.Println("触发！")
})
if err != nil { /* MaxTimeout 超限 */ }

// repeating：每 5 秒触发一次（fixed-delay）
heartbeat, _ := tw.EveryFunc(5*time.Second, sendHeartbeat)

// 取消
timer.Cancel()
heartbeat.Cancel()

// 异步化阻塞回调，不阻塞定时器引擎
tw.AfterFunc(time.Second, timingwheel.GoAsync(heavyWork))
```

## API

```go
func New(opts ...Option) *TimingWheel
func (tw *TimingWheel) Start()
func (tw *TimingWheel) Stop()
func (tw *TimingWheel) AfterFunc(d time.Duration, f func()) (*Timer, error)
func (tw *TimingWheel) EveryFunc(d time.Duration, f func()) (*Timer, error)
func (t *Timer) Cancel()
func GoAsync(f func()) func()

func WithMaxTimeout(d time.Duration) Option
```

| 操作 | 复杂度 |
|------|--------|
| AfterFunc / EveryFunc | O(1) |
| Cancel | O(1) |

## 注意事项

- **回调必须非阻塞**：taskExecutor 是单 goroutine，阻塞回调会延迟所有定时器。使用 `GoAsync` 异步化。
- **Cancel 的 best-effort 语义**：与 `time.Timer.Stop()` 一致，Cancel 与"刚好到期"存在竞态窗口，Cancel 返回后回调仍可能执行一次。
- **repeating 使用 fixed-delay**：下次触发时间 = 实际执行时刻 + interval，防止 GC pause 后连锁补发。
- 使用完毕必须调用 `Stop()`，否则内部 goroutine 泄漏。
