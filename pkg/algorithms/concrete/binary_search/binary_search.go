// Package algorithm.

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

// 二分搜索

// 作者:  yangyuan
// 创建日期:2023/7/7
package binary_search

import (
	"github.com/stormYuanYang/yytools/pkg/common/concrete/assert"
	"github.com/stormYuanYang/yytools/pkg/common/concrete/base"
)

// 正确执行的条件: 数组是有序的(允许重复的元素出现)
// nums必须是非降序(即排除重复元素的话，就是升序)

// 经典实现一: 在有序数组里查找指定的元素
// 可以用力扣题目验证方法的正确性: https://leetcode.cn/problems/binary-search/submissions/
func BinarySearch[T base.Integer](nums []T, target T) int {
	// 左右均为闭区间
	left := 0
	right := len(nums) - 1
	for left <= right {
		// 有另外一种计算方式: (left+right)/2
		// 但是这种方式不太好，当left和right都比较大时容易出现整型溢出
		// 如下实现可以避免整型溢出
		mid := left + (right-left)/2
		// 二分搜索代码实现框架并不难，难在边界上的处理
		// 边界处理不好，程序的正确性就是玄学
		// 把每种情况的条件和处理都写清楚
		if target == nums[mid] {
			// 找到
			return mid
		} else if target < nums[mid] {
			// 目标值小于nums[mid]——目标值位于nums[mid]左侧
			// 左边界不变，收缩右边界，在范围[left, mid-1]中继续查找
			right = mid - 1
		} else if target > nums[mid] {
			// 目标值大于nums[mid]——目标值位于nums[mid]右侧
			// 右边界不变，收缩左边界，在范围[mid+1,right]中继续查找
			left = mid + 1
		} else {
			// 不可能执行到这里! target(小于|等于|大于)nums[mid] 三种情况前文都已列举
			assert.Assert(false, "logic error, mid:", mid, " target:", target)
		}
	}
	// 结束循环时, right 小于 left
	// 不停缩小的查找范围里没有元素了,仍然没有找到目标值,则目标值不存在
	return -1
}

// 经典实现二(一): (当有重复元素时)查找其左边界
func LeftBound[T base.Integer](nums []T, target T) int {
	left := 0
	right := len(nums) - 1
	for left <= right {
		mid := left + (right-left)/2
		if target < nums[mid] {
			// 目标值在nums[mid]左侧
			// 查找范围变为[left, mid-1]
			right = mid - 1
		} else if target > nums[mid] {
			// 目标值在nums[mid]右边
			// 查找范围变为[mid+1, right]
			left = mid + 1
		} else if target == nums[mid] {
			// 找到目标值序列的某一个下标mid
			// 目标值序列的左边界(下标)应当小于等于下标mid
			// 而我们的目的是找到目标值序列的左边界(下标)
			// 那么应当收缩查找范围的右端
			// 在区间[left, mid-1]中继续查找
			right = mid - 1
		} else {
			// 不可能执行到这里!
			assert.Assert(false)
		}
	}
	// 当退出循环时，right等于left-1
	// 此时left指向的就是左边界(如果存在)
	// 判断下标是否合法
	// 这里其实没必要判断left小于0，因为left总是进行加法不可能比0小
	// 但是为了判断统一和方便记忆,这里一并判断
	if left < 0 || left >= len(nums) {
		return -1
	}

	// 判断是否找到目标值,能找到则left就是左边界
	if nums[left] == target {
		// 返回有效的left
		return left
	}
	// 找不到目标值，也就没有边界
	return -1
}

// 经典实现二(二): (当有重复元素时)查找其右边界
func RightBound[T base.Integer](nums []T, target T) int {
	return rightBound(nums, target, 0, len(nums)-1)
}

func rightBound[T base.Integer](nums []T, target T, left int, right int) int {
	for left <= right {
		mid := left + (right-left)/2
		if target < nums[mid] {
			// 目标值在nums[mid]左侧
			// 那么收缩查找范围的左端
			// 在区间[left, mid-1]中继续查找
			right = mid - 1
		} else if target > nums[mid] {
			// 目标值在nums[mid]右侧
			// 那么收缩查找范围的左端
			// 在区间[mid+1, right]中继续查找
			left = mid + 1
		} else if target == nums[mid] {
			// 找到目标值序列的某一个下标mid
			// 目标值序列的右边界(下标)应当大于等于下标mid
			// 而我们的目的是找到目标值序列的右边界(下标)
			// 那么应当收缩查找范围的左端
			// 在区间[mid+1, right]中继续查找
			left = mid + 1
		} else {
			// 不可能执行到这里!
			assert.Assert(false)
		}
	}
	// 退出循环时, left等于right+1
	// 判断right是否越界
	// 这里其实只需要判断right是否小于0，因为right总是做减法，是不可能大于len(nums)的
	if right < 0 || right >= len(nums) {
		return -1
	}
	// 如果nums[right]等于target, 此时nums[right]就是右边界
	if nums[right] == target {
		return right
	}
	// 否则，找不到目标值,也就没有边界
	return -1

}

// 查找元素的左右边界
// 可以用力扣题目验证方法的正确性: https://leetcode.cn/problems/find-first-and-last-position-of-element-in-sorted-array/submissions/
func SearchBound[T base.Integer](nums []T, target T) (int, int) {
	left := LeftBound(nums, target)
	if left == -1 {
		return -1, -1
	}
	// 这里可以优化
	// 左边界找到了，则查找右边界时，进一步收缩查其左端
	right := RightBound(nums, target)
	assert.Assert(right != -1)
	return left, right
}

// 优化
func SearchBoundOpt[T base.Integer](nums []T, target T) (int, int) {
	left := LeftBound(nums, target)
	if left == -1 {
		return -1, -1
	}
	// 左边界找到了，则查找右边界时，进一步收缩查其左端
	right := rightBound(nums, target, left, len(nums)-1)
	assert.Assert(right != -1)
	return left, right
}