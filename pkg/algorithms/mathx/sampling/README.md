# sampling 使用文档

随机采样算法，从整数范围中无偏均匀采样，支持不重复采样和带最小间隔约束的采样。

## API

### `SampleKDistinctFloyd[T Integer](lo, hi T, k int, r *rand.Rand) []T`

从 `[lo, hi]`（含端点）中均匀采样 **k 个不重复**的整数。

- 时间复杂度 O(k)，空间复杂度 O(k)
- 使用 Floyd 算法，避免 Fisher-Yates 需要完整数组的问题
- 返回的切片**无序**（不保证排序）

### `SampleWithMinGap[T Integer](L, R T, k, gap int, r *rand.Rand) []T`

从 `[L, R]` 中采样 **k 个不重复**的整数，相邻值之间间隔**至少 gap**。

- 返回**有序切片**（升序）
- 时间复杂度 O(k log k)，空间复杂度 O(k)
- 可行性前提：`(R - L + 1) - (k-1)*gap >= k`，不满足时触发断言 panic

## 使用示例

```go
import (
    "math/rand/v2"
    "github.com/motocat46/yytools/pkg/algorithms/mathx/sampling"
)

r := rand.New(rand.NewPCG(42, 0)) // 可重现的随机源

// 从 1~100 中采样 5 个不重复的数
nums := sampling.SampleKDistinctFloyd(1, 100, 5, r)
// 例如：[73, 12, 58, 3, 91]（无序）

// 从 1~50 中采样 4 个，相邻至少间隔 3
nums = sampling.SampleWithMinGap(1, 50, 4, 3, r)
// 例如：[2, 8, 15, 24]（有序，相邻差 ≥ 3）
```

## 适用场景

- 抽奖、随机关卡生成：从 ID 范围中抽取 k 个不重复的奖励/关卡
- 生成随机题目集合：从题库编号范围采样
- 带间隔约束的资源分配：如随机放置敌人，保证不过于密集

## 注意事项

- `k > hi-lo+1` 时触发断言 panic（采样数量超过范围）
- `SampleWithMinGap` 的间隔约束不满足时触发断言 panic，调用前需自行验证可行性
- `r *rand.Rand` 传入调用方管理的随机源，便于测试复现和多 goroutine 隔离
- **平台限制**：`int(hi) - int(lo)` 使用 `int` 中间类型，在 64 位平台（`int` = 64 位）范围跨度最大为 `math.MaxInt64`；**32 位平台不支持跨度超过 `math.MaxInt32` 的范围**
