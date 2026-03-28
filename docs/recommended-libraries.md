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

## 原则

遇到已有成熟实现的场景，优先引入经过生产验证的库，不在 yytools 中重复造轮子。

判断是否应自己实现的标准：yytools 的实现是否比现有成熟库更有价值（如更轻量、更贴合 yytools API 风格、减少外部依赖）？若否，直接用成熟库。
