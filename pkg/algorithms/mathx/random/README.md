# random

泛型整数随机工具，支持所有整数类型的闭区间均匀随机，处理全范围溢出边界情况。

## 快速上手

```go
// 全局随机源（goroutine 安全，OS 熵自动初始化）
n := random.RandInt(1, 100)          // int，[1, 100]
n8 := random.RandInt[int8](-128, 127) // int8，全范围
u := random.RandInt[uint32](0, 1000)  // uint32
f := random.RandFloat64()             // float64，[0.0, 1.0)

// 固定种子（可重放，适合单元测试/仿真）
rng := random.NewRand(42)
n := random.RandIntWith(rng, 1, 100)  // 相同 seed + 相同调用序列 = 相同结果
```

## API

| 函数 | 说明 |
|------|------|
| `RandInt[T](low, high T) T` | 全局源，闭区间 [low, high] 均匀整数，goroutine 安全 |
| `RandIntWith[T](rng, low, high T) T` | 指定源，用于确定性重放，非 goroutine 安全 |
| `NewRand(seed uint64) *rand.Rand` | 创建固定种子的本地随机源 |
| `RandFloat64() float64` | 全局源，[0.0, 1.0) 均匀浮点数 |

## 支持类型

所有 `base.Integer` 约束的类型：`int`、`int8`、`int16`、`int32`、`int64`、`uint`、`uint8`、`uint16`、`uint32`、`uint64`，以及这些类型的派生类型（如 `type Score int32`）。

## 溢出处理

内部对有符号全范围 `[MinIntN, MaxIntN]` 和无符号全范围 `[0, MaxUintN]` 做了专项处理，计算区间大小时通过 uint64 补码算术避免有符号溢出。

## 注意

- `low > high` 时触发 `assert`（可被 `-tags assertion_off` 关闭）
- 固定种子的 `*rand.Rand` 非 goroutine 安全，多 goroutine 共享需自行加锁
