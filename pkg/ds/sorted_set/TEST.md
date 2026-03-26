# sorted_set 测试文档

## 测试文件

| 文件 | 说明 |
|------|------|
| `sortedset_test.go` | 单元测试 + Benchmark；含 `checkSkiplistOrder`（白盒跳表链表全序校验）和 `checkInvariants`（随机测试不变量校验）|
| `common_test.go` | `randomLevel` 边界测试（Mock 注入 + 高概率采样 + Fuzz）|

## 运行方式

```bash
# 运行所有单元测试（包含 10 万次随机混合测试，约 30s）
go test ./pkg/ds/sorted_set/

# 跳过大规模随机测试（CI 快速通道，约 1s）
go test -short ./pkg/ds/sorted_set/

# 详细输出（查看每个用例）
go test -v ./pkg/ds/sorted_set/

# 竞态检测
go test -race ./pkg/ds/sorted_set/

# 运行所有 Benchmark（含 1M 规模，约 2 分钟）
go test -bench=. -benchmem ./pkg/ds/sorted_set/

# 指定运行时长（每项跑 3 秒，结果更稳定）
go test -bench=. -benchtime=3s -benchmem ./pkg/ds/sorted_set/

# 只跑特定操作的特定规模
go test -bench=BenchmarkSortedSet_Get/n=1000000 ./pkg/ds/sorted_set/

# 重复多次取均值（配合 benchstat 做统计对比）
go test -bench=. -count=5 -benchmem ./pkg/ds/sorted_set/

# Fuzz 测试 randomLevel（探索概率参数的未知边界，建议定期跑）
go test -fuzz=FuzzRandomLevel -fuzztime=30s ./pkg/ds/sorted_set/
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
| `TestSortedSet_GetByRank` | 按排名取值、超出范围（rank > Length）返回 nil |
| `TestSortedSet_GetByRank_NegativeAndZero` | rank ≤ 0 是合约违反，assert 触发 panic |
| `TestSortedSet_GetByRankDesc_NegativeAndZero` | GetByRankDesc rank ≤ 0 触发 panic |

### GetRankDesc / GetByRankDesc / GetRangeByRankDesc

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_GetRankDesc` | score 最大的元素 rank=1、不存在 Key 返回 0 |
| `TestSortedSet_GetByRankDesc` | 按降序排名取值、超出范围返回 nil |
| `TestSortedSet_GetRangeByRankDesc` | 正常范围（降序）、start>end 自动交换 |
| `TestSortedSet_GetRangeByRankDesc_BeyondLength` | end 超出总长度时截断返回，不 panic |
| `TestSortedSet_DescAscRankSymmetry` | 正序 rank + 逆序 rank == Length+1（对称性验证）|

### DeleteRangeByRankDesc

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_DeleteRangeByRankDesc` | 正常删除（降序）、返回结果降序、删后不可查、start>end 自动交换 |
| `TestSortedSet_DeleteRangeByRankDesc_BeyondLength` | end 超出总长度时截断，不 panic |

### GetRangeByRank / DeleteRangeByRank

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_GetRangeByRank` | 正常范围（升序）、start>end 自动交换 |
| `TestSortedSet_GetRangeByRank_BeyondLength` | end 超出总长度时截断返回，不 panic |
| `TestSortedSet_DeleteRangeByRank` | 正常删除、start>end 自动交换、删后不可查 |
| `TestSortedSet_DeleteRangeByRank_BeyondLength` | end 超出总长度时截断返回，不 panic |

### UpdateScore

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_UpdateScore` | 更新后排名真正重排（多元素验证）、不存在 Key 返回 false |
| `TestSortedSet_UpdateScore_SameValue` | 更新为相同分数，排名不变、长度不变 |
| `TestSortedSet_UpdateScore_EqualNeighbor` | 新分数等于前驱/后继，触发删除重插路径，seq 稳定排序仍正确 |

### GetRangeByScore / DeleteRangeByScore / CountByScore

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_GetRangeByScore` | 全包含 [min, max]，结果在范围内 |
| `TestSortedSet_GetRangeByScore_Exclusive` | 全包含 / 左排他 / 右排他 / 两端排他 四种组合 |
| `TestSortedSet_GetRangeByScore_EmptyResult` | 范围内无元素，返回空切片不 panic |
| `TestSortedSet_GetRangeByScore_SinglePoint` | [x,x] 返回 1 个；(x,x) / [x,x) 返回空 |
| `TestSortedSet_DeleteRangeByScore` | 四种排他组合（含两端排他），验证删后元素不可查 |
| `TestSortedSet_CountByScore` | 四种排他组合计数与 len(GetRangeByScore) 对比，空集合，范围外无元素 |

### GetMin / GetMax

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_GetMin` | 空集合返回 nil、多元素返回最小值、单元素 |
| `TestSortedSet_GetMax` | 空集合返回 nil、多元素返回最大值、单元素 |
| `TestSortedSet_GetMin_SameScore` | Score 相同时取 seq 最小（最先插入）的元素 |
| `TestSortedSet_GetMax_SameScore` | Score 相同时取 seq 最大（最后插入）的元素 |

### 特殊场景

| 测试函数 | 覆盖场景 |
|----------|---------|
| `TestSortedSet_NegativeScore` | 负数分数排在最前，rank=1 为最小值 |
| `TestSortedSet_EmptySet_Operations` | 空集合上所有操作返回空/零值，无 panic |
| `TestSortedSet_SameScore_StableOrder` | 相同分数按插入顺序（seq 升序）稳定排列 |
| `TestSortedSet_KeyValDifferentTypes` | Key 和 Val 使用不同类型（string/struct） |
| `TestSortedSet_SingleElement` | 只含 1 个元素时的增删查改全路径，含 Desc API 全覆盖 |
| `TestSortedSet_InfScore` | ±Inf 作为 score 正常插入、排序、范围查询 |
| `TestSortedSet_NaN_Rejected` | NaN Score 在 Insert / UpdateScore / GetRangeByScore / DeleteRangeByScore 均触发 panic，原有数据不受影响 |

### 随机混合操作压力测试

| 测试函数 | 数据量 | 覆盖场景 |
|----------|--------|---------|
| `TestSortedSet_RandomOps` | 10 万次 | 参考模型逐步对比 + 全量不变量校验 |
| `TestSortedSet_StressOps` | 100 万次 | 大规模不变量校验（maxKeys=10k，压测 O(log n) 路径）|

`TestSortedSet_RandomOps` 通过随机混合操作序列验证整体正确性，弥补单方法测试无法覆盖的**跨操作状态一致性**：

- 维护与 SortedSet 语义完全一致的参考模型（`refModel`，基于 map）
- 随机执行 Insert / Delete / UpdateScore 混合序列（**20 轮 × 5000 次操作 = 10 万次**）
- key 池大小 1000，score 范围 [-500, 500)
- 每轮结束后验证全量不变量：
  1. `Length` == ref 元素数
  2. ref 中每个 key 通过 `Get` 可查到，且 score 与 ref 一致
  3. `GetByRank` 遍历结果 score 单调不降
  4. `GetRank(GetByRank(r).Key) == r`（排名双向一致）
  5. `GetByRank` 覆盖的 key 集合 == ref key 集合（无多无少）
  6. `GetRankDesc(key) + GetRank(key) == Length+1`（升降序对称）
  7. `GetMin/GetMax` 与 `GetByRank(1)/GetByRank(n)` 一致
  8. 白盒：跳表底层 level-0 链表 `lessOrder` 全序（`checkSkiplistOrder`）
- 额外验证每轮随机 `GetRangeByScore` 的结果数量、范围正确性、有序性
- 固定随机种子（PCG 42）保证失败可复现
- 使用 `-short` 跳过（CI 快速通道）

`TestSortedSet_StressOps` 在 100 万次操作下验证不变量，重点：
- maxKeys=10,000，集合规模更大，暴露大集合下跳表边界 bug
- 使用 keyPool 切片实现 O(1) 随机键选取（避免 O(n) map 遍历成为瓶颈）
- 每轮结束后调用 `checkInvariants` 验证全量不变量（与 RandomOps 相同的校验逻辑）
- 固定随机种子（PCG 99）保证失败可复现
- 使用 `-short` 跳过（CI 快速通道）

## Benchmark

各基准均按集合规模分组（n=100 / 1000 / 10000 / 100000 / 1000000），便于观察复杂度曲线。

### 单操作基准（稳定集合规模）

所有单操作基准在**预填充的稳定规模集合**上运行，每次迭代只测一次目标操作，确保代价反映的是"规模为 n 时的成本"。

| 函数 | 复杂度 | 说明 |
|------|--------|------|
| `BenchmarkSortedSet_Get` | O(1) | 哈希表查找 |
| `BenchmarkSortedSet_GetMin` | O(1) | Head.Levels[0].Forward 直接访问，无分配 |
| `BenchmarkSortedSet_GetMax` | O(1) | Tail 指针直接访问，无分配 |
| `BenchmarkSortedSet_GetRank` | O(log n) | 排名查询 |
| `BenchmarkSortedSet_UpdateScore` | O(log n) | 分数更新（使用随机新值触发真实重排序路径） |
| `BenchmarkSortedSet_Insert` | O(log n) | 插入（每次同步删除 rank=1 元素以维持集合规模） |
| `BenchmarkSortedSet_GetRangeByRank` | O(log n + k) | 排名范围查询（k = n/2） |
| `BenchmarkSortedSet_GetRangeByScore` | O(log n + k) | 分数范围查询（k ≈ n/2） |
| `BenchmarkSortedSet_CountByScore` | O(log n + k) | 范围计数（k ≈ n/2），1 次结构体分配，避免结果切片开销 |

### 混合负载基准

| 函数 | 说明 |
|------|------|
| `BenchmarkSortedSet_Mixed` | 模拟排行榜场景的混合负载，反映整体吞吐量 |

`BenchmarkSortedSet_Mixed` 的操作比例：

| 操作 | 比例 | 说明 |
|------|------|------|
| UpdateScore | 50% | 玩家分数更新（最高频） |
| GetRank | 25% | 查询自己排名 |
| GetRangeByRank(1,10) | 15% | 查看排行榜前 10 |
| GetRangeByScore | 5% | 按积分段查询 |
| Insert+Delete | 5% | 玩家进出场（维持集合规模） |

### 性能基线（2025-03-16，Apple M 系列，-benchtime=2s）

```
BenchmarkSortedSet_Get/n=100          	586681833	   3.6 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_Get/n=1000         	643231018	   3.6 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_Get/n=10000        	523853334	   4.5 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_Get/n=100000       	423544951	   5.5 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_Get/n=1000000      	136782262	  18.1 ns/op	  0 B/op	 0 allocs/op

BenchmarkSortedSet_GetRank/n=100      	 81626406	  33.6 ns/op	  8 B/op	 1 allocs/op
BenchmarkSortedSet_GetRank/n=1000     	 47062476	  45.0 ns/op	  8 B/op	 1 allocs/op
BenchmarkSortedSet_GetRank/n=10000    	 29888462	  81.2 ns/op	  8 B/op	 1 allocs/op
BenchmarkSortedSet_GetRank/n=100000   	 25137810	  99.2 ns/op	  8 B/op	 1 allocs/op
BenchmarkSortedSet_GetRank/n=1000000  	 13103905	 182.5 ns/op	  8 B/op	 1 allocs/op

BenchmarkSortedSet_UpdateScore/n=100      	 64251007	  38.6 ns/op	 15 B/op	 1 allocs/op
BenchmarkSortedSet_UpdateScore/n=1000     	 26633152	  93.4 ns/op	 31 B/op	 3 allocs/op
BenchmarkSortedSet_UpdateScore/n=10000    	 19241169	 117.6 ns/op	 32 B/op	 4 allocs/op
BenchmarkSortedSet_UpdateScore/n=100000   	 13771917	 154.7 ns/op	 32 B/op	 4 allocs/op
BenchmarkSortedSet_UpdateScore/n=1000000  	  8337150	 276.9 ns/op	 43 B/op	 4 allocs/op

BenchmarkSortedSet_Insert/n=100      	 14970192	 155.5 ns/op	 143 B/op	  8 allocs/op
BenchmarkSortedSet_Insert/n=1000     	 10765479	 189.4 ns/op	 176 B/op	 12 allocs/op
BenchmarkSortedSet_Insert/n=10000    	  9622390	 241.8 ns/op	 176 B/op	 12 allocs/op
BenchmarkSortedSet_Insert/n=100000   	  8998192	 270.8 ns/op	 176 B/op	 12 allocs/op
BenchmarkSortedSet_Insert/n=1000000  	  6004932	 407.2 ns/op	 176 B/op	 12 allocs/op

BenchmarkSortedSet_GetRangeByRank/n=100      	 13111365	    183 ns/op	    992 B/op	  5 allocs/op
BenchmarkSortedSet_GetRangeByRank/n=1000     	  1331265	   1795 ns/op	   9320 B/op	  9 allocs/op
BenchmarkSortedSet_GetRangeByRank/n=10000    	    49881	  48972 ns/op	 154730 B/op	 15 allocs/op
BenchmarkSortedSet_GetRangeByRank/n=100000   	     3446	 587034 ns/op	2169965 B/op	 24 allocs/op
BenchmarkSortedSet_GetRangeByRank/n=1000000  	      376	6183945 ns/op	22789228 B/op	34 allocs/op

BenchmarkSortedSet_GetRangeByScore/n=100      	 11401147	    211 ns/op	   1016 B/op	  6 allocs/op
BenchmarkSortedSet_GetRangeByScore/n=1000     	  1237050	   1978 ns/op	   9336 B/op	  9 allocs/op
BenchmarkSortedSet_GetRangeByScore/n=10000    	    49326	  48768 ns/op	 154746 B/op	 15 allocs/op
BenchmarkSortedSet_GetRangeByScore/n=100000   	     4274	 550979 ns/op	2169979 B/op	 24 allocs/op
BenchmarkSortedSet_GetRangeByScore/n=1000000  	      368	6470009 ns/op	22789241 B/op	34 allocs/op

BenchmarkSortedSet_GetMin/n=100       	1000000000	   0.59 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMin/n=1000      	1000000000	   0.59 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMin/n=10000     	1000000000	   0.81 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMin/n=100000    	1000000000	   1.06 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMin/n=1000000   	1000000000	   0.80 ns/op	  0 B/op	 0 allocs/op

BenchmarkSortedSet_GetMax/n=100       	1000000000	   0.41 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMax/n=1000      	1000000000	   0.41 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMax/n=10000     	1000000000	   0.41 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMax/n=100000    	1000000000	   0.42 ns/op	  0 B/op	 0 allocs/op
BenchmarkSortedSet_GetMax/n=1000000   	1000000000	   0.41 ns/op	  0 B/op	 0 allocs/op

BenchmarkSortedSet_CountByScore/n=100      	  9807984	   120.4 ns/op	  24 B/op	 1 allocs/op
BenchmarkSortedSet_CountByScore/n=1000     	  1000000	  1139   ns/op	  24 B/op	 1 allocs/op
BenchmarkSortedSet_CountByScore/n=10000    	    87076	 14061   ns/op	  24 B/op	 1 allocs/op
BenchmarkSortedSet_CountByScore/n=100000   	     8236	137944   ns/op	  24 B/op	 1 allocs/op
BenchmarkSortedSet_CountByScore/n=1000000  	      586	 2238419 ns/op	  24 B/op	 1 allocs/op

BenchmarkSortedSet_Mixed/n=100      	 52882267	   48.6 ns/op	    43 B/op	 0 allocs/op
BenchmarkSortedSet_Mixed/n=1000     	 55394017	   46.8 ns/op	    45 B/op	 1 allocs/op
BenchmarkSortedSet_Mixed/n=10000    	 13908218	  157.0 ns/op	   154 B/op	 1 allocs/op
BenchmarkSortedSet_Mixed/n=100000   	   325668	  29929 ns/op	 34858 B/op	 6 allocs/op
BenchmarkSortedSet_Mixed/n=1000000  	    33601	  82511 ns/op	235549 B/op	 6 allocs/op
```

## 注意事项

- `assert` 断言**始终开启**，合约违反（如 `GetByRank(0)`）会 panic
- 并发安全性不在单元测试范围内，如需验证请使用 `-race` 标志
