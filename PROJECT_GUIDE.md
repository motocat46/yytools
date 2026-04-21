# Project Guide

Project-level guidance for AI coding agents in this repo.

@~/.config/ai/shared/go.md
@~/.config/ai/shared/infra_trust_boundary.md
@~/.config/ai/shared/infra_trust_boundary_go.md

用中文回答.

## Project Overview

yytools: Go utilities library. Cover algorithms, data structures, common tools. Flat package layout. Prioritize performance + correctness.

## Architecture

### Core Structure
- `pkg/`: core packages by domain
  - `algorithms/`: `binary_search/`, `mathx/` (`bits/`, `overflow/`, `probability_distribution/`, `random/`), `sort/`, `idgen/snowflake/`
  - `slicex/`: slice helpers (`MinInSlice`, `MaxInSlice`, `MinBy`, `MaxBy`, etc.)
  - `common/`: `assert/`, `base/`, `cpu/`
  - `numconst/`: numeric + time constants
  - `ds/`: `heap/`, `queue/`, `stack/`, `sorted_set/`, `lru/`, `trie/`
  - `infra/`: `safeexec/`, `timeutil/`, `timecond/`, `os/`, `concurrency/unbounded_channel/`, `concurrency/workerpool/`, `timer/timingwheel/`, `timer/delayqueue/`
    - `unbounded_channel/`: 机制层，调用方管生命周期
    - `workerpool/`: 机制层，调用方管生命周期；内部业务项目经工程基础层 `TaskExecutor`
    - `timingwheel/`: 机制层，调用方管 `Start/Stop`；内部业务项目经工程基础层 `AppScheduler`
    - `delayqueue/`: 同上
- `internal/bench/`: benchmark/demo runners. Cover `heap`, `queue`, `stack`, `sort`, `sorted_set`, `mathx`, `probability_distribution`
- `cmd/demo/`: HTTP demo + benchmark visualization server

## Common Commands

### Building and Running
```bash
# Build the project
go build ./...

# Start performance visualization server (http://localhost:8081)
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

### 遇到障碍时的行为规范

**遇到障碍必须停下来报告，不得自行绕过。**

判断标准：
- 障碍在**我写的代码内部**（编译错、测试失败）→ 自修
- 障碍涉及**目录结构、配置、.gitignore、git 操作失败**等设计决策 → 停下报告，等指示

git 失败通常反映意图，不当偶发错误绕过。

### 叶子包交付强制规范（pre-commit hook 执行）

**新建 `pkg/` 叶子包时，以下三个文件必须在同一次 commit 中提交，缺一不可：**

| 文件 | 内容 |
|------|------|
| `README.md` | 使用者视角：功能、快速上手、完整 API、注意事项 |
| `TEST.md` | 测试文件列表、分层命令（快速/全量/压测）、性能基线 |
| `DESIGN.md` | 选型理由、关键决策、踩坑 |

pre-commit hook 自动检查。缺任一文件直接拒绝 commit。纯常量包等例外可用 `git commit --no-verify`，message 说明原因。

### 数据结构 API 边界规范

`pkg/ds/` 下的数据结构根据**返回类型**采用不同的空/越界策略：

#### 指针返回方法 → 返回 nil，不 panic

适用：`sorted_set`、`heap`、`priority_queue` 等返回 `*NodeData`、`*Item`、`*PriorityItem` 的方法。

- `nil` 明确表示"没有该元素"，调用方直接做 nil 检查
- 例：`GetByRankDesc(999)` 在只有 3 个元素时返回 nil，不 panic
- 例：`Heap.PopItem()` 空堆返回 nil，不 panic

这是有序集合类 API 惯例（参考 Redis ZADD/ZRANK 语义）。

#### 泛型 T 返回方法 → panic，不返回零值

适用：`Stack[T].Pop()`、`Stack[T].Top()`、`Queue[T].Dequeue()`、`Queue[T].Peek()` 等返回泛型 `T` 的方法。

- `T` 的零值（`0`、`""`、`false` 等）可能是合法元素，调用方**无法区分**"空集合"和"真实零值元素"
- 返回零值是**静默错误**，比 panic 更危险
- 先用 `Empty()` 检查；空时操作直接 panic，逼调用方写防御代码

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
- 单测文件：`*_test.go`
- 正确性命题：`correctness_test.go`
- 基准：`bench_test.go`
- 性能可视化：go-echarts，`http://localhost:8081`

### Assertion System
- 全仓统一用 `pkg/common/assert/`
- Assertion 永远开启，不能关闭
- 合约违例（caller bug）用 `assert.Assert(condition, message...)`
- 简单布尔检查用 `assert.AssertFast(condition)`
- 运行时边界保护直接 `panic(...)`，如时间戳溢出、类型范围越界

### Code Conventions
- 泛型约束用 `pkg/common/base/`
- 中文注释
- 所有文件保留 Apache 2.0 license header
- 保持 flat package；除非 generic + concrete 都真实存在，否则不要加 `concrete/generic` 中间目录

### 新组件评估：先查外部库

**禁止在完成外部库搜索前建议自行实现任何新组件。** 搜索步骤：先查 awesome-go.com，再 GitHub/Google 验证维护状态。成熟库已覆盖 → 加 `docs/recommended-libraries.md`，不在 yytools 实现。

### 复用优先：实现前先搜索

内联实现工具代码前，先确认以下包是否已有现成实现：

| 需求 | 检查位置 |
|------|---------|
| panic 恢复 / 安全执行 | `pkg/infra/safeexec/` — `Safe`, `SafeExec`, `SafeExecErr`, `SafeExecVal[T]` |
| 运行时断言 | `pkg/common/assert/` |
| 时间解析 / 工具 | `pkg/infra/timeutil/` |
| 数值 / 时间常量 | `pkg/numconst/` |

禁止重复造轮子；会把维护成本翻倍，也容易出现语义分叉。

### Performance Considerations
- 排序算法要和 Go 标准库做性能对比
- 可视化服务启动命令：`go run ./cmd/demo http`，地址 `http://localhost:8081`
- 可视化入口在 `cmd/demo/registry.go`；各 `api_*.go` 通过 `init()` 注册
- Counting sort / quick sort 针对不同数据范围做优化

### yytools 边界判断标准

判断一个模块是否属于 yytools，依次回答三个问题：

**Q1：是否无策略、无生命周期、低依赖？**
- 无策略：不做系统级决策（不决定日志格式、重试次数、超时配置）
- 无生命周期：不需要与应用启动/关闭钩子绑定（内部资源管理不算）
- 低依赖：不引入重量级第三方依赖（不传递性污染调用方依赖图）

全满足 → 属于 yytools。有一项不满足 → 进 Q2。

**Q2：是"机制"还是"集成"？**
- 机制：提供算法或并发原语，调用方自己决定怎么用、何时启停
- 集成：需要被绑定到系统入口，或依赖系统级上下文才能正确工作

是机制 → 属于 yytools，文档明确"调用方负责生命周期管理"。是集成 → 属于工程基础层。

**Q3：强制所有使用者依赖工程基础层才能用到此模块，代价是否可接受？**

可接受 → 可考虑迁移到工程基础层。不可接受（工程基础层有重依赖或强约束）→ 保留在 yytools，靠 depguard 约束内部使用。
