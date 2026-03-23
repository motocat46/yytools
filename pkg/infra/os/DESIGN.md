# infra/os 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、IsFileExist 的三态语义：(bool, error)

标准库 `os.Stat` 返回的 error 有两种含义：
- **路径不存在**：`os.IsNotExist(err) == true`
- **系统错误**（权限不足、IO 错误等）：需要调用方处理

对调用方而言，"文件不存在"是正常情况（通常表示"尚未创建"），并非错误；而权限错误需要另外处理。

`IsFileExist` 将这两种情形显式分离：

| 返回值 | 含义 |
|--------|------|
| `(true, nil)` | 路径存在 |
| `(false, nil)` | 路径不存在（正常情况）|
| `(false, err)` | 系统错误，err 非 nil |

这让调用方代码更清晰：
```go
// 无需解析 err 类型，直接用 bool
if exists, err := yos.IsFileExist(path); err != nil {
    // 处理系统错误
} else if exists {
    // 路径已存在
}
```

---

## 二、BackupFile 的命名策略

备份文件名格式：`原路径_YYYYMMDDHHmmssXXXX`（时间戳 + 序号）。

序号范围 `[1, 9999]`：同一秒内最多支持 9999 次备份，超出返回 `os.ErrExist`。这是有意为之的防护，避免意外的无限循环导致文件系统被填满。
