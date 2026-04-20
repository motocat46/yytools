# slidingwindow

固定容量泛型滑动窗口，提供 O(1) 的 Sum/Avg/Max/Min 统计。
Max/Min 通过单调双端队列实现 O(1) 均摊，优于朴素 O(N) 全窗口扫描。

## 适用场景

- 基准测试噪声消除（最近 N 次耗时的滚动均值）
- 游戏 DPS 计算（最近 N 次伤害的 Sum）
- 监控指标平滑（最近 N 个采样点的 Avg）
- 请求延迟观测（最近 N 次请求的 Max）

## 快速上手

```go
import sw "github.com/motocat46/yytools/pkg/ds/slidingwindow"

w := sw.New[int](5) // 最近 5 个值

w.Add(10)
w.Add(20)
w.Add(30)

fmt.Println(w.Sum())  // 60
fmt.Println(w.Avg())  // 20.0
fmt.Println(w.Max())  // 30
fmt.Println(w.Min())  // 10
fmt.Println(w.Len())  // 3
fmt.Println(w.Full()) // false

w.Add(40)
w.Add(50)
w.Add(60) // 淘汰最旧的 10，窗口变为 [20,30,40,50,60]

fmt.Println(w.Max())  // 60
fmt.Println(w.Min())  // 20
fmt.Println(w.Full()) // true
```

## API

```go
func New[T base.Number](size int) *Window[T]

func (w *Window[T]) Add(v T)          // O(1) 均摊
func (w *Window[T]) Sum() T           // O(1)
func (w *Window[T]) Avg() float64     // O(1)；空时 panic
func (w *Window[T]) Max() T           // O(1) 均摊；空时 panic
func (w *Window[T]) Min() T           // O(1) 均摊；空时 panic
func (w *Window[T]) Len() int
func (w *Window[T]) Full() bool
func (w *Window[T]) Empty() bool
```

## 性能基准（Apple M4）

| 操作 | ns/op | allocs/op |
|------|-------|-----------|
| Add（稳态满窗口） | ~7 | 0 |
| Max | ~2.3 | 0 |
| Min | ~2.3 | 0 |
| Sum | ~2.0 | 0 |

ns/op 在窗口大小 100~100000 范围内保持常数，验证 O(1) 均摊复杂度。

## 注意事项

- 非并发安全，并发访问由调用方加锁
- `Avg/Max/Min` 在空窗口调用时 panic，调用前用 `Empty()` 检查
- 支持所有整数和浮点类型（`base.Number`）
- 不提供时间窗口、Percentile、Reset
