# trie

并发安全的前缀树（Trie），支持 Unicode，`sync.RWMutex` 读写分离。

## 适用场景

- 前缀查询 / 自动补全（输入法、命令提示）
- 路由前缀匹配
- 字典查找

## 快速上手

```go
import "github.com/motocat46/yytools/pkg/ds/trie"

tr := trie.New()
tr.Insert("apple")
tr.Insert("application")
tr.Insert("app")

tr.Search("apple")      // true
tr.Search("ap")         // false（仅前缀，不是完整词）
tr.HasPrefix("app")     // true
tr.WithPrefix("app")    // 返回 ["app", "apple", "application"]（顺序不保证，示例已排序）

tr.Delete("apple")
tr.Len()                // 2
```

## API

### `New() *Trie`

创建空 Trie。

### `Insert(word string) bool`

插入词。新插入返回 `true`，词已存在返回 `false`。支持空字符串。

### `Search(word string) bool`

精确匹配。词存在返回 `true`，仅是前缀（或不存在）返回 `false`。

### `HasPrefix(prefix string) bool`

判断是否存在以 `prefix` 开头的词。`prefix=""` 等价于 `Len() > 0`。

### `WithPrefix(prefix string) []string`

返回所有以 `prefix` 开头的词，**顺序不保证**。需要字典序时自行 `sort.Strings(result)`。`prefix=""` 返回全部词。无匹配时返回 `nil`。

### `Delete(word string) bool`

删除词并剪枝空节点。词存在返回 `true`，不存在返回 `false`。

### `Len() int`

返回当前词数。

## 注意事项

- `WithPrefix` 不保证顺序，如需排序请自行调用 `sort.Strings`
- 并发安全：读操作（Search/HasPrefix/WithPrefix/Len）使用读锁，写操作（Insert/Delete）使用写锁
- `Delete` 会剪枝：被删词的路径上如有不再使用的节点，同步回收，无内存泄漏
