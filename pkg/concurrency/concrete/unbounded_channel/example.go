package main

import (
	"fmt"
	"sync"
	"time"
)

// 简单的使用示例
func main() {
	// 创建无界通道
	uc := NewUnboundedChannel()
	defer uc.Close()

	// 启动消费者
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range uc.Channel {
			fmt.Printf("处理消息: ID=%d, 数据=%s\n", msg.ID, string(msg.Data))
			// 模拟处理时间
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// 发送消息
	fmt.Println("开始发送消息...")
	for i := int64(0); i < 10; i++ {
		msg := &Msg{
			ID:   i,
			Data: []byte(fmt.Sprintf("消息内容_%d", i)),
		}

		if err := uc.SendMsg(msg); err != nil {
			fmt.Printf("发送消息失败: %v\n", err)
			continue
		}
		fmt.Printf("发送消息: ID=%d\n", i)
	}

	// 等待一段时间让消息被处理
	time.Sleep(100 * time.Millisecond)

	fmt.Println("示例完成")
}
