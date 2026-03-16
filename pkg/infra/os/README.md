# os 使用文档

对标准库 `os` 的扩展，提供文件存在性检测和备份功能。

## API

### `IsFileExist(file string) (error, bool)`

检测路径是否存在。结合 `os.Stat` 和 `os.IsExist` 双重判断，覆盖边缘情况。

- `bool=true`：路径存在
- `bool=false`：路径不存在或出错，`error` 含具体原因

### `IsFileNormalStat(file string) (error, bool)`

通过 `os.Stat` 判断路径是否正常可访问。

- `error=nil, bool=true`：路径存在且可访问
- 其他情况：`bool=false`，`error` 含具体原因

### `BackupFile(file string) (error, bool)`

将指定文件重命名为带时间戳的备份文件名，原位置文件被移走。

**命名格式：** `原路径_年月日时分秒序号`

例如：`~/work/config.json` → `~/work/config.json_202306071537010001`

- 序号从 1 开始，最多尝试 9999 个后缀（防止同秒内多次备份冲突）
- 原文件不存在时返回 `bool=false`

## 使用示例

```go
import yos "github.com/motocat46/yytools/pkg/infra/os"

// 检测文件是否存在
err, exists := yos.IsFileExist("/path/to/file.txt")
if exists {
    // 文件存在
}

// 备份配置文件后再写入新配置
err, ok := yos.BackupFile("/etc/app/config.json")
if !ok {
    log.Printf("backup failed: %v", err)
    return
}
// 写入新配置...
```

## 注意事项

- 函数签名返回 `(error, bool)`，注意顺序与标准库惯例 `(value, error)` **相反**
- `BackupFile` 执行的是 `os.Rename`（移动），不是复制，原路径文件操作后**不再存在**
