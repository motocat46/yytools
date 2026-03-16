# binary_search 使用文档

对**非降序整数数组**进行二分搜索，支持精确查找和边界查找。

**前提条件**：输入数组必须已按非降序排列（允许重复元素）。

## API

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `BinarySearch[T](nums []T, target T) int` | 精确查找 | 找到返回下标，未找到返回 -1 |
| `LeftBound[T](nums []T, target T) int` | 查找最左边界 | 找到返回下标，未找到返回 -1 |
| `RightBound[T](nums []T, target T) int` | 查找最右边界 | 找到返回下标，未找到返回 -1 |
| `SearchBound[T](nums []T, target T) (int, int)` | 同时查找左右边界 | (左边界, 右边界)，未找到返回 (-1, -1) |
| `SearchBoundOpt[T](nums []T, target T) (int, int)` | SearchBound 优化版 | 同上，找到左边界后收缩右侧搜索范围 |

类型约束：`T` 满足 `base.Integer`。

## 使用示例

```go
import bs "github.com/motocat46/yytools/pkg/algorithms/binary_search"

nums := []int{1, 2, 2, 3, 3, 3, 4, 5}

// 精确查找
idx := bs.BinarySearch(nums, 3)   // 返回 3、4 或 5 中的某个（无重复时精确）
idx  = bs.BinarySearch(nums, 99)  // -1

// 查找重复元素的范围
left  := bs.LeftBound(nums, 3)   // 3（第一个 3 的下标）
right := bs.RightBound(nums, 3)  // 5（最后一个 3 的下标）

// 同时获取左右边界
l, r := bs.SearchBound(nums, 3)      // (3, 5)
l, r  = bs.SearchBound(nums, 99)     // (-1, -1)

// 优化版（左边界找到后缩小右侧搜索范围）
l, r = bs.SearchBoundOpt(nums, 3) // (3, 5)
```

## 注意事项

- 输入数组**必须非降序**，否则结果不可预期
- `BinarySearch` 在有重复元素时返回的是其中任意一个匹配下标，不保证最左或最右
- 需要确定重复元素范围时，用 `SearchBound` 或 `SearchBoundOpt`
- `SearchBoundOpt` 相比 `SearchBound` 在左边界已确定后进一步收缩搜索，性能略好
