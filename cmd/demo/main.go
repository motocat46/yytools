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
	"strings"

	benchheap "github.com/stormYuanYang/yytools/internal/bench/heap"
	benchmathx "github.com/stormYuanYang/yytools/internal/bench/mathx"
	benchprobdist "github.com/stormYuanYang/yytools/internal/bench/probability_distribution"
	benchqueue "github.com/stormYuanYang/yytools/internal/bench/queue"
	benchsort "github.com/stormYuanYang/yytools/internal/bench/sort"
	benchsortedset "github.com/stormYuanYang/yytools/internal/bench/sorted_set"
	benchstack "github.com/stormYuanYang/yytools/internal/bench/stack"
	"github.com/stormYuanYang/yytools/pkg/common/assert"
)

var commandsMap = map[string]int{}

type Command struct {
	Key     string
	Note    string
	Handler func(int)
}

var commands []*Command

func init() {
	commands = append(commands, &Command{
		Key:     "http",
		Note:    "可视化",
		Handler: nil,
	})
	commands = append(commands, &Command{
		Key:     "heap",
		Note:    "最小堆",
		Handler: benchheap.HeapTest,
	})
	commands = append(commands, &Command{
		Key:     "mathcommon",
		Note:    "公共数学方法（比如gcd）",
		Handler: benchmathx.MathCommonTest,
	})
	commands = append(commands, &Command{
		Key:     "maxheap",
		Note:    "最大堆",
		Handler: benchheap.MaxHeapTest,
	})
	commands = append(commands, &Command{
		Key:     "prob",
		Note:    "概率分布",
		Handler: benchprobdist.ProbabilityDistributionTest,
	})
	commands = append(commands, &Command{
		Key:     "pq",
		Note:    "优先级队列",
		Handler: benchheap.PriorityQueueTest,
	})
	commands = append(commands, &Command{
		Key:     "queue",
		Note:    "队列",
		Handler: benchqueue.QueueTest,
	})
	commands = append(commands, &Command{
		Key:     "sort",
		Note:    "排序",
		Handler: benchsort.SortTest,
	})
	commands = append(commands, &Command{
		Key:     "sortedset",
		Note:    "有序集合",
		Handler: benchsortedset.SortedSetTest,
	})
	commands = append(commands, &Command{
		Key:     "stack",
		Note:    "栈",
		Handler: benchstack.StackTest,
	})

	for i, c := range commands {
		commandsMap[c.Key] = i
	}
}

func testAll(num int) {
	for i := 0; i < len(commands); i++ {
		handler := commands[i].Handler
		handler(num)
	}
	println("\n所有测试完毕...")
}

func main() {
	assert.SetAssert(true)

	// 第一个参数是可执行文件本身的路径
	// 后续的参数是通过控制台传递的参数
	args := os.Args[1:]
	if len(args) == 0 {
		println("需要传入参数,如需要帮助, 请使用: yytools help")
		return
	}
	command := strings.ToLower(args[0])
	if command == "help" {
		println("使用参考：yytools sorted_set 5\n表示执行5轮sorted_set相关测试代码\n yytools all 5\n表示对所有测试进行5轮测试\n")
		println("已支持的命令:")
		fmt.Printf("%-20s\t说明:执行所有命令\n", "all")
		for i := 0; i < len(commands); i++ {
			fmt.Printf("%-20s\t说明:%s\n", commands[i].Key, commands[i].Note)
		}
		return
	}

	if command == "http" {
		http.HandleFunc("/", graphHttpServer)
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			panic(err)
		}
		return
	}

	num, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("Command: %s Error: %+v\n", command, err)
		return
	}

	if command == "all" {
		testAll(num)
	} else {
		index, ok := commandsMap[command]
		if !ok {
			println("不支持的命令")
			return
		}
		handler := commands[index].Handler
		handler(num)
	}
}
