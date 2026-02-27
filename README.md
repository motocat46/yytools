# yytools

yytools 的 API 必须可预测、可搜索、可扩展；宁可少也要稳定。

提供算法、数据结构和常用工具方法，方便日常开发使用。

## 目录结构

```
yytools/
├── cmd/
│   └── demo/                  # 演示程序（CLI + HTTP 可视化）
├── internal/
│   └── bench/                 # 各模块的基准/压测函数（仅供内部使用）
│       ├── heap/
│       ├── queue/
│       ├── stack/
│       ├── sort/
│       ├── sorted_set/
│       ├── mathx/
│       └── probability_distribution/
├── pkg/                       # 核心功能模块
│   ├── algorithms/
│   │   ├── binary_search/     # 二分查找
│   │   ├── mathx/             # 数学工具（GCD、Fibonacci等）
│   │   │   ├── bits/          # 位运算工具
│   │   │   ├── overflow/      # 数值溢出检查
│   │   │   ├── probability_distribution/  # 概率分布（普通/Vose别名/动态权重）
│   │   │   └── random/        # 随机数生成
│   │   └── sort/              # 排序算法（快排、计数排序、基数排序等）
│   ├── ds/                    # 数据结构
│   │   ├── heap/              # 最小堆 / 最大堆 / 优先级队列
│   │   ├── queue/             # 环形队列（自动扩缩容）
│   │   ├── stack/             # 栈
│   │   └── sorted_set/        # 有序集合（跳表实现，类 Redis ZADD）
│   ├── infra/                 # 基础设施工具
│   │   ├── concurrency/
│   │   │   └── unbounded_channel/  # 无大小限制的 Channel
│   │   ├── os/                # OS 工具封装
│   │   ├── safeexec/          # panic 安全执行（Safe、SafeCall 等）
│   │   └── timeutil/          # 时间工具函数
│   ├── common/
│   │   ├── assert/            # 运行时断言框架（可全局开关）
│   │   └── base/              # 泛型类型约束（Integer、Ordered 等）
│   ├── numconst/              # 数字和时间常量（千/万/亿、时间单位）
│   └── slicex/                # 切片工具函数（MinBy、MaxBy 等）
├── examples/                  # 示例代码（待补充）
├── docs/
│   └── decisions/             # 架构决策记录（待补充）
└── README.md
```

## 快速开始

```bash
# 构建
go build ./...

# 运行所有单元测试
go test ./...

# 运行指定模块测试
go test ./pkg/ds/sorted_set/
go test ./pkg/algorithms/sort/

# 启动性能可视化 HTTP 服务（浏览器访问 http://localhost:8081）
cd cmd/demo && go run . http
```

## 演示程序

`cmd/demo` 提供命令行入口，可对各模块运行压测：

```bash
cd cmd/demo

go run . help          # 查看所有命令
go run . sort 5        # 排序算法压测，执行 5 轮
go run . heap 3        # 堆操作压测，执行 3 轮
go run . sortedset 2   # 有序集合压测，执行 2 轮
go run . all 1         # 所有模块压测
go run . http          # 启动排序性能可视化服务
```

## 模块说明

### `pkg/algorithms/mathx`

数学工具，包括最大公约数（GcdR/GcdI/Gcd）、绝对值（Abs）、Fibonacci 数列（带记忆化）。

子包：
- `bits`：位运算工具
- `overflow`：加减乘除溢出检查
- `probability_distribution`：三种概率分布实现（遍历法、Vose 别名法、动态权重）
- `random`：随机整数生成

### `pkg/algorithms/sort`

多种排序算法实现：
- `BubbleSort` / `BubbleSortDesc`
- `InsertionSort` / `InsertionSortDesc`
- `QuickSort` / `QuickSortDesc`（随机 pivot + 小数组切换插入排序）
- `QuickSortTraversal` / `QuickSortDescTraversal`（栈辅助迭代版快排）
- `CountingSort`（计数排序）
- `RadixSort`（LSD 基数排序，适用于非负整数）

### `pkg/ds`

泛型数据结构：
- **heap**：最小堆（`Heap[T]`）、最大堆（`MaxHeap[T]`）、优先级队列（`PriorityQueue[T]`）
- **queue**：环形队列（`Queue[T]`），自动扩容（翻倍）和缩容（缩半）
- **stack**：栈（`Stack[T]`）
- **sorted_set**：有序集合（`SortedSet[T]`），基于跳表实现，支持 Insert / Delete / UpdateScore / GetRank / GetByRank / GetRangeByScore 等操作

### `pkg/infra`

基础设施工具：
- **safeexec**：`Safe`、`SafeCall`、`SafeExecWithError` 等 panic 安全执行包装
- **timeutil**：时间工具函数
- **os**：OS 相关工具封装
- **concurrency/unbounded_channel**：无大小限制的 Channel，多种实现变体

### `pkg/common`

极简公共依赖：
- **assert**：运行时断言框架，通过 `assert.SetAssert(true)` 启用，`assert.Assert(cond, msg...)` 使用
- **base**：泛型类型约束定义

### `pkg/slicex`

切片工具函数：`MinInSlice`、`MaxInSlice`、`MinBy`、`MaxBy` 等。

### `pkg/numconst`

常用数字常量（千/万/亿）和时间单位常量。

## 开发规范

- 泛型实现使用 `pkg/common/base` 中的类型约束
- 代码注释使用中文
- 所有文件包含 Apache 2.0 License 头
- 使用 `assert.Assert` 做运行时契约检查，生产环境可关闭

## License

[Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0)