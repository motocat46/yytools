# stack 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `stack_test.go` | Push/Pop/Top（含空栈 panic）、LIFO 顺序、自动缩容、指针/结构体元素 |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/ds/stack/

# 竞态检测
go test -race ./pkg/ds/stack/

# 基准测试
go test -bench=. -benchmem -benchtime=3s ./pkg/ds/stack/
```

## 性能基准（Apple M4，benchtime=2s）

| 操作 | ns/op | B/op | allocs/op |
|------|-------|------|-----------|
| Push | 9.92 | 49 | 0 |
| Pop | 5.40 | 8 | 0 |
| Top | 0.37 | 0 | 0 |

## 注意

- Stack **非 goroutine-safe**，并发访问需调用方加锁
- 空栈 Pop/Top 会 panic；调用前须先用 `Empty()` 检查（泛型 T 的零值无法区分"无数据"和"真实零值元素"）
