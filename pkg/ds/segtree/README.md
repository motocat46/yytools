# pkg/ds/segtree

ACL 风格泛型线段树（Segment Tree with Lazy Propagation）。支持两个独立类型参数 `T`（节点值）和 `L`（lazy 标记），调用方注入 merge/apply/compose 函数，一次实现覆盖 sum/min/max/GCD 等任意 monoid 场景。提供 O(log n) 的单点赋值、区间 lazy 更新和区间查询；对外 0-indexed，非并发安全。

## 快速上手

### 区间加法 + 区间求和

```go
s := segtree.New[int, int](5, 0,
    func(a, b int) int { return a + b },       // merge
    0,                                          // lazyZero
    func(val, lazy, size int) int { return val + lazy*size }, // apply
    func(newL, oldL int) int { return newL + oldL },          // compose
)

for i, v := range []int{1, 2, 3, 4, 5} {
    s.Set(i, v)
}
s.Apply(1, 3, 10)        // [1, 12, 13, 14, 5]
s.Query(1, 3)            // 39
s.QueryAll()             // 45
```

### 区间赋值 + 区间最小值

```go
type Assign struct{ val int; hasVal bool }

s := segtree.New[int, Assign](5, 1<<62,
    func(a, b int) int {
        if a < b { return a }
        return b
    },
    Assign{},
    func(val int, lazy Assign, _ int) int {
        if lazy.hasVal { return lazy.val }
        return val
    },
    func(newL, oldL Assign) Assign {
        if newL.hasVal { return newL }
        return oldL
    },
)

for i, v := range []int{5, 3, 8, 1, 7} {
    s.Set(i, v)
}
s.Apply(1, 3, Assign{val: 4, hasVal: true}) // [5,4,4,4,7]
s.Query(0, 4)                               // 4
```

## API

```go
func New[T, L any](
    n        int,
    identity T,
    merge    func(T, T) T,
    lazyZero L,
    apply    func(T, L, int) T,
    compose  func(L, L) L,
) *SegTree[T, L]

func (s *SegTree[T, L]) Set(i int, val T)
func (s *SegTree[T, L]) Apply(l, r int, lazy L)
func (s *SegTree[T, L]) Query(l, r int) T
func (s *SegTree[T, L]) QueryAll() T
func (s *SegTree[T, L]) Len() int
```

## 参数约束（调用方须保证）

| 约束 | 说明 |
|------|------|
| `merge(identity, x) == x` | identity 是 merge 的单位元 |
| `apply(val, lazyZero, size) == val` | lazyZero 是 apply 的单位元 |
| `compose(lazyZero, x) == x` | lazyZero 是 compose 的左单位元 |
| `merge` 满足结合律 | 线段树正确性依赖 |

违反上述约束行为未定义。

## 复杂度

| 操作 | 时间 | allocs/op |
|------|------|-----------|
| New | O(n) | O(n) |
| Set | O(log n) | 0 |
| Apply | O(log n) | 0 |
| Query | O(log n) | 0 |
| QueryAll | O(1) | 0 |

## 注意事项

- 非并发安全，并发访问须调用方加锁
- 容量固定，构造时确定，不支持动态扩容
- `apply` 第三个参数为区间长度；sum 场景须乘以 size，max/min 场景可忽略
- `compose(newL, oldL)` 新 lazy 在前、旧 lazy 在后
