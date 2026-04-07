# random_split

通用随机份额分配引擎，适用于红包拆分、掉落奖励分配等游戏场景。

## 适用场景

将总量 `S` 分成 `N` 份，每份不低于 `min`，分配过程按序进行。保证：

- **守恒性**：所有份额之和恰好等于 `S`
- **合法性**：每份 >= `min`
- **随机性**：分布在合法范围内尽量均匀，避免确定性

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/mechanics/distribution/random_split"

state := random_split.State{
    RemainAmount: 100,
    RemainCount:  10,
    MinPerPart:   1,
}

// 方式一：预生成（推荐并发场景）
d, err := random_split.New(state, random_split.DoubleMean(), nil)
if err != nil { /* 处理错误 */ }
allocations, err := d.Allocate()
// allocations 是只读 slice，可原子分发给多个 goroutine

// 方式二：按需生成
d, _ = random_split.New(state, random_split.DoubleMean(), nil)
for !d.Done() {
    amount, err := d.Next()
    if err != nil { /* 处理错误 */ }
    // 处理 amount
}
```

## 内置策略

| 策略 | 说明 | 推荐场景 |
|------|------|---------|
| `DoubleMean()` | 二倍均值法，上界 = min(safeUpper, 2×当前均值) | 通用红包，符合直觉 |
| `Uniform()` | 均匀可行域 [min, safeUpper] | 分布最均匀，尾部更集中 |
| `Fixed()` | 每次返回 currentAmount/currentCount（整除） | 统计基准、确定性测试 |
| `MeanBounded(m)` | 上界 = min(safeUpper, floor(m×均值))，m>=1.0 | 自定义方差控制 |

## API 参考

### State

```go
type State struct {
    RemainAmount int64 // 剩余总量
    RemainCount  int64 // 剩余份数
    MinPerPart   int64 // 每份最小值（>= 1）
}

func (s State) Validate() error
```

### Distributor

```go
func New(s State, fn SampleFunc, rng *rand.Rand) (*Distributor, error)
func (d *Distributor) Next() (int64, error)
func (d *Distributor) Allocate() ([]int64, error)
func (d *Distributor) Remaining() State
func (d *Distributor) Done() bool
```

### 错误类型

| 错误 | 触发时机 |
|------|---------|
| `ErrDone` | `Done()==true` 后调用 `Next()` |
| `ErrInvalidState` | `State.Validate()` 失败（由 `New()` 返回） |
| `ErrInvalidParam` | `MeanBounded(m<1.0)`、`New(fn=nil)`、`Simulate(fn=nil)`、`Simulate(rounds<=0)` |
| `ErrAlreadyStarted` | `Allocate()` 在非全新 Distributor 上调用 |

`ErrInvalidState`、`ErrInvalidParam` 以 `fmt.Errorf("...: %w", ErrXxx)` 包装；`ErrDone`、`ErrAlreadyStarted` 直接返回哨兵值。均可用 `errors.Is` 识别。

## 并发使用

`Distributor` 非线程安全。并发场景推荐预生成：

```go
// 主 goroutine 预生成，不加锁
allocations, _ := d.Allocate()

// 多个 goroutine 安全读取（只读）
var idx atomic.Int64
go func() {
    i := idx.Add(1) - 1
    if i < int64(len(allocations)) {
        process(allocations[i])
    }
}()
```

## 统计验证

```go
result, _ := random_split.Simulate(state, random_split.DoubleMean(), 100_000, 42)
result.CheckConservation() // 守恒性：每轮 sum == S
result.CheckLegality()     // 合法性：每值 >= min
fmt.Print(result.Summary())
```

## 常见误用

- ❌ 混用 `Next()` 和 `Allocate()`：先调用 `Next()` 后再调用 `Allocate()` → `ErrAlreadyStarted`
- ❌ 在 `Done()==true` 后调用 `Next()` → `ErrDone`（不 panic）
- ❌ 多 goroutine 共享同一个 `Distributor`：非线程安全，使用 `Allocate()` 预生成
