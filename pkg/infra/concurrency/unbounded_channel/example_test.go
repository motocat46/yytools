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

package unbounded_channel_test

import (
	"fmt"
	"sync"

	uc "github.com/motocat46/yytools/pkg/infra/concurrency/unbounded_channel"
)

// ExampleUnboundedChannel_basic 展示基本的生产者-消费者用法。
func ExampleUnboundedChannel_basic() {
	ch := uc.NewUnboundedChannel[int](16, 10000)
	defer ch.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// 消费者
	go func() {
		defer wg.Done()
		for range 5 {
			val, ok := ch.Receive()
			if !ok {
				return
			}
			fmt.Println(val)
		}
	}()

	// 生产者
	for i := 1; i <= 5; i++ {
		ch.Send(i)
	}

	wg.Wait()
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

// ExampleUnboundedChannel_select 展示通过 Out() 在 select 中使用无界通道。
func ExampleUnboundedChannel_select() {
	ch := uc.NewUnboundedChannel[string](16, 10000)
	defer ch.Close()

	ch.Send("hello")
	ch.Send("world")

	for range 2 {
		msg, ok := <-ch.Out()
		if ok {
			fmt.Println(msg)
		}
	}
	// Output:
	// hello
	// world
}

// ExampleUnboundedChannel_backpressure 展示背压：buffer 超过 limit 时生产者阻塞，
// 消费者消费后生产者自动继续。
func ExampleUnboundedChannel_backpressure() {
	// chanSize=2, limit=3：buffer 超过 3 条时生产者阻塞
	ch := uc.NewUnboundedChannel[int](2, 3)
	defer ch.Close()

	done := make(chan struct{})

	// 慢速消费者
	go func() {
		defer close(done)
		for range 6 {
			val, ok := ch.Receive()
			if !ok {
				return
			}
			fmt.Println(val)
		}
	}()

	// 生产者：第 4 条起会因背压短暂阻塞，消费者消费后继续
	for i := range 6 {
		ch.Send(i + 1)
	}

	<-done
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
}
