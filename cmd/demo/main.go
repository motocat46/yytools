// Package main.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2022/6/15
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	benchheap "github.com/stormYuanYang/yytools/internal/bench/heap"
	benchmathx "github.com/stormYuanYang/yytools/internal/bench/mathx"
	benchprobdist "github.com/stormYuanYang/yytools/internal/bench/probability_distribution"
	benchqueue "github.com/stormYuanYang/yytools/internal/bench/queue"
	benchsort "github.com/stormYuanYang/yytools/internal/bench/sort"
	benchsortedset "github.com/stormYuanYang/yytools/internal/bench/sorted_set"
	benchstack "github.com/stormYuanYang/yytools/internal/bench/stack"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
)

// benchCmd 描述一个 bench 子命令
type benchCmd struct {
	use     string
	short   string
	handler func(int)
}

var benchCmds = []benchCmd{
	{"heap", "最小堆", benchheap.HeapTest},
	{"maxheap", "最大堆", benchheap.MaxHeapTest},
	{"mathcommon", "公共数学方法（比如gcd）", benchmathx.MathCommonTest},
	{"prob", "概率分布", benchprobdist.ProbabilityDistributionTest},
	{"pq", "优先级队列", benchheap.PriorityQueueTest},
	{"queue", "队列", benchqueue.QueueTest},
	{"sort", "排序", benchsort.SortTest},
	{"sortedset", "有序集合", benchsortedset.SortedSetTest},
	{"stack", "栈", benchstack.StackTest},
}

func parseNum(args []string) int {
	num, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "iterations 必须是整数: %v\n", err)
		os.Exit(1)
	}
	return num
}

func main() {
	assert.SetAssert(true)

	rootCmd := &cobra.Command{
		Use:   "yytools",
		Short: "yytools demo — 算法与数据结构演示",
	}

	// http 子命令：启动可视化服务
	rootCmd.AddCommand(&cobra.Command{
		Use:   "http",
		Short: "启动可视化 HTTP 服务（:8081）",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			http.HandleFunc("/", graphHttpServer)
			if err := http.ListenAndServe(":8081", nil); err != nil {
				panic(err)
			}
		},
	})

	// all 子命令：运行全部 bench
	rootCmd.AddCommand(&cobra.Command{
		Use:   "all <iterations>",
		Short: "运行所有 bench 命令",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			num := parseNum(args)
			for _, bc := range benchCmds {
				bc.handler(num)
			}
			println("\n所有测试完毕...")
		},
	})

	// 逐一注册各 bench 子命令
	for _, bc := range benchCmds {
		rootCmd.AddCommand(&cobra.Command{
			Use:   bc.use + " <iterations>",
			Short: bc.short,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				bc.handler(parseNum(args))
			},
		})
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
