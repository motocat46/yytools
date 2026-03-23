# cpu 测试说明

## 无独立测试文件

`cpu` 包提供 CPU 特性检测（缓存行大小等硬件参数），通过平台特定文件实现（`cpu_arm64.go`、`cpu_x86.go`、`cpu_other.go`）。

无独立测试的原因：
- 返回值是硬件常量，无法在测试中断言具体数值（不同机器不同）
- 正确性由使用方验证（如 `unbounded_channel` 用 `CacheLinePad` 避免 false sharing，其并发测试通过 `-race` 覆盖）

## 关联验证命令

```bash
# 验证可正常编译（跨平台）
go build ./pkg/common/cpu/

# 运行使用 cpu 包的并发测试
go test -race ./pkg/infra/concurrency/...
```
