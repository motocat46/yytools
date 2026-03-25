# DESIGN.md — workerpool 设计决策记录

面向维护者。记录关键设计决策及背后的理由，避免日后"为什么这样写"的重复推导。

---

## 1. Submit 接收 ctx 而非直接阻塞

**问题**：队列满时，Submit 必须等待。如果直接阻塞（无超时），调用方无法控制等待上界，程序在高负载下可能卡死。

**决策**：Submit 接受 `context.Context`，通过 `select` 同时监听队列空位和 `ctx.Done()`。

```go
select {
case p.queue <- task:
    return nil
case <-ctx.Done():
    p.wg.Done()
    return ctx.Err()
}
```

**好处**：
- 调用方可以用 `context.WithTimeout` 限制单次提交的最大等待时间。
- 调用方可以用 `context.WithCancel` 在外部事件（如收到 SIGTERM）时批量取消排队中的 Submit。
- 不强制调用方使用 timeout，传 `context.Background()` 即退化为原来的阻塞语义，灵活性最高。

---

## 2. 有界 queue 的背压设计——为何不用 unbounded_channel

`queue` 使用 `make(chan func(), queueSize)`，而非 `unbounded_channel`。

**理由**：

- **背压是核心目的**：WorkerPool 的主要价值之一是防止无界积压。如果使用无界队列，生产者可以无限提交，堆内存无上限增长，pool 的"限流"语义名存实亡。
- **有界 channel 天然提供背压**：队列满时 Submit 阻塞，自动形成反压，迫使生产者减速，这正是想要的行为。
- **unbounded_channel 的适用场景不同**：`unbounded_channel` 解决的是"生产者和消费者速率短期不匹配但长期均衡"的场景，不适合需要强限流的 WorkerPool。

`queueSize=0` 是合法值，此时队列无缓冲，每次 Submit 都直接等待 worker 空闲，背压最强。

---

## 3. Pipeline 作薄封装的理由

Pipeline 仅封装"从 channel 读取 → Submit → 收集结果 → 关闭输出 channel"这个固定模式，不重新实现调度逻辑。

**理由**：
- 避免重复实现 WorkerPool 已经处理好的并发安全、WaitGroup 管理、graceful shutdown。
- 职责单一：WorkerPool 管调度，Pipeline 管数据流。
- 保持代码量最小，减少 bug 面。

Pipeline 内部直接持有 `*WorkerPool` 而非接口，因为目前没有替换实现的需求，过早抽象增加维护成本。

---

## 4. Wait 和 Close 分离的原因

两个方法职责不同：

| 方法 | 职责 |
|------|------|
| `Wait()` | 等待已提交任务全部完成，worker 继续运行，可以再提交新任务 |
| `Close()` | 关闭 pool，等待任务完成，然后退出所有 worker goroutine |

**为什么需要 Wait**：在某些场景下，需要等待一批任务完成（如分批处理、阶段性同步），然后继续提交下一批，而不是关闭整个 pool。`Wait` 提供了这种"批次屏障"能力。

**为什么 Close 内部也调用 Wait**：Close 的语义是"完整关闭"，必须保证所有已提交任务都完成后才退出 worker，否则可能丢任务。

---

## 5. Submit 与 Close 并发安全：mu 保护 check+wg.Add 的原子性（核心决策）

这是整个实现中最容易出错的地方，也是最关键的设计决策。

### 问题描述

`Submit` 和 `Close` 并发时存在以下竞态窗口：

```
Submit                          Close
------                          -----
1. check closed == false
                                2. set closed = true
                                3. wg.Wait()  ← 此时 wg 计数为 0，立即返回
4. wg.Add(1)
   // Close 已经返回，wg.Add(1) 在 Close 之后执行
   // worker 已退出（close(stop)），任务永远不会被执行
```

后果：`wg.Add(1)` 在 `wg.Wait()` 返回之后执行，任务被放入队列但 worker 已退出，任务丢失。更危险的是，`sync.WaitGroup` 文档明确规定：**在 Wait 返回后调用 Add 是未定义行为**（可能 panic 或数据竞争）。

### 解决方案：mu 保护 check+Add 的原子性

```go
// Submit
p.mu.Lock()
if p.closed {
    p.mu.Unlock()
    return ErrPoolClosed
}
p.wg.Add(1)   // check 和 Add 在同一把锁内，不可分割
p.mu.Unlock()

// Close
p.once.Do(func() {
    p.mu.Lock()
    p.closed = true   // 持锁设置，与 Submit 的 check+Add 互斥
    p.mu.Unlock()
    p.wg.Wait()       // 此后不会再有新的 wg.Add(1)
    close(p.stop)
})
```

**关键不变量**：`p.closed = true` 之后，不会再有任何 `wg.Add(1)` 执行。因为：
- Close 持锁设置 `closed = true`。
- Submit 持同一把锁检查 `closed`，若为 true 直接返回错误。
- 两者互斥，check-then-Add 是原子操作。

`sync.Once` 保证 Close 的 body 只执行一次，使 Close 幂等。

---

## 6. submitLocker 接口：默认 RWMutex，保留 Mutex 版用于对比

**背景**：Submit 与 Close 对锁的需求不对称——Submit 只需保证"check+Add 原子"，多个并发 Submit 之间没有互斥需求；Close 需要独占，防止新的 Add 进入。

**决策**：将锁行为抽象为 `submitLocker` 接口，Submit 调用 `lockSubmit/unlockSubmit`，Close 调用 `lockClose/unlockClose`。两种实现：

| 实现 | Submit 锁 | Close 锁 |
|------|-----------|---------|
| `mutexLocker` | 排他锁 | 排他锁 |
| `rwMutexLocker` | 读锁（并发不互斥）| 写锁（独占）|

**benchmark 结果**（Apple M4，`-benchtime=3s -count=3`）：

| 场景 | Mutex | RWMutex | 差异 |
|------|-------|---------|------|
| 顺序基线 | ~163 ns | ~159 ns | 持平 |
| 并发 p=1 | ~151 ns | ~106 ns | RW 快 ~30% |
| 并发 p=16 | ~134 ns | ~107 ns | RW 快 ~20% |
| 并发 p=64 | ~136 ns | ~108 ns | RW 快 ~20% |

**结论**：RWMutex 在并发场景下稳定提升约 20%，顺序场景持平，改动低风险，默认切换为 RWMutex。Mutex 版以 `NewWorkerPoolMutex` 保留，供性能回归对比使用。

---

## 7. 为什么用 stop channel 而非 close(queue)

**close(queue) 的问题**：多个 worker 并发执行 `close(queue)` 或向已关闭的 channel 发送数据都会 panic。即便只由 Close 关闭，`Submit` 在 Close 之后若还持有任务并执行 `p.queue <- task` 也会 panic。

**stop channel 的优势**：

```go
// worker 主循环
select {
case task := <-p.queue:
    // 执行任务
case <-p.stop:
    return  // 收到关闭信号，退出
}
```

- `stop` 由 `make(chan struct{})` 创建，只有 `Close` 调用 `close(p.stop)` 一次（`sync.Once` 保证）。
- `close(stop)` 是广播：所有阻塞在 `<-p.stop` 的 worker 同时收到信号，无需逐个通知。
- `queue` 只用于正向数据流，永远不被关闭，避免向已关闭 channel 发送数据的 panic 风险。

**为什么 Close 在 close(stop) 之前先 wg.Wait()**：确保所有任务执行完毕后才退出 worker。若先 close(stop)，worker 可能在任务还在队列中时就退出，导致任务丢失。
