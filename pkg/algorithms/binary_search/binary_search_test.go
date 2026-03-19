// Package binary_search.

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

package binary_search

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"testing"
)

// TestBinarySearch 测试基础二分查找
func TestBinarySearch(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int // -1 表示不存在，否则验证 nums[want]==target
	}{
		{name: "正常找到单次出现", nums: []int{1, 3, 5, 7, 9}, target: 5, want: 2},
		{name: "找到头部元素", nums: []int{1, 3, 5, 7, 9}, target: 1, want: 0},
		{name: "找到尾部元素", nums: []int{1, 3, 5, 7, 9}, target: 9, want: 4},
		{name: "目标小于最小值", nums: []int{1, 3, 5, 7, 9}, target: 0, want: -1},
		{name: "目标大于最大值", nums: []int{1, 3, 5, 7, 9}, target: 10, want: -1},
		{name: "目标在中间空缺", nums: []int{1, 3, 5, 7, 9}, target: 4, want: -1},
		{name: "空切片", nums: []int{}, target: 5, want: -1},
		{name: "单元素等于目标", nums: []int{7}, target: 7, want: 0},
		{name: "单元素不等于目标", nums: []int{7}, target: 3, want: -1},
		{name: "全相同元素命中", nums: []int{5, 5, 5, 5, 5}, target: 5, want: -2}, // -2 表示任意有效下标
		{name: "有重复元素命中其中一个", nums: []int{1, 3, 3, 3, 7}, target: 3, want: -2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BinarySearch(tc.nums, tc.target)
			if tc.want == -1 {
				if got != -1 {
					t.Errorf("期望 -1，实际 %d", got)
				}
			} else if tc.want == -2 {
				// 只验证找到了，且命中的下标确实等于 target
				if got < 0 || got >= len(tc.nums) || tc.nums[got] != tc.target {
					t.Errorf("期望命中目标 %d，实际下标 %d", tc.target, got)
				}
			} else {
				if got != tc.want {
					t.Errorf("期望下标 %d，实际 %d", tc.want, got)
				}
			}
		})
	}
}

// TestLeftBound 测试查找左边界
func TestLeftBound(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{name: "无重复找到", nums: []int{1, 3, 5, 7, 9}, target: 5, want: 2},
		{name: "有重复返回最左下标", nums: []int{1, 3, 3, 3, 7}, target: 3, want: 1},
		{name: "全相同返回下标0", nums: []int{5, 5, 5, 5, 5}, target: 5, want: 0},
		{name: "目标不存在", nums: []int{1, 3, 5, 7, 9}, target: 4, want: -1},
		{name: "空切片", nums: []int{}, target: 5, want: -1},
		{name: "目标比所有元素小", nums: []int{1, 3, 5, 7, 9}, target: 0, want: -1},
		{name: "目标比所有元素大", nums: []int{1, 3, 5, 7, 9}, target: 10, want: -1},
		{name: "单元素等于目标", nums: []int{7}, target: 7, want: 0},
		{name: "单元素不等于目标", nums: []int{7}, target: 3, want: -1},
		{name: "左边界在最左", nums: []int{2, 2, 3, 4}, target: 2, want: 0},
		{name: "左边界在最右", nums: []int{1, 2, 3, 5}, target: 5, want: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := LeftBound(tc.nums, tc.target)
			if got != tc.want {
				t.Errorf("期望 %d，实际 %d", tc.want, got)
			}
		})
	}
}

// TestRightBound 测试查找右边界
func TestRightBound(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{name: "无重复找到", nums: []int{1, 3, 5, 7, 9}, target: 5, want: 2},
		{name: "有重复返回最右下标", nums: []int{1, 3, 3, 3, 7}, target: 3, want: 3},
		{name: "全相同返回最后下标", nums: []int{5, 5, 5, 5, 5}, target: 5, want: 4},
		{name: "目标不存在", nums: []int{1, 3, 5, 7, 9}, target: 4, want: -1},
		{name: "空切片", nums: []int{}, target: 5, want: -1},
		{name: "目标比所有元素小", nums: []int{1, 3, 5, 7, 9}, target: 0, want: -1},
		{name: "目标比所有元素大", nums: []int{1, 3, 5, 7, 9}, target: 10, want: -1},
		{name: "单元素等于目标", nums: []int{7}, target: 7, want: 0},
		{name: "单元素不等于目标", nums: []int{7}, target: 3, want: -1},
		{name: "右边界在最左", nums: []int{2, 3, 4, 5}, target: 2, want: 0},
		{name: "右边界在最右", nums: []int{1, 3, 5, 5}, target: 5, want: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RightBound(tc.nums, tc.target)
			if got != tc.want {
				t.Errorf("期望 %d，实际 %d", tc.want, got)
			}
		})
	}
}

type boundCase struct {
	name      string
	nums      []int
	target    int
	wantLeft  int
	wantRight int
}

var searchBoundCases = []boundCase{
	{name: "找到无重复left==right", nums: []int{1, 3, 5, 7, 9}, target: 5, wantLeft: 2, wantRight: 2},
	{name: "找到有重复left<right", nums: []int{1, 3, 3, 3, 7}, target: 3, wantLeft: 1, wantRight: 3},
	{name: "全相同", nums: []int{5, 5, 5, 5, 5}, target: 5, wantLeft: 0, wantRight: 4},
	{name: "不存在", nums: []int{1, 3, 5, 7, 9}, target: 4, wantLeft: -1, wantRight: -1},
	{name: "空切片", nums: []int{}, target: 5, wantLeft: -1, wantRight: -1},
	{name: "目标比最小值小", nums: []int{1, 3, 5}, target: 0, wantLeft: -1, wantRight: -1},
	{name: "目标比最大值大", nums: []int{1, 3, 5}, target: 6, wantLeft: -1, wantRight: -1},
	{name: "单元素命中", nums: []int{7}, target: 7, wantLeft: 0, wantRight: 0},
	{name: "单元素未命中", nums: []int{7}, target: 3, wantLeft: -1, wantRight: -1},
}

// TestSearchBound 测试同时查找左右边界
func TestSearchBound(t *testing.T) {
	for _, tc := range searchBoundCases {
		t.Run(tc.name, func(t *testing.T) {
			gotLeft, gotRight := SearchBound(tc.nums, tc.target)
			if gotLeft != tc.wantLeft || gotRight != tc.wantRight {
				t.Errorf("期望 (%d, %d)，实际 (%d, %d)", tc.wantLeft, tc.wantRight, gotLeft, gotRight)
			}
		})
	}
}

// TestSearchBoundOpt 验证优化版与标准版结果完全一致
func TestSearchBoundOpt(t *testing.T) {
	for _, tc := range searchBoundCases {
		t.Run(tc.name, func(t *testing.T) {
			gotLeft, gotRight := SearchBoundOpt(tc.nums, tc.target)
			if gotLeft != tc.wantLeft || gotRight != tc.wantRight {
				t.Errorf("期望 (%d, %d)，实际 (%d, %d)", tc.wantLeft, tc.wantRight, gotLeft, gotRight)
			}
		})
	}
}

// TestBinarySearch_RandomLarge 随机大规模一致性测试（10 万次查询，与 sort.SearchInts 对比）
func TestBinarySearch_RandomLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大规模测试")
	}
	const n = 100_000
	rng := rand.New(rand.NewPCG(42, 0))

	// 构造有序数组（允许重复）
	nums := make([]int, n)
	for i := range nums {
		nums[i] = rng.IntN(n * 2)
	}
	sort.Ints(nums)

	// 随机查询 10 万次，与标准库对比
	for range n {
		target := rng.IntN(n * 2)

		// BinarySearch：找到时验证值正确，找不到时 stdlib 也应找不到
		got := BinarySearch(nums, target)
		stdIdx := sort.SearchInts(nums, target)
		if got == -1 {
			if stdIdx < len(nums) && nums[stdIdx] == target {
				t.Fatalf("BinarySearch 未找到 target=%d，但 stdlib 在 %d 找到", target, stdIdx)
			}
		} else {
			if nums[got] != target {
				t.Fatalf("BinarySearch 返回下标 %d，但 nums[%d]=%d != target=%d", got, got, nums[got], target)
			}
		}

		// LeftBound / RightBound：左右边界对称验证
		left, right := SearchBound(nums, target)
		leftOpt, rightOpt := SearchBoundOpt(nums, target)
		if left != leftOpt || right != rightOpt {
			t.Fatalf("SearchBound vs SearchBoundOpt 不一致：target=%d left=%d/%d right=%d/%d",
				target, left, leftOpt, right, rightOpt)
		}
		if left != -1 {
			if nums[left] != target || nums[right] != target {
				t.Fatalf("边界值错误：target=%d left=%d right=%d", target, left, right)
			}
			if left > 0 && nums[left-1] == target {
				t.Fatalf("左边界不是最左：target=%d left=%d", target, left)
			}
			if right < len(nums)-1 && nums[right+1] == target {
				t.Fatalf("右边界不是最右：target=%d right=%d", target, right)
			}
		}
	}
}

var benchSizes = []int{100, 1000, 10_000, 100_000, 1_000_000}

func BenchmarkBinarySearch(b *testing.B) {
	for _, n := range benchSizes {
		nums := make([]int, n)
		for i := range nums {
			nums[i] = i * 2 // 偶数有序数组
		}
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				BinarySearch(nums, i%n*2+1) // 查找奇数（不存在），触发完整遍历
			}
		})
	}
}

func BenchmarkLeftBound(b *testing.B) {
	for _, n := range benchSizes {
		nums := make([]int, n)
		for i := range nums {
			nums[i] = i / 2 // 含大量重复
		}
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				LeftBound(nums, i%n)
			}
		})
	}
}

func BenchmarkSearchBoundOpt(b *testing.B) {
	for _, n := range benchSizes {
		nums := make([]int, n)
		for i := range nums {
			nums[i] = i / 2
		}
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				SearchBoundOpt(nums, i%n)
			}
		})
	}
}
