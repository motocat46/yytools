# random_split 设计决策记录

## 问题定义

将总量 S 分成 N 份，每份 >= min，按序分配（第 i 个人拿到后第 i+1 个才开始）。
目标：守恒性 + 合法性 + 随机性，三者缺一不可。

## 关键决策

### 1. SampleFunc 函数类型 vs 接口

选择**函数类型**（`func(state State, rng *rand.Rand) (int64, error)`）而非接口。

- 函数类型是 Go 惯用法，适合策略注入：`DoubleMean()`、`Uniform()` 等都是工厂函数返回函数值
- 无需定义接口和实现结构体，调用方可以用 lambda 直接注入自定义策略
- 扩展性：接口只能以方法扩展，函数类型可以高阶组合（如 `WithLogging(fn)`）

### 2. 最后一份在 Next() 统一处理

`SampleFunc` 永远不会被调用到最后一份（`RemainCount==1` 时 `Next()` 直接返回剩余全部）。

原因：最后一份必须返回全部剩余，否则守恒性无法保证。让策略函数处理这一边界会导致：
- 所有自定义策略都需要重复实现同一段逻辑
- 策略逻辑和守恒逻辑耦合

### 3. Welford 在线算法替代存储全量值

统计模块使用 Welford 在线算法（O(1) 内存）而非 `[]int64`（O(N) 内存）。

对比：
- 10 万轮 × 1000 个位置 × 8 字节 = **800MB** 内存（不可接受）
- Welford：每个 PositionStat 固定约 56 字节，1000 个位置 = **56KB**

代价：无法事后回溯原始值，但 `CheckConservation` 通过 `roundSums` 单独记录（每轮仅 8 字节）解决。

### 4. Allocate() 要求全新 Distributor

`Allocate()` 在 `Next()` 被调用后拒绝执行（返回 `ErrAlreadyStarted`），而非静默继续。

原因：混用 `Next()` 和 `Allocate()` 的语义不清晰——调用方很可能是误用。快速失败优于静默返回残留结果。

### 5. rng 由调用方注入

`New(s, fn, rng)` 中 `rng=nil` 时使用随机种子，传入固定种子的 `rng` 用于测试复现。

原因（显式层保险原则）：库代码不应依赖全局随机源；测试需要确定性；显式注入支持两者。

### 6. MeanBounded 的上界选择

上界 = `min(safeUpper, floor(multiplier × avg))`

- `safeUpper` 保证剩余每人都能拿到 min（守恒安全线）
- `floor(multiplier × avg)` 控制分布方差（倍数越大越分散）
- 两者取 min：既不破坏守恒，又限制方差

`multiplier < 1.0` 时上界低于均值，分布无意义，直接拒绝（返回 `ErrInvalidParam`）。
