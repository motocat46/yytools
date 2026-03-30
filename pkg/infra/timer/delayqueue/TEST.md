# delayqueue 测试说明

## 快速验证

```bash
# 功能 + 竞态检测
go test -race -count=1 ./pkg/infra/timer/delayqueue/

# 详细输出
go test -race -v ./pkg/infra/timer/delayqueue/

# 跳过大规模测试（无 -short 标记的测试均较快）
go test -race -short ./pkg/infra/timer/delayqueue/
```

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `delayqueue_test.go` | Offer/TryPoll 堆顺序、Poll 阻塞/到期唤醒/ctx 取消、Offer 更早元素唤醒 Poll、多生产者并发无丢失（10万） |

`TestMain` 使用 `goleak.VerifyTestMain`，所有测试结束后自动验证无残留 goroutine。
