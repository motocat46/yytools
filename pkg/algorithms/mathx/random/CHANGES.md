# random 包重构说明

## 必要性

原实现（2022 年）存在以下问题，随着 Go 版本迭代和泛型能力的成熟，有必要整体重写：

| 问题 | 具体表现 |
|------|---------|
| API 冗余 | 5 个几乎相同的类型特定函数（`RandInt8/16/32/64/Int`）+ 2 个泛型包装器，维护成本高 |
| reflect 性能差 | `RandInteger` 每次调用执行 `reflect.ValueOf(low).Kind()`，在堆上分配对象，且注释自相矛盾 |
| 类型系统失误 | `RandInteger1(interface{}) interface{}` 返回值需要二次类型断言，调用方不友好 |
| 错误的参数约束 | `assert(low >= 0)` 阻止了合法的负数范围（如 `RandInt(-5, 5)` 骰子偏移）|
| 静默参数交换 | `low > high` 时自动对调，隐藏调用方的逻辑错误 |
| 逻辑死代码 | `RandInt` 中 `result < 0` 检查永不成立（`rand.Int()` 不可能返回负数）|
| 废弃 API | `rand.Seed` 在 Go 1.20 起已废弃；`math/rand/v2` 自 Go 1.22 纳入标准库 |

---

## 优化结果

### API 对比

| 旧 API | 新 API | 说明 |
|--------|--------|------|
| `RandInt8(low, high int8) int8` | ✗ 删除 | 由泛型函数覆盖 |
| `RandInt16(low, high int16) int16` | ✗ 删除 | 由泛型函数覆盖 |
| `RandInt32(low, high int32) int32` | ✗ 删除 | 由泛型函数覆盖 |
| `RandInt64(low, high int64) int64` | ✗ 删除 | 由泛型函数覆盖 |
| `RandInt(low, high int) int` | ✗ 删除 | 由泛型函数覆盖 |
| `RandInteger[T Signed](low, high T) T` | ✗ 删除 | reflect 实现，性能差 |
| `RandInteger1(low, high interface{})` | ✗ 删除 | 类型不安全 |
| `RandSeed(seed int64)` | ✗ 删除 | 基于废弃的 `rand.Seed` |
| — | ✅ `RandInt[T Integer](low, high T) T` | 统一泛型接口，goroutine 安全 |
| — | ✅ `RandIntWith[T Integer](rng *rand.Rand, low, high T) T` | 支持确定性重放 |
| — | ✅ `NewRand(seed uint64) *rand.Rand` | 创建固定种子随机源 |
| `RandFloat64() float64` | ✅ 保留 | 有外部调用依赖 |

### 技术改进

#### 1. 类型派发：reflect → unsafe.Sizeof（零开销）

```
// 旧：每次调用在堆上分配 reflect.Value
kind := reflect.ValueOf(low).Kind()   // ~10-50 ns + 堆分配

// 新：编译期常量折叠，运行时零开销
signed := T(0)-T(1) < 0              // 编译期确定
size   := unsafe.Sizeof(zero)        // 编译期确定
```

性能提升（int64, n=1000）：

```
BenchmarkRandInt_Global    约 10 ns/op   0 allocs/op
```

#### 2. 支持负数范围（修正错误约束）

```go
// 旧：assert(low >= 0) 阻止合法用法
RandInt32(-5, 5)  // ← panic!

// 新：支持任意 low <= high
RandInt(-5, 5)    // ✓ 正常工作
RandInt[int8](math.MinInt8, math.MaxInt8)  // ✓ 全范围
```

#### 3. 正确的参数验证：静默交换 → assert 快速失败

```go
// 旧：静默交换，隐藏调用方逻辑错误
RandInt32(10, 1)  // 自动变成 RandInt32(1, 10)，调用方可能不知情

// 新：assert(low <= high)，及早暴露问题
RandInt(10, 1)    // panic，调用方立即发现参数顺序错误
```

#### 4. 确定性重放（新功能）

```go
// 全局随机源（不可重放，goroutine 安全）
result := random.RandInt[int64](0, 100)

// 本地固定种子（可重放，非 goroutine 安全）
rng := random.NewRand(42)          // 固定种子
a := random.RandIntWith[int64](rng, 0, 100)   // 第 1 次：固定值
b := random.RandIntWith[int64](rng, 0, 100)   // 第 2 次：固定值

rng2 := random.NewRand(42)         // 相同种子
c := random.RandIntWith[int64](rng2, 0, 100)  // c == a，序列完全一致
```

适用场景：单元测试中固定随机序列、物理仿真的复现、算法调试。

#### 5. 迁移到 math/rand/v2（Go 1.22+ 标准库）

- 使用 PCG 算法（比旧版 Mersenne Twister 统计特性更好）
- 全局源自动以 OS 熵初始化，无需手动 `Seed`
- API 更清晰（`Int64N`、`Uint64N` 等命名一致）

#### 6. 修复 int64/uint64 全范围溢出

旧版 `RandInt64(0, MaxInt64)` 的溢出处理是特判（仅覆盖 `low==0`），新版用 uint64 二补数算术覆盖所有情形：

```go
// 新：通用 uint64 算术，正确处理 low<0 的情况
n := uint64(high) - uint64(low) + 1   // 二补数安全
```

---

## 向后兼容性

本次重构为**破坏性变更**，发生在 `refactor` 分支上。调用方迁移方式：

- `RandInt32(a, b)` → `RandInt[int32](a, b)` 或直接 `RandInt(int32(a), int32(b))`
- `RandInteger(a, b)` → `RandInt(a, b)`（Signed 类型可直接使用，约束兼容）
- `RandSeed(x)` → 删除（无需手动 seed，或使用 `NewRand(seed)` 获取本地随机源）
