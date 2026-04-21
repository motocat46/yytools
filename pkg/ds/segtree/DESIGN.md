# segtree 设计文档

## 选型：为何选 ACL 风格

Go 生态无成熟泛型线段树。ACL（AtCoder Library）是日本竞赛社区 10 年打磨的标准实现，核心设计是将 monoid 操作（merge/apply/compose）完全外置，库本身不依赖任何具体操作语义。

对比其他方案：
- **接口方案（interface Monoid）**：调用方需定义满足接口的类型，对 int 等基础类型不友好，需 wrapper
- **硬编码 sum/min/max**：覆盖面窄，每种操作需单独实现
- **ACL 函数注入**：调用方传闭包，类型安全，泛型推断自然，与 Go 风格吻合

## 关键决策

### 始终 pushdown（不判 lazy 零值）

标准实现判断 `lazy != lazyZero` 再 pushdown，但 Go 泛型 `L any` 无法对任意类型做相等比较（`comparable` 约束会排除 struct 含 slice 的情形）。两种替代方案：

1. 引入 `comparable` 约束 → 限制 L 类型，不能含 slice/map/func
2. **始终 pushdown** → 无约束，依赖调用方保证 `apply(val, lazyZero, size)==val`

选方案 2：灵活性更高，约束作为文档说明，与 ACL 原版行为一致。

### 数组大小 4n

递归线段树节点数上界为 `2 * nextPow2(n) - 1 ≤ 4n`，`4n` 是业界标准安全上界，不需要复杂的动态计算。

### 对外 0-indexed，内部 1-indexed

与 FenwickTree 惯例一致，根节点在 `tree[1]`，左子 `2v`，右子 `2v+1`。`tree[0]` 不使用（初始化为 identity）。

### 不支持 Build（从初始数组构建）

spec 未要求，YAGNI。需要时调用方可逐一 Set，O(n log n)。

## 不在范围内

- 持久化线段树：需指针结构，内存模型完全不同
- 二维线段树：调用方可自行嵌套
- 并发安全：调用方加锁
- 动态开点：固定容量
