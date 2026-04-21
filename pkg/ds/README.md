# pkg/ds

泛型数据结构集合。

| 子模块 | 功能 |
|--------|------|
| [heap](heap/README.md) | 最小堆、最大堆、优先级队列（支持 UpdatePriority） |
| [queue](queue/README.md) | 环形队列，自动扩缩容 |
| [ring_buffer](ring_buffer/README.md) | 固定容量环形缓冲区，写满时覆盖最旧元素 |
| [stack](stack/README.md) | LIFO 栈，自动缩容 |
| [sorted_set](sorted_set/README.md) | 有序集合（跳表实现），类 Redis ZADD/ZRANK/ZRANGE |
| [lru](lru/README.md) | LRU 缓存（双向链表实现），带 TTL，并发安全；学习目的，生产推荐 hashicorp/golang-lru v2 |
| [trie](trie/README.md) | 前缀树，支持 Unicode；Search/HasPrefix/WithPrefix/Delete，并发安全 |
| [slidingwindow](slidingwindow/README.md) | 固定容量滑动窗口，O(1) Sum/Avg/Max/Min（单调双端队列） |
| [unionfind](unionfind/README.md) | 泛型并查集（DSU），O(α) Union/Find/Connected/Size/Count，支持任意 comparable 类型 |
| [fenwicktree](fenwicktree/README.md) | 泛型树状数组（BIT），O(log n) 单点更新 + 前缀和 / 区间和，支持任意 Number 类型 |
| [sparsetable](sparsetable/README.md) | 泛型稀疏表（Sparse Table），O(n log n) 预处理后 O(1) 区间 min/max/GCD 查询（仅静态数据） |
