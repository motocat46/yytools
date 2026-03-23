# base 测试说明

## 无独立测试文件

`base` 包只包含泛型类型约束定义（`Signed`、`Unsigned`、`Integer`、`Float`、`Number`、`Ordered`），没有可执行逻辑，不需要独立测试。

约束的正确性由两处保证：

1. **编译期**：所有使用 `base` 约束的包（heap、queue、stack、sorted_set、slicex 等）在 `go build` 时隐式验证约束的覆盖范围
2. **运行期**：上述包的测试覆盖了各种类型参数实例化的场景

## 关联验证命令

```bash
# 验证所有使用 base 约束的包可正常编译
go build ./...

# 运行使用 base 约束的包的测试
go test ./pkg/ds/... ./pkg/slicex/... ./pkg/algorithms/...
```
