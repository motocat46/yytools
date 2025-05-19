# yytools
提供一些工具方法和数据结构，方便日常使用。
目录结构大致为:
yytools/
├── cmd/                   # 可执行程序入口（如果有）
├── internal/              # 内部基础工具库（assert, common, timeutil等）
├── pkg/                   # 核心功能模块
│   ├── algorithms/        # 算法模块
│   │   ├── concrete/      # 普通类型版
│   │   └── generic/       # 泛型版
│   ├── datastructures/    # 数据结构模块
│   │   ├── concrete/
│   │   └── generic/
│   ├── concurrency/       # 并发控制模块
│   │   ├── concrete/
│   │   └── generic/
│   ├── gameutils/         # 游戏业务常见功能
│   │   ├── concrete/
│   │   └── generic/
├── examples/              # 示例代码
├── tests/                 # 集成测试、基准测试
└── README.md              # 项目总览文档


algorithm —— 提供算法相关的方法 

common —— 提供公共的一些方法和参数

datastructure ——提供常用的一些数据结构

template —— 提供一些代码的实现模版

main.go —— 入口，提供相关命令(测试等)


## 备注
template下的代码一般是不好直接调用(可能是一个片段、不好实现成通用方法等)，但是可以作为实现参考的代码。当然template下的代码也保证经过了测试，能正确运行的。