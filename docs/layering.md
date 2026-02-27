docs/layering.md

目标

yytools 是个人长期复用的低层组件库，覆盖算法、数据结构、工程基础设施与可移植的玩法组件。其首要目标是：结构可预测、依赖可控、扩展可持续。

本文件定义 yytools 的分层与依赖规则，用于防止库在长期演进中"业务化腐烂"。

分层定义

yytools 的代码按抽象层级从低到高划分如下：

L0：common（全局基础）
•	内容：断言、泛型约束、极少量通用小工具、纯常量
•	要求：必须极度克制，禁止成为"万物依赖中心"

建议路径：
•	pkg/common/assert
•	pkg/common/base
•	pkg/numconst/（纯数值/时间常量，无任何内部依赖）

L1：algorithms（纯算法）
•	内容：纯函数/纯逻辑算法（数学、图、搜索、slice 选择/聚合等）
•	特征：无副作用；可测试；可证明；不依赖业务语义

建议路径：
•	pkg/algorithms/mathx
•	pkg/slicex/（slice 工具；逻辑归 L1，顶层短路径便于使用）
•	pkg/algorithms/binary_search/
•	pkg/algorithms/sort/
•	pkg/algorithms/graph（规划中）

L2：ds（数据结构）
•	内容：容器与结构（heap、queue、stack、sorted set 等）
•	特征：关注性能与内存语义；不引入业务概念

建议路径：
•	pkg/ds/heap
•	pkg/ds/queue（环形缓冲实现，支持自动扩缩容）
•	pkg/ds/stack
•	pkg/ds/sorted_set

L3：infra（工程基础设施）
•	内容：并发原语/worker pool/限流、时间工具、panic 安全执行、OS 工具等
•	特征：面向工程复用；允许依赖 ds 与 algorithms

建议路径：
•	pkg/infra/concurrency/unbounded_channel
•	pkg/infra/safeexec
•	pkg/infra/timeutil
•	pkg/infra/os
•	pkg/infra/timer（规划中，暂未实现）
•	pkg/infra/obs（规划中，暂未实现）

L4：gameplay（玩法组件）

当前 yytools 已实现 L0–L3；L4 为预留层，待需求明确后实现。

•	内容：可移植玩法组件（条件系统、奖励选择、活动日历等）
•	特征：偏业务但必须可配置、可观测、可替换；不得绑定某个项目的具体常量/结构

建议路径：
•	pkg/gameplay/condition（规划中，暂未实现）
•	pkg/gameplay/reward（规划中，暂未实现）
•	pkg/gameplay/schedule（规划中，暂未实现）

依赖规则

允许依赖关系（只允许从高层依赖低层）
•	L4 gameplay → L3 infra / L2 ds / L1 algorithms / L0 common
•	L3 infra → L2 ds / L1 algorithms / L0 common
•	L2 ds → L1 algorithms / L0 common
•	L1 algorithms → L0 common
•	L0 common → 标准库（仅）

禁止依赖关系
•	algorithms/ds/infra 禁止依赖 gameplay
•	common 禁止依赖除标准库之外的任何内部包
•	gameplay 禁止引入具体项目的配置结构、全局单例、项目常量（必须通过配置/接口注入）

algorithms 与 ds 的依赖原则（软约束）

algorithms 原则上不依赖 ds；如有例外，必须在文件头注释中说明原因，
且只允许使用 ds 中的通用无业务语义容器（stack/queue），不允许依赖 ds 的业务特性接口。

已知例外：
•	pkg/algorithms/mathx/probability_distribution → pkg/ds/stack（算法内部使用通用栈，合理工程决策）
•	pkg/algorithms/sort/quick_sort → pkg/ds/stack（迭代版快排使用显式栈替代系统调用栈）

包职责边界

algorithms：必须满足
•	不读写外部状态（无全局变量作为状态）
•	不依赖网络、IO、时间等外部环境
•	不包含日志/指标上报（可通过回调由上层包装）
•	注释中给出时间复杂度与空间复杂度（可选但推荐）

ds：必须满足
•	明确并发语义（是否线程安全）
•	明确容量/扩容策略（如适用）
•	提供基准测试或至少可基准的接口（推荐）

infra：必须满足
•	明确资源生命周期（Start/Stop/Close）
•	明确时钟/调度策略（真实时间、可注入时钟等）
•	可观测性接口可注入（logger/metrics/tracer）

gameplay：必须满足
•	配置驱动（cfg 或 options）
•	核心行为可测试（可注入随机源、时钟、依赖接口）
•	对外 API 稳定且语义清晰（尤其空输入、并列规则、稳定性）

internal 的使用规则
•	internal/ 下的包仅 yytools 自身使用，不对外承诺兼容性
•	测试辅助、基准数据生成、脚手架工具建议放在 internal/

迁移策略
•	新代码必须遵守本分层规则
•	旧路径用 Deprecated: 注释标记，新代码全部用新路径，视情况直接删除旧路径