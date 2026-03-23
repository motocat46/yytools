# probability_distribution 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `probability_distribution_test.go` | 统计分布正确性、边界情形、工厂函数、DynamicWeights 状态机 |
| `example_test.go` | 可运行的使用示例（`go test -run Example`） |

## 测试设计要点

### 统计验证策略
采样次数 `sampleCount = 10000`，容差率 `toleranceRate = ±10%`。
每个下标的命中次数与理论期望偏差超出 10% 即报错。
10000 次采样在 ±10% 容差下失败概率极低（约 10^-6 量级）。

### 覆盖的关键场景
| 场景 | 测试目标 |
|------|---------|
| 单权重 | 永远返回下标 0 |
| 等权重 | 均匀分布，偏差 ≤ 10% |
| 不等权重 `[1,2,7]` | 比例正确，偏差 ≤ 10% |
| 全零权重 | 退化为等概率随机，不 panic |
| NormalMethod 与 VoseAlias 趋同 | 两种实现对同一权重集的分布结果一致 |

### DynamicWeights 状态机
| 测试 | 验证目标 |
|------|---------|
| CanGenerate 初始/耗尽 | 初始总权重 > 0 时可生成；耗尽后不可生成 |
| Generate 总次数 = 总权重 | 每次 reduce 1，恰好生成 totalWeight 次 |
| 每个 key 生成次数 = 初始权重 | 验证 key 级别的消耗正确性 |
| SetReduce / 运行中修改 reduce | 修改步长后生成次数正确 |

## 运行方式

```bash
# 快速验证（含统计测试）
go test ./pkg/algorithms/mathx/probability_distribution/

# 含示例
go test -run Example ./pkg/algorithms/mathx/probability_distribution/
```

## 注意

统计测试存在极小概率随机失败（约 10^-6）。若 CI 偶发失败，重跑即可排除。
