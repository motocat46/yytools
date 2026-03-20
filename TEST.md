# TEST.md — yytools 测试指南

## 命令速查

### 日常开发 / CI（推荐）

```bash
go test ./... -race -short
```

- `-race`：竞态检测，并发代码必须通过
- `-short`：跳过各包中标注了 `testing.Short()` 的耗时测试（千万级压测、长时间运行等）
- 耗时约 30–60 秒，适合提交前验证

### 打 tag / 上线前完整验证

```bash
go test ./... -race -count=1
```

- 去掉 `-short`，运行全部测试，包括大规模压测
- **`-count=1` 强制禁用缓存**，确保每个测试真实执行，见下方说明
- 各包完整耗时见对应 `TEST.md`

### 只跑某个包

```bash
go test ./pkg/ds/sorted_set/... -v -race -short
```

### 筛查失败结果

测试失败时默认**不中断**，所有包跑完后汇总。面对大量输出时，用 grep 快速定位：

```bash
go test ./... -race -short 2>&1 | grep -E "FAIL|ok"
```

输出示例：
```
ok    github.com/motocat46/yytools/pkg/ds/sorted_set        7.9s
ok    github.com/motocat46/yytools/pkg/infra/safeexec       6.8s
FAIL  github.com/motocat46/yytools/pkg/algorithms/sort      [build failed]
```

全部为 `ok` 即可打 tag；出现 `FAIL` 则定位该包单独排查。

> 保留 `ok` 行的原因：只 grep `FAIL` 时，若全部通过则输出为空，无法确认命令是否真正执行。

### CI 快速失败模式

```bash
go test ./... -race -short -failfast
```

`-failfast`：遇到第一个失败立即终止，适合 CI 快速反馈。

---

## 测试缓存

Go 会缓存测试结果：测试代码和被测代码均未变化时，直接复用上次结果，输出 `(cached)`，**不重新执行**。

```
ok  github.com/motocat46/yytools/pkg/ds/sorted_set  (cached)
```

**禁用缓存的方式：**

```bash
# 方式一：-count=1（官方推荐，语义明确）
go test ./... -race -count=1

# 方式二：清空全局缓存
go clean -testcache
```

**打 tag 前必须用 `-count=1`**，原因：缓存只看代码哈希，不感知外部状态（系统时间、环境变量、文件系统）。
代码没变但环境变了的情况下，缓存会掩盖真实问题。

---

## 测试失败行为说明

| 层次 | 行为 |
|------|------|
| `t.Error` / `t.Errorf` | 标记当前测试失败，**继续执行**函数剩余代码 |
| `t.Fatal` / `t.Fatalf` | 标记失败，**立即停止**当前测试函数 |
| 子测试（`t.Run`）中 Fatal | 只停止当前子测试，兄弟子测试和父测试**继续运行** |
| 跨测试函数 | 一个函数失败**不影响**其他函数，全部跑完 |
| 跨包（`./...`）| 某包失败，其他包**继续运行**，最后汇总报告 |

---

## 预期输出中的 slog 噪声处理

`safeexec` 等包的实现在捕获 panic 时会调用 `slog.Error` 写日志。
测试中故意触发 panic 以验证恢复行为时，这些日志会出现在测试输出里，容易误读为测试错误。

处理方式：在测试文件中用 `TestMain` 将 slog 重定向到 `io.Discard`：

```go
func TestMain(m *testing.M) {
    // panic 恢复日志是预期行为，静默输出避免污染测试结果
    slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
    os.Exit(m.Run())
}
```

若将来需要断言日志内容，改为自定义 `slog.Handler` 写入 `t.Log`——
输出通过测试框架走，只在 `-v` 或测试失败时显示。

---

## 各包测试文档

每个包的详细测试命令、数据规模说明和基准基线见对应 `TEST.md`：

| 包 | TEST.md |
|----|---------|
| `pkg/algorithms/idgen/snowflake` | [TEST.md](pkg/algorithms/idgen/snowflake/TEST.md) |
| `pkg/algorithms/binary_search` | [TEST.md](pkg/algorithms/binary_search/TEST.md) |
| `pkg/algorithms/sort` | [TEST.md](pkg/algorithms/sort/TEST.md) |
| `pkg/algorithms/mathx` | [TEST.md](pkg/algorithms/mathx/TEST.md) |
| `pkg/algorithms/mathx/sampling` | [TEST.md](pkg/algorithms/mathx/sampling/TEST.md) |
| `pkg/ds/heap` | [TEST.md](pkg/ds/heap/TEST.md) |
| `pkg/ds/queue` | [TEST.md](pkg/ds/queue/TEST.md) |
| `pkg/ds/stack` | [TEST.md](pkg/ds/stack/TEST.md) |
| `pkg/ds/sorted_set` | [TEST.md](pkg/ds/sorted_set/TEST.md) |
| `pkg/infra/concurrency/unbounded_channel` | [TEST.md](pkg/infra/concurrency/unbounded_channel/TEST.md) |
| `pkg/mechanics/distribution/tiered_cycle` | [TEST.md](pkg/mechanics/distribution/tiered_cycle/TEST.md) |
| `pkg/mechanics/distribution/progressive_weight_cycle` | [TEST.md](pkg/mechanics/distribution/progressive_weight_cycle/TEST.md) |
| `pkg/slicex` | [TEST.md](pkg/slicex/TEST.md) |
