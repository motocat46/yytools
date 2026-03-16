# UnboundedChannel 使用文档

有序（FIFO）、无界、带背压的泛型消息通道。

---

## 适用场景

| 场景 | 说明 |
|------|------|
| 生产速度 > 消费速度的削峰 | buffer 自动吸收突发流量，不丢消息 |
| 需要严格 FIFO 顺序 | 所有消息按发送顺序被消费 |
| 需要防止内存无限增长 | 通过 `limit` 参数触发背压，阻塞生产者 |
| 需要在 `select` 中使用 | `Out()` 返回只读 channel，可直接用于多路复用 |

**不适合的场景：** 对吞吐量极度敏感且不需要 FIFO 的场景，直接用 Go 原生 channel 性能更好（参见 TEST.md 性能基准）。

---

## 快速上手

```go
import uc "github.com/motocat46/yytools/pkg/infra/concurrency/unbounded_channel"

// 创建通道：底层 channel 大小 1024，buffer 上限 10 万条
ch := uc.NewUnboundedChannel[int](1024, 100_000)
defer ch.Close() // 必须调用，否则内部 goroutine 泄漏

// 生产者（并发安全）
ch.Send(42)

// 消费者（阻塞，直到有消息或通道关闭）
val, ok := ch.Receive()
if !ok {
    // 通道已关闭且无剩余消息
}
```

---

## API 说明

### `NewUnboundedChannel[T](chanSize int, limit int) *UnboundedChannel[T]`

创建无界通道。

| 参数 | 类型 | 说明 |
|------|------|------|
| `chanSize` | `int` | 底层 channel 的缓冲大小，必须 > 0。建议设为预期并发生产者数的 2~4 倍 |
| `limit` | `int` | buffer 中允许积压的消息上限，必须 > 0。超过后生产者阻塞（背压）。按实际内存预算设置 |

---

### `Send(msg T) bool`

发送消息，并发安全。

- 返回 `true`：消息已投递
- 返回 `false`：通道已关闭，消息被丢弃
- 当 buffer 积压超过 `limit` 时，**阻塞**直到消费者消费或通道关闭

```go
ok := ch.Send(msg)
if !ok {
    // 通道已关闭
}
```

---

### `Receive() (T, bool)`

接收消息，**阻塞**直到有消息可用或通道关闭。

- `ok = true`：成功接收到消息
- `ok = false`：通道已关闭且 buffer 中无剩余消息

```go
for {
    val, ok := ch.Receive()
    if !ok {
        break // 通道关闭，退出
    }
    // 处理 val
}
```

---

### `Out() <-chan T`

返回只读的底层 channel，供 `select` 多路复用使用。

```go
for {
    select {
    case val, ok := <-ch.Out():
        if !ok {
            return
        }
        // 处理 val
    case <-ctx.Done():
        return
    }
}
```

> **注意**：使用 `Out()` 时，worker 有 1ms ticker 兜底搬运 buffer，不需要手动处理。

---

### `Close()`

关闭通道。

- 标记关闭后，`Send()` 立即返回 `false`
- 已在 buffer 中的消息仍可被消费，**不会丢失**
- 所有因背压阻塞的生产者会被唤醒并返回 `false`
- 内部 worker goroutine 在消息全部消费完毕后自动退出

```go
defer ch.Close()
```

---

## 完整示例

### 基本生产者-消费者

```go
ch := uc.NewUnboundedChannel[string](16, 10000)
defer ch.Close()

var wg sync.WaitGroup
wg.Add(1)

// 消费者
go func() {
    defer wg.Done()
    for {
        val, ok := ch.Receive()
        if !ok {
            return
        }
        fmt.Println(val)
    }
}()

// 生产者
ch.Send("hello")
ch.Send("world")
ch.Close() // 关闭后消费者会读完剩余消息再退出

wg.Wait()
```

### 多生产者

```go
ch := uc.NewUnboundedChannel[int](64, 100_000)
defer ch.Close()

var wg sync.WaitGroup
for p := range 10 {
    wg.Add(1)
    go func(pid int) {
        defer wg.Done()
        for i := range 1000 {
            ch.Send(pid*1000 + i)
        }
    }(p)
}

// 消费者读取所有消息
go func() {
    wg.Wait()
    ch.Close()
}()

for {
    _, ok := ch.Receive()
    if !ok {
        break
    }
}
```

### 结合 select 使用

```go
ch := uc.NewUnboundedChannel[Event](128, 50_000)
defer ch.Close()

go func() {
    for {
        select {
        case event, ok := <-ch.Out():
            if !ok {
                return
            }
            handle(event)
        case <-ctx.Done():
            return
        }
    }
}()
```

---

## 参数选择建议

| 场景 | chanSize | limit |
|------|----------|-------|
| 低并发（< 10 生产者） | 64 ~ 256 | 10_000 ~ 100_000 |
| 中等并发（10~100 生产者） | 256 ~ 1024 | 100_000 ~ 1_000_000 |
| 高并发（> 100 生产者） | 1024 ~ 4096 | 按实际内存预算设置 |

`limit` 的内存开销约为 `limit × sizeof(T)`，设置前估算峰值内存用量。

---

## 常见误用

### ❌ 不调用 Close()

```go
ch := uc.NewUnboundedChannel[int](16, 1000)
// 忘记 ch.Close() → 内部 worker goroutine 永久泄漏
```

### ❌ Close() 后继续 Send()

```go
ch.Close()
ch.Send(1) // 返回 false，消息被丢弃，不会 panic
```

### ❌ 混用 Receive() 和 Out()

`Receive()` 和 `Out()` 读取的是同一个底层 channel，混用会导致消息被随机分流到两个消费路径，破坏业务逻辑。选其一即可。

### ❌ chanSize 设为 0

```go
uc.NewUnboundedChannel[int](0, 1000) // panic：chanSize must be > 0
```
