# Design

## 核心选型：level-major 二维数组 + bits.Len 求 k

### 问题

静态区间 min/max/GCD：预处理一次，O(1) 查询。

### 方案

`st[j][i]` = merge(data[i..i+2^j-1])，两段区间重叠覆盖目标区间，幂等性保证重叠不影响结果。

**bits.Len 替代 log2 查找表**：`bits.Len(uint(r-l+1)) - 1` 单条 BSR/CLZ 指令，零分配，比 `[]int` 查找表更简洁，比 `math.Log2` 更快。

**level-major 布局 `st[j][i]`**（vs element-major `st[i][j]`）：
- Build：外层 j，内层 i → 顺序写 `st[j]`，cache 友好
- Query：访问 `st[k][l]` 和 `st[k][r-2^k+1]`，同行两次随机读，仍 cache 友好
- element-major 的 build 需要跨列访问，cache miss 更多

### 关键约束

**merge 必须满足幂等性**：`merge(x, x) == x`。Query 将目标区间分成两段重叠区间，重叠部分 merge 两次，幂等性保证结果不变。违反时（如 sum）结果悄然错误，API 注释中明确标注。

## 复杂度

| 操作 | 时间 | 空间 |
|------|------|------|
| New/Build | O(n log n) | O(n log n) |
| Query | O(1) | O(1) |

## 与 SegTree/FenwickTree 的定位

三者共同构成“区间查询三件套”：
- FenwickTree：动态前缀和，O(log n) 更新+查询，最轻量
- SegTree：动态区间任意操作，O(log n) 更新+查询，最灵活
- SparseTable：静态幂等操作，O(1) 查询，最快（但不支持更新）
