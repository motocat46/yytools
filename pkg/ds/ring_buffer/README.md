# ring_buffer

固定容量的泛型环形缓冲区。写满时静默覆盖最旧元素，不扩容，适合只关心最新 N 个元素的场景。

## 适用场景

- 滑动窗口统计（最近 N 次操作的数据）
- 日志/事件缓冲（只关心最新的 N 条）
- 游戏操作历史记录（回放最近 N 帧）

不适用于不能丢数据的场景，那种情况请用 [`queue`](../queue/README.md)。

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/ds/ring_buffer"

r := ring_buffer.NewRingBuffer[int](3)

r.Enqueue(1)
r.Enqueue(2)
r.Enqueue(3)
r.Enqueue(4) // 缓冲区已满，覆盖最旧的 1，现在保留 [2, 3, 4]

fmt.Println(r.Len()) // 3
fmt.Println(r.Full()) // true

// 遍历（从最旧到最新）
r.Range(func(v int) bool {
    fmt.Println(v) // 2, 3, 4
    return true    // 返回 false 可提前终止
})

// 读取并移除队首
fmt.Println(r.Dequeue()) // 2

// 查看队首但不移除
fmt.Println(r.Peek()) // 3
```

## API

```go
// 创建容量为 capacity 的空缓冲区，capacity 必须 > 0
func NewRingBuffer[T any](capacity int) *RingBuffer[T]

// 入队；已满时覆盖最旧元素，Len() 和 Cap() 保持不变
func (r *RingBuffer[T]) Enqueue(item T)

// 出队，返回并移除队首元素；空时 panic，调用前先检查 Empty()
func (r *RingBuffer[T]) Dequeue() T

// 返回队首元素但不移除；空时 panic，调用前先检查 Empty()
func (r *RingBuffer[T]) Peek() T

// 当前元素数量
func (r *RingBuffer[T]) Len() int

// 缓冲区容量（固定值）
func (r *RingBuffer[T]) Cap() int

func (r *RingBuffer[T]) Empty() bool
func (r *RingBuffer[T]) Full() bool

// 从最旧到最新遍历；f 返回 false 时提前终止
func (r *RingBuffer[T]) Range(f func(T) bool)
```

## 与 queue 的区别

| | queue | ring_buffer |
|---|---|---|
| 写满时 | 扩容，不丢数据 | 覆盖最旧元素 |
| 容量 | 动态增长 | 固定 |

## 常见误用

```go
// ❌ 错误：空时直接 Dequeue/Peek 会 panic
item := r.Dequeue()

// ✅ 正确：先检查
if !r.Empty() {
    item := r.Dequeue()
}
```

非并发安全，并发访问由调用方负责加锁。
