# random 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、泛型派发：用 `unsafe.Sizeof` + 有符号检测替代类型断言

Go 泛型不支持在同一函数内对类型参数做 `switch T.(type)`。本包的解决方案：在运行时通过两个零开销操作做派发：

```go
signed := T(0)-T(1) < 0      // 有符号？（编译期常量折叠）
size   := unsafe.Sizeof(zero) // 字节宽度（编译期常量折叠）
```

这两个值在编译期即可确定，运行时 switch 会被优化掉。派生类型（如 `type Score int8`）也能正确走 int8 路径。

---

## 二、int64 的特殊处理：uint64 补数算术

**问题**：`high - low + 1` 对于 int64 全范围 `[MinInt64, MaxInt64]` 会溢出（结果为 0，有符号）。

**解决方案**：用 uint64 补数算术计算范围 `n = uint64(high) - uint64(low) + 1`，当 n=0 时为全范围，直接返回任意 uint64 转型；否则用 `int64(uint64(low) + Uint64N(n))` 避免加法溢出。

```go
n := uint64(high) - uint64(low) + 1  // uint64 不溢出
if n == 0 { return int64(src.Uint64()) }   // 全范围特判
return int64(uint64(low) + src.Uint64N(n)) // 避免有符号加法溢出
```

类似地，uint64 全范围 `[0, MaxUint64]` 时 n 也溢出为 0，同样特判。

---

## 三、全局源 vs 本地实例的权衡

| | `RandInt`（全局源）| `RandIntWith`（本地实例）|
|--|------------------|------------------------|
| goroutine 安全 | ✓（自动加锁）| ✗（调用方保证）|
| 可复现 | ✗ | ✓（固定种子）|
| 性能 | ~7.7 ns（含锁）| ~5.2 ns（无锁）|

**使用原则**：业务随机用 `RandInt`（简单、安全）；测试回放、仿真用 `NewRand(seed)` + `RandIntWith`（可复现）。

---

## 四、为什么 RandFloat64 直接委托给标准库

浮点随机无泛型需求（只需 float64），直接暴露为 `rand.Float64()` 的薄封装，保持接口一致（调用方无需直接依赖 `math/rand/v2`）。
