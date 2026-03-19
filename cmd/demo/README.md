# cmd/demo

命令行演示工具，用于运行各模块的基准测试和性能可视化。

## 用法

```bash
# 查看所有可用命令
go run ./cmd/demo help

# 运行指定模块的演示（iterations 为迭代次数）
go run ./cmd/demo heap 5
go run ./cmd/demo sortedset 10
go run ./cmd/demo sort 3

# 运行所有模块
go run ./cmd/demo all 3

# 启动性能可视化 HTTP 服务（http://localhost:8081）
go run ./cmd/demo http
```

## 可用命令

| 命令 | 说明 |
|------|------|
| `heap` | 最小堆 |
| `maxheap` | 最大堆 |
| `mathcommon` | 公共数学方法（GCD 等） |
| `prob` | 概率分布 |
| `pq` | 优先级队列 |
| `queue` | 队列 |
| `sort` | 排序算法对比 |
| `sortedset` | 有序集合 |
| `stack` | 栈 |
| `all` | 运行所有模块 |
| `http` | 启动 go-echarts 性能可视化服务 |

## 说明

- 此工具仅供开发调试使用，不对外暴露 API
- 底层实现在 `internal/bench/` 各子目录
- HTTP 可视化使用 go-echarts，排序对比图等可在浏览器中查看
