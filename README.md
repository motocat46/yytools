# yytools

[![CI](https://github.com/motocat46/yytools/actions/workflows/ci.yml/badge.svg)](https://github.com/motocat46/yytools/actions/workflows/ci.yml)

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
│   │   ├── idgen/
│   │   │   └── snowflake/     # 雪花算法唯一 ID 生成器（无锁、零分配）
│   │   ├── mathx/             # 数学工具（GCD、Fibonacci等）
│   │   │   ├── bits/          # 位运算工具
│   │   │   ├── overflow/      # 数值溢出检查
│   │   │   ├── probability_distribution/  # 概率分布（普通/Vose别名/动态权重）
│   │   │   ├── random/        # 随机数生成
│   │   │   └── sampling/      # Floyd 不重复采样、带最小间隔约束采样
│   │   └── sort/              # 排序算法（快排、计数排序、基数排序等）
│   ├── ds/                    # 数据结构
│   │   ├── heap/              # 最小堆 / 最大堆 / 优先级队列
│   │   ├── lru/               # LRU 缓存（参考实现，生产建议用 hashicorp/golang-lru v2）
│   │   ├── queue/             # 环形队列（自动扩缩容）
│   │   ├── ring_buffer/       # 固定容量环形缓冲区，写满时覆盖最旧元素
│   │   ├── stack/             # 栈
│   │   ├── sorted_set/        # 有序集合（跳表实现，类 Redis ZADD）
│   │   ├── trie/              # 前缀树
│   │   ├── slidingwindow/     # 固定容量滑动窗口，O(1) Sum/Avg/Max/Min（单调双端队列）
│   │   ├── unionfind/         # 泛型并查集（DSU），O(α) Union/Find/Connected/Size/Count
│   │   ├── fenwicktree/       # 泛型树状数组（BIT），O(log n) 单点更新 + 前缀和 / 区间和
│   │   ├── segtree/           # 泛型线段树（ACL 风格），O(log n) 区间更新 + 区间查询
│   │   └── sparsetable/       # 泛型稀疏表，O(n log n) 预处理后 O(1) 区间 min/max/GCD（仅静态数据）
│   ├── infra/                 # 基础设施工具
│   │   ├── concurrency/
│   │   │   ├── unbounded_channel/  # 无大小限制的 Channel
│   │   │   └── workerpool/         # 固定大小 goroutine 池 + 泛型 Pipeline
│   │   ├── os/                # OS 工具封装
│   │   ├── safeexec/          # panic 安全执行（Safe、SafeExec、SafeExecErr、SafeExecVal[T]）
│   │   ├── timecond/          # 时间条件判断工具
│   │   ├── timeutil/          # 时间工具函数
│   │   └── timer/
│   │       ├── timingwheel/   # 时间轮（带生命周期的通用并发定时原语）
│   │       └── delayqueue/    # 延迟队列
│   ├── common/
│   │   ├── assert/            # 运行时断言框架（始终启用，正向条件语法糖）
│   │   ├── base/              # 泛型类型约束（Integer、Ordered 等）
│   │   └── cpu/               # CPU 特性检测（缓存行大小等）
│   ├── numconst/              # 数字和时间常量（时间单位等）
│   ├── mechanics/             # 规则机制（分布策略、调度策略等）
│   └── slicex/                # 切片工具函数（MinBy、MaxBy 等）
├── examples/                  # 各包 example_test.go 索引（go test -run Example ./...）
├── docs/
│   ├── decisions/             # 架构决策记录（ADR）
│   ├── layering.md            # 项目分层说明
│   ├── style.md               # 代码风格规范
│   └── recommended-libraries.md  # 推荐外部库清单
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
go run ./cmd/demo http
```

## 演示程序

`cmd/demo` 提供命令行入口，可对各模块运行压测：

```bash
go run ./cmd/demo --help          # 查看所有命令
go run ./cmd/demo sort --help     # 查看指定命令说明
go run ./cmd/demo sort 5          # 排序算法压测，执行 5 轮
go run ./cmd/demo heap 3          # 堆操作压测，执行 3 轮
go run ./cmd/demo sortedset 2     # 有序集合压测，执行 2 轮
go run ./cmd/demo all 1           # 所有模块压测
go run ./cmd/demo http            # 启动排序性能可视化服务
```

## 模块索引

| 模块 | 功能简介 |
|------|---------|
| [pkg/algorithms](pkg/algorithms/README.md) | 排序、二分查找、数学工具、唯一 ID 生成 |
| [pkg/ds](pkg/ds/README.md) | 堆、队列、栈、有序集合、LRU 缓存、前缀树、滑动窗口、并查集、树状数组、线段树 |
| [pkg/mechanics](pkg/mechanics/README.md) | 游戏规则机制：渐进式权重周期、双层保底抽卡引擎、随机份额分配 |
| [pkg/infra](pkg/infra/README.md) | panic 安全执行、时间工具、时间条件判断、OS 封装、无界 Channel、时间轮、延迟队列 |
| [pkg/common](pkg/common/README.md) | 泛型类型约束、运行时断言框架、CPU 特性检测 |
| [pkg/slicex](pkg/slicex/README.md) | 切片工具函数（MinBy、MaxBy 等） |
| [pkg/numconst](pkg/numconst/README.md) | 常用数字常量（千/万/亿）和时间单位常量 |

## 开发规范

- 泛型实现使用 `pkg/common/base` 中的类型约束
- 代码注释使用中文
- 所有文件包含 Apache 2.0 License 头
- 使用 `assert.Assert` 做运行时契约检查，始终启用（正向条件语法糖，见 `pkg/common/assert/DESIGN.md`）

## 文档参考

 [项目分层](docs/layering.md)

 [项目风格](docs/style.md)

## License

[Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0)
