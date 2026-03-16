# overflow 使用文档

对有符号整数四则运算进行溢出检测，解决 Go 整数溢出**静默**（无报错、无 panic）的问题。

## API

每种运算提供两个版本：

| 函数 | 行为 |
|------|------|
| `AddInt[T Signed](a, b T) (T, bool)` | 计算 a+b，bool=true 表示溢出 |
| `AddIntAssert[T Signed](a, b T) T` | 溢出时触发断言 panic |
| `SubInt[T Signed](a, b T) (T, bool)` | 计算 a-b，bool=true 表示溢出 |
| `SubIntAssert[T Signed](a, b T) T` | 溢出时触发断言 panic |
| `MulInt[T Signed](a, b T) (T, bool)` | 计算 a*b，bool=true 表示溢出 |
| `MulIntAssert[T Signed](a, b T) T` | 溢出时触发断言 panic |
| `DivInt[T Signed](a, b T) (T, bool)` | 计算 a/b，bool=true 表示溢出（仅 MinT/-1 情形）|
| `DivIntAssert[T Signed](a, b T) T` | 溢出时触发断言 panic |

类型约束：`T` 满足 `base.Signed`（有符号整数）。

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/algorithms/mathx/overflow"

// 检测版本：由调用方处理溢出
result, ovf := overflow.MulInt(int64(1e18), int64(10))
if ovf {
    return fmt.Errorf("计算溢出")
}

// 断言版本：开发期快速失败（生产需关闭 assert）
days := overflow.MulIntAssert(int64(time.Hour)*24, int64(365))

// 除法溢出（MinInt64 / -1）
res, ovf := overflow.DivInt(math.MinInt64, int64(-1)) // ovf=true
```

## 适用场景

- 时间戳计算（乘以天数可能溢出 int64）
- 金融计算（金额相乘）
- 任何涉及较大整数的乘加运算

## 注意事项

- 仅支持**有符号**整数，无符号整数溢出语义不同（模运算）
- 除以 0 会触发 Go 运行时 panic，本包不处理
