# ADR-001：集合类 API 的空/越界边界策略

**状态**：已采纳
**日期**：2026-03
**影响范围**：`pkg/ds/`（heap、queue、stack、sorted_set）

---

## 背景

泛型集合类在空状态或越界时应返回什么？两种方案各有代价：

- **返回零值**：`T` 的零值（`0`、`""`、`false`）可能是合法元素，调用方无法区分"空集合"和"真实零值"，是静默 bug 的温床。
- **返回 nil / panic**：明确，调用方被迫处理异常情况，bug 在开发阶段立刻暴露。

---

## 决策

根据**返回类型**采用不同策略：

### 指针返回方法 → 返回 nil，不 panic

适用：`sorted_set`、`heap`（返回 `*NodeData`、`*Item`、`*PriorityItem`）。

- `nil` 在语义上是明确的"没有该元素"，与任何合法元素不冲突
- 调用方可直接做 nil 检查，无需预先调用 `Empty()`

```go
item := h.PopItem()
if item == nil { /* 空堆 */ }
```

### 泛型 T 返回方法 → assert/panic，不返回零值

适用：`Stack[T].Pop()`、`Queue[T].Dequeue()`、`Queue[T].Peek()` 等。

- `T` 的零值可能是合法元素，返回零值是静默错误
- assert 迫使调用方先用 `Empty()` 检查，让 bug 在开发阶段立刻暴露

```go
// ✅ 正确
if !stack.Empty() {
    item := stack.Pop()
}
```

---

## 权衡

放弃的方案：`(T, bool)` 二元组。优点是类型安全且不 panic，缺点是每次调用都要解构，日常使用冗余。项目选择"先检查再操作"的惯用法，与标准库 `map` 的 `val, ok := m[key]` 语义一致但更严格（不检查直接 panic）。

---

## 相关文档

- `pkg/ds/heap/DESIGN.md` §三
- `pkg/ds/queue/DESIGN.md` §三
- `pkg/common/assert/DESIGN.md`
- `CLAUDE.md` §数据结构 API 边界规范
