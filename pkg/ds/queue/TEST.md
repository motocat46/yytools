# queue 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `queue_test.go` | Enqueue/Dequeue/Peek/Empty/Range 正确性；空队列 Dequeue/Peek 必须 panic；自动扩容与缩容；指针/结构体元素；Benchmark |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/ds/queue/

# 竞态检测
go test -race ./pkg/ds/queue/

# 基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/ds/queue/
```

## 性能基准（Apple M4，benchtime=2s）

| 操作 | ns/op | allocs/op |
|------|-------|-----------|
| Enqueue | 8.43 | 0 |
| Dequeue | 8.63 | 0 |
| Peek | 0.37 | 0 |
| Range (n=100) | 78.6 | 0 |

## 注意

- Queue **非 goroutine-safe**，并发访问需调用方加锁
- 空队列 Dequeue/Peek 会 panic；调用前须先用 `Empty()` 检查（泛型 T 的零值无法区分"无数据"和"真实零值元素"）
