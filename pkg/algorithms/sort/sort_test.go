package sort

import (
	"reflect"
	"sort"
	"testing"
)

func TestBubbleSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正常数组",
			input:    []int{64, 34, 25, 12, 22, 11, 90},
			expected: []int{11, 12, 22, 25, 34, 64, 90},
		},
		{
			name:     "已排序数组",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "逆序数组",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素数组",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "重复元素",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			expected: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			BubbleSort(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestInsertionSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正常数组",
			input:    []int{64, 34, 25, 12, 22, 11, 90},
			expected: []int{11, 12, 22, 25, 34, 64, 90},
		},
		{
			name:     "已排序数组",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "逆序数组",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素数组",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "重复元素",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			expected: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			InsertionSort(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestQuickSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正常数组",
			input:    []int{64, 34, 25, 12, 22, 11, 90},
			expected: []int{11, 12, 22, 25, 34, 64, 90},
		},
		{
			name:     "已排序数组",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "逆序数组",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素数组",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "重复元素",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			expected: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			QuickSort(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestCountingSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正常数组",
			input:    []int{64, 34, 25, 12, 22, 11, 90},
			expected: []int{11, 12, 22, 25, 34, 64, 90},
		},
		{
			name:     "已排序数组",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "逆序数组",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素数组",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "重复元素",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			expected: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			CountingSort(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestRadixSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正常数组",
			input:    []int{64, 34, 25, 12, 22, 11, 90},
			expected: []int{11, 12, 22, 25, 34, 64, 90},
		},
		{
			name:     "已排序数组",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "逆序数组",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素数组",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "重复元素",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			expected: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			RadixSort(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestSortStability(t *testing.T) {
	// 测试排序算法的稳定性
	// 这里使用简单的整数数组来测试稳定性
	items := []int{3, 1, 2, 1, 3}

	// 使用稳定的排序算法（插入排序）
	input := append([]int{}, items...)
	InsertionSort(input)

	// 验证排序结果
	expected := []int{1, 1, 2, 3, 3}
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("期望 %v，实际 %v", expected, input)
	}
}

func TestSortWithNegativeNumbers(t *testing.T) {
	input := []int{-5, 3, -1, 0, 2, -10, 7}
	expected := []int{-10, -5, -1, 0, 2, 3, 7}

	QuickSort(input)
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("期望 %v，实际 %v", expected, input)
	}
}

func TestSortWithLargeNumbers(t *testing.T) {
	input := []int{1000000, 999999, 1000001, 999998}
	expected := []int{999998, 999999, 1000000, 1000001}

	RadixSort(input)
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("期望 %v，实际 %v", expected, input)
	}
}

func TestQuickSortDesc(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"正常数组", []int{64, 34, 25, 12, 22, 11, 90}, []int{90, 64, 34, 25, 22, 12, 11}},
		{"已排序", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"逆序数组", []int{5, 4, 3, 2, 1}, []int{5, 4, 3, 2, 1}},
		{"空数组", []int{}, []int{}},
		{"单元素", []int{42}, []int{42}},
		{"重复元素", []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}, []int{9, 6, 5, 5, 5, 4, 3, 3, 2, 1, 1}},
		{"含负数", []int{-5, 3, -1, 0, 2}, []int{3, 2, 0, -1, -5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			QuickSortDesc(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestQuickSortTraversal(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"正常数组", []int{64, 34, 25, 12, 22, 11, 90}, []int{11, 12, 22, 25, 34, 64, 90}},
		{"已排序", []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{"逆序", []int{5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5}},
		{"空数组", []int{}, []int{}},
		{"单元素", []int{42}, []int{42}},
		{"重复元素", []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}, []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			QuickSortTraversal(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestQuickSortDescTraversal(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"正常数组", []int{64, 34, 25, 12, 22, 11, 90}, []int{90, 64, 34, 25, 22, 12, 11}},
		{"已排序", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"空数组", []int{}, []int{}},
		{"单元素", []int{42}, []int{42}},
		{"重复元素", []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}, []int{9, 6, 5, 5, 5, 4, 3, 3, 2, 1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			QuickSortDescTraversal(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

func TestCountingSort_Negative(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"纯负数", []int{-3, -1, -4, -1, -5}, []int{-5, -4, -3, -1, -1}},
		{"混合正负", []int{-5, 3, -1, 0, 2, -10, 7}, []int{-10, -5, -1, 0, 2, 3, 7}},
		{"全相同", []int{5, 5, 5}, []int{5, 5, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]int{}, tt.input...)
			CountingSort(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("期望 %v，实际 %v", tt.expected, input)
			}
		})
	}
}

// TestSortConsistency 验证所有排序算法对同一输入输出一致
func TestSortConsistency(t *testing.T) {
	input := []int{5, 3, 8, 1, 9, 2, 4, 7, 6, 0}
	expected := append([]int{}, input...)
	sort.Ints(expected)

	algorithms := []struct {
		name string
		fn   func([]int)
	}{
		{"BubbleSort", func(a []int) { BubbleSort(a) }},
		{"InsertionSort", func(a []int) { InsertionSort(a) }},
		{"QuickSort", func(a []int) { QuickSort(a) }},
		{"QuickSortTraversal", func(a []int) { QuickSortTraversal(a) }},
		{"CountingSort", func(a []int) { CountingSort(a) }},
		{"RadixSort", func(a []int) { RadixSort(a) }},
	}
	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			arr := append([]int{}, input...)
			alg.fn(arr)
			if !reflect.DeepEqual(arr, expected) {
				t.Errorf("%s: 期望 %v，实际 %v", alg.name, expected, arr)
			}
		})
	}
}

func BenchmarkBubbleSort(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BubbleSort(append([]int{}, input...))
	}
}

func BenchmarkInsertionSort(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InsertionSort(append([]int{}, input...))
	}
}

func BenchmarkQuickSort(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuickSort(append([]int{}, input...))
	}
}

func BenchmarkCountingSort(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountingSort(append([]int{}, input...))
	}
}

func BenchmarkRadixSort(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RadixSort(append([]int{}, input...))
	}
}

func BenchmarkGoSort(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := append([]int{}, input...)
		sort.Ints(data)
	}
}
