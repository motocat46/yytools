# heap 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、为什么封装 `container/heap` 而不是自己实现

Go 标准库 `container/heap` 提供了完整的堆算法（siftUp/siftDown），但 API 是基于接口的非泛型设计：调用方需要实现 `heap.Interface`（Len/Less/Swap/Push/Pop），每次使用都要写大量模板代码，且操作时需要手动类型断言。

封装目标：
- 一次性吸收模板代码，调用方只面对 `PushItem` / `PopItem` / `PeekItem`
- 通过泛型消除类型断言
- 统一 API 边界：空堆返回 `nil`，不 panic

**决策**：复用标准库算法，只封装使用层。不重写堆算法——标准库经过充分测试，重写收益为零。

---

## 二、三种变体的分工

| 类型 | 文件 | 排序语义 | 特点 |
|------|------|---------|------|
| `Heap[T]` | `heap.go` | Weight 小 → 先出（最小堆）| 通用，最简 |
| `MaxHeap[T]` | `max_heap.go` | Weight 大 → 先出（最大堆）| 翻转 Less 实现 |
| `PriorityQueue[T]` | `priority_queue.go` | Priority 大 → 先出，支持动态更新 | 额外维护 Index 字段 |

**MaxHeap 为什么单独一个类型而不是参数化？**

选项一：`NewHeap[T](order Order)` 通过参数控制升/降序。
选项二：独立类型 `MaxHeap[T]`。

选项一让类型签名携带了运行时状态，且在编译期无法区分最小堆和最大堆——传错参数无任何提示。选项二让意图在类型名中直接可见，不需要看构造函数参数。

**PriorityQueue 为什么需要 Index 字段？**

`UpdatePriority` 需要调用 `heap.Fix(this, index)` 重新调整堆序。`heap.Fix` 需要元素在底层切片中的下标，而堆每次 Swap 都会改变元素位置。唯一可靠的方法是在 Swap 时同步更新元素的 Index 字段，让元素始终"知道自己在哪"。

代价：每次 Swap 多两次写（`Items[i].Index = i`，`Items[j].Index = j`），对于不需要 `UpdatePriority` 的场景是冗余开销，这也是 `Heap` 和 `MaxHeap` 不维护 Index 的原因。

---

## 三、空堆 API 行为：返回 nil，不 panic

`PopItem` / `PeekItem` 返回 `*Item[T]`（指针），`nil` 在语义上是明确的"无元素"信号，与合法元素不冲突。

```go
item := h.PopItem()
if item == nil { ... } // 调用方可直接 nil 检查
```

这与项目中泛型 T 返回方法（如 `Stack[T].Pop()`）的 panic 策略不同——那些方法返回值类型是 `T`，零值无法区分"空集合"和"真实零值元素"，所以选择 panic 迫使调用方先检查。指针返回时没有这个歧义，返回 nil 更友好。

---

## 四、踩坑：直接调用 `Push`/`Pop` 的问题

`Heap[T]` 实现了 `heap.Interface`，因此 `Push` 和 `Pop` 是**公开**方法（Go 接口要求首字母大写）。但这两个方法是给 `container/heap` 包内部调用的，直接从外部调用会绕过堆序维护，破坏堆的不变量。

解决方案：README 中明确禁止直接调用，并提供了替代方法。未来可考虑将 `Heap` 的内部实现改为私有类型（外部只暴露接口），但当前写法够清晰，暂不重构。
