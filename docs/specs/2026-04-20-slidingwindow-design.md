# SlidingWindow 设计文档

**日期：** 2026-04-20
**状态：** 已批准，待实现

## 背景

yytools 已有 `pkg/ds/ring_buffer`（固定容量环形缓冲区），但它只负责存取，不提供统计语义。  
Go 生态中无任何 >500 stars 的泛型滑动窗口统计库（已核查 asecurityteam/rolling 59 stars，RobinUS2/golang-moving-average 44 stars，均无泛型）。

**核心价值：** 让「最近 N 个值的统计」成为 O(1) 操作，尤其是 Max/Min——朴素实现是 O(N) 全窗口扫描，单调双端队列实现是 O(1) 均摊。

## 使用场景

| 场景 | 代码意图 |
|------|---------|
| 基准测试噪声消除 | 最近 20 次耗时的滚动均值，忽略偶发抖动 |
| 游戏 DPS 计算 | 最近 10 次伤害的 Sum，除以时间间隔 |
| 监控指标平滑 | 最近 50 个采样点的 Avg，曲线不跳动 |
| 请求延迟观测 | 最近 100 次请求的 Max（尾延迟） |

## 包路径

```
pkg/ds/slidingwindow/
```

包名：`slidingwindow`，类型名：`Window[T]`。

## API

```go
// Number 直接使用 pkg/common/base.Number（Integer | Float），无需重新定义。

// Window 是固定容量的滑动窗口，维护增量统计量。
// 零值不可用，必须通过 New 创建；非并发安全。
type Window[T base.Number] struct { ... }

// New 创建容量为 size 的滑动窗口。
// size 须 > 0，否则 panic。
func New[T base.Number](size int) *Window[T]

// Add 向窗口追加一个值；窗口满时自动淘汰最旧元素。O(1) 均摊。
func (w *Window[T]) Add(v T)

// Sum 返回当前窗口内所有元素之和。O(1)。
func (w *Window[T]) Sum() T

// Avg 返回当前窗口均值（float64）。O(1)。
// 窗口为空时 panic；调用前用 Empty() 检查。
func (w *Window[T]) Avg() float64

// Max 返回当前窗口最大值。O(1) 均摊。
// 窗口为空时 panic；调用前用 Empty() 检查。
func (w *Window[T]) Max() T

// Min 返回当前窗口最小值。O(1) 均摊。
// 窗口为空时 panic；调用前用 Empty() 检查。
func (w *Window[T]) Min() T

// Len 返回当前窗口元素数量（0 ≤ Len ≤ size）。
func (w *Window[T]) Len() int

// Full 报告窗口是否已满（Len() == size）。
func (w *Window[T]) Full() bool

// Empty 报告窗口是否为空（Len() == 0）。
func (w *Window[T]) Empty() bool
```

**不提供：** Reset()（YAGNI，重建代价极低）、Percentile（需排序，非增量）、并发安全（调用方加锁）。

## 内部实现

### 数据结构

```go
type Window[T base.Number] struct {
    size   int
    buf    []T   // 环形缓冲区，长度 == size
    head   int   // 下一个写入位置（物理索引）
    count  int   // 当前元素数
    sum    T     // 增量维护的总和
    seq    int   // 逻辑序号，每次 Add 自增，用于单调队列淘汰判断
    maxDeq []int // 单调递减队列（存逻辑序号），队首 = 最大值所在序号
    minDeq []int // 单调递增队列（存逻辑序号），队首 = 最小值所在序号
    // buf 中逻辑序号 s 对应物理索引：(head - count + s_offset) % size
    // 实际存储方式：maxDeq/minDeq 存逻辑序号，对应值通过 bufAt(seq) 查询
}
```

**注意：** 单调队列存逻辑序号（全局自增），而非物理索引，避免环绕计算复杂度。`bufAt(s int) T` 将逻辑序号映射到 `buf` 物理位置：

```
物理索引 = (head - count + (s - (seq - count))) % size
         = (head - (seq - s)) % size  （化简后）
```

### Add 逻辑

```
newSeq = seq          // 新元素的逻辑序号
seq++
writePos = head       // 新元素写入位置（满时同时是被淘汰元素的位置）
head = (head+1) % size

若 count == size（窗口已满，需淘汰最旧元素）：
    evictedSeq = newSeq - size    // 最旧元素的逻辑序号
    sum -= buf[writePos]          // writePos 此时存放最旧元素
    若 maxDeq 非空 && maxDeq[0] == evictedSeq：弹出队首
    若 minDeq 非空 && minDeq[0] == evictedSeq：弹出队首
else：
    count++

buf[writePos] = v
sum += v

// 维护单调递减队列（Max）
while maxDeq 非空 && bufAt(maxDeq[尾]) <= v：弹出队尾
maxDeq 入队尾 newSeq

// 维护单调递增队列（Min）
while minDeq 非空 && bufAt(minDeq[尾]) >= v：弹出队尾
minDeq 入队尾 newSeq
```

**关键不变量：** 当 `count == size` 时，`writePos`（即旧 `head`）存放的恰好是最旧元素——这是环形缓冲区的固有特性。

### 复杂度汇总

| 操作 | 时间 | 空间 |
|------|------|------|
| Add  | O(1) 均摊 | — |
| Sum  | O(1) | — |
| Avg  | O(1) | — |
| Max  | O(1) 均摊 | O(N) deque |
| Min  | O(1) 均摊 | O(N) deque |
| Len/Full/Empty | O(1) | — |
| 总空间 | — | O(N) |

## 正确性命题（并发无关，聚焦算法正确性）

| 命题 | 描述 |
|------|------|
| Sum 精确 | 任意操作序列后，Sum() == 当前窗口内所有元素之和 |
| Max 精确 | Max() == 当前窗口内所有元素的最大值 |
| Min 精确 | Min() == 当前窗口内所有元素的最小值 |
| Avg 精确 | Avg() == Sum() / float64(Len())，误差在浮点精度范围内 |
| Len 一致 | 0 ≤ Len() ≤ size，Full() == (Len()==size)，Empty() == (Len()==0) |
| 淘汰正确 | 窗口满时，最旧元素被精确淘汰，新元素正确入窗 |
| 空 panic | Avg/Max/Min 在 Empty() 时 panic |

命题测试集中在 `correctness_test.go`，使用参考模型（`[]T` 全量扫描）对比验证，随机操作序列，规模 ≥ 10 万次操作。

## 测试策略

- **TDD**：先写失败测试，再实现
- **单方法测试**：每个 API 的边界（空窗口、单元素、恰好满、满后继续 Add）
- **correctness_test.go**：随机操作序列 + 参考模型对比，规模 10 万次
- **bench_test.go**：Add/Max/Min/Sum/Avg，规模 100/1000/10000/100000
- **-race**：不涉及并发，但标准流水线仍跑

## 不在范围内

- 时间窗口（time-based eviction）
- Percentile / Median（需排序，非增量）
- 并发安全（`sync.Mutex` 由调用方决定）
- Reset()

## 受影响文档

- 新建 `pkg/ds/slidingwindow/README.md`
- 更新 `pkg/ds/README.md`（新增 slidingwindow 行）
- 更新根 `README.md` 目录树（新增 slidingwindow/）
- 新增 `cmd/demo/api_ds_slidingwindow.go`（可视化）
- 更新 `cmd/demo/groups.go` 的 `pkg/ds` 描述（加"滑动窗口"）
