// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2025/5/19

package unbounded_channel

import (
	"fmt"
	"sync"
	"time"
)

// Example 展示无界通道的基本用法
func Example() {
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
