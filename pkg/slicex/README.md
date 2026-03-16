# slicex 使用文档

切片工具函数，提供查找最小/最大值及其下标的功能。相比标准库 `slices.Min/Max`，额外返回元素下标，并支持自定义比较规则。

## API

### 可比较类型（T 满足 Ordered）

| 函数 | 空切片行为 | 返回值 |
|------|-----------|--------|
| `MinInSlice[T](s []T) (int, T)` | 触发断言 panic | 最小值下标、最小值 |
| `MaxInSlice[T](s []T) (int, T)` | 触发断言 panic | 最大值下标、最大值 |
| `MinInSliceOK[T](s []T) (int, T, bool)` | ok=false | 最小值下标、最小值、是否找到 |
| `MaxInSliceOK[T](s []T) (int, T, bool)` | ok=false | 最大值下标、最大值、是否找到 |

### 自定义比较规则（T 为任意类型）

| 函数 | 空切片行为 | 说明 |
|------|-----------|------|
| `MinBy[T](s []T, better func(a,b T) bool) (int, T)` | 触发断言 panic | better(a,b)=true 表示 a 优于 b |
| `MaxBy[T](s []T, better func(a,b T) bool) (int, T)` | 触发断言 panic | 同上 |
| `MinByOK / MaxByOK` | ok=false | 对应的安全版本 |

所有函数均保证**稳定**：存在多个并列最优值时，返回**第一个**的下标。

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/slicex"

// 基本用法
nums := []int{3, 1, 4, 1, 5, 9}
idx, val := slicex.MinInSlice(nums)  // idx=1, val=1
idx, val  = slicex.MaxInSlice(nums)  // idx=5, val=9

// 安全版本（可能为空的切片）
idx, val, ok := slicex.MinInSliceOK(nums)
if !ok {
    // 切片为空
}

// 自定义结构体
type Player struct {
    Name  string
    Score int
}
players := []Player{{"Alice", 90}, {"Bob", 85}, {"Carol", 95}}

// 找分数最高的玩家
idx, best := slicex.MaxBy(players, func(a, b Player) bool {
    return a.Score > b.Score
})
// idx=2, best={Carol, 95}

// 找名字字典序最小的玩家
idx, first := slicex.MinBy(players, func(a, b Player) bool {
    return a.Name < b.Name
})
// idx=0, first={Alice, 90}
```

## 注意事项

- 切片含 NaN 时不满足全序语义（与 Go 内置 `<`、`>` 行为一致）
- 需要处理空切片时，使用 `*OK` 版本；确定非空时用无后缀版本更简洁
