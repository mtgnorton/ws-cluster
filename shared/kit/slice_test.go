package kit

import (
	"slices"
	"testing"
)

func TestSliceRemoveIndex(t *testing.T) {
	t.Run("keepOrder", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		slice = SliceRemoveIndex(slice, 1, true)
		if len(slice) != 4 {
			t.Errorf("expect %v, actual %v", 4, len(slice))
		}
		if slice[0] != 1 || slice[1] != 3 || slice[2] != 4 || slice[3] != 5 {
			t.Errorf("expect %v, actual %v", []int{1, 3, 4, 5}, slice)
		}
	})

	t.Run("notKeepOrder", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		slice = SliceRemoveIndex(slice, 1, false)
		if len(slice) != 4 {
			t.Errorf("expect %v, actual %v", 4, len(slice))
		}
		if !slices.Contains(slice, 1) || !slices.Contains(slice, 3) || !slices.Contains(slice, 4) || !slices.Contains(slice, 5) {
			t.Errorf("expect %v, actual %v", []int{1, 3, 4, 5}, slice)
		}
	})
}

func TestSliceRemoveFirstElement(t *testing.T) {
	t.Run("keepOrder", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		slice = SliceRemoveFirstElement(slice, 1, true)
		if len(slice) != 4 {
			t.Errorf("expect %v, actual %v", 4, len(slice))
		}
		if slice[0] != 2 || slice[1] != 3 || slice[2] != 4 || slice[3] != 5 {
			t.Errorf("expect %v, actual %v", []int{2, 3, 4, 5}, slice)
		}
	})
	t.Run("notKeepOrder", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		slice = SliceRemoveFirstElement(slice, 1, false)
		if len(slice) != 4 {
			t.Errorf("expect %v, actual %v", 4, len(slice))
		}
		if !slices.Contains(slice, 2) || !slices.Contains(slice, 3) || !slices.Contains(slice, 4) || !slices.Contains(slice, 5) {
			t.Errorf("expect %v, actual %v", []int{2, 3, 4, 5}, slice)
		}
	})

}

func TestSliceRangeRemoveElements(t *testing.T) {
	t.Run("keepOrder", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		slice = SliceRangeRemoveElements(slice, func(item int) bool {
			return item%2 == 0
		}, true)
		if len(slice) != 3 {
			t.Errorf("expect %v, actual %v", 3, len(slice))
		}
		if slice[0] != 1 || slice[1] != 3 || slice[2] != 5 {
			t.Errorf("expect %v, actual %v", []int{1, 3, 5}, slice)
		}
	})
	t.Run("notKeepOrder", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		slice = SliceRangeRemoveElements(slice, func(item int) bool {
			return item%2 == 0
		}, false)
		if len(slice) != 3 {
			t.Errorf("expect %v, actual %v", 3, len(slice))
		}
		if !slices.Contains(slice, 1) || !slices.Contains(slice, 3) || !slices.Contains(slice, 5) {
			t.Errorf("expect %v, actual %v", []int{1, 3, 5}, slice)
		}
	})
}

func BenchmarkSliceRemoveElement(b *testing.B) {
	b.Run("keepOrder", func(b *testing.B) {
		slice := make([]int, 1000)
		for i := range slice {
			slice[i] = i
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(slice))
			copy(tmp, slice)
			_ = SliceRemoveIndex(tmp, 500, true)
		}
	})

	b.Run("notKeepOrder", func(b *testing.B) {
		slice := make([]int, 1000)
		for i := range slice {
			slice[i] = i
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(slice))
			copy(tmp, slice)
			_ = SliceRemoveIndex(tmp, 500, false)
		}
	})

}

func BenchmarkSliceRangeRemoveElements(b *testing.B) {
	b.Run("keepOrder", func(b *testing.B) {
		slice := make([]int, 1000)
		for i := range slice {
			slice[i] = i
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(slice))
			copy(tmp, slice)
			_ = SliceRangeRemoveElements(tmp, func(item int) bool {
				return item%2 == 0
			}, true)
		}
	})

	b.Run("notKeepOrder", func(b *testing.B) {
		slice := make([]int, 1000)
		for i := range slice {
			slice[i] = i
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(slice))
			copy(tmp, slice)
			_ = SliceRangeRemoveElements(tmp, func(item int) bool {
				return item%2 == 0
			}, false)
		}
	})
}

func TestSliceIntersection(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{
			name:     "基本交集测试",
			a:        []int{1, 2, 3, 4},
			b:        []int{3, 4, 5, 6},
			expected: []int{3, 4},
		},
		{
			name:     "空切片测试",
			a:        []int{},
			b:        []int{1, 2, 3},
			expected: []int{},
		},
		{
			name:     "无交集测试",
			a:        []int{1, 2, 3},
			b:        []int{4, 5, 6},
			expected: []int{},
		},
		{
			name:     "重复元素测试",
			a:        []int{1, 2, 2, 3, 3},
			b:        []int{2, 2, 3, 3, 4},
			expected: []int{2, 3},
		},
		{
			name:     "完全相同测试",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceIntersection(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Errorf("长度不匹配, 期望 %v, 得到 %v", tt.expected, result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("结果不匹配, 期望 %v, 得到 %v", tt.expected, result)
					return
				}
			}
		})
	}
}

func TestSliceUnion(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{
			name:     "基本并集测试",
			a:        []int{1, 2, 3},
			b:        []int{3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空切片测试",
			a:        []int{},
			b:        []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "完全不重叠测试",
			a:        []int{1, 2, 3},
			b:        []int{4, 5, 6},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "重复元素测试",
			a:        []int{1, 2, 2, 3, 3},
			b:        []int{2, 2, 3, 3, 4},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "完全相同测试",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "一个空切片测试",
			a:        []int{1, 2, 3},
			b:        []int{},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceUnion(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Errorf("长度不匹配, 期望 %v, 得到 %v", tt.expected, result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("结果不匹配, 期望 %v, 得到 %v", tt.expected, result)
					return
				}
			}
		})
	}
}

func TestSliceDifference(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{
			name:     "基本差集测试",
			a:        []int{1, 2, 3, 4},
			b:        []int{3, 4, 5, 6},
			expected: []int{1, 2},
		},
		{
			name:     "空切片测试",
			a:        []int{},
			b:        []int{1, 2, 3},
			expected: []int{},
		},
		{
			name:     "完全不重叠测试",
			a:        []int{1, 2, 3},
			b:        []int{4, 5, 6},
			expected: []int{1, 2, 3},
		},
		{
			name:     "重复元素测试",
			a:        []int{1, 2, 2, 3, 3},
			b:        []int{2, 3},
			expected: []int{1},
		},
		{
			name:     "完全相同测试",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: []int{},
		},
		{
			name:     "b为空切片测试",
			a:        []int{1, 2, 3},
			b:        []int{},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceDifference(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Errorf("长度不匹配, 期望 %v, 得到 %v", tt.expected, result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("结果不匹配, 期望 %v, 得到 %v", tt.expected, result)
					return
				}
			}
		})
	}
}
