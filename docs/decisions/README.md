# 架构决策记录（ADR）

记录影响整个项目的跨包架构决策。各包内部的设计决策见对应包的 `DESIGN.md`。

| 编号 | 标题 | 状态 |
|------|------|------|
| [ADR-001](adr-001-api-boundary-empty-collections.md) | 集合类 API 的空/越界边界策略（nil vs panic vs 零值） | 已采纳 |
| [ADR-002](adr-002-assert-always-on.md) | 断言框架始终启用，不提供运行时开关 | 已采纳 |
| [ADR-003](adr-003-flat-package-no-concrete-generic-split.md) | 扁平包结构，不拆分 concrete/generic 子包 | 已采纳 |
