# archive — unbounded_channel 历史版本

本目录存放 unbounded_channel 的历史演进版本，仅供参考。所有文件均标注 `//go:build ignore`，不参与任何构建。

## 版本说明

| 文件 | 说明 |
|------|------|
| `pure_queue.go` | 最早的纯队列原型，无 goroutine 并发 |
| `ucv1.go` ~ `ucv4.go` | 初期迭代版本，逐步引入无界发送能力 |
| `ucv5_original.go` | V5 原始版本，引入背压机制 |
| `ucv51.go` | V5.1：微调背压阈值逻辑 |
| `ucv53.go` | V5.3：修复并发竞争问题，接近当前主版本 |

当前生产实现位于上级目录 `unbounded_channel/`。
