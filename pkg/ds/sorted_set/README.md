# sorted_set

基于跳表（Skip List）+ 哈希表实现的有序集合，语义类似 Redis ZADD/ZRANK/ZRANGE。

- O(log n) 插入 / 删除 / 排名查询
- O(1) 按 Key 查找
- 相同 Score 的元素按插入顺序稳定排序

## 核心概念

每个元素由三部分组成：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Key` | `T` | 唯一标识，用于哈希表快速查找（comparable 约束） |
| `Score` | `float64` | 排序依据，分数越小排名越靠前（rank=1 为最小值） |
| `Val` | `T` | 业务数据 |

## API

| 方法 | 说明 |
|------|------|
| `NewSortedSet[T comparable]()` | 创建有序集合 |
| `NewNodeData(key T, score float64, val T)` | 创建节点数据 |
| `Insert(data *NodeData[T]) bool` | 插入元素，Key 重复返回 false |
| `Delete(key T) (*NodeData[T], bool)` | 按 Key 删除，不存在返回 (nil, false) |
| `Get(key T) *NodeData[T]` | 按 Key 查找，不存在返回 nil |
| `Length() int` | 元素总数 |
| `GetRank(key T) int` | 按 Key 查询排名（从 1 开始），不存在返回 0 |
| `GetByRank(rank int) *NodeData[T]` | 按排名查询元素，超出范围返回 nil |
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
ss.GetRank("bob")   // 1（最低分）
ss.GetRank("alice") // 2
ss.GetRank("carol") // 3（最高分）

// 按排名获取
ss.GetByRank(1) // bob 的 NodeData

// 获取排名 1~2
ss.GetRangeByRank(1, 2) // [bob, alice]

// 更新分数后自动重排序
ss.UpdateScore("alice", 2000.0)
ss.GetRank("alice") // 3

// 按分数范围查询 [1000, 1500)（含左不含右）
ss.GetRangeByScore(1000, false, 1500, true)

// 按 Key 删除
ss.Delete("bob")
```

## 排名说明

排名从 **1** 开始，rank=1 对应**分数最小**的元素（升序）。如需实现"排行榜"语义（高分靠前），存入负分数即可：`score = -actualScore`。

## 注意事项

- Key 不可重复，重复插入返回 false（不覆盖已有元素）
- 更新分数使用 `UpdateScore`，不要删除再插入（会改变内部 seq 影响稳定排序）
- `GetByRank` 的 rank 必须 > 0，否则触发断言 panic
- 非并发安全，多 goroutine 使用时需自行加锁
