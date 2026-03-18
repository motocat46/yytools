# DESIGN.md — snowflake

## 设计目标

为游戏服务器提供**无锁、线程安全、零分配**的雪花 ID 生成器。
每毫秒每节点最多 4096 个 ID，支持 1024 个节点，有效期约 69 年（2025–2094）。

---

## 核心设计：state 原子打包

### 问题

雪花算法需要同时原子地读写两个字段：`lastMs`（上一次生成的毫秒时间戳）和 `sequence`（同毫秒内的序号）。朴素实现用 mutex 保护两个字段，有锁竞争开销。

### 方案

将 `(lastMs, sequence)` 打包进一个 `int64`，用 CAS 原子操作整体更新：

```
int64 state = (lastMs << 12) | sequence
```

- `lastMs`：高 52 位（实际只用到 41 位）
- `sequence`：低 12 位（0–4095）

打包后的 state 最多占用 53 位（41 + 12），完全在 int64 范围内，CAS 一条指令完成。

### 为什么不用 sync.Mutex

- mutex 在高并发下有内核态切换开销
- CAS 在无竞争时只需 1 条原子指令（~5 ns），在竞争时自旋重试，不阻塞
- 实测：单线程 244 ns/op，多线程（10 核）266 ns/op，CAS 竞争代价极小

---

## 时钟回拨处理

### 问题

NTP 校时可能导致系统时钟短暂回拨。若 ID 的时间戳回退，同一节点相同时间戳下可能生成重复 ID。

### 方案

`state` 记录的 `lastMs` **只增不减**：

```go
if now <= oldMs {
    // 回拨或同一毫秒：沿用 lastMs，递增 sequence
    newMs = oldMs
    newSeq = oldSeq + 1
}
```

时钟回拨期间：
- ID 的时间戳字段固定为 `lastMs`（回拨前的值），不跟随回拨
- sequence 继续递增，保证唯一性
- ID 数值大于正常时期的 ID（因为 `lastMs` 更大），不破坏单调性

**代价**：回拨幅度大时，可能短暂产生"未来时间戳"的 ID（timestamp > 实际当前时间）。游戏服务器可接受，ID 唯一性不受影响。

---

## 序号耗尽处理

同一毫秒内 sequence 超过 4095 时，自旋等待进入下一毫秒：

```go
if newSeq > MaxSequence {
    for currentMillis() <= oldMs {
        runtime.Gosched()
    }
    continue
}
```

**为什么用 Gosched 而不是 time.Sleep(1ms)**

`time.Sleep` 在 1ms 尺度精度仅 1–5ms（OS 调度粒度），会大幅过冲，导致恶性循环（过冲 → 下一毫秒序号耗尽 → 再过冲）。`Gosched()` 单次约 50–500 ns，可以精确检测毫秒边界，过冲风险极低。

**边界情况：序号耗尽与时钟回拨同时发生**

单独任一条件对性能无明显影响：
- 仅序号耗尽 → 自旋 < 1ms，等待时钟推进
- 仅时钟回拨 → 沿用 `lastMs` 递增序号，不触发自旋

但当两个条件**同时发生**时：

```
state = (lastMs=1000, seq=4095)  ← 序号恰好耗尽
此时时钟回拨 → currentMillis() 跌至 500
自旋：for currentMillis() <= 1000，需等待真实时钟追上 lastMs
```

goroutine 将自旋约等于回拨幅度的时长。NTP 典型调整为毫秒级，可忽略不计；手动大幅调时（分钟级）才会产生可感知的停顿，运维调时后若进程看似停顿请勿误判为死锁——不会永久卡死，时钟追上后自动恢复。

---

## 包级 defaultGen 与线程安全

`Init` 写入 `defaultGen`（普通指针），`NewID()` 读取，均无额外同步原语。

这依赖 Go 内存模型的 goroutine 创建 happens-before 保证：

```go
// main goroutine
snowflake.Init(nodeID)     // 写 defaultGen
go worker()                // goroutine 创建是同步点
// worker 内 NewID() 读 defaultGen，可见 Init 的写入
```

**不变量**：`Init` 必须在任何调用 `NewID()` 的 goroutine 启动之前完成。

若需支持热重载（运行中更换 nodeID），应改用 `atomic.Pointer[Generator]`，当前版本不需要。

---

## 位布局选择

标准雪花算法（Twitter 原版）：`timestamp(41) | datacenter(5) | worker(5) | sequence(12)`，分 datacenter + worker 两级。

本实现合并为单个 `nodeID(10)`，原因：
- 游戏服务器通常由运维直接分配节点 ID，不需要两级层次
- 合并减少配置复杂度，节点上限 1024 足够
- 实现更简单，不引入额外的 bit 操作

---

## 自定义纪元

选择 `2025-01-01 00:00:00.000 UTC` 作为纪元（Epoch）：
- 对应项目创建时间，减小时间戳数值，延长有效期上限（41 位可用到 2094 年）
- 相比 Twitter 原版纪元（2010 年），有效期多约 15 年

**有效期限制**

41 位时间戳的上限为 `MaxTimestamp = 2^41 - 1 = 2199023255551`，对应 2094 年。
`currentMillis()` 同时检测下界（< 2025-01-01）和上界（> 2094），两者均无条件 panic，
不允许静默产生符号位损坏的负数 ID。2094 年前需迁移至新纪元或扩展时间戳位数。