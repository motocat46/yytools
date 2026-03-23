# ADR-003：扁平包结构，不拆分 concrete/generic 子包

**状态**：已采纳
**日期**：2026-03
**影响范围**：`pkg/` 下所有包

---

## 背景

Go 泛型（1.18+）引入后，有一种常见模式是将包拆分为：

```
pkg/ds/heap/
    concrete/   # 具体类型实现（如 IntHeap）
    generic/    # 泛型实现（Heap[T]）
```

项目曾讨论是否采用这一结构。

---

## 决策

**保持扁平结构，不创建 concrete/generic 中间目录，除非两者都真实存在且各自有足够的代码量。**

目前所有数据结构均直接使用泛型实现，无具体类型变体。包结构为：

```
pkg/ds/heap/
    heap.go
    max_heap.go
    priority_queue.go
```

---

## 理由

### 1. 调用方认知成本最低

调用方 `import "github.com/motocat46/yytools/pkg/ds/heap"` 即可获得所有变体。拆子包后调用方需要判断"我需要的是 concrete 还是 generic"，增加不必要的心智负担。

### 2. 避免为假设的未来需求设计

项目中没有具体类型变体（如 `IntHeap`）的实际需求。创建空的 `concrete/` 目录是过度设计，且可能误导维护者认为"这里应该放具体类型实现"。

### 3. 中间目录没有承载信息

`concrete/` 和 `generic/` 这类技术性分组目录描述的是"如何实现"，不是"做什么"。领域驱动的包名（`heap`、`queue`）比技术分组更能反映意图。

---

## 例外情况

若某个领域确实同时存在：
- 具体类型实现（为了零分配或特化性能）
- 泛型实现（通用场景）

且两者代码量均不小，可以创建子包。但应以领域名而非 concrete/generic 命名（如 `heap/int` 而非 `heap/concrete`）。

---

## 相关文档

- `CLAUDE.md` §Code Conventions："Flat package structure: no concrete/generic intermediate directories except where both truly exist"
