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
    - `concurrency/unbounded_channel/`: Unbounded channel implementation with multiple variants
    - `concurrency/workerpool/`: Goroutine worker pool with pipeline support
    - `timer/timingwheel/`: Time wheel based timer
    - `timer/delayqueue/`: Delay queue implementation

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