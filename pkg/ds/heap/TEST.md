# heap 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `heap_test.go` | 最小堆：Push、Pop、Peek（含空堆零值返回）、大数据集堆序验证 |
| `max_heap_test.go` | 最大堆：降序弹出、空堆零值返回、结构体元素 |
| `priority_queue_test.go` | 优先级队列：Push/Pop 顺序、UpdatePriority、Index 跟踪、空队列零值返回 |
| `correctness_test.go` | 命题1 MinHeap最小值语义（10万次随机Push/Pop与参考模型逐步对比）；命题2 MaxHeap最大值语义（同上）；命题3 PriorityQueue最高优先级语义（同上）；命题4 UpdatePriority不破坏堆序（5k元素+2k次随机更新后Pop全部验证降序） |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/ds/heap/

# 竞态检测
go test -race ./pkg/ds/heap/

# 基准测试
go test -bench=. -benchmem -benchtime=3s ./pkg/ds/heap/
```

## 性能基准（Apple M4，benchtime=2s）

| 操作 | ns/op | B/op | allocs/op |
|------|-------|------|-----------|
| MinHeap Push | 35.96 | 63 | 1 |
| MinHeap Pop | 242.4 | 0 | 0 |
| MinHeap Peek | 0.23 | 0 | 0 |
| MaxHeap Push | 32.44 | 56 | 1 |
| MaxHeap Pop | 248.5 | 0 | 0 |
| PriorityQueue Push | 39.75 | 73 | 1 |
| PriorityQueue Pop | 271.7 | 0 | 0 |

## 注意

- Heap/MaxHeap/PriorityQueue **非 goroutine-safe**，并发访问需调用方加锁
- 空堆 Pop/Peek 返回 nil，不 panic（与项目越界约定一致）
