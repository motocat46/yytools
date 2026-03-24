# concurrency — 并发原语

## 子模块

| 包 | 简介 |
|----|------|
| [unbounded_channel](unbounded_channel/README.md) | 有序（FIFO）、无界、带背压的泛型消息通道，解决生产者/消费者速率不匹配问题 |
| [workerpool](workerpool/README.md) | 固定大小 goroutine 池 + 泛型 Pipeline，提供有背压的并发任务调度 |
