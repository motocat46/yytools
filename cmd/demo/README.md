# cmd/demo

可视化演示服务，在浏览器中展示 yytools 各模块的性能与行为特征。

## 用法

```bash
# 启动可视化 HTTP 服务（http://localhost:8081）
go run ./cmd/demo http
```

浏览器访问 http://localhost:8081，首页按包分组列出所有可视化页面。

## 当前可视化（32 个）

### pkg/algorithms
| 页面 | 说明 |
|------|------|
| 二分搜索性能 | BinarySearch vs 线性搜索 O(n)；三种变体（BinarySearch/LeftBound/RightBound）耗时 vs 规模 |
| GCD 递归 vs 迭代 | GcdR（递归）vs GcdI（迭代）耗时 vs 输入量级 |
| Bits 操作性能 | CountingBits（Kernighan O(k)）vs math/bits.OnesCount64（POPCNT O(1)）位密度自适应性 |
| Overflow 检测耗时 | AddInt/SubInt/MulInt/DivInt vs 裸运算基准 |
| RandInt 均匀性与性能 | RandInt 分布均匀性；各宽度类型耗时 vs 标准库 |
| 概率分布 — 准确性与性能 | O(n)/O(log n)/O(1) 三种实现耗时 vs 权重项数对比 |
| Sampling 均匀性与 O(k) 特性 | Floyd 采样均匀性；耗时 vs 范围大小 N（O(k) 特性） |
| Snowflake ID 并发吞吐 | 生成吞吐 vs 并发度；ID 单调性验证 |
| 高效排序对比 | TimSort / QuickSort / CountingSort vs 标准库耗时 |
| 简单排序对比 | BubbleSort / InsertionSort / SelectionSort 耗时 |
| 快排 vs pdqsort | 不同数据分布下快排 vs Go 标准库 pdqsort |
| 概率分布对比 | NormalMethod vs VoseAliasMethod 100 万次采样 |

### pkg/ds
| 页面 | 说明 |
|------|------|
| Heap 各操作耗时 | MinHeap / PriorityQueue Push、Pop、UpdatePriority 均摊 ns/op |
| LRU 命中率 vs 缓存容量 | 不同缓存容量的命中率曲线 |
| LRU Get/Put 耗时 | Get / Put 均摊耗时 vs 缓存规模 |
| Queue 稳态吞吐 | Enqueue / Dequeue 均摊 ns/op |
| Queue 扩缩容行为 | Len vs Capacity 阶梯变化 |
| RingBuffer vs Queue 吞吐对比 | 固定容量 vs 动态扩缩容吞吐差异 |
| RingBuffer 覆盖写语义 | 满写时覆盖最旧元素的行为验证 |
| SortedSet 各操作耗时 | Insert / Delete / GetByRank 等 O(log n) 耗时 vs 规模 |
| 跳表 vs 有序切片 | 跳表 O(log n) vs 有序切片 O(n) 插入性能对比 |
| Stack 稳态吞吐 | Push / Pop 均摊 ns/op |
| Stack 扩缩容行为 | Len vs Capacity 阶梯变化 |
| Trie 操作耗时 | Insert / Search / Delete 耗时 vs 词典大小 |

### pkg/infra
| 页面 | 说明 |
|------|------|
| DelayQueue 吞吐 | Offer O(log n) 耗时 vs 规模；TryPoll 命中/未命中/空队列对比 |
| TimingWheel 吞吐 | Cancel+AfterFunc O(1) 稳态耗时 vs 规模；并发度吞吐对比 |
| UnboundedChannel 吞吐对比 | 各容量配置吞吐；并发生产者吞吐 vs 并发度 |
| WorkerPool 并发吞吐 | 吞吐 vs worker 数；RWMutex vs Mutex Submit 吞吐 |

### pkg/mechanics
| 页面 | 说明 |
|------|------|
| 分层周期引擎 | 奖励分布 + 特殊位置散布（500 周期） |
| 渐进权重周期 | 奖励分布实际 vs 期望（1000 周期） |
| RandomSplit 策略对比 | 各分配策略实际分布 vs 期望 |

### pkg/slicex
| 页面 | 说明 |
|------|------|
| slicex 切片工具性能 | MinInSlice / MaxInSlice / MinBy vs 标准库 slices.Min 耗时对比 |

## 说明

- 此工具仅供开发调试使用，不对外暴露 API
- 各可视化页面数据由 `cmd/demo/api_*.go` 生成，通过 `registry.go` 统一注册
- 每次请求实时计算，结果反映当前机器的性能特征
