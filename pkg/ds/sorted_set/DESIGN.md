# sorted_set 设计文档

## 设计目标

提供一个类 Redis ZADD 语义的有序集合：按 Score 排序、O(log n) 增删查、O(1) 按 Key 查找，支持排名查询和范围操作。

---

## 核心结构选型

### 为什么是跳表 + 哈希表的组合？

| 操作 | 只用平衡树 | 只用跳表 | 跳表 + 哈希表 |
|------|-----------|---------|-------------|
| 按 Key 查找 | O(log n) | O(log n) | **O(1)** |
| 按 Score 排序查找 | O(log n) | O(log n) | O(log n) |
| 排名查询 | 需扩展（order statistics tree）| 天然支持（Span 字段）| O(log n) |
| 实现复杂度 | 高（旋转平衡）| 中 | 中 |

Redis 的 zset 也是同样的选择。哈希表负责 Key → NodeData 的 O(1) 映射，跳表负责 Score 排序和排名运算。

### 为什么选跳表而不是红黑树？

- 跳表的 **Span 字段**天然记录层间跨度，排名查询（GetRank/GetByRank）可在 O(log n) 内完成；红黑树实现排名需要额外的 size 域并维护，实现复杂度更高。
- 跳表代码可读性更高，调试和验证更容易。
- 范围删除（DeleteRangeByRank/DeleteRangeByScore）在跳表上可以顺序遍历完成，效率和实现都更简单。

---

## 关键设计决策

### 1. `seq` 实现相同分数的稳定排序

**问题**：分数相同时，跳表需要全序才能正确定位节点（否则无法区分同分数的两个节点）。

**方案**：每次 `Insert` 时分配自增 `seq`，跳表内部排序使用 `(Score, seq)` 二元组，保证全序且稳定。

**`seq` 不对外暴露**：业务代码只关心 Score，seq 是内部实现细节，故设为小写字段。

**注意**：`UpdateScore` 不改变 `seq`，因此相同分数内的相对顺序不变。如果用删除再插入代替 UpdateScore，新插入会分配新的 seq，破坏稳定性——这是 README 中明确告知的注意事项。

### 2. 比较方法仅保留内部实现

跳表排序使用两个内部方法：
- `lessOrder`：按 `(Score, seq)` 比较，保证全序，用于节点定位和插入。
- `equalOrder`：按 `seq` 比较，用于精确定位同一个节点。

两者均为小写（包内可见），不对外暴露。外部调用方直接访问 `.Score` 和 `.Key` 字段即可，无需包装方法。

> 曾经存在公开的 `LessThan` 和 `EqualTo` 方法，但两者在实现中均未被使用，且 `EqualTo` 比较的是 `Val` 而非 `Key`，语义有误。另外 `EqualTo` 要求 `Val` 满足 `comparable`，阻碍了 `V any` 的扩展。已全部删除。

### 3. `UpdateScore` 的原地优化

如果新分数不改变节点在跳表中的相对位置（前驱分数 < newScore < 后继分数），则直接修改 `Score` 字段，无需删除重插，节省一次完整的 O(log n) 操作。

```go
if (prev.Score < newScore) && (next.Score > newScore) {
    current.Data.Score = newScore  // 原地更新
    return current, true
}
// 否则：删除 + 重新插入
```

### 4. `SKIPLIST_MAXLEVEL = 32`，`LevelUpProb = 0.25`

沿用 Redis 参数：
- 最大层数 32 可支持约 4^32 ≈ 18 亿个节点，远超实际使用上限。
- 晋升概率 0.25（而非常见的 0.5）在时间复杂度和空间占用之间取得更好的平衡：平均节点高度为 1/(1-0.25) ≈ 1.33 层。

### 5. 双泛型参数 `[K comparable, V any]`

初始设计使用单泛型参数 `[T comparable]`，导致 Key 和 Val 必须是同一类型，Val 也被迫满足 `comparable` 约束，实际使用中只能存 `key` 本身作为 val，造成冗余。

改为 `NodeData[K comparable, V any]` 后：
- Key 负责哈希表查找（需要 comparable）
- Val 可以是任意类型（any），包括 struct、指针、slice 等业务数据
- 跳表本身的排序逻辑只使用 Score 和 seq，对 K、V 类型完全透明，改动是纯机械的泛型参数透传

### 6. `SortedSet` 与 `SkipList` 分层

- `SkipList` 是纯跳表，不感知哈希表，可独立复用。
- `SortedSet` 组合 `SkipList` + `hash`，提供面向业务的 API，并在每次变更后用 `lengthMustEqual()` 做一致性断言。

这种分层让跳表的单元测试和有序集合的单元测试可以独立进行，也方便未来替换底层实现。

---

## 已知限制

- 非并发安全，调用方需自行加锁。
