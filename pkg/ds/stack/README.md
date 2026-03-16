# stack 使用文档

基于切片实现的泛型栈，LIFO（后进先出）语义，支持自动扩缩容。

## API

### 接口 `IStack[T any]`

| 方法 | 说明 |
|------|------|
| `Length() int` | 栈中的元素数量 |
| `Empty() bool` | 栈是否为空 |
| `Push(item T)` | 入栈 |
| `Pop() T` | 出栈，**空时触发断言 panic** |
| `Top() T` | 查看栈顶元素（不出栈），**空时触发断言 panic** |

### 构造函数

```go
stack.NewStack[T]()                // 默认容量 16
stack.NewStackWithSize[T](size int) // 指定初始容量
```

### 扩缩容策略

- **扩容**：底层 `append` 自动处理
- **缩容**：元素数 < 容量 1/4 时缩为一半，最小保持 16

## 使用示例

```go
import "github.com/motocat46/yytools/pkg/ds/stack"

s := stack.NewStack[string]()

// 入栈
s.Push("a")
s.Push("b")
s.Push("c")

// 查看栈顶
fmt.Println(s.Top()) // "c"

// 出栈（LIFO 顺序）
for !s.Empty() {
    fmt.Println(s.Pop()) // c, b, a
}

// 括号匹配示例
func isValid(str string) bool {
    s := stack.NewStack[rune]()
    for _, ch := range str {
        switch ch {
        case '(', '[', '{':
            s.Push(ch)
        case ')', ']', '}':
            if s.Empty() {
                return false
            }
            s.Pop()
        }
    }
    return s.Empty()
}
```

## 注意事项

- `Pop()` 和 `Top()` 在空栈时触发断言 panic，调用前需确保 `!Empty()`
- 非并发安全，多 goroutine 使用时需自行加锁
