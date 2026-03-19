# cpu

CPU 平台常量与工具类型，提供跨平台的 cache line 大小常量和 padding 类型。

## 提供的功能

### `CacheLinePadSize`（编译期常量）

当前平台的 cache line 字节数：
- x86/x86_64：64
- ARM64：128
- ARM：32
- 其他平台：64（保守默认值）

### `CacheLinePad`（结构体 padding 类型）

将相邻字段隔离到不同 cache line，避免多核并发写时的伪共享（false sharing）：

```go
type Counter struct {
    val atomic.Int64
    _   cpu.CacheLinePad // 将 val 与后续字段隔离到不同 cache line
}

type ShardedCounters struct {
    counters [NumShards]Counter // 每个 Counter 独占一个 cache line
}
```

## 适用场景

- 高并发计数器、锁数组等需要避免 false sharing 的结构
- 性能调优时消除热点字段的 cache line 争用

## 与标准库的关系

命名和数值与 Go 标准库 `internal/cpu` 完全对齐，但标准库的 `internal` 包不对外暴露，本包以 build tag 分文件的方式提供相同功能。
