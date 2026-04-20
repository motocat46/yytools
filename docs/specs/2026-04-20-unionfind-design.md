# Union-Find（并查集）设计文档

**日期**：2026-04-20
**状态**：已批准

---

## 背景

Go 生态中不存在成熟的泛型 Union-Find 独立库（所有已知实现 < 20 stars，emirpasic/gods 16k stars 亦不含此结构）。Union-Find 在游戏开发中有真实场景：组队/公会系统、消消乐连通分量、程序化地图生成（Kruskal 迷宫）、围棋气判断等。

---

## 目标

实现 `pkg/ds/unionfind`——固定 API 的泛型并查集，支持任意 `comparable` 类型元素，提供 O(α) 均摊的 Union/Find/Connected/Size/Count。

---

## API

```go
// New 创建空的并查集。
func New[T comparable]() *UnionFind[T]

// Union 合并 a 和 b 所在的组。
// 返回 true 表示发生了实际合并（两者原本不在同一组）；
// 返回 false 表示已在同一组，无操作。
// a、b 若未注册，自动注册为单独组后再合并。
func (uf *UnionFind[T]) Union(a, b T) bool

// Find 返回 a 所在组的代表元（根节点）。O(α) 均摊（路径压缩）。
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

---

## 内部实现

### 数据结构

```go
type UnionFind[T comparable] struct {
    parent map[T]T   // parent[x] = x 的父节点；根节点的父节点是自身
    size   map[T]int // size[x] 仅在 x 为根时有意义，表示组大小
    count  int       // 当前独立组数
}
```

### 算法

**路径压缩（Find）**：递归查找根，返回途中将所有中间节点的 parent 直接指向根，使后续查找更快。

```
Find(x):
  if parent[x] != x:
    parent[x] = Find(parent[x])   // 路径压缩
  return parent[x]
```

**按大小合并（Union）**：将小树的根接到大树的根下，保持树高度平缓。

```
Union(a, b):
  ra = Find(a); rb = Find(b)
  if ra == rb: return false
  if size[ra] < size[rb]: swap(ra, rb)
  parent[rb] = ra
  size[ra] += size[rb]
  count--
  return true
```

**自动注册（register）**：
```
register(x):
  if x not in parent:
    parent[x] = x
    size[x] = 1
    count++
```
所有公开方法在访问元素前调用 `register`。

### 复杂度

| 操作 | 时间 | 空间 |
|------|------|------|
| Union | O(α(n)) 均摊 | O(1) |
| Find | O(α(n)) 均摊 | O(1) |
| Connected | O(α(n)) 均摊 | O(1) |
| Count | O(1) | O(1) |
| Size | O(α(n)) 均摊 | O(1) |
| 整体空间 | — | O(n) |

α(n) 为反阿克曼函数，实际使用中 ≤ 4，视为常数。

---

## 正确性命题

| 命题 | 验证方式 |
|------|---------|
| **Union 后连通**：`Union(a,b)` 后 `Connected(a,b) == true` | 随机 Union 序列后逐对验证 |
| **Find 一致性**：同组元素 `Find` 结果相同 | 同上，验证同组所有元素 Find 相同 |
| **Count 精确**：等于当前独立组数 | 与参考模型（`map[T]struct{}` 组集合）对比 |
| **Size 精确**：等于组内元素数 | 与参考模型对比 |
| **自动注册**：未见过的元素操作后 Count 增加 1，Size 为 1 | 单元测试覆盖 |
| **传递连通**：a-b 连通、b-c 连通 → a-c 连通 | 随机链式 Union 后验证传递性 |

---

## 受影响文档

| 文档 | 操作 |
|------|------|
| `pkg/ds/unionfind/README.md` | 新建 |
| `pkg/ds/README.md` | 新增 unionfind 行 |
| `README.md` | 目录树 + 模块索引更新 |

---

## 不在范围内

- `Members(a T) []T`（枚举组成员）：O(N) 扫描，调用方自行实现
- `Remove(a T)`（删除元素）：Union-Find 经典不支持删除
- 并发安全：非并发安全，调用方加锁
- 持久化 / 序列化
