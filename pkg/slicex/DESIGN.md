# slicex 设计记录

面向维护者，记录关键决策及其背后的理由。使用文档见 README.md。

---

## 一、双 API 设计：panic 版 + OK 版

每种操作提供两个版本：

| 函数 | 空切片行为 | 适用场景 |
|------|-----------|---------|
| `MinInSlice` / `MaxInSlice` / `MinBy` / `MaxBy` | 触发 assert（panic）| 调用方能保证切片非空 |
| `MinInSliceOK` / `MaxInSliceOK` / `MinByOK` / `MaxByOK` | 返回 `ok=false` | 切片可能为空，调用方自行处理 |

**为什么不只提供 OK 版本？**

泛型 T 的零值（`0`、`""`、`false`）可能是切片中的合法元素。空切片时返回零值是静默错误——调用方收到一个看似正常的值却无法得知切片是空的。`assert.Assert(ok)` 版本将这类调用方 bug 在开发阶段立刻暴露。

---

## 二、MinBy / MaxBy 使用 `any` 而非 `Ordered`

`MinInSlice` / `MaxInSlice` 约束为 `base.Ordered`（可直接比较）；`MinBy` / `MaxBy` 约束为 `any`（接受任意类型）。

原因：`MinBy` 通过调用方传入的 `better(a, b T) bool` 函数定义"优先级"，这使得：
- 元素本身不需要实现 `<`（如结构体）
- 可以按字段子集比较（如只比较 `Score` 字段）
- `MinBy` 与 `MaxBy` 共用同一套底层逻辑（`opInSliceByOK`），函数签名完全对称

---

## 三、稳定性保证

当存在多个并列最优值时，所有函数均返回**第一个**出现的下标（迭代使用严格 `<` / `>` 不含等号，不替换相同优先级的当前最优值）。
