# binary_search 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、五个函数的分工

| 函数 | 重复元素时的行为 | 典型用途 |
|------|--------------|---------|
| `BinarySearch` | 返回任意匹配下标 | 只需确认"存在/不存在" |
| `LeftBound` | 返回最左下标 | 统计某值在排序数组中的起始位置 |
| `RightBound` | 返回最右下标 | 统计某值的结束位置 |
| `SearchBound` | 同时返回左右边界 | 获取重复元素的完整范围 |
| `SearchBoundOpt` | 同上，但复用左边界缩小右侧搜索 | `SearchBound` 的性能优化版 |

`SearchBoundOpt` 与 `SearchBound` 的区别：`SearchBound` 先调 `LeftBound`（全范围搜索），再调 `RightBound`（全范围搜索）。`SearchBoundOpt` 找到左边界后，把左边界作为右侧搜索的起点，减少约一半比较次数。两者语义完全相同。

---

## 二、边界处理：闭区间 `[left, right]`

所有函数统一使用**双闭区间**：

- 循环条件：`left <= right`（两端都包含，允许单元素区间）
- mid 计算：`left + (right-left)/2`（避免 `left+right` 溢出）
- 缩小边界时：`right = mid-1` 或 `left = mid+1`（排除已知不匹配的 mid）

**为什么不用半开区间 `[left, right)`？**

闭区间的终止条件 `left > right` 更直觉：循环结束意味着搜索范围为空。半开区间 `left == right` 时区间仍非空（包含一个元素），终止条件变成 `left >= right`，与区间的语义分离，容易混淆。

---

## 三、LeftBound / RightBound 的关键差异

两者在命中 `target == nums[mid]` 时的行为相反：

- `LeftBound`：命中后继续向左收缩（`right = mid-1`），寻找更左的位置
- `RightBound`：命中后继续向右收缩（`left = mid+1`），寻找更右的位置

循环结束后：
- `LeftBound`：检查 `left` 是否越界且 `nums[left] == target`
- `RightBound`：检查 `right` 是否越界且 `nums[right] == target`

源码注释解释了为什么 `LeftBound` 不需要检查 `left < 0`（left 只做加法），但为统一起见仍然保留该检查。

---

## 四、仅支持 `base.Integer`，不支持浮点

与 `sort` 包理由相同：浮点数的 NaN 破坏全序关系，二分搜索依赖严格全序。整数没有这个问题。浮点数排序/搜索请使用标准库 `slices`。
