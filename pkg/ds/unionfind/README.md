# pkg/ds/unionfind

泛型并查集（Disjoint Set Union）。支持任意 `comparable` 类型元素，提供 O(α) 均摊的 Union/Find/Connected/Size/Count。元素首次使用时自动注册为单独组，无需显式 Add。非并发安全。

## 快速上手

```go
u := unionfind.New[string]()

// 合并两个组
u.Union("alice", "bob")   // true — 发生了合并
u.Union("bob", "alice")   // false — 已在同一组

// 查询连通性
u.Connected("alice", "bob")  // true
u.Connected("alice", "eve")  // false（eve 自动注册为新单独组）

// 组的代表元
u.Find("alice")  // "alice" 或 "bob"（同组元素 Find 结果相同）

// 统计
u.Count()        // 2（alice+bob 一组，eve 一组）
u.Size("alice")  // 2
u.Size("eve")    // 1
```

## API

```go
// New 创建空的并查集。
func New[T comparable]() *UnionFind[T]

// Union 合并 a 和 b 所在的组。
// 返回 true 表示发生了实际合并（两者原本不在同一组）；
// 返回 false 表示已在同一组，无操作。
// a、b 若未注册，自动注册为单独组后再合并。
func (uf *UnionFind[T]) Union(a, b T) bool

// Find 返回 a 所在组的代表元（根节点）。O(α) 均摊（迭代路径压缩）。
// a 若未注册，自动注册为单独组，返回 a 自身。
func (uf *UnionFind[T]) Find(a T) T

// Connected 报告 a 和 b 是否在同一组。O(α) 均摊。
// a、b 若未注册，自动注册为单独组（两个新单独组必然不连通，返回 false）。
func (uf *UnionFind[T]) Connected(a, b T) bool

// Count 返回当前独立组的数量。O(1)。
func (uf *UnionFind[T]) Count() int

// Size 返回 a 所在组的元素数量。O(α) 均摊。
// a 若未注册，自动注册为单独组，返回 1。
func (uf *UnionFind[T]) Size(a T) int
```

## 使用场景

- **组队 / 公会系统**：玩家加入公会后，查询任意两个玩家是否同公会
- **消消乐连通分量**：合并相邻同色格子，统计连通块大小
- **程序化地图生成（Kruskal 迷宫）**：随机连通房间，避免成环
- **图的连通性判断**：实时判断两个节点是否属于同一连通分量

## 注意事项

- **非并发安全**：并发访问由调用方负责加锁
- **不支持删除元素**：Union-Find 经典不支持 Remove
- **不支持枚举组成员**：无 Members() 方法，调用方自行维护映射
- **代表元不稳定**：路径压缩后 Find 的返回值可能变化，不应依赖具体的根节点身份，只依赖"同组元素 Find 结果相同"这一不变量

## 复杂度

| 操作 | 时间（均摊） | 空间 |
|------|------------|------|
| Union | O(α(n)) | O(1) |
| Find | O(α(n)) | O(1) |
| Connected | O(α(n)) | O(1) |
| Count | O(1) | O(1) |
| Size | O(α(n)) | O(1) |
| 整体空间 | — | O(n) |

α(n) 为反阿克曼函数，实际使用中 ≤ 4，视为常数。

## 性能基准（Apple M4）

| 操作 | n=100 | n=10k | n=1M |
|------|-------|-------|------|
| Union | ~40 ns | ~45 ns | ~136 ns |
| Find | ~22 ns | ~24 ns | ~67 ns |
| Connected | ~42 ns | ~48 ns | ~133 ns |
| Mixed | ~45 ns | ~47 ns | ~148 ns |

所有操作 0 allocs/op。

## 运行测试

```bash
go test ./pkg/ds/unionfind/
go test -race ./pkg/ds/unionfind/
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/unionfind/
```
