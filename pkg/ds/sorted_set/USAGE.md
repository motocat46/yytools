# sorted_set 使用文档

基于跳表（Skip List）+ 哈希表实现的有序集合，语义类似 Redis ZADD/ZRANK/ZRANGE。

支持按 `Score`（分数）排序，O(log n) 插入/删除/查询，O(1) 按 Key 查找。

## 核心概念

每个元素由三部分组成：
- `Key T`：唯一标识，用于哈希表快速查找（comparable 约束）
- `Score float64`：排序依据，分数越小排名越靠前（rank=1 为最小值）
- `Val T`：业务数据（当前实现中与 Key 同类型）

相同 Score 的元素按插入顺序稳定排序。

## API

| 方法 | 说明 |
|------|------|
| `NewSortedSet[T comparable]()` | 创建有序集合 |
| `Insert(data *NodeData[T]) bool` | 插入元素，Key 重复返回 false |
| `Delete(key T) (*NodeData[T], bool)` | 按 Key 删除，不存在返回 false |
| `Get(key T) *NodeData[T]` | 按 Key 查找，不存在返回 nil |
| `Length() int` | 元素总数 |
| `GetRank(key T) int` | 按 Key 查询排名（从 1 开始），不存在返回 0 |
| `GetByRank(rank int) *NodeData[T]` | 按排名查询元素（rank 从 1 开始） |
| `GetRangeByRank(start, end int) []*NodeData[T]` | 按排名范围查询（含端点，自动处理 start>end）|
| `DeleteRangeByRank(start, end int) []*NodeData[T]` | 按排名范围删除 |
| `UpdateScore(key T, newScore float64) (*NodeData[T], bool)` | 更新分数，触发重新排序 |
| `GetRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[T]` | 按分数范围查询 |
| `DeleteRangeByScore(min float64, minEx bool, max float64, maxEx bool) []*NodeData[T]` | 按分数范围删除 |

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/ds/sorted_set"

ss := sorted_set.NewSortedSet[string]()

// 插入玩家分数
ss.Insert(sorted_set.NewNodeData("alice", 1500.0, "alice"))
ss.Insert(sorted_set.NewNodeData("bob",   1200.0, "bob"))
ss.Insert(sorted_set.NewNodeData("carol", 1800.0, "carol"))

// 查询排名（分数越小排名越靠前）
ss.GetRank("bob")   // 1（最低分，排名第 1）
ss.GetRank("alice") // 2
ss.GetRank("carol") // 3（最高分，排名最后）

// 按排名获取
ss.GetByRank(1) // {bob, 1200.0, "bob"}

// 获取前 2 名
ss.GetRangeByRank(1, 2) // [{bob,...}, {alice,...}]

// 更新分数（alice 超越 carol）
ss.UpdateScore("alice", 2000.0)
ss.GetRank("alice") // 3

// 按分数范围查询 [1000, 1500)（含左不含右）
ss.GetRangeByScore(1000, false, 1500, true) // [{bob, 1200}, ...]

// 按 Key 删除
ss.Delete("bob")
```

## 排名说明

排名从 **1** 开始，rank=1 对应**分数最小**的元素（升序）。如需实现"排行榜"语义（高分靠前），建议存入负分数：`score = -actualScore`。

## 注意事项

- Key 不可重复，重复插入返回 false（不覆盖）
- 需要更新分数时使用 `UpdateScore`，不要删除再插入（会改变 seq 影响稳定性）
- `GetByRank` 的 rank 必须 > 0，否则触发断言 panic
- 非并发安全，多 goroutine 使用时需自行加锁
