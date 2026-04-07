# random_split 测试说明

## 测试文件

| 文件 | 内容 |
|------|------|
| `random_split_test.go` | State.Validate、New/Done/Remaining、Next、Allocate、策略单测、随机混合（300k 次）、压力测试、基准测试 |
| `correctness_test.go` | 5 条正确性命题（10 万轮，固定种子 42，必须 -race 运行） |

## 测试命令

```bash
# 快速验证（跳过压力测试）
go test -short -count=1 ./pkg/mechanics/distribution/random_split/

# 全量测试（含竞态检测，约 10s）
go test -race -count=1 ./pkg/mechanics/distribution/random_split/

# 仅命题测试（-race 必须）
go test -race -count=1 -run TestCorrectness -v ./pkg/mechanics/distribution/random_split/

# 压力测试（约 30-60s）
go test -count=1 -run TestStress_LargeScale -timeout 120s ./pkg/mechanics/distribution/random_split/

# 基准测试
go test -bench=BenchmarkAllocate -benchmem -benchtime=3s ./pkg/mechanics/distribution/random_split/
go test -bench=BenchmarkSimulate -benchmem -benchtime=10s ./pkg/mechanics/distribution/random_split/
```

## 性能基线

环境：Apple M4，Go 1.24，`-benchtime=3s`

| 基准 | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| BenchmarkAllocate/strategy=Fixed/n=10 | 47.55 | 128 | 2 |
| BenchmarkAllocate/strategy=Fixed/n=100 | 344 | 944 | 2 |
| BenchmarkAllocate/strategy=Fixed/n=1000 | 3,329 | 8,240 | 2 |
| BenchmarkAllocate/strategy=Uniform/n=10 | 77.61 | 128 | 2 |
| BenchmarkAllocate/strategy=Uniform/n=100 | 519 | 944 | 2 |
| BenchmarkAllocate/strategy=Uniform/n=1000 | 4,690 | 8,240 | 2 |
| BenchmarkAllocate/strategy=DoubleMean/n=10 | 113.7 | 128 | 2 |
| BenchmarkAllocate/strategy=DoubleMean/n=100 | 1,105 | 944 | 2 |
| BenchmarkAllocate/strategy=DoubleMean/n=1000 | 11,016 | 8,240 | 2 |
| BenchmarkSimulate/rounds=100000 | 14,048,709 | 13,603,667 | 200,006 |

**说明：**
- Allocate 每次操作始终 2 allocs：Distributor 本身 + 结果 slice，与策略和规模无关
- Fixed < Uniform < DoubleMean：DoubleMean 多一次 `math.Floor` + `float64` 计算
- Simulate 100k 轮 约 14ms，内存 13.6MB（主要是每轮的 Distributor + slice 分配）
