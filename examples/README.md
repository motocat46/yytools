# 示例索引

各包的示例代码以 `example_test.go` 形式存放在包目录内，可直接用 `go test -run Example` 运行，也会出现在 `go doc` 输出中。

## 可运行示例

### 数据结构

| 包 | 示例文件 | 示例函数 |
|----|---------|---------|
| `pkg/ds/sorted_set` | [example_test.go](../pkg/ds/sorted_set/example_test.go) | 游戏积分排行榜、积分更新后排名调整 |
| `pkg/ds/heap` | [example_test.go](../pkg/ds/heap/example_test.go) | 最小堆按权重出堆、优先级队列动态更新 |
| `pkg/ds/queue` | [example_test.go](../pkg/ds/queue/example_test.go) | FIFO 基本用法、Range 非消费遍历 |
| `pkg/ds/stack` | [example_test.go](../pkg/ds/stack/example_test.go) | LIFO 基本用法、Top 查看栈顶 |

### 算法

| 包 | 示例文件 | 示例函数 |
|----|---------|---------|
| `pkg/algorithms/sort` | [example_test.go](../pkg/algorithms/sort/example_test.go) | QuickSort、CountingSort、QuickSortDesc |
| `pkg/algorithms/idgen/snowflake` | [example_test.go](../pkg/algorithms/idgen/snowflake/example_test.go) | 标准布局生成 ID、自定义位布局 |

### 基础设施

| 包 | 示例文件 | 示例函数 |
|----|---------|---------|
| `pkg/infra/safeexec` | [example_test.go](../pkg/infra/safeexec/example_test.go) | panic 恢复、panic 转 error |
| `pkg/infra/concurrency/unbounded_channel` | [example_test.go](../pkg/infra/concurrency/unbounded_channel/example_test.go) | 生产者消费者、select 用法、背压演示 |

### 规则机制

| 包 | 示例文件 | 示例函数 |
|----|---------|---------|
| `pkg/mechanics/distribution/tiered_cycle` | [example_test.go](../pkg/mechanics/distribution/tiered_cycle/example_test.go) | 双层保底抽卡引擎 |
| `pkg/mechanics/distribution/progressive_weight_cycle` | [example_test.go](../pkg/mechanics/distribution/progressive_weight_cycle/example_test.go) | 渐进式权重周期 |

## 运行示例

```bash
# 运行指定包的所有示例
go test -run Example ./pkg/ds/sorted_set/
go test -run Example ./pkg/ds/heap/
go test -run Example ./pkg/infra/safeexec/

# 运行所有包的示例
go test -run Example ./...
```

## 演示程序

`cmd/demo` 提供 CLI + HTTP 可视化服务器，对各模块运行压测：

```bash
go run ./cmd/demo --help          # 查看所有命令
go run ./cmd/demo sort 5          # 排序算法压测
go run ./cmd/demo http            # 启动可视化服务（访问 http://localhost:8081）
```
