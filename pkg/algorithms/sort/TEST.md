# sort 测试说明

## 测试文件

| 文件 | 覆盖范围 |
|------|---------|
| `sort_test.go` | SelectionSort/InsertionSort/QuickSort/CountingSort/RadixSort 正确性；稳定性验证；多算法一致性交叉对比；多规模 Benchmark |

## 分层执行命令

```bash
# 快速验证
go test ./pkg/algorithms/sort/

# 单算法正确性
go test -run TestQuickSort ./pkg/algorithms/sort/

# 多规模基准（含复杂度曲线）
go test -bench='Multi' -benchmem -benchtime=2s ./pkg/algorithms/sort/

# 单规模对比基准（固定 n=1000 逆序）
go test -bench='^Benchmark[^_]' -benchmem -benchtime=2s ./pkg/algorithms/sort/
```

## 性能基准（Apple M4，benchtime=2s，随机输入）

### 多规模复杂度曲线

| 算法 | n=100 | n=1K | n=10K | n=100K | n=1M |
|------|-------|------|-------|--------|------|
| QuickSort (µs) | 0.001 | 0.026 | 0.354 | 4.43 | 51.8 |
| CountingSort (µs) | 0.0003 | 0.003 | 0.029 | 0.549 | 6.33 |
| Go stdlib sort (µs) | 0.001 | 0.009 | 0.221 | 4.33 | 53.3 |

- QuickSort 与 Go stdlib 性能相近（均为 O(n log n) 随机化快排）
- CountingSort 在随机整数场景下比快排快 5~8 倍（O(n+k)，值域与 n 同量级）
- **CountingSort 限制**：max-min 差值超过 1e7 会 panic，差值过大时请改用 QuickSort

## 注意

- SelectionSort / InsertionSort 为 O(n²)，不适合 n > 1 万的场景
- RadixSort 仅支持非负整数，传入负数会 panic
