# sorted_set 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `sortedset_test.go` | 单元测试 + Benchmark |

## 运行方式

```bash
# 运行所有单元测试
go test ./pkg/ds/sorted_set/

# 详细输出
go test -v ./pkg/ds/sorted_set/

# 运行 Benchmark
go test -bench=. ./pkg/ds/sorted_set/

# 开启竞态检测
go test -race ./pkg/ds/sorted_set/
```

## 测试用例覆盖

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestNewSortedSet` | 初始化，长度为 0 |
| `TestNewNodeData` | 构造函数字段正确性 |
| `TestNodeData_LessThan` | 分数不同 / 相同的比较结果 |
| `TestNodeData_EqualTo` | 分数+值均等 / 任一不同的结果 |
| `TestSortedSet_Insert` | 正常插入、重复 Key 拒绝 |
| `TestSortedSet_Get` | 存在的 Key / 不存在的 Key |
| `TestSortedSet_Delete` | 存在 Key 删除成功、不存在 Key 返回 false |
| `TestSortedSet_GetRank` | 多元素排名正确性、不存在 Key 返回 0 |
| `TestSortedSet_GetByRank` | 按排名取值、超出范围返回 nil |
| `TestSortedSet_GetRangeByRank` | 正常范围、start>end 自动交换 |
| `TestSortedSet_UpdateScore` | 多元素场景下更新分数后排名真正重排 |
| `TestSortedSet_GetRangeByScore` | 全包含区间 [min, max] |
| `TestSortedSet_GetRangeByScore_Exclusive` | 全包含 / 左排他 / 右排他 / 两端排他 四种组合 |
| `TestSortedSet_DeleteRangeByScore` | 全包含 / 左排他 / 右排他，验证删除后元素不可查 |
| `TestSortedSet_DeleteRangeByRank` | 正常删除、start>end 自动交换 |
| `TestSortedSet_SameScore_StableOrder` | 相同分数按插入顺序（seq 升序）稳定排列 |
| `TestSortedSet_WithStruct` | struct 类型作为 Key/Val |

## Benchmark

| 函数 | 说明 |
|------|------|
| `BenchmarkSortedSet_Insert` | 持续插入性能（O(log n)） |
| `BenchmarkSortedSet_Get` | 1000 元素下按 Key 查找（O(1)） |
| `BenchmarkSortedSet_Delete` | 持续删除性能（O(log n)） |

## 注意事项

- `assert` 断言默认关闭，测试中如需验证 panic 行为需先调用 `assert.SetAssert(true)`
- 并发安全性不在单元测试范围内，如需验证请使用 `-race` 标志
