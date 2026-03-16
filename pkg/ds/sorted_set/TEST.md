# sorted_set 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `sortedset_test.go` | 单元测试 + Benchmark |

## 运行方式

```bash
# 运行所有单元测试
go test ./pkg/ds/sorted_set/

# 详细输出（查看每个用例）
go test -v ./pkg/ds/sorted_set/

# 竞态检测
go test -race ./pkg/ds/sorted_set/

# 运行所有 Benchmark（每项默认 1 秒）
go test -bench=. -benchmem ./pkg/ds/sorted_set/

# 指定运行时长（每项跑 3 秒，结果更稳定）
go test -bench=. -benchtime=3s ./pkg/ds/sorted_set/

# 只跑特定操作的特定规模
go test -bench=BenchmarkSortedSet_Get/n=10000 ./pkg/ds/sorted_set/

# 重复多次取均值
go test -bench=. -count=3 ./pkg/ds/sorted_set/
```

## 正确性测试用例

### 构造

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestNewSortedSet` | 初始化，长度为 0 |
| `TestNewNodeData` | Key/Score/Val 字段赋值正确 |

### Insert

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_Insert` | 正常插入、长度递增、重复 Key 返回 false 且不覆盖原值 |
| `TestSortedSet_ReinsertAfterDelete` | 删除后同 Key 应能重新插入，新 score 生效 |

### Delete

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_Delete` | 存在 Key 删除成功、不存在返回 false、删后 Get 返回 nil |
| `TestSortedSet_DeleteAll` | 逐一删光所有元素后 Length=0，再插入正常工作 |

### Get / GetRank / GetByRank

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_Get` | 存在 / 不存在 Key |
| `TestSortedSet_GetRank` | 多元素排名正确性、不存在 Key 返回 0 |
| `TestSortedSet_GetByRank` | 按排名取值、超出范围返回 nil |

### GetRangeByRank / DeleteRangeByRank

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_GetRangeByRank` | 正常范围（升序）、start>end 自动交换 |
| `TestSortedSet_GetRangeByRank_BeyondLength` | end 超出总长度时截断返回，不 panic |
| `TestSortedSet_DeleteRangeByRank` | 正常删除、start>end 自动交换、删后不可查 |

### UpdateScore

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_UpdateScore` | 更新后排名真正重排（多元素验证）、不存在 Key 返回 false |
| `TestSortedSet_UpdateScore_SameValue` | 更新为相同分数，排名不变、长度不变 |
| `TestSortedSet_UpdateScore_EqualNeighbor` | 新分数等于前驱/后继，触发删除重插路径，seq 稳定排序仍正确 |

### GetRangeByScore / DeleteRangeByScore

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_GetRangeByScore` | 全包含 [min, max]，结果在范围内 |
| `TestSortedSet_GetRangeByScore_Exclusive` | 全包含 / 左排他 / 右排他 / 两端排他 四种组合 |
| `TestSortedSet_GetRangeByScore_EmptyResult` | 范围内无元素，返回空切片不 panic |
| `TestSortedSet_GetRangeByScore_SinglePoint` | [x,x] 返回 1 个；(x,x) / [x,x) 返回空 |
| `TestSortedSet_DeleteRangeByScore` | 三种排他组合，验证删后元素不可查 |

### 特殊场景

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_NegativeScore` | 负数分数排在最前，rank=1 为最小值 |
| `TestSortedSet_EmptySet_Operations` | 空集合上所有操作返回空/零值，无 panic |
| `TestSortedSet_SameScore_StableOrder` | 相同分数按插入顺序（seq 升序）稳定排列 |
| `TestSortedSet_KeyValDifferentTypes` | Key 和 Val 使用不同类型（string/struct） |

### 随机混合操作压力测试

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_RandomOps` | 见下文 |

`TestSortedSet_RandomOps` 通过随机混合操作序列验证整体正确性，弥补单方法测试无法覆盖的**跨操作状态一致性**：

- 维护与 SortedSet 语义完全一致的参考模型（`refModel`，基于 map）
- 随机执行 Insert / Delete / UpdateScore 混合序列（20 轮 × 500 次操作）
- 每轮结束后验证全量不变量：
  1. `Length` == ref 元素数
  2. ref 中每个 key 通过 `Get` 可查到，且 score 与 ref 一致
  3. `GetByRank` 遍历结果 score 单调不降
  4. `GetRank(GetByRank(r).Key) == r`（排名双向一致）
  5. `GetByRank` 覆盖的 key 集合 == ref key 集合（无多无少）
- 额外验证每轮随机 `GetRangeByScore` 的结果数量、范围正确性、有序性
- 固定随机种子（PCG 42）保证失败可复现

## Benchmark

各操作均按集合规模分组（n=100 / 1000 / 10000），便于观察 O(log n) 特性。

| 函数 | 说明 |
|------|------|
| `BenchmarkSortedSet_Insert` | 批量插入（每轮重建集合） |
| `BenchmarkSortedSet_Get` | 按 Key 查找（O(1) 哈希表）|
| `BenchmarkSortedSet_GetRank` | 按 Key 查询排名（O(log n)）|
| `BenchmarkSortedSet_UpdateScore` | 更新分数触发重排（O(log n)）|
| `BenchmarkSortedSet_GetRangeByRank` | 查询前半范围（O(log n + k)）|
| `BenchmarkSortedSet_GetRangeByScore` | 按分数范围查询（O(log n + k)）|
| `BenchmarkSortedSet_Delete` | 全量删除（每轮重建集合）|

## 注意事项

- `assert` 断言默认关闭，如需验证 panic 路径需先调用 `assert.SetAssert(true)`
- 并发安全性不在单元测试范围内，如需验证请使用 `-race` 标志
