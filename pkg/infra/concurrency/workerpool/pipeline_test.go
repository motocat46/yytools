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

package workerpool_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/motocat46/yytools/pkg/infra/concurrency/workerpool"
)

func TestPipeline_正常处理(t *testing.T) {
	p := workerpool.NewPipeline(4, 100, func(n int) (string, error) {
		return strconv.Itoa(n), nil
	})
	defer p.Close()

	in := make(chan int, 5)
	for i := range 5 {
		in <- i
	}
	close(in)

	var results []string
	for r := range p.Process(in) {
		if r.Err != nil {
			t.Fatalf("意外错误: %v", r.Err)
		}
		results = append(results, r.Value)
	}
	if len(results) != 5 {
		t.Errorf("结果数 = %d, 期望 5", len(results))
	}
}

func TestPipeline_错误传递(t *testing.T) {
	sentinel := errors.New("处理失败")
	p := workerpool.NewPipeline(2, 10, func(n int) (int, error) {
		if n == 3 {
			return 0, sentinel
		}
		return n * 2, nil
	})
	defer p.Close()

	in := make(chan int, 5)
	for i := range 5 {
		in <- i
	}
	close(in)

	var errCount int
	for r := range p.Process(in) {
		if errors.Is(r.Err, sentinel) {
			errCount++
		}
	}
	if errCount != 1 {
		t.Errorf("错误数 = %d, 期望 1", errCount)
	}
}

func TestPipeline_输入关闭后输出自动关闭(t *testing.T) {
	p := workerpool.NewPipeline(2, 10, func(n int) (int, error) { return n, nil })
	defer p.Close()

	in := make(chan int)
	close(in)

	count := 0
	for range p.Process(in) {
		count++
	}
	if count != 0 {
		t.Errorf("期望 0 个结果，得到 %d", count)
	}
}

func TestPipeline_集成_大规模(t *testing.T) {
	const n = 100_000
	p := workerpool.NewPipeline(8, 1000, func(i int) (int, error) { return i * 2, nil })
	defer p.Close()

	in := make(chan int, 1000)
	go func() {
		for i := range n {
			in <- i
		}
		close(in)
	}()

	var count int
	for r := range p.Process(in) {
		if r.Err != nil {
			t.Fatalf("意外错误: %v", r.Err)
		}
		count++
	}
	if count != n {
		t.Errorf("结果数 = %d, 期望 %d", count, n)
	}
}
