# fenwicktree 设计文档

## 设计目标

O(log n) 单点更新 + 前缀和查询，泛型 `base.Number`，0-indexed 对外接口，零分配。

---

## 关键设计决策

### 0-indexed 对外，1-indexed 对内

经典 BIT 教材全部使用 1-indexed，因为 `lowbit(i) = i & (-i)` 在 `i=0` 时退化为死循环（`lowbit(0)=0`，`pos` 永不前进）。

对外选择 0-indexed 的原因：Go 调用方的数据来自 slice，`for` 循环从 0 开始，强迫记住 `+1` 是制造越界 bug 的温床。`yourbasic/fenwick`（Go 生态常见的 BIT 实现）也选择了 0-indexed 对外。

实现方式：所有公开方法在访问 `tree` 前做 `i+1` 偏移，完全透明。

### 为什么不支持区间更新（RangeAdd）

区间更新需要两棵 BIT 配合（差分树技巧），语义上 `RangeSum` 变成 `PrefixSum(r) * r - PrefixSum(r-1) * (r-1)` 的形式，实现复杂度翻倍，API 不直观。此场景更适合 Segment Tree，yytools 不在此引入半成品。

### 为什么不支持 min/max 查询

BIT 的 `RangeSum` 依赖运算的**可逆性**（`rangeSum = prefixSum(r) - prefixSum(l-1)`）。`min/max` 不可逆，无法做区间查询。需要 `min/max` 区间查询请使用 Segment Tree。

### 为什么提供 Build(nums)

调用方常见场景不是“先 New 再逐个 Add”，而是“已有一整段初始数组，直接建树”。若只能循环 `Add`，建树成本为 `O(n log n)`；Fenwick Tree 存在经典的 `O(n)` 建树方法，因此提供 `Build(nums)` 作为独立构造函数。

选择独立构造函数而不是实例方法的原因：`Build` 和 `New` 都是“构造新树”，语义平行；若做成 `(*FenwickTree).Build(nums)`，会把“新建对象”和“重置已有对象”混在一起，引入不必要的状态语义。

---

## 复杂度

| 操作 | 时间 | allocs |
|------|------|--------|
| Add | O(log n) | 0 |
| PrefixSum | O(log n) | 0 |
| RangeSum | O(log n) | 0 |
| 整体空间 | O(n) | — |
