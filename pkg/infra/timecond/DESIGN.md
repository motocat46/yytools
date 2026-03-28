# timecond 设计文档

## 来源

从游戏服务器业务代码中提炼。原始代码实现活动开放条件判断（注册时间触发、开服时间触发），提炼时剥离业务常量，保留通用的时间比较语义。

## 核心抽象

原始代码的本质是两件事：

1. **Parse**：把 `(type int32, value string)` 解析成带类型的值（绝对时间戳 / 相对时长 / 时间区间）
2. **Check**：把运行时 subject 时间戳和解析出的值做比较

比较运算只有五种：`<`、`>=`、区间、相对 `<`、相对 `>=`，加上无条件成立共六种 Op。

## 关键决策

### 接口 vs struct+switch

**决策：struct+switch**

考察接口方案：
```go
type TimeCondition interface { Check(subjectMs, nowMs int64) bool }
type absCondition struct{ absTs int64 }
type relCondition struct{ relDur time.Duration }
// ...
```

两种方案扩展代价相当（都需要改两处）。Op 集合在逻辑上是完备的（数学比较关系），不预期频繁扩展。struct+switch 逻辑集中，一眼看完所有 case，维护更简单。

原始代码的问题不是用了接口，而是 `GetVal() interface{}` 丢失类型信息导致外部类型断言。两种方案都消除了这个问题（Check 内嵌行为）。

### Check(subjectMs, nowMs int64) 显式传入 nowMs

**决策：调用方传入 nowMs，不在内部调用 time.Now()**

优点：
- 可测试：测试无需 mock 时钟
- 职责清晰：timecond 只做比较，不关心"现在是几点"
- 绝对条件（OpLT/GE/Within）不使用 nowMs，调用方可传任意值（如 0）

代价：多个同类型 int64 参数存在顺序混淆风险，注释中显式说明语义。

### OpWithin 左闭右开

**决策：`absRange[0] <= subject < absRange[1]`**

与原始代码保持一致（原注释："绝对时间1 ≤ 开服时间 < 绝对时间2"）。左闭右开是区间的行业惯例，与 Go 的 slice 语义一致。

自动排序：Parse 时若 t1 > t2 则交换，兼容配置填反的情况。

### Op.String() 与 Op() 访问器

**决策：两者都加**

- `Op.String()`：`Parse` 返回错误时错误信息可读（`"unsupported OpRelLT"` 而非 `"unsupported 4"`）
- `Op()` 访问器：业务层打日志、序列化时需要知道条件类型，成本为零

### 私有 helper 分层

```
Parse（调度器）
├── parseAbsMs     → 单个时间字符串 → int64 ms（OpLT/GE 用）
├── parseAbsRange  → "t1,t2" → (lo, hi)（OpWithin 用，内部复用 parseAbsMs ×2）
└── timeutil.ParseDuration（OpRelLT/GE 直接复用）
```

`Parse` 只做分支分发，解析细节下沉到 helper。

### 依赖方向

```
timecond → timeutil（单向，正确）
```

`timecond` 是 `timeutil` 的消费者，放在 `pkg/infra/timecond/` 与 `timeutil` 平级，避免循环依赖。
