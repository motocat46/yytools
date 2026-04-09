# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

@~/.claude/go.md
@~/.claude/infra_trust_boundary.md
@~/.claude/infra_trust_boundary_go.md

用中文回答.

## Project Overview

yytools is a Go utilities library providing algorithms, data structures, and common tools for daily development. The project uses a flat package structure, with a focus on performance and correctness.

## Architecture

### Core Structure
- `pkg/`: Core functionality organized by domain
  - `algorithms/`: Algorithm implementations (sorting, math utilities, binary search)
    - `binary_search/`: Binary search implementation
    - `mathx/`: GCD, probability distributions, common math (package: `mathx`)
      - `bits/`: Bit operation utilities
      - `overflow/`: Numeric overflow check utilities
      - `probability_distribution/`: Probability distribution utilities
      - `random/`: Random number generation
    - `sort/`: Multiple sorting algorithm implementations
    - `idgen/`: ID generation utilities
      - `snowflake/`: Snowflake algorithm based distributed ID generator
  - `slicex/`: Slice utility functions (MinInSlice, MaxInSlice, MinBy, MaxBy, etc.)
  - `common/`: Minimal common utilities
    - `assert/`: Runtime assertion framework (always enabled, cannot be disabled)
    - `base/`: Type definitions and constraints for generics (Integer, Ordered, etc.)
    - `cpu/`: CPU architecture detection utilities
  - `numconst/`: Numeric and time constants (千/万/亿, time unit constants)
  - `ds/`: Data structure implementations
    - `heap/`: Min-heap and max-heap implementations
    - `queue/`: Ring-buffer queue with auto expand/shrink
    - `stack/`: Stack implementation
    - `sorted_set/`: Skip-list based sorted set (similar to Redis ZADD)
    - `lru/`: LRU cache with TTL
    - `trie/`: Trie (prefix tree) implementation
  - `infra/`: Infrastructure utilities
    - `safeexec/`: panic-safe function execution wrappers (Safe, SafeExec, SafeExecErr, SafeExecVal[T])
    - `timeutil/`: Time utility functions (package: `timeutil`)
    - `timecond/`: Time condition utilities
    - `os/`: OS utility wrappers
    - `concurrency/unbounded_channel/`: Unbounded channel implementation with multiple variants（机制层，调用方负责生命周期管理）
    - `concurrency/workerpool/`: Goroutine worker pool with pipeline support（机制层，调用方负责生命周期管理；内部业务项目须通过工程基础层 TaskExecutor 使用）
    - `timer/timingwheel/`: Time wheel based timer（机制层，调用方负责 Start/Stop；内部业务项目须通过工程基础层 AppScheduler 使用）
    - `timer/delayqueue/`: Delay queue implementation（机制层，同上）

- `internal/bench/`: Benchmark/demo runner functions (not for external use)
  - `heap/`, `queue/`, `stack/`, `sort/`, `sorted_set/`, `mathx/`, `probability_distribution/`

- `cmd/demo/`: Demo application with CLI and HTTP visualization server

## Common Commands

### Building and Running
```bash
# Build the project
go build ./...

# Run specific functionality tests (entry: cmd/demo/)
go run ./cmd/demo <command> <iterations>

# Available commands (get help)
go run ./cmd/demo help
```

### Testing Individual Components
```bash
# Test heap operations (5 iterations)
go run ./cmd/demo heap 5

# Test sorting algorithms (10 iterations)
go run ./cmd/demo sort 10

# Test all components (3 iterations)
go run ./cmd/demo all 3

# Start performance visualization server
go run ./cmd/demo http
```

### Standard Go Commands
```bash
# Run all Go tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test file
go test ./pkg/algorithms/mathx/random/

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

## Development Notes

### 数据结构 API 边界规范

`pkg/ds/` 下的数据结构根据**返回类型**采用不同的空/越界策略：

#### 指针返回方法 → 返回 nil，不 panic

适用：`sorted_set`、`heap`、`priority_queue` 等返回 `*NodeData`、`*Item`、`*PriorityItem` 的方法。

- `nil` 在语义上是明确的"没有该元素"，调用方可以直接做 nil 检查
- 例：`GetByRankDesc(999)` 在只有 3 个元素时返回 nil，不 panic
- 例：`Heap.PopItem()` 空堆返回 nil，不 panic

这是有序集合类 API 的惯例（参考 Redis ZADD/ZRANK 语义）。

#### 泛型 T 返回方法 → panic，不返回零值

适用：`Stack[T].Pop()`、`Stack[T].Top()`、`Queue[T].Dequeue()`、`Queue[T].Peek()` 等返回泛型 `T` 的方法。

- `T` 的零值（`0`、`""`、`false` 等）可能是合法元素，调用方**无法区分**"空集合"和"真实零值元素"
- 返回零值是**静默错误**，比 panic 更危险——调用方收到一个看似正常的值却不知道操作失败了
- 正确做法：调用前先用 `Empty()` 检查，空时操作直接 panic，迫使调用方写防御代码

```go
// 正确
if !stack.Empty() {
    item := stack.Pop()
    // ...
}

// 错误：0 可能是真实元素，也可能是空栈的静默返回
item := stack.Pop()
```

### Testing Strategy
- Unit tests follow `*_test.go` naming convention
- Test files: `_test.go` suffix; correctness propositions in `correctness_test.go`; benchmarks in `bench_test.go`
- Performance testing includes visualization via go-echarts at `http://localhost:8081`

### Assertion System
- All code uses the custom assertion framework in `pkg/common/assert/`
- Assertions are always enabled and cannot be disabled
- Use `assert.Assert(condition, message...)` for contract violations (caller bug)
- Use `assert.AssertFast(condition)` for simple boolean checks
- Use unconditional `panic(...)` for runtime boundary protection (e.g. timestamp overflow, type range exceeded)

### Code Conventions
- Generic implementations use type constraints from `pkg/common/base/`
- Chinese comments are used throughout the codebase
- Apache 2.0 license header in all files
- Flat package structure: no `concrete/generic` intermediate directories except where both truly exist

### 复用优先：实现前先搜索

内联实现工具代码前，先确认以下包是否已有现成实现：

| 需求 | 检查位置 |
|------|---------|
| panic 恢复 / 安全执行 | `pkg/infra/safeexec/` — `Safe`, `SafeExec`, `SafeExecErr`, `SafeExecVal[T]` |
| 运行时断言 | `pkg/common/assert/` |
| 时间解析 / 工具 | `pkg/infra/timeutil/` |
| 数值 / 时间常量 | `pkg/numconst/` |

重复实现已有功能不可接受：维护成本翻倍，且两份实现容易产生语义分歧。

### Performance Considerations
- Sorting algorithms include performance comparison with Go's standard library
- Use `cmd/demo/graph.go` for performance visualization at `http://localhost:8081`
- Counting sort and quick sort implementations are optimized for different data ranges

### yytools 边界判断标准

判断一个模块是否属于 yytools，依次回答三个问题：

**Q1：是否无策略、无生命周期、低依赖？**
- 无策略：不做系统级决策（不决定日志格式、重试次数、超时配置）
- 无生命周期：不需要与应用启动/关闭钩子绑定（内部资源管理不算）
- 低依赖：不引入重量级第三方依赖（不传递性污染调用方依赖图）

全部满足 → 属于 yytools。有一项不满足 → 进 Q2。

**Q2：是"机制"还是"集成"？**
- 机制：提供算法或并发原语，调用方自己决定怎么用、何时启停
- 集成：需要被绑定到系统入口，或依赖系统级上下文才能正确工作

是机制 → 属于 yytools，文档明确"调用方负责生命周期管理"。是集成 → 属于工程基础层。

**Q3：强制所有使用者依赖工程基础层才能用到此模块，代价是否可接受？**

可接受 → 可考虑迁移到工程基础层。不可接受（工程基础层有重依赖或强约束）→ 保留在 yytools，靠 depguard 约束内部使用。

### 第三方库使用策略

按库的性质分三类：

**第一类：纯工具库（算法、数据结构）**
> 例：`hashicorp/golang-lru`、`shopspring/decimal`、`google/btree`

直接使用，不加适配层。这类库无 I/O、无副作用，加适配层是过度设计。接口层的合理性由替换场景的真实存在决定，而不是由"万一需要替换"决定。

**第二类：基础设施库（数据库、缓存服务、消息队列、HTTP 客户端）**
> 例：`gorm`、`go-redis`、`kafka-go`

分两层处理：
- 连接/客户端管理（连接池、超时、重试策略）→ 工程基础层统一负责
- 业务域适配（repository 实现）→ 业务项目自己负责；接口由业务域（消费方）定义，不由工程基础层强加

**第三类：横切关注点（日志、监控、重试、熔断）**
> 例：`zap`、`opentelemetry`、`hystrix`

通过工程基础层统一封装，业务代码不直接导入。接口命名面向系统语义，不透传第三方参数签名，外部类型不泄漏到业务层。

**Wrapper 最低要求**：封装必须收回系统语义控制权，不只是改名字——只改名不收回控制权的 wrapper 没有价值。

### 内部访问控制（公司内部业务项目）

yytools 作为公开库不设访问限制。公司内部业务项目须在 CI 中配置 `depguard`，禁止直接导入以下路径，必须通过工程基础层对应封装使用：

| 禁止直接导入 | 应通过 |
|-------------|--------|
| `yytools/pkg/infra/concurrency/workerpool` | 工程基础层 `TaskExecutor` |
| `yytools/pkg/infra/concurrency/unbounded_channel` | 工程基础层对应封装 |
| `yytools/pkg/infra/timer/timingwheel` | 工程基础层 `AppScheduler` |
| `yytools/pkg/infra/timer/delayqueue` | 工程基础层 `AppScheduler` |
| `yytools/pkg/infra/safeexec` | 工程基础层 panic 上报集成 |

以下路径业务项目可直接使用，无约束：`algorithms/*`、`ds/*`、`mechanics/*`、`infra/timeutil`、`infra/timecond`、`infra/os`。