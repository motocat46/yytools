# DESIGN.md — ring_buffer

## 与 queue 的本质区别

queue 的契约是"不丢数据"（写满时扩容）；ring_buffer 的契约是"只保留最新 N 个"（写满时覆盖最旧元素）。两者语义相反，独立包，不作为 queue 的变体。

## 关键决策

### 1. Enqueue 满时静默覆盖，不返回旧值

旧数据本来就是要丢的——Linux kfifo、大多数音视频 SDK 均采用此策略。提供 `Full()` 供调用方在写入前主动判断。若调用方需要感知每次覆盖，通常意味着它不应该用 ring buffer。

### 2. Dequeue/Peek 为空时 panic，不返回零值

与 queue 策略一致。泛型 `T` 的零值（`0`、`""`、`false`）可能是合法元素，返回零值是静默错误。panic 迫使调用方先用 `Empty()` 检查。

### 3. 非并发安全，由调用方负责

不同并发场景对同步策略需求不同，数据结构内部加锁是过度设计。

### 4. 用独立 length 字段区分满和空

queue 用"留一个空位"区分满和空（实际可用容量是 `Capacity-1`）。ring_buffer 满时仍然写入，不能留空位，改用独立 `length` 字段：`length == 0` 为空，`length == capacity` 为满，实际可用容量等于 `capacity`。

三字段不变量（任何公开方法执行前后均须成立）：

```
0 <= head < capacity
0 <= tail < capacity
0 <= length <= capacity
tail == (head + length) % capacity
```

Enqueue 满时覆盖如何维护不变量：写入 `items[tail]`，tail 前进一步，此时 `length == capacity`，head 也前进一步——tail 和 head 同步前进，length 不变，不变量持续成立。

### 5. Range 签名为 func(T) bool，支持提前终止

与 Go 标准库 `slices.All` 等惯例一致，f 返回 false 时立即停止遍历，支持 break 语义。现在确定签名，避免发布后的破坏性变更。

### 6. Range 的环绕判断

`Range` 先用 `Empty()` 守门，进入遍历后只需判断是否环绕：

- `tail > head`：数据连续存放在 `[head, tail)`，无环绕。
- `tail <= head`：数据跨越数组末尾，先遍历 `[head, capacity)`，再遍历 `[0, tail)`。

满且 `tail == head` 时走第二个分支，`[head, capacity)` + `[0, tail)` 恰好覆盖全部元素。`tail == head` 表示空的情况已被 `Empty()` 拦截，不会进入此分支。

### 7. Cap() 命名

有意对齐 Go 内置的 `cap()` 语义（返回固定容量），而非沿用 `queue` 的 `Capacity()`——ring_buffer 的容量是构造时确定的常量，`Cap()` 更简洁且与 Go 惯用法一致。
