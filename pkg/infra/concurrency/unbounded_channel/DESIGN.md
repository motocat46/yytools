# UnboundedChannel 设计复盘

> 本文记录从 V1 到 V6 的完整决策路径，方便日后回顾每一个设计取舍及其背后的原因。

---

## 问题背景

Go 原生 channel 是有界的：当 channel 满时，发送者阻塞。在生产者突发流量场景下，这会导致生产者堆积、系统延迟抖动。目标是设计一个：

1. **FIFO 顺序**：消息严格按发送顺序被消费
2. **无界缓冲**：channel 满时不阻塞生产者，消息暂存 buffer
3. **背压**：buffer 超过上限时生产者阻塞，避免内存无限增长
4. **高性能**：快速路径接近 native channel

---

## 核心架构：双层队列 + Worker

```
Send()
  ├─ 快速路径：buffer 为空 && channel 未满 → 直接投入 channel（无锁 select）
  └─ 慢路径：sendSlow()
       ├─ channel 满：enqueue buffer，return
       └─ channel 未满但 buffer 非空：enqueue buffer + inline transfer

Worker goroutine（后台搬运）
  └─ buffer → channel（事件驱动，ticker 兜底）

Receive() / Out()
  └─ 直接从 channel 消费
```

**FIFO 不变式**：若 buffer 非空，一切新消息必须先进 buffer 排队，再由 worker 按序搬入 channel。绝不允许新消息绕过 buffer 直接投入 channel。

---

## V1 → V6 完整演进路径

### V1（`archive/ucv1.go`）—— 原型，不保证 FIFO

**目标**：验证"双层队列"基本可行。

**设计**：
- 使用 `container/list` 作为 buffer
- `Send()`：channel 满 → `PushFront`（头插）暂存，否则直接投入
- `Receive()`：list 非空 → `list.Back()`（尾取）返回，否则从 channel 取
- **无后台 worker**：buffer 中的消息只有 `Receive()` 主动取才会被消费，永远不会搬入 channel

**遗留问题**：
1. **FIFO 未保证**：注释明确标注"不保证消息顺序"，buffer 走 LIFO（头插尾取）
2. **buffer 孤岛**：无 worker 搬运，buffer 中的消息不会流向 channel，消费者若只用 channel 会丢消息
3. **类型硬编码**：`chan int`，不支持泛型
4. **无 Close 安全**：直接 `close(channel)`，无等待 buffer 排空逻辑

---

### V2（`archive/ucv2.go`）—— 首次实现 FIFO + 后台 worker

**目标**：修复 V1 的两个核心问题：FIFO 顺序 + buffer 孤岛。

**关键改进**：
1. **FIFO 修复**：`Send()` 仍用 `PushFront`，但 `transfer()` 从 `list.Back()`（最老的消息）开始搬运，实现先进先出
2. **后台 worker（`listCheck`）**：独立 goroutine 定期搬运 buffer → channel，buffer 不再孤立
3. **Close 增加排空等待**：`listCheck` 检测 `closed && buffer空 && channel空` 后才真正 `close(channel)`

**遗留问题**：
1. **固定 10ms 轮询**：`time.Sleep(10ms)` 空转，高负载时延迟最高 10ms；低负载时 CPU 浪费
2. **测试代码混入生产代码**：`seq` 字段和 `uc.seq++` 强制覆盖 msg，不应出现在正式实现中
3. **`container/list`**：链表节点堆分配，GC 压力大，缓存不友好
4. **Close 信号缺失**：`Close()` 只设标志，不唤醒 worker；若 worker 恰好在 Sleep，最多延迟 10ms 才退出

---

### V3（`archive/ucv3.go`）—— 信号驱动替换固定轮询

**目标**：消除 10ms 固定延迟，改为事件驱动。

**关键改进**：
1. **`sync.Cond` 替换 Sleep**：`listCheck` 阻塞在 `cond.Wait()`；`Send()` 入 buffer 后 `cond.Signal()`；`Receive()` 消费后 `cond.Signal()`
2. **环形队列替换链表**：`queue.IQueue`（ring buffer），预分配内存，零 GC 分配，缓存友好

**遗留问题**：
1. **`bufferLen` 定义了但未维护**：字段存在，但 `Send()`/`transfer()` 均未调用 `bufferLen.Add()`，快速路径无法使用，等于废字段
2. **worker 等待条件有漏洞**：`for buffer.Len()==0 || channel满`，若 transfer 执行到一半 channel 填满，worker 阻塞等新信号，而此时 buffer 仍有积压却无法继续搬运
3. **测试代码仍混入生产代码**：`seq` 字段残留
4. **Close 竞态**：`close(done)` 和 `close(channel)` 在 worker 内持锁执行，但 `Receive()` 用 `select` 同时监听两者，存在双关闭竞态

---

### V4（`archive/ucv4.go`）—— 快速路径 + 背压 + 泛化

**目标**：性能优化（快速路径）+ 内存安全（背压）+ 类型扩展（`any`）。

**关键改进**：
1. **快速路径**：`bufferLen.Load()==0 && chanLen.Load()<cap` → 直接 `select case channel<-msg`，绕过锁
2. **背压**：`bufferLen > 1e6` 时生产者阻塞在 `condNotFull.Wait()`
3. **比例唤醒**：`transfer()` 按搬运量占 limit 百分比唤醒阻塞生产者，避免雪崩
4. **类型升级**：`chan int` → `chan any`，支持任意消息类型

**遗留问题（严重）**：
1. **`chanLen` 与 channel 实际长度漂移**：维护独立 `atomic.Int32 chanLen` 与 Go runtime 的 `len(channel)` 并行，在并发下极易不一致，是快速路径判断错误的根源
2. **死锁风险**（注释原文："这个基于信号量的版本还有死锁的风险"）：`cond` 和 `condNotFull` 共用同一把 `mutex`，`Receive()` 在持锁情况下发 `cond.Signal()`，与 worker 持同一锁的代码路径存在循环等待可能
3. **`Receive()` 获锁发信号**：`uc.mutex.Lock(); uc.cond.Signal(); uc.mutex.Unlock()` — 不必要的加锁，`sync.Cond.Signal()` 无需持锁调用
4. **Close 不唤醒 worker**：`Close()` 仅 `closed.Store(true)`，若 worker 阻塞在 `cond.Wait()`，须等外部 signal 才会检测关闭，可能永久泄漏

---

### V5（`archive/ucv5_original.go`）—— 彻底修复快速路径代理

**目标**：清除 V4 的 `chanLen` 漂移问题，建立正确的快速路径代理。

**关键改进**：
1. **废弃 `chanLen`**：直接用 `len(uc.channel)` 读 Go runtime 维护的 channel 内部计数，消除漂移
2. **`bufferLen` 正式成为快速路径代理**：在 `sendSlow()`/`transfer()` 中完整维护 `bufferLen.Add(±1)`
3. **背压机制完善**：`condNotFull` 更名为 `condSendWaiter`，语义更清晰；`waiters` 计数跳过无等待者时的无效 Signal
4. **`listCheck` 改为自旋 + 指数退避**：先自旋 10 次（`mutex+Gosched`）探测，失败后指数退避（1ms→2ms→…→100ms）

**遗留问题**：
1. **退避积累导致延迟尖刺**：channel 满 → 搬运失败 → 退避增大 → buffer 积压加剧（正反馈）
2. **空闲时自旋浪费 CPU**：自旋阶段无意义地持锁/释锁 + Gosched
3. **类型不安全**：仍用 `any`，无编译期类型检查

---

### V51（`archive/ucv51.go`）—— 自旋策略变种对照

**目标**：量化不同自旋策略的吞吐差异。

**变化**：与 V5 快慢路径、背压设计完全相同，仅 `listCheck` 改为**自旋 10 次 + 指数退避（1ms→…→100ms）**（与 V52 无自旋+原子预检+退避 对照）。

**结论**：退避上限 100ms 在高负载时仍会触发 10ms 级延迟尖刺，退避积累的根因未解决。

---

### V53（`archive/ucv53.go`）—— 去掉指数退避

**目标**：消除退避积累引起的延迟尖刺。

**变化**：在 V51 基础上去掉指数退避，改为**固定 1ms sleep**。

**改进**：消除"搬运失败→退避增大→积压加剧"的正反馈。
**残留问题**：1ms 定时唤醒本质上仍是轮询，空闲时无效唤醒。

---

### V6（`unbounded_channel_v6.go`）—— 事件驱动 + 泛型 + 关闭安全

**目标**：从根本上解决轮询问题；同时引入泛型和可靠的并发安全关闭。

**核心改变**：

| | V5 系列 | V6 |
|---|---|---|
| worker 触发 | 定时轮询（自旋/sleep） | 事件信号 + 1ms ticker 兜底 |
| 空闲 CPU | 有（自旋/轮询） | 零（阻塞在 select） |
| 延迟 | 退避可能导致尖刺 | 事件驱动，精准唤醒 |
| 关闭安全 | 无 activeSenders 保护 | `activeSenders` 计数防 panic |
| 类型安全 | `any` | `[T any]` 泛型 |

**V6 三个信号源**：
1. 消费者 `Receive()` 后若 buffer 非空 → 主动 `signal()`
2. `Close()` 调用时 → `signal()`，确保 worker 及时感知关闭
3. 1ms ticker 兜底（覆盖使用 `Out()` 直接消费的场景）

> 注意：`sendSlow` 本身**不发** signal。channel 满时直接 return，依靠消费者 `Receive()` 触发信号；buffer 非空时调用 `transfer()` 内联搬运，同样不发 signal。

---

## 关键设计决策

### 决策 1：`bufferLen` 原子计数作为快速路径代理（V3 引入意图，V5 正式落地）

**问题**：快速路径需要在不加锁的情况下判断"buffer 是否为空"。直接读 `buffer.Len()` 需要加锁。

**决策**：维护 `atomic.Int32 bufferLen` 作为 buffer 长度的**高估代理**：
- `bufferEnqueue`：**先** `bufferLen.Add(1)`，**再** `buffer.Enqueue()`
- `bufferDequeue`：**先** `buffer.Dequeue()`，**再** `bufferLen.Add(-1)`

**不变式**：`bufferLen >= buffer.Len()`（只高估，不低估）

**理由**：高估的代价是偶尔多走慢路径（性能损失可忽略）；低估的代价是误判"buffer 为空"绕过 buffer，**破坏 FIFO**，不可接受。

**V4 的错误**：维护独立 `chanLen` 替代 `len(channel)`，在并发下漂移，快速路径判断失效。V5 废弃 `chanLen`，直接用 Go runtime 维护的 `len(channel)`。

---

### 决策 2：`sendSlow()` 中 channel 满时直接 return（V6 bug 修复）

```go
if len(uc.channel) == cap(uc.channel) {
    uc.bufferEnqueue(msg)
    return   // ← 关键：不能 fallthrough
}
```

**背景**：此处曾因代码修改丢失 `return`，产生 **double-enqueue bug**：
- 第一个 `if` 进入 buffer 后，buffer 非空
- 第二个 `if (buffer.Len() != 0)` 条件成立，再次 enqueue 同一条消息
- 消费者收到重复消息，Porcupine 检测到 FIFO 违反

**决策**：channel 满时 enqueue buffer 后**立即 return**，不执行 transfer（channel 已满，transfer 无意义）。消费者 `Receive()` 后会主动 signal，触发 worker 搬运。

**额外收益**：更早 return 减少临界区持锁时间，性能略优。

---

### 决策 3：背压唤醒——按比例唤醒，非全量广播（V4 引入，V6 沿用）

```go
percent := float32(movedCount) / float32(uc.limit) * 100
if percent > 50 {
    uc.condSendWaiter.Broadcast()
    return
}
wakenCount := max(1, int(percent))
for i := 0; i < wakenCount; i++ {
    uc.condSendWaiter.Signal()
}
```

**问题**：buffer 搬运后若唤醒全部阻塞生产者，它们会同时发送，channel 可能立即再次堆满，造成"雪崩唤醒"。

**决策**：搬运量占 limit 百分比决定唤醒数量。搬运量少 → 少唤醒；搬运量多（>50%）→ 广播。

**边界处理**：`buffer.Len() == 0`（排空）时强制 Broadcast，否则不再有任何唤醒源，产生死锁。

---

### 决策 4：`activeSenders` 计数防止 close-during-send race（V6 新增）

**问题**：`Close()` 时若直接 close channel，而并发 `Send()` 仍在执行快速路径的 `select case channel <- msg`，会 panic（向已关闭 channel 发送）。V4/V5 均无此保护。

**决策**：
- `Send()` 第一步 `activeSenders.Add(1)`（持有"读锁"）
- worker 关闭 channel 的条件：`closed && activeSenders==0 && buffer空 && channel空`
- 通过 happens-before 链保证：worker 见到 `activeSenders==0` 时，所有 `Send()` 已完成

---

### 决策 5：`notify` channel 容量为 1（信号合并）

```go
notify: make(chan struct{}, 1)

func (uc *UnboundedChannelV6[T]) signal() {
    select {
    case uc.notify <- struct{}{}:
    default:  // 已有信号，合并，不阻塞
    }
}
```

**理由**：高并发下多个生产者可能同时触发 signal，若 notify 无缓冲则 signal() 可能丢失。容量 1 保证：不管多少个 signal() 并发调用，worker 最多积压 1 个信号，多余的自动丢弃（幂等）。worker 仍会处理所有积压数据（循环处理），无正确性问题。

---

## 各版本解决问题对照表

| 版本 | 解决的核心问题 | 引入的新问题 |
|------|--------------|------------|
| V1 | 双层队列基本结构可行 | FIFO 未保证；buffer 孤岛；无 worker |
| V2 | FIFO 正确；后台 worker 排空 buffer | 10ms 固定轮询高延迟；链表 GC 压力 |
| V3 | 信号驱动替换轮询；环形队列替换链表 | `bufferLen` 定义未维护；worker 等待条件漏洞 |
| V4 | 快速路径；背压；比例唤醒；`any` 类型 | `chanLen` 漂移；死锁风险；Close 不唤醒 worker |
| V5 | 废弃 `chanLen`；`bufferLen` 正确维护 | 自旋 + 退避仍有延迟尖刺；无泛型；无关闭安全 |
| V51/V53 | 退避策略对照实验 | 本质仍是轮询 |
| **V6** | 事件驱动零轮询；泛型；`activeSenders` 关闭安全；double-enqueue bug 修复 | — |

---

## 并发安全证明（V6 核心路径）

### FIFO 保证

1. **快速路径前置条件**：`bufferLen == 0`（buffer 为空），因此直接投入 channel 不破坏顺序
2. **慢路径**：在 mutex 保护下，channel 满 → 入 buffer；channel 未满但 buffer 非空 → 入 buffer + transfer（按序搬运）
3. **不变式**：buffer 非空时，任何新消息必须先入 buffer，再由 transfer 按 FIFO 搬入 channel

### Close 安全

```
Send():  activeSenders.Add(1) → closed.Load() → ... → activeSenders.Add(-1)
Close(): closed.Store(true) → signal() → Broadcast()
worker:  canClose() = closed && activeSenders==0 && buffer空 && channel空 → close(channel)
```

- 若 `Send()` 在 `closed.Store(true)` 之前完成 `activeSenders.Add(1)`：worker 等待 activeSenders 归零后才关闭
- 若 `Send()` 在 `closed.Store(true)` 之后执行 `closed.Load()`：看到 true，直接返回 false

---

## 性能基准（Apple M4，arm64）

| 场景 | NativeChan | V6 | 差距 |
|------|-----------|-----|------|
| Large chanSize（buffer 几乎不触发）| 83.89 ns/op | 88.50 ns/op | +5.5% |
| Small chanSize（buffer 频繁触发）| 85.28 ns/op | 140.6 ns/op | +64.9% |
| 10k chanSize | 82.62 ns/op | 82.31 ns/op | ≈持平 |
| 吞吐量（100k 消息）| 28.2M msg/s | 20.2M msg/s | -28.3% |

**结论**：
- buffer 路径不触发时，V6 与 native channel 几乎持平
- buffer 频繁触发时，V6 慢 ~65%（worker goroutine 调度 + sync.Mutex 开销）
- V6 的价值在于提供 native channel **没有的能力**：无界缓冲 + 背压控制

---

## 正确性验证（Porcupine 线性一致性）

使用 [Porcupine](https://github.com/anishathalye/porcupine) 对 V6 FIFO 语义进行形式化验证：

| 测试 | 场景 | 操作数 | 结果 |
|------|------|--------|------|
| SingleProducer | 1P×20msg, 1C | 40 ops | PASS |
| MultiProducer | 2P×8msg, 2C | 32 ops | PASS |
| HighBackpressure | 5P×6msg, 2C, limit=10 | 60 ops | PASS |
| Stress | 100轮 × 2P×6msg, 2C | 24 ops/轮 | PASS (100/100) |

**Porcupine 复杂度说明**：Porcupine 是 NP-hard 算法，复杂度指数级于**并发重叠操作数**。背压（小 limit）迫使生产者串行化，减少操作时间窗口重叠，显著降低搜索空间。

---

## 文件结构

```
unbounded_channel/
├── unbounded_channel.go          # 公开 API（类型别名 + NewUnboundedChannel）
├── unbounded_channel_v6.go       # V6 实现
├── example_test.go               # 用法示例（可运行）
├── compare_test.go               # NativeChan vs V6 性能基准
├── linearizability_test.go       # Porcupine 线性一致性测试
├── DESIGN.md                     # 本文档
└── archive/                      # 历史版本存档（//go:build ignore）
    ├── ucv1.go                   # V1：原型，FIFO 未保证，无 worker
    ├── ucv2.go                   # V2：首次 FIFO + 后台 worker（10ms 固定轮询）
    ├── ucv3.go                   # V3：信号驱动 + 环形队列替换链表
    ├── ucv4.go                   # V4：快速路径 + 背压 + any 类型（chanLen 漂移 bug）
    ├── ucv5_original.go          # V5：废弃 chanLen，bufferLen 正确维护（自旋+退避）
    ├── ucv51.go                  # V51：自旋 + 指数退避（V5 变种，退避策略对照）
    ├── ucv53.go                  # V53：自旋 + 固定 1ms sleep（去掉指数退避）
    └── pure_queue.go             # 早期纯队列实验
```

---

## 教训：WaitDone() 的增删经历（测试驱动 API 污染的反面案例）

**经过：**
1. 时间轮组件测试使用 `goleak.VerifyTestMain`，检测到 V6 的 worker goroutine 在 `Close()` 后短暂存活
2. 为消除报警，给 V6 增加了公开方法 `WaitDone()`，并在时间轮 `Stop()` 中调用
3. 分析后发现这是错误决策，将 `WaitDone()` 全部回退

**问题根因：**

goleak 检测到的是 worker goroutine 的**调度时序窗口**——`Close()` 发出信号到 worker 收到信号并退出之间需要一次调度，不是真实的资源泄漏。在正确使用下，worker 会在微秒内自然退出。

**为什么 `WaitDone()` 是错误的 API：**

- 暴露了内部实现细节（调用方不应感知 worker goroutine 的存在）
- `Close()` 的语义类同 Go 原生 `close(ch)`——非阻塞，不等内部清理完成
- `WaitDone()` 存在死锁风险：channel 未排干时调用会永久阻塞
- 唯一用途是满足 goleak，没有独立的生产价值

**正确处理方式：**

时间轮测试改用 `goleak.IgnoreTopFunction(...)` 过滤 worker 的调度窗口；unbounded_channel 包不使用 `goleak.VerifyTestMain`（非阻塞 `Close()` 的 channel 原语与 `VerifyTestMain` 不兼容）。

**教训：**

测试工具报告问题时，先问：**这是真实的生产 bug，还是测试工具的配置问题？**

禁止为了让测试工具满意而新增公开 API。测试基础设施应适应正确的生产代码，而不是反向污染接口设计。
