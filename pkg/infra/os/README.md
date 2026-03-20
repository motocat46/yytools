# os 使用文档

对标准库 `os` 的扩展，提供文件存在性检测和备份功能。

## API

### `IsFileExist(file string) (bool, error)`

检测路径是否存在。

- `(true, nil)`：路径存在
- `(false, nil)`：路径不存在（不存在是正常情况，不是错误）
- `(false, err)`：权限不足等系统错误

### `IsFileNormalStat(file string) (bool, error)`

通过 `os.Stat` 判断路径是否正常可访问，直接透传 Stat 错误。

- `(true, nil)`：路径存在且可访问
- `(false, err)`：路径不存在或不可访问，`err` 含具体原因

### `BackupFile(file string) (bool, error)`

将指定文件重命名为带时间戳的备份文件名，原位置文件被移走。

**命名格式：** `原路径_年月日时分秒序号`

例如：`~/work/config.json` → `~/work/config.json_202306071537010001`

- 序号从 1 开始，最多尝试 9999 个后缀（防止同秒内多次备份冲突）
- 原文件不存在时返回 `(false, nil)`

## 使用示例

```go
import yos "github.com/motocat46/yytools/pkg/infra/os"

// 检测文件是否存在
exists, err := yos.IsFileExist("/path/to/file.txt")
if err != nil {
    log.Printf("stat error: %v", err) // 权限不足等系统错误
}
if exists {
    // 文件存在
}

// 备份配置文件后再写入新配置
ok, err := yos.BackupFile("/etc/app/config.json")
if !ok || err != nil {
    log.Printf("backup failed: %v", err)
    return
}
// 写入新配置...
```

## 注意事项

- `IsFileExist` 文件不存在时返回 `(false, nil)`，不返回 error；仅系统级错误（如权限不足）才返回非 nil error
- `BackupFile` 执行的是 `os.Rename`（移动），不是复制，原路径文件操作后**不再存在**
- `BackupFile` 非并发安全，多 goroutine 同时备份同一文件需自行加锁
