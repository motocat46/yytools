# queue 使用文档

基于环形缓冲区（Ring Buffer）实现的泛型队列，支持自动扩缩容，FIFO 语义。

## API

### 接口 `IQueue[T any]`

| 方法 | 说明 |
|------|------|
| `Len() int` | 队列中的元素数量 |
| `Empty() bool` | 队列是否为空 |
| `Enqueue(item T)` | 入队 |
| `Dequeue() T` | 出队，**调用方需保证非空** |
| `Peek() T` | 查看队首元素（不出队），空时 panic |

### 具体类型 `Queue[T any]`，额外提供：

| 方法 | 说明 |
|------|------|
| `Capacity() int` | 当前底层数组容量 |
| `Range(f func(T))` | 按 FIFO 顺序遍历，不出队 |

### 构造函数

```go
queue.NewQueue[T]()                // 默认容量 16
queue.NewQueueWithSize[T](size int) // 指定初始容量
```

### 扩缩容策略

- **扩容**：满时自动翻倍
- **缩容**：元素数 < 容量 1/4 时缩为一半，最小保持 16

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/ds/queue"

q := queue.NewQueue[int]()

// 入队
q.Enqueue(1)
q.Enqueue(2)
q.Enqueue(3)

// 查看队首
fmt.Println(q.Peek()) // 1

// 出队
for !q.Empty() {
    fmt.Println(q.Dequeue()) // 1, 2, 3（FIFO 顺序）
}

// 遍历（不消费）
q.Enqueue(10)
q.Enqueue(20)
q.Range(func(v int) {
    fmt.Println(v) // 10, 20
})
```

## 注意事项

- `Dequeue()` 不检查空队列，调用前需确保 `!Empty()` 或 `Len() > 0`
- `Peek()` 空队列时触发 panic
- 非并发安全，多 goroutine 使用时需自行加锁
