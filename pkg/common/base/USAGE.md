# base 使用文档

提供跨模块复用的泛型类型约束，用于编写类型安全的泛型函数。

## 约束列表

| 约束 | 包含类型 |
|------|---------|
| `Signed` | ~int、~int8、~int16、~int32、~int64 |
| `Unsigned` | ~uint、~uint8、~uint16、~uint32、~uint64、~uintptr |
| `Integer` | Signed \| Unsigned |
| `Float` | ~float32、~float64 |
| `Number` | Integer \| Float |
| `Ordered` | = cmp.Ordered（string、integer、float） |

`~T` 语义：不仅匹配 `T` 本身，还匹配**底层类型为 T 的自定义类型**。例如 `type Age int` 满足 `Signed`。

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/common/base"

// 对任意整数类型求绝对值
func Abs[T base.Integer](a T) T {
    if a < 0 {
        return -a
    }
    return a
}

// 对任意可比较类型取最大值
func Max[T base.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

// 自定义类型也满足约束
type Score int
var s Score = Abs(Score(-42)) // 编译通过
```

## 说明

- 这是内部基础包，供 yytools 其他模块使用
- `Ordered` 直接复用 `cmp.Ordered`，与标准库保持一致
