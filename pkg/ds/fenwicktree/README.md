# pkg/ds/fenwicktree

泛型树状数组（Fenwick Tree / Binary Indexed Tree）。支持任意 `base.Number` 类型，提供 O(log n) 的单点更新、前缀和及区间和查询。0-indexed 接口，内部 1-indexed。非并发安全。

## 快速上手

```go
f := fenwicktree.New[int](5) // 管理 5 个元素，下标 0..4

// 单点更新
f.Add(0, 10)
f.Add(1, 20)
f.Add(2, 30)

// 前缀和：[0..2] = 60
f.PrefixSum(2) // 60

// 区间和：[1..2] = 50
f.RangeSum(1, 2) // 50

// 容量
f.Len() // 5

f2 := fenwicktree.Build([]int{1, 2, 3, 4, 5})
f2.PrefixSum(4) // 15
```

## API

```go
func New[T base.Number](n int) *FenwickTree[T]
func Build[T base.Number](nums []T) *FenwickTree[T]
func (f *FenwickTree[T]) Add(i int, delta T)
func (f *FenwickTree[T]) PrefixSum(i int) T
func (f *FenwickTree[T]) RangeSum(l, r int) T
func (f *FenwickTree[T]) Len() int
```

## 使用场景

- **排行榜区间计数**：快速统计 `[rankL, rankR]` 内的玩家数
- **累积概率**：构建概率前缀数组，O(log n) 查询累积概率
- **动态频率统计**：实时更新元素频率并查询前 k 个的总和

## 注意事项

- **非并发安全**：并发访问由调用方加锁
- **不支持区间更新**（RangeAdd）：需要区间更新请使用 Segment Tree
- **固定容量**：构造时确定 `n`，不支持动态扩容
- **越界 panic**：`Add`/`PrefixSum` 要求 `i ∈ [0, n)`；`RangeSum` 要求 `l ≤ r` 且均 `∈ [0, n)`

## 复杂度

| 操作 | 时间 | 空间 | allocs |
|------|------|------|--------|
| Build | O(n) | O(n) | 0 |
| Add | O(log n) | O(1) | 0 |
| PrefixSum | O(log n) | O(1) | 0 |
| RangeSum | O(log n) | O(1) | 0 |
| Len | O(1) | O(1) | 0 |

## 运行测试

```bash
go test ./pkg/ds/fenwicktree/
go test -race ./pkg/ds/fenwicktree/
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/fenwicktree/
```
