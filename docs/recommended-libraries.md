# 推荐外部库

yytools 专注于轻量、无外部依赖的通用工具。以下场景已有成熟的外部库覆盖，直接使用即可，无需在 yytools 中重复实现。

## 列表

| 场景 | 推荐库 | 使用建议 |
|------|--------|---------|
| 重试 / 指数退避 | `github.com/cenkalti/backoff/v5` | 直接用 `Retry` 函数；有定制需求时复制 `retry.go` 到项目中修改 |
| 限流（令牌桶） | `golang.org/x/time/rate` | 标准库扩展，令牌桶实现，直接集成 |
| LRU 缓存 | `github.com/hashicorp/golang-lru/v2` | 生产级，支持 onEvict 回调、per-key TTL、2Q Cache |
| 请求合并（singleflight） | `golang.org/x/sync/singleflight` | 防缓存击穿，相同 key 的并发请求只执行一次 |
| 并发度控制（semaphore） | `golang.org/x/sync/semaphore` | 加权信号量，控制同时执行的 goroutine 数量 |
| 布隆过滤器 | `github.com/bits-and-blooms/bloom/v3` | 误判率计算、动态扩容、序列化均已支持 |
| 位图（稠密小宇宙） | `github.com/bits-and-blooms/bitset` | 稠密位操作首选；内存模型为连续 `[]uint64`（N 位占 N/8 字节），Set/Clear/Test/AND/OR/XOR 直接映射 CPU 指令，稠密场景速度最快；适合权限掩码、特征标志等位宽有限的场景；~1.5k stars，同作者维护 |
| 位图（稀疏大宇宙） | `github.com/RoaringBitmap/roaring` | 压缩整数集合；自适应容器（稀疏用数组、稠密用 bitmap、连续段用 run），稀疏时内存远优于 bitset；支持 Rank/Select、跨语言序列化格式；InfluxDB、Netflix、Elasticsearch、Apache Spark 均在用；~2.9k stars。**选法：** 存「第 i 位是否为 1」用 bitset；存「整数 N 是否在集合里」且整数范围大或稀疏时用 roaring |
| 图算法 | `github.com/dominikbraun/graph` | BFS/DFS/Dijkstra/拓扑排序/环检测/强连通分量/最小生成树；原生泛型 API，支持有向/无向/加权图，含 Graphviz 可视化输出；~90% 测试覆盖率；~2.1k stars |
| 双端队列（Deque） | `github.com/gammazero/deque` | Ring buffer，所有操作 O(1) 均摊；PushFront/PushBack/PopFront/PopBack |
| 综合数据结构库 | `github.com/emirpasic/gods` | Go 生态最全面的数据结构综合库；含 ArrayList/LinkedList、HashSet/TreeSet、HashMap/TreeMap/BTreeMap、RedBlackTree/AVLTree/BTree/BinaryHeap、PriorityQueue 等；v2 提供泛型 API；~16k stars。**不含**：Union-Find、Fenwick Tree、Segment Tree、Trie（专项需求仍需单独实现） |
| 切片/Map 工具函数 | `github.com/samber/lo` | 泛型版 Chunk/Flatten/Unique/GroupBy/Zip/Keys/Values/Filter/Map 等，覆盖 slicex/mapx 常见需求 |
| 一致性哈希 | `github.com/buraksezer/consistent` | 虚拟节点、权重、bounded loads 全支持；适用于缓存分片、负载均衡 |
| singleflight 泛型封装 | `golang.org/x/sync/singleflight`（标准库）+ 自行封装泛型 wrapper | 标准库返回 `any`，需类型安全时自行写 ~10 行泛型 `Do[V any]` 包装即可，无需引入外部库 |
| 并行任务错误聚合 | `golang.org/x/sync/errgroup` | 多 goroutine 并发执行、聚合第一个错误；标准库扩展，Go 团队维护 |
| UUID 生成 | `github.com/google/uuid` | 标准 UUID v4/v7；Google 维护，业界标准 |
| 结构化并发 | `github.com/sourcegraph/conc` | 比 errgroup 更高层：bounded pool、panic 安全、有序 stream；~10k stars，Sourcegraph 出品 |
| 熔断器 | `github.com/sony/gobreaker` | 经典三态状态机（Closed→Open→Half-Open）；单文件实现，代码可直接审计；~3.6k stars，Sony 出品 |
| 高性能 JSON | `github.com/bytedance/sonic` | JIT + SIMD，Unmarshal ~5x、Marshal ~3x 快于标准库；API 与 `encoding/json` 完全兼容；非 amd64/arm64 自动 fallback；~8k stars，字节出品 |
| 错误增强 | `github.com/cockroachdb/errors` | 自动堆栈追踪、兼容 `errors.Is/As` 和 `pkg/errors`、支持 Sentry 集成；是 `pkg/errors` 的真正替代品 |
| 测试断言 | `github.com/stretchr/testify` | `assert.Equal`/`require.NoError` 等，失败信息自动格式化；含 mock、suite；~25.9k stars，Go 生态使用率最高的测试库 |
| Struct 校验 | `github.com/go-playground/validator` | struct tag 驱动字段校验（`validate:"required,email,min=2"`），支持跨字段、自定义校验器、i18n；Gin 默认集成；~19.9k stars |
| 带 TTL 的内存缓存 | `github.com/jellydator/ttlcache` | 时间驱动过期（区别于 hashicorp/lru 的容量驱动淘汰），泛型 API，支持懒加载、eviction 回调；~1.2k stars；该场景目前维护状态最好的选项（老牌替代品 patrickmn/go-cache 已事实废弃，见下方） |
| 环境变量解析 | `github.com/caarlos0/env` | struct tag 解析环境变量，零依赖，支持 `time.Duration`、自定义类型、required/default；~6.1k stars |
| 高性能结构化日志 | `go.uber.org/zap` | 比 `slog` 默认实现快 5-10x；可通过 `zapslog` 桥接作为 `slog.Handler` backend，业务代码仍面向标准 `slog` API |
| 精确十进制计算 | `github.com/shopspring/decimal` | 避免浮点精度问题；金融/计费/汇率场景必备；支持四则运算、比较、序列化 |
| goroutine 泄漏检测 | `go.uber.org/goleak` | 测试结束后检测残留 goroutine；用法：`defer goleak.VerifyNone(t)` 或 `TestMain` 中 `goleak.VerifyTestMain(m)`；**可作为 yytools test-only 依赖** |
| goroutine pool | `github.com/panjf2000/ants` | 高性能有界 goroutine 池；`Release()`/`ReleaseTimeout()` 均等 worker 全部退出才返回，`Close()` 语义完整；~14k stars |
| 定时任务调度 | `github.com/robfig/cron` | 标准 cron 表达式 + 秒级调度；`Stop()` 返回 `context.Context`，调用方通过 `<-ctx.Done()` 等待所有 job 完成；~14k stars |

## 观察中（尚不稳定，谨慎引入）

已知问题或成熟度不足，暂不推荐生产使用，但值得持续关注。

| 场景 | 库 | 说明 |
|------|-----|------|
| 高性能并发 Map | `github.com/puzpuzpuz/xsync` | 性能优于 `sync.Map`，有泛型 `MapOf[K,V]`；但 stars 较少（~1.5k），v4 要求 Go 1.24，存在未关闭的并发正确性 [issue #145](https://github.com/puzpuzpuz/xsync/issues/145) |

## 原则

遇到已有成熟实现的场景，优先引入经过生产验证的库，不在 yytools 中重复造轮子。

判断是否应自己实现的标准：yytools 的实现是否比现有成熟库更有价值（如更轻量、更贴合 yytools API 风格、减少外部依赖）？若否，直接用成熟库。

## 如何寻找候选库

在决定引入或自己实现之前，按顺序执行以下两步：

1. **先查 [awesome-go.com](https://awesome-go.com)**：Go 生态最权威的社区精选目录，按场景分类，覆盖全面，是查找候选库的第一入口。
2. **再结合经验或 Google 搜索验证**：确认候选库的维护状态（最近提交、stars 趋势、open issues）、生产用户、与场景的契合度，选出最成熟可靠的选项后再决定是否引入。
