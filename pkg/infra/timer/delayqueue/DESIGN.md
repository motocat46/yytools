# DelayQueue 设计说明

## min-heap 选型

基于 `container/heap`，O(log n) Offer/Poll。DelayQueue 在 TimingWheel 中存储的是 bucket（约几千个），而非全量 timer，规模有限，log n 可接受。

## wakeC 缓冲为 1

`wakeC chan struct{}` 缓冲为 1，`Offer` 加入更早元素时通过 non-blocking send 通知 `Poll` 重新计算等待时间。

- **缓冲为 1（而非 0）**：`Offer` 不阻塞，多次通知可合并（只保留一个信号）。
- **非阻塞 send（`select { default: }`）**：若 `wakeC` 已有信号（上次通知未被消费），丢弃此次通知——`Poll` 醒来后会重新检查堆顶，效果等价。

## nil channel 技巧

队列为空时 `sleepC` 为 nil（未初始化的 `<-chan time.Time`）。Go 的 select 对 nil channel 对应的 case 永久阻塞，实现"队列为空时 Poll 无限等待，直到 Offer 或 ctx 取消"——无需额外分支判断。

## time.NewTimer vs time.After

`Poll` 内每次循环用 `time.NewTimer` 而非 `time.After`：

- `time.After` 在 `wakeC` 或 `ctx.Done()` 先触发时不会立即释放内部 timer 资源
- `time.NewTimer + Stop()` 在非定时器触发路径上显式释放，避免高频循环（TimingWheel reaper 毫秒级）的无效内存分配累积

Go 1.23 已修复 `time.After` 的泄漏问题，但内存分配仍存在；使用 `NewTimer` 在高频场景下减少 GC 压力。

## 并发安全模型

- `Offer`：持 `mu.Lock`，heap.Push 后释放；发 signal 在锁外（non-blocking send 不阻塞）
- `Poll`：持 `mu.Lock` 检查堆顶，计算 sleep 时长后释放；sleep/select 在锁外，不持锁等待
- `TryPoll`：持 `mu.Lock`，非阻塞检查后返回

生产者和消费者通过 `mu.Mutex` 串行化对 heap 的访问。`wakeC` 是 Offer→Poll 的单向通知通道，无需额外锁保护（channel 自身并发安全）。
