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
- 最大层数 32 可支持约 4^32 ≈ 1.8 × 10^19 个节点，远超实际使用上限。
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

### 7. 逆序排名接口（GetRankDesc / GetByRankDesc / GetRangeByRankDesc）

初始实现只有正序（score 升序）接口，排行榜场景（高分靠前）需要传负分数，反直觉。

后续补充了三个逆序接口，全部在 `SortedSet` 层做坐标转换（`descRank = Length - ascRank + 1`），skiplist 不需要改动。

`GetRangeByRankDesc` 内部调用 `GetRangeByRank` 后反转切片，O(k) 额外开销（k 为返回元素数），可接受。

---

### 8. API 分组与数量管理

随着方法增多（当前 18 个），采用**按语义分组 + 独立接口文件**的策略，而非拆包：

| 接口 | 方法 |
|------|------|
| `BasicOps` | `Insert` / `Delete` / `Get` / `Length` / `UpdateScore` / `GetMin` / `GetMax` |
| `AscRankOps` | `GetRank` / `GetByRank` / `GetRangeByRank` / `DeleteRangeByRank` |
| `DescRankOps` | `GetRankDesc` / `GetByRankDesc` / `GetRangeByRankDesc` / `DeleteRangeByRankDesc` |
| `ScoreOps` | `GetRangeByScore` / `DeleteRangeByScore` / `CountByScore` |
| `SortedSetOps` | 以上全部（供整体 mock 用） |

**不拆包的理由**：`SortedSet` 是内聚概念，拆包后调用方不知道"查排名用哪个类"。分组接口文件解决的是使用侧的依赖范围问题，而非实现侧的拆分。

**编译期保障**：`interfaces.go` 末尾有 `var _ SortedSetOps[int, int] = (*SortedSet[int, int])(nil)`。新增公开方法时若未加入任何子接口，编译即报错，防止接口与实现悄悄脱节。

### 9. `GetMin` / `GetMax`

语义比数字 rank 更直观，避免调用方写 `GetByRankDesc(1)` 时需要心算"最高分就是降序第 1"。

实现直接利用跳表已有的结构指针，**O(1)**：

- `GetMin`：`Head.Levels[0].Forward`，底层链表 level-0 的第一个真实节点，始终是全局最小
- `GetMax`：`Tail`，跳表在每次插入/删除时同步维护的尾指针，始终是全局最大

两者均只读一个指针，无需遍历任何索引层。基准数据（Apple M4）：GetMin ~0.6 ns/op，GetMax ~0.4 ns/op，0 allocs，与集合规模无关。

**维护注意**：修改跳表插入/删除逻辑时，必须同时保证 `Head.Levels[0].Forward` 和 `Tail` 的正确性，否则 GetMin/GetMax 会静默返回错误结果。

### 10. `CountByScore`

在 skiplist 层实现，只遍历节点、不构建切片。

`GetRangeByScore` 每次调用会分配结果切片（n=100 时 6 allocs/1016 B，随 n 增长）；`CountByScore` 只有 1 次结构体分配（24 B，固定），节省的是结果切片部分。适合只需要数量、不需要元素列表的高频计数场景。

---

## 已知限制

- 非并发安全，调用方需自行加锁。
- Score 不支持 NaN。NaN 的比较语义（`NaN != NaN`、`NaN < NaN == false`）会破坏跳表依赖的全序关系，`Insert`、`UpdateScore`、`GetRangeByScore`、`DeleteRangeByScore` 均对 NaN 触发 assert panic。`±Inf` 是合法值。
- seq 使用 `uint64` 单调递增，每次 `Insert` 消耗一个值。理论上限 2^64 ≈ 1.8×10¹⁹ 次插入，远超任何实际场景，不需要处理溢出。
