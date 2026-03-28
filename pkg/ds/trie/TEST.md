# trie 测试说明

## 测试文件

| 文件 | 包 | 覆盖范围 |
|------|-----|---------|
| `trie_test.go` | `trie`（内部）| 单元测试：Insert/Search/HasPrefix/WithPrefix/Delete/边界；基准测试 |
| `correctness_test.go` | `trie_test`（外部）| 正确性命题：随机混合参考模型对比（100k ops）、并发安全 |
| `example_test.go` | `trie_test`（外部）| 可运行示例 |

## 快速验证

```bash
# 单元测试
go test ./pkg/ds/trie/ -v

# 含竞态检测
go test -race ./pkg/ds/trie/ -v

# 跳过大规模测试（CI 快速通道）
go test -short ./pkg/ds/trie/ -v
```

## 全量测试（含正确性命题）

```bash
go test -race ./pkg/ds/trie/ -v -run "TestCorrectness"
```

## 基准测试

```bash
go test -bench=. -benchmem -benchtime=3s ./pkg/ds/trie/
```

## 基准基线（Apple M4，Go 1.24，-benchtime=3s）

```
BenchmarkSearch/n=100          30.82 ns/op    0 B/op   0 allocs/op
BenchmarkSearch/n=1000         49.16 ns/op    0 B/op   0 allocs/op
BenchmarkSearch/n=10000        66.96 ns/op    0 B/op   0 allocs/op
BenchmarkSearch/n=100000      117.2  ns/op    0 B/op   0 allocs/op
BenchmarkSearch/n=1000000     281.9  ns/op    0 B/op   0 allocs/op

BenchmarkInsert/n=100          66.70 ns/op    0 B/op   0 allocs/op
BenchmarkInsert/n=1000        125.1  ns/op    0 B/op   0 allocs/op
BenchmarkInsert/n=10000       184.0  ns/op    0 B/op   0 allocs/op
BenchmarkInsert/n=100000      370.1  ns/op    6 B/op   0 allocs/op
BenchmarkInsert/n=1000000     908.9  ns/op  117 B/op   1 allocs/op

BenchmarkHasPrefix/n=100       40.70 ns/op    0 B/op   0 allocs/op
BenchmarkHasPrefix/n=1000      61.62 ns/op    0 B/op   0 allocs/op
BenchmarkHasPrefix/n=10000     73.94 ns/op    0 B/op   0 allocs/op
BenchmarkHasPrefix/n=100000    97.96 ns/op    0 B/op   0 allocs/op
BenchmarkHasPrefix/n=1000000  161.8  ns/op    0 B/op   0 allocs/op

BenchmarkMixed/n=100           64.26 ns/op      3 B/op    0 allocs/op
BenchmarkMixed/n=1000         178.2  ns/op     30 B/op    0 allocs/op
BenchmarkMixed/n=10000       1259    ns/op    332 B/op    5 allocs/op
BenchmarkMixed/n=100000     18186    ns/op   3836 B/op   52 allocs/op
BenchmarkMixed/n=1000000   183717    ns/op  43547 B/op  487 allocs/op

BenchmarkConcurrent_ReadHeavy/p=1    237.6 ns/op   0 B/op   0 allocs/op
BenchmarkConcurrent_ReadHeavy/p=4    297.8 ns/op   0 B/op   0 allocs/op
BenchmarkConcurrent_ReadHeavy/p=16   345.2 ns/op   0 B/op   0 allocs/op
BenchmarkConcurrent_ReadHeavy/p=64   374.6 ns/op   0 B/op   0 allocs/op
```

**观察：**
- Search/HasPrefix 零分配，随规模增长符合 O(len) 而非 O(n) 特征
- 并发读写（ReadHeavy）p=1→p=64 仅 237→374 ns/op（约 1.6×），读写锁竞争代价较小
- BenchmarkMixed 的内存分配来自 WithPrefix 构建结果切片，与前缀命中词数成正比
