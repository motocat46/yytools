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

package workerpool

import "context"

// Result 携带 Pipeline 处理结果或错误。
// 调用方自行决定如何处理 Err（忽略、记录或终止）。
type Result[R any] struct {
	Value R
	Err   error
}

// Pipeline 将输入 channel 中的元素并发交给 fn 处理，结果写入输出 channel。
// Pipeline 是 WorkerPool 的薄封装，不重新实现调度逻辑。
//
// 用法：
//
//	p := NewPipeline(4, 100, func(n int) (string, error) {
//	    return strconv.Itoa(n), nil
//	})
//	defer p.Close()
//	out := p.Process(in)
//	for r := range out {
//	    if r.Err != nil { ... }
//	    fmt.Println(r.Value)
//	}
type Pipeline[T, R any] struct {
	pool *WorkerPool
	fn   func(T) (R, error)
}

// NewPipeline 创建泛型 pipeline。
// workers、queueSize 透传给内部 WorkerPool；fn 为每个元素的处理函数。
func NewPipeline[T, R any](workers, queueSize int, fn func(T) (R, error)) *Pipeline[T, R] {
	return &Pipeline[T, R]{
		pool: NewWorkerPool(workers, queueSize),
		fn:   fn,
	}
}

// Process 消费输入 channel，并发执行 fn，结果写入返回的输出 channel。
// 输入 channel 关闭后，所有任务完成，输出 channel 自动关闭。
//
// 注意：调用方必须消费完输出 channel，否则会导致 worker goroutine 阻塞。
func (p *Pipeline[T, R]) Process(in <-chan T) <-chan Result[R] {
	out := make(chan Result[R], cap(in)+1)
	go func() {
		defer func() {
			p.pool.Wait()
			close(out)
		}()
		for item := range in {
			item := item
			_ = p.pool.Submit(context.Background(), func() {
				val, err := p.fn(item)
				out <- Result[R]{Value: val, Err: err}
			})
		}
	}()
	return out
}

// Close 关闭内部 WorkerPool，释放所有 worker goroutine。
// 调用前应确保输入 channel 已关闭且输出 channel 已消费完毕。
func (p *Pipeline[T, R]) Close() {
	p.pool.Close()
}
