# sort 使用文档

多种排序算法实现，支持整数类型（`base.Integer`）的切片原地排序。

## 算法列表

| 算法 | 函数 | 时间复杂度 | 适用场景 |
|------|------|-----------|---------|
| 快速排序（升序）| `QuickSort[T]` | 平均 O(n log n) | 通用，大数据量 |
| 快速排序（降序）| `QuickSortDesc[T]` | 平均 O(n log n) | 通用，大数据量 |
| 快速排序-迭代版（升序）| `QuickSortTraversal[T]` | 平均 O(n log n) | 避免栈溢出的超大数组 |
| 快速排序-迭代版（降序）| `QuickSortDescTraversal[T]` | 平均 O(n log n) | 避免栈溢出的超大数组 |
| 冒泡排序 | `BubbleSort[T]` | O(n²) | 教学、极小数组 |
| 插入排序 | `InsertionSort[T]` | O(n²)，近乎有序时接近 O(n) | 小数组（≤12）或近乎有序 |
| 计数排序 | `CountingSort[T]` | O(n+k) | 元素值域集中的大数组 |
| 基数排序 | `RadixSort[T]` | O(n·k) | 元素值域较大的整数排序 |

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/algorithms/sort"

nums := []int{5, 3, 8, 1, 9, 2}

sort.QuickSort(nums)      // [1 2 3 5 8 9]
sort.QuickSortDesc(nums)  // [9 8 5 3 2 1]

// 值域集中时，计数排序更快
scores := []int{85, 92, 78, 95, 88}
sort.CountingSort(scores) // [78 85 88 92 95]
```

## 算法选型建议

- **通用场景**：用 `QuickSort`，性能与标准库 `slices.Sort` 接近
- **重复元素多**：`QuickSort` 采用三路划分（荷兰国旗），大量重复时 O(n²) 退化得到优化
- **数组元素值域 [min, max] 差值 ≤ 10^6 量级**：用 `CountingSort`，O(n+k) 比快排快
- **小数组（≤12 元素）**：快排内部自动切换插入排序，无需手动选择

## 类型约束

所有排序函数要求 `T` 满足 `base.Integer`（有符号或无符号整数及其自定义底层类型）。浮点类型不支持，请使用标准库 `slices.Sort`。

## 注意事项

- 所有排序均为**原地排序**（in-place），会修改传入的切片
- `CountingSort` 的额外内存开销为 `O(max - min + 1)`，值域过大时不适用
- `QuickSort` 使用随机化 pivot，避免有序输入退化为 O(n²)
