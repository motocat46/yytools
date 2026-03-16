# heap 使用文档

基于 Go 标准库 `container/heap` 封装的泛型堆实现，提供三种变体：最小堆、最大堆、优先级队列。

## 核心概念

堆元素通过 `Weight`（权重）或 `Priority`（优先级）决定出堆顺序。

---

## 最小堆 `Heap[T]`

**Weight 越小越先出堆。**

```go
import "github.com/motocat46/yytools/pkg/ds/heap"

h := heap.NewHeap[string]()

h.PushItem(&heap.Item[string]{Data: "task-c", Weight: 3})
h.PushItem(&heap.Item[string]{Data: "task-a", Weight: 1})
h.PushItem(&heap.Item[string]{Data: "task-b", Weight: 2})

h.PeekItem()   // {task-a, 1}，不弹出
h.PopItem()    // {task-a, 1}
h.PopItem()    // {task-b, 2}
```

---

## 最大堆 `MaxHeap[T]`

**Weight 越大越先出堆。**

```go
h := heap.NewMaxHeap[string]()

h.PushItem(&heap.Item[string]{Data: "low",  Weight: 1})
h.PushItem(&heap.Item[string]{Data: "high", Weight: 100})

h.PopItem() // {high, 100}
```

---

## 优先级队列 `PriorityQueue[T]`

**Priority 越大优先级越高（越先出队）。** 额外支持动态修改优先级。

```go
pq := heap.NewPriorityQueue[string]()

itemA := &heap.PriorityItem[string]{Data: "normal", Priority: 5}
itemB := &heap.PriorityItem[string]{Data: "urgent", Priority: 10}

pq.PushItem(itemA)
pq.PushItem(itemB)

pq.PopItem() // {urgent, 10}

// 动态更新优先级（已在队列中的元素）
pq.PushItem(itemA) // 重新入队
pq.UpdatePriority(itemA, 20) // 提升为最高优先级
pq.PopItem() // {normal, 20}
```

---

## API 对比

| 操作 | Heap（最小堆）| MaxHeap（最大堆）| PriorityQueue |
|------|-------------|----------------|--------------|
| 创建 | `NewHeap[T]()` | `NewMaxHeap[T]()` | `NewPriorityQueue[T]()` |
| 元素类型 | `*Item[T]` | `*Item[T]` | `*PriorityItem[T]` |
| 排序字段 | `Weight`（小→大）| `Weight`（大→小）| `Priority`（大→小）|
| 入堆 | `PushItem` | `PushItem` | `PushItem` |
| 出堆 | `PopItem` | `PopItem` | `PopItem` |
| 查看堆顶 | `PeekItem` | `PeekItem` | `PeekItem` |
| 动态更新 | ✗ | ✗ | `UpdatePriority` |
| 堆大小 | `Length` | `Length` | `Length` |

## 注意事项

- **不要直接调用 `Push`/`Pop`**，这是实现 `container/heap` 接口的内部方法；使用 `PushItem`/`PopItem`
- `PeekItem` 在空堆时会 panic（index out of range）
- `UpdatePriority` 要求元素当前**在队列中**，否则触发断言 panic
- 非并发安全，多 goroutine 使用时需自行加锁
