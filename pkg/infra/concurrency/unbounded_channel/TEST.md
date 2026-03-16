# UnboundedChannel V6 测试指南

---

## 测试文件总览

| 文件 | build tag | 内容 |
|------|-----------|------|
| `correctness_test.go` | 无 | FIFO 语义、消息完整性、背压、生命周期安全 |
| `linearizability_test.go` | 无 | Porcupine 形式化线性一致性验证 |
| `compare_test.go` | 无 | 单/多生产者 FIFO、慢消费者、NativeChan vs V6 性能基准 |
| `stress_test.go` | `stress` | 长时间随机化压力测试（需 `-tags stress` 才编译） |

---

## 层次 1：快速验证（CI / 每次提交前）

覆盖全部正确性测试，约 10 秒。

```bash
go test ./pkg/infra/concurrency/unbounded_channel/ \
    -run "TestFIFO|TestIntegrity|TestBackpressure|TestLifecycle|TestLinearizability|TestCompare" \
    -race -v -timeout 300s
```

**覆盖范围：**
- FIFO：快速路径、慢路径（chanSize=1）、分批发送、double-enqueue 回归
- 完整性：多生产者无丢失/无重复、单生产者 per-producer 顺序验证
- 背压：生产者确实阻塞、消费后确实解除、Close() 解除所有阻塞（死锁测试）
- 生命周期：goroutine 无泄漏、Close 后 buffer 消息仍可消费、并发 Send+Close 无 panic
- Porcupine：单/多生产者线性一致性验证、高背压场景、100 轮压测

---

## 层次 2：压力测试（上线前 / 服务器长跑）

`stress_test.go` 带有 `//go:build stress`，不加 `-tags stress` 时**不参与编译**，
对 `go test ./...` 零干扰。

### `-timeout` 即运行时长

压力测试内部使用 `t.Deadline()` 自适应，只有一条规则：

```
实际运行时长 = -timeout - 30s（最后一轮的清理余量）
```

| 想跑多久 | 命令 |
|---------|------|
| 30 分钟 | `go test -tags stress -run TestStress -timeout 30m30s` |
| 1 小时  | `go test -tags stress -run TestStress -timeout 1h0m30s` |
| 1 天    | `go test -tags stress -run TestStress -timeout 24h0m30s` |
| 1 个月  | `go test -tags stress -run TestStress -timeout 720h0m30s` |

**建议加 `-v` 查看每分钟进度报告：**

```bash
go test -tags stress -run TestStress -v \
    ./pkg/infra/concurrency/unbounded_channel/ \
    -timeout 30m30s
```

**压力测试场景（每轮随机选一种）：**
- FIFO round：单生产者，严格验证顺序
- Integrity round：多生产者，验证无丢失/无重复
- Backpressure round：小 limit + 慢消费者，反复触发背压与解压
- Chaos round：并发 Send/Receive/Close，验证无 panic/死锁

---

## 层次 3：全量（含性能基准）

### 正确性全量（不含压力测试）

```bash
go test ./pkg/infra/concurrency/unbounded_channel/ -race -timeout 300s
```

### 性能基准（NativeChan vs V6）

```bash
go test ./pkg/infra/concurrency/unbounded_channel/ \
    -bench="BenchmarkQPS_NativeChan|BenchmarkQPS_V6|BenchmarkQPS_Throughput" \
    -benchmem -benchtime=3s -run "^$" -timeout 300s
```

**参考结果（Apple M4，arm64）：**

| 场景 | NativeChan | V6 | 差距 |
|------|-----------|-----|------|
| Large chanSize | 83.89 ns/op | 88.50 ns/op | +5.5% |
| Small chanSize | 85.28 ns/op | 140.6 ns/op | +64.9% |
| 10k chanSize | 82.62 ns/op | 82.31 ns/op | ≈持平 |
| 吞吐量（100k 消息）| 28.2M msg/s | 20.2M msg/s | -28.3% |

---

## 注意事项

**不要把压力测试与正确性测试混在同一次 `go test` 调用里。**

`-timeout` 是整个测试二进制的强杀时间，由所有测试共享。
若混跑，`TestStress` 会耗尽几乎所有时间，其他测试可能来不及执行。

```bash
# ✅ 正确：分两次运行
go test ./pkg/infra/concurrency/unbounded_channel/ -race -timeout 300s
go test -tags stress -run TestStress -timeout 30m30s ./pkg/infra/concurrency/unbounded_channel/

# ❌ 警告：混在一起，stress 可能会吞噬所有时间
go test -tags stress ./pkg/infra/concurrency/unbounded_channel/ -timeout 30m
```