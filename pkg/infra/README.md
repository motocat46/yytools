# pkg/infra

基础设施工具，面向横切关注点（安全执行、时间、OS、并发原语）。

| 子模块 | 功能 |
|--------|------|
| [safeexec](safeexec/README.md) | panic 安全执行包装（Safe、SafeCall、SafeExecWithError 等） |
| [timeutil](timeutil/README.md) | 时间工具函数（`ParseDuration` 支持天单位；`Parse`/`ParseUnixMilli` 解析日期时间；日历边界计算；`IsSameDay/Week`、`DaysBetween` 等比较函数） |
| [os](os/README.md) | OS 工具封装（文件存在检测、备份等） |
| [concurrency/unbounded_channel](concurrency/unbounded_channel/README.md) | 无大小限制的 Channel，多种实现变体 |
| [concurrency/workerpool](concurrency/workerpool/README.md) | 固定大小 goroutine 池 + 泛型 Pipeline |
| [timecond](timecond/README.md) | 基于配置的时间条件判断（绝对/相对时间戳，OpLT/GE/Within/RelLT/RelGE） |
| [timer/timingwheel](timer/timingwheel/README.md) | 分层时间轮定时器（O(1) add/cancel，毫秒精度，约 50 天覆盖）+ [DelayQueue](timer/delayqueue/README.md) 子模块 |
