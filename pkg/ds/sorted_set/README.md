# sorted_set

基于跳表（Skip List）+ 哈希表实现的有序集合，语义类似 Redis ZADD/ZRANK/ZRANGE。

- O(log n) 插入 / 删除 / 排名查询
- O(1) 按 Key 查找
- 相同 Score 的元素按插入顺序稳定排序

## 核心概念

每个元素由三部分组成：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Key` | `K comparable` | 唯一标识，用于哈希表快速查找 |
| `Score` | `float64` | 排序依据，值越小升序排名越靠前 |
| `Val` | `V any` | 业务数据，类型独立于 Key，可为任意类型 |

### 排名说明

提供两套排名接口，从 1 开始计数：

| 接口 | rank=1 对应 | 适合场景 |
|------|------------|---------|
| `GetRank` / `GetByRank` / `GetRangeByRank` | Score 最小的元素 | 时间竞速、延迟排序等 |
| `GetRankDesc` / `GetByRankDesc` / `GetRangeByRankDesc` | Score 最大的元素 | 游戏排行榜、积分榜等 |

两套接口满足对称关系：`GetRank(key) + GetRankDesc(key) == Length + 1`。

> **rank 约定**：无论正序还是逆序，rank 均从 **1** 开始（而非 0），范围查询均为**闭区间** `[start, end]`（而非传统的 `[0, length)`）。这与"第 1 名、第 2 名"的自然语言语义一致，减少调用时的心智负担。
>
> **底层说明**：逆序接口仅在 `SortedSet` 层做坐标转换（`descRank = Length - ascRank + 1`），底层跳表的物理结构始终按 Score 升序排列（Score 越大，升序 rank 越大）。逆序接口不改变跳表结构，只是换了一个观察视角，两套接口的性能完全相同。

---

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/ds/sorted_set"

ss := sorted_set.NewSortedSet[string, int]()

// 插入玩家积分
ss.Insert(sorted_set.NewNodeData("alice", 1500.0, 1500))
ss.Insert(sorted_set.NewNodeData("bob",   1200.0, 1200))
ss.Insert(sorted_set.NewNodeData("carol", 1800.0, 1800))

// 排行榜：carol 第 1，alice 第 2，bob 第 3
ss.GetRankDesc("carol") // 1
ss.GetRangeByRankDesc(1, 3) // [carol, alice, bob]

// 更新积分后自动重排
ss.UpdateScore("bob", 2000.0)
ss.GetRankDesc("bob") // 1（现在最高）
```

---

## API

### 接口分组

所有公开方法按语义分为四个接口，定义在 `interfaces.go`：

| 接口 | 包含方法 | 适用场景 |
|------|---------|---------|
| `BasicOps[K, V]` | `Get` / `Insert` / `Delete` / `Length` / `UpdateScore` / `GetMin` / `GetMax` | 只需按 Key 操作的代码 |
| `AscRankOps[K, V]` | `GetRank` / `GetByRank` / `GetRangeByRank` / `DeleteRangeByRank` | 正序排名（时间竞速等） |
| `DescRankOps[K, V]` | `GetRankDesc` / `GetByRankDesc` / `GetRangeByRankDesc` / `DeleteRangeByRankDesc` | 降序排行榜 |
| `ScoreOps[K, V]` | `GetRangeByScore` / `DeleteRangeByScore` / `CountByScore` | 按积分段查询 / 清理 |
| `SortedSetOps[K, V]` | 以上全部 | 整体 mock / 替换实现 |

**使用建议**：函数参数优先声明最小接口，而非 `*SortedSet`：

```go
// 只需要排行榜展示——依赖 DescRankOps，不暴露写操作
func ShowLeaderboard(r sorted_set.DescRankOps[string, int], top int) {
    result := r.GetRangeByRankDesc(1, top)
    // ...
}

// 需要清理数据——依赖 ScoreOps
func ExpireByScore(s sorted_set.ScoreOps[string, int], deadline float64) {
    s.DeleteRangeByScore(math.Inf(-1), false, deadline, false)
}

// 多数情况：直接用具体类型，不需要接口
ss := sorted_set.NewSortedSet[string, int]()
ss.Insert(...)
```

---

### 构造

#### `NewSortedSet[K comparable, V any]() *SortedSet[K, V]`

创建一个空的有序集合。

#### `NewNodeData(key K, score float64, val V) *NodeData[K, V]`

创建节点数据，用于传入 `Insert`。

---

### 写操作

#### `Insert(data *NodeData[K, V]) bool`

插入元素。O(log n)。

| 情况 | 返回值 |
|------|-------|
| Key 不存在 | 插入成功，返回 `true` |
| Key 已存在 | 原有元素不受影响，返回 `false` |
| `data` 为 nil | 触发断言 panic |

> **注意**：不要用 `Delete + Insert` 来更新分数，应使用 `UpdateScore`。重新插入会分配新的内部序列号，破坏相同 Score 下原有的稳定排序顺序。

---

#### `Delete(key K) (*NodeData[K, V], bool)`

按 Key 删除元素。O(log n)。

| 情况 | 返回值 |
|------|-------|
| Key 存在 | 返回被删除的节点数据，`true` |
| Key 不存在 | 返回 `nil, false` |

---

#### `UpdateScore(key K, newScore float64) (*NodeData[K, V], bool)`

更新元素的分数并触发重新排序。O(log n)。

| 情况 | 返回值 |
|------|-------|
| Key 存在 | 返回更新后的节点数据，`true` |
| Key 不存在 | 返回 `nil, false` |

> 与 `Delete + Insert` 的区别：`UpdateScore` 保留原有内部序列号，相同 Score 下的稳定排序不受影响。
> 若新分数未改变节点的相对位置（前驱 Score < newScore < 后继 Score），则原地修改，无需删除重插。

---

#### `DeleteRangeByRank(start, end int) []*NodeData[K, V]`

删除升序排名范围 `[start, end]` 内的所有元素。O(log n + k)，k 为删除元素数。

- 返回被删除的元素列表（升序）
- `start > end` 时自动交换
- 范围无交集时返回空切片

---

#### `DeleteRangeByRankDesc(start, end int) []*NodeData[K, V]`

删除降序排名范围 `[start, end]` 内的所有元素。O(log n + k)，k 为删除元素数。

- 返回被删除的元素列表（降序，Score 从大到小）
- `start > end` 时自动交换；`end` 超出总长度时截断，不 panic

```go
// 删除排行榜前 3 名（积分最高的 3 人）
ss.DeleteRangeByRankDesc(1, 3)

// 踢出积分最低的后 3 名（降序排名末尾）
ss.DeleteRangeByRankDesc(ss.Length()-2, ss.Length())
```

---

#### `DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V]`

删除 Score 在指定范围内的所有元素。O(log n + k)，k 为删除元素数。

- `minEx=false` 包含 min 端点，`minEx=true` 排除 min 端点；`maxEx` 同理作用于 max
- 返回被删除的元素列表（升序）；范围内无元素时返回空切片

```go
// 删除积分在 (1000, 1500] 之间的所有玩家
deleted := ss.DeleteRangeByScore(1000, true, 1500, false)
```

---

### 读操作

#### `Get(key K) *NodeData[K, V]`

按 Key 查找元素。O(1)。Key 不存在时返回 `nil`。

---

#### `Length() int`

返回集合中的元素总数。O(1)。

---

#### `GetMin() *NodeData[K, V]`

返回排序最靠前的元素。**O(1)**。集合为空时返回 `nil`。
Score 最小；Score 相同时取最先插入的元素（seq 最小）。

#### `GetMax() *NodeData[K, V]`

返回排序最靠后的元素。**O(1)**。集合为空时返回 `nil`。
Score 最大；Score 相同时取最后插入的元素（seq 最大）。

---

#### `GetRank(key K) int`

查询 key 的**升序**排名（Score 最小为 rank=1）。O(log n)。Key 不存在时返回 `0`。

#### `GetRankDesc(key K) int`

查询 key 的**降序**排名（Score 最大为 rank=1）。O(log n)。Key 不存在时返回 `0`。

---

#### `GetByRank(rank int) *NodeData[K, V]`

按**升序**排名取元素。O(log n)。

| 情况 | 返回值 |
|------|-------|
| rank 在 `[1, Length]` 内 | 对应元素 |
| rank 超出范围 | `nil` |
| rank ≤ 0 | 触发断言 panic |

#### `GetByRankDesc(rank int) *NodeData[K, V]`

按**降序**排名取元素（rank=1 为 Score 最大）。O(log n)。返回值规则同上。

---

#### `GetRangeByRank(start, end int) []*NodeData[K, V]`

返回**升序**排名范围 `[start, end]` 内的元素。O(log n + k)，k 为返回元素数。

- 结果按 Score 从小到大排列
- `start > end` 时自动交换；`end` 超出总长度时截断，不 panic

#### `GetRangeByRankDesc(start, end int) []*NodeData[K, V]`

返回**降序**排名范围 `[start, end]` 内的元素。O(log n + k)，k 为返回元素数。

- 结果按 Score 从大到小排列
- `start > end` 时自动交换；`end` 超出总长度时截断，不 panic

```go
// 排行榜前 10 名
top10 := ss.GetRangeByRankDesc(1, 10)
```

---

#### `GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[K, V]`

返回 Score 在指定范围内的所有元素。O(log n + k)，k 为返回元素数。

- 结果按 Score 从小到大排列
- `minEx=false` 包含 min 端点，`minEx=true` 排除；`maxEx` 同理
- 范围内无元素时返回空切片，不 panic

```go
// 查询积分在 [1000, 2000) 之间的玩家
players := ss.GetRangeByScore(1000, false, 2000, true)
```

---

#### `CountByScore(min float64, minEx bool, max float64, maxEx bool) int`

统计 Score 在指定范围内的元素数量。O(log n + k)，k 为命中元素数。

- `minEx`/`maxEx` 语义与 `GetRangeByScore` 相同
- 与 `len(GetRangeByScore(...))` 结果相同，但只需 1 次结构体分配，避免结果切片的多次分配，适合高频计数场景

```go
// 统计积分在 (1000, 2000] 之间的玩家数量
count := ss.CountByScore(1000, true, 2000, false)
```

---

## 注意事项

- **Key 不可重复**：重复插入返回 `false`，不覆盖已有元素
- **相同 Score 按插入顺序稳定排序**：Score 相同的元素，先插入的升序 rank 更小（在跳表中靠前）。此顺序由内部序列号（`seq`）保证，`UpdateScore` 也不会改变它。
- **持久化后顺序可恢复**：存盘和恢复都应以**跳表升序 rank（`GetByRank(1)`、`GetByRank(2)`…）为基准**，而非逆序 rank。按升序 rank 存盘、按同样顺序 Insert 恢复，则恢复后的排名与存盘前完全一致——包括 Score 相同元素之间的相对顺序。逆序 rank（排行榜视图）只是访问视角，不影响底层存储顺序，因此即使业务侧通过 `GetByRankDesc` 展示排行榜，存盘和恢复仍应使用升序 rank 完成，两者不要混淆。
- **更新分数用 `UpdateScore`**，不要 `Delete + Insert`（会改变内部序列号，破坏稳定排序）
- **rank ≤ 0 触发 panic**：`GetByRank`、`GetByRankDesc` 的 rank 参数必须 ≥ 1
- **Score 不支持 NaN**：`Insert`、`UpdateScore`、`GetRangeByScore`、`DeleteRangeByScore` 对 NaN 触发 assert panic。原因：NaN 的比较语义（`NaN != NaN`、`NaN < NaN == false`）会破坏跳表依赖的全序关系，导致结构静默损坏。`±Inf` 是合法的 Score 值。
- **调用方负责保证输入合法性**：`rank ≤ 0` 和 NaN Score 等非法输入会触发 panic，assert 始终开启，无法关闭。
- **非并发安全**：多 goroutine 访问时需自行加锁
