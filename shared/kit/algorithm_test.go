package kit

import (
	"reflect"
	"testing"
)

func Test_QuickSort(t *testing.T) {
	cases := []struct {
		name string
		arr  []int
		want []int
	}{
		{
			name: "empty slice",
			arr:  []int{},
			want: []int{},
		},
		{
			name: "already sorted",
			arr:  []int{1, 2, 3, 4, 5},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "reverse sorted",
			arr:  []int{5, 4, 3, 2, 1},
			want: []int{1, 2, 3, 4, 5},
		},	
		{
			name: "random order",
			arr:  []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			want: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
		{
			name: "duplicate elements",
			arr:  []int{5, 5, 5, 5, 5},
			want: []int{5, 5, 5, 5, 5},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			QuickSort(c.arr, 0, len(c.arr)-1)
			if !reflect.DeepEqual(c.arr, c.want) {
				t.Errorf("QuickSort(%v) = %v, want %v", c.arr, c.arr, c.want)
			}
		})
	}
}
