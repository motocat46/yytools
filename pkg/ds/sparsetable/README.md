# sparsetable

泛型稀疏表（Sparse Table），支持 O(n log n) 预处理后 **O(1)** 区间幂等查询。
适用于静态数据的区间 min/max/GCD 查询；有更新需求请用 `pkg/ds/segtree`。

非并发安全。

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/ds/sparsetable"

data := []int{3, 1, 4, 1, 5, 9, 2, 6}

// 区间最小值
st := sparsetable.New(data, func(a, b int) int {
	if a < b {
		return a
	}
	return b
})
fmt.Println(st.Query(2, 5)) // 1（data[2..5] = [4,1,5,9] 的最小值）
fmt.Println(st.Query(0, 7)) // 1

// 区间 GCD
stGCD := sparsetable.New([]int{12, 8, 6, 4}, func(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
})
fmt.Println(stGCD.Query(0, 3)) // 2
```

## API

### `New[T any](data []T, merge func(T, T) T) *SparseTable[T]`

构建 Sparse Table。

- `data`：源数据，构建后不可修改；`len(data) == 0` 时 assert panic
- `merge`：必须满足**幂等性**：`merge(x, x) == x`（min/max/gcd 均满足；sum 不满足）

时间：O(n log n)，空间：O(n log n)。

### `Query(l, r int) T`

返回区间 `[l, r]` 的 merge 结果（闭区间，0-indexed）。O(1)。

- `l > r` 或越界时 assert panic；调用前用 `Len()` 验证下标

### `Len() int`

返回元素数量 n。O(1)。

## 幂等性约束

Sparse Table 的 O(1) 查询依赖区间重叠：

```text
Query(l, r) = merge(st[k][l], st[k][r - 2^k + 1])
```

两段区间重叠，若 merge 不满足 `merge(x, x) == x`（如 sum），重叠部分会被计算两次，结果错误。

| merge | 幂等？ | 可用 Sparse Table |
|-------|--------|------------------|
| min   | ✓      | ✓                |
| max   | ✓      | ✓                |
| gcd   | ✓      | ✓                |
| sum   | ✗      | ✗（用 FenwickTree）|
| xor   | ✗      | ✗                |

## 与其他区间结构对比

| 结构 | Build | Query | Update | 适用 |
|------|-------|-------|--------|------|
| FenwickTree | O(n) | O(log n) | O(log n) | 动态前缀和 |
| SegTree | O(n) | O(log n) | O(log n) | 动态区间更新+查询 |
| **SparseTable** | O(n log n) | **O(1)** | ❌ | 静态区间 min/max/GCD |

## 注意事项

- 构建后数据不可修改；若数据会变化，改用 SegTree
- 非并发安全，并发访问由调用方加锁
- 对于 n=1M，Build 约需 ~100ms（Go arm64），Query 约 10ns/op
