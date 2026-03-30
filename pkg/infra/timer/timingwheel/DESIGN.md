# TimingWheel 设计说明

## 层级参数选取

5 层固定结构：L1=256 slots（8 bit），L2-L5=64 slots（6 bit）。

| 层 | Shift | 粒度 | 覆盖范围 |
|----|-------|------|---------|
| L1 | 0 | 1 ms | 0 ~ 255 ms |
| L2 | 8 | 256 ms | 256 ms ~ ~16 s |
| L3 | 14 | ~16 s | ~16 s ~ ~17 min |
| L4 | 20 | ~17 min | ~17 min ~ ~18 h |
| L5 | 26 | ~18 h | ~18 h ~ ~50 day |

L1 选 256 slots（8 bit）而非 64：1ms tick 下 256 格覆盖 256ms 窗口，减少跨层降级频率，降低 Flush 路径的开销。L2-L5 选 64 slots 平衡内存与覆盖范围。

## overflow heap

超出 L5 覆盖范围（~50 天）的 timer 入 overflow min-heap（`timerHeap`）。`advanceClock` 推进时钟时批量将进入调度窗口的 overflow timer 提升到时间轮。

**边界条件（严格 `<` 而非 `<=`）：**

```go
heapPeek.expireAt < currentTime + l5Interval
```

若用 `<=`，`delta == l5Interval` 的 timer 提升后路由回 `addInternal`，再次落入 overflow，造成死循环。

## 两层锁设计

| 锁 | 持有者 | 保护目标 |
|----|--------|---------|
| `mu sync.RWMutex`（写锁） | `advanceClock`（由 reaper goroutine 持有） | 排他访问时钟和时间轮结构 |
| `mu sync.RWMutex`（读锁） | `add()`（多 goroutine 并发） | 允许并发插入，与推进互斥 |
| `overflowMu sync.Mutex` | `add()`（持 RLock 的路径） | 串行化多个并发 add() 对 overflow heap 的写入 |
| `bucket.mu sync.Mutex` | bucket 的 Add/Flush/Cancel | 保护 bucket 内双向链表的并发操作 |

`advanceClock` 持 writeLock，已独占访问所有共享状态，不需要再加 `overflowMu`。

## Cancel O(1) 设计

`Timer` 持有自身所在 `bucket` 的 `atomic.Pointer[bucket]`。`Cancel` 直接拿目标 bucket 锁，O(1) 从链表摘除，无需遍历任何层级。

`Flush` 清零 bucket 指针时持 bucket 锁，`Cancel` 通过 double-check 处理窗口期竞态；`cancelled atomic.Bool` 作为 bucket==nil（overflow 中或 Flush 窗口期）时的备份标志。

## taskQueue 和 taskExecutor

`taskExecutor` 是单 goroutine，回调串行执行。此设计使回调天然无需额外同步，但要求回调必须非阻塞（通过 `GoAsync` 包装阻塞操作）。

`taskQueue`（`UnboundedChannel`）携带 `*Timer` 而非 `func()`，`taskExecutor` 据此判断是否需要重注册 repeating timer。

## Stop() 排水时序

```
Stop()
  → cancel()            通知 reaper 退出
  → wg.Wait()           等待 reaper + taskExecutor 均退出
      reaper defer:
        → taskQueue.Close()     标记关闭，signal worker
      worker goroutine:
        → close(channel)        排水语义：channel 关闭
        → close(workerDone)     通知 WaitDone
      taskExecutor:
        → range Out() 结束      channel 关闭后退出
        → wg.Done()
  → taskQueue.WaitDone()  等待 worker goroutine 完全退出（消除 goleak）
```

`wg.Wait()` 在 `taskExecutor.wg.Done()` 后返回，此时 worker 已 `close(channel)` 但可能尚未执行 `close(workerDone)`（goroutine 尚未完全退出）。`taskQueue.WaitDone()` 确保 worker goroutine 彻底退出，避免 goroutine 泄漏检测误报。

## Cancel 的 best-effort 语义

与 `time.Timer.Stop()` 一致：

- timer 在 bucket 中时 Cancel 可阻止执行（通过链表摘除）
- timer 已通过 `taskQueue.Send` 投入队列后，`taskExecutor` 对 one-shot timer 不检查 `cancelled` 标志，回调仍会执行一次

这是有意为之的权衡：检查 `cancelled` 需要额外原子读且语义模糊（回调可能已执行到一半）。使用者需要通过 `GoAsync` + 业务层幂等性处理此竞态窗口。
