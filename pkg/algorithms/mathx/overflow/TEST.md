# overflow 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `overflow_test.go` | MulInt、DivInt、AddInt、SubInt 及其 Assert 变体，覆盖 int8/int16/int32/int64 |
| `example_test.go` | 可运行的使用示例（`go test -run Example`） |

## 测试设计要点

### 覆盖的关键场景

每个运算均覆盖：正常值、零值、边界值（MaxT/MinT 各类型）、溢出情形。

| 运算 | 关键测试点 |
|------|-----------|
| `MulInt` | `MaxT * 2`（正向溢出）、`MinT * 2`（负向溢出）、`MinInt64 * -1`（int64 特有溢出路径）、零乘任意数 |
| `DivInt` | `MinT / -1`（唯一溢出情形）、正常除法、零被除、负数除法四象限 |
| `AddInt` | `MaxT + 1`（正向溢出）、`MinT + (-1)`（负向溢出）、一正一负（永不溢出）|
| `SubInt` | `MinT - 1`（负向溢出）、`MaxT - (-1)`（正向溢出）|

### Assert 变体
每个 Assert 变体均验证：
- 正常值返回正确结果（不 panic）
- 溢出时 panic（`recover` 捕获）

### 类型覆盖
`MulInt` 和 `DivInt` 独立测试 int8 / int32 / int64，确保小类型提升路径和 int64 除法边界检查路径各自正确。

## 运行方式

```bash
# 快速验证
go test ./pkg/algorithms/mathx/overflow/

# 含示例
go test -run Example ./pkg/algorithms/mathx/overflow/
```

## 注意

除以 0 由 Go 运行时处理（panic），本包不覆盖此情形。
