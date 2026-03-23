# assert 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `assertion_test.go` | `Assert`：条件为真正常执行、条件为假触发 panic 且消息格式正确；`AssertFast`：条件为真/假的基础行为 |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/common/assert/

# 竞态检测
go test -race ./pkg/common/assert/
```

## 注意

- assert 包**始终启用**，不受 build tag 控制（详见 DESIGN.md）
- 断言失败以 `panic(string)` 形式抛出，消息格式：`assertion failed: <args...>`
- 无基准测试——断言路径是热路径，性能由调用方场景决定，不在此单独测量
