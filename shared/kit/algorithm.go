package kit

import "golang.org/x/exp/constraints"

// QuickSort 快速排序（泛型版本）
// 思路参考 https://www.youtube.com/watch?v=duln2xAZhBA
// sort: SortAsc 升序排序, SortDesc 降序排序
func QuickSort[T constraints.Ordered](arr []T, l, r int, sort ...Sort) {
	if l >= r {
		return
	}

	s := SortAsc
	if len(sort) > 0 {
		s = sort[0]
	}
	q := partition(arr, l, r, s)

	QuickSort(arr, l, q-1, s)
	QuickSort(arr, q+1, r, s)
}

func partition[T constraints.Ordered](arr []T, l, r int, sort Sort) int {
	var (
		i = l
		j = l
	)
	for ; j < r; j++ {
		if sort == SortAsc {
			if arr[j] <= arr[r] {
				arr[i], arr[j] = arr[j], arr[i]
				i++
			}
		} else {
			if arr[j] >= arr[r] {
				arr[i], arr[j] = arr[j], arr[i]
				i++
			}
		}
	}
	arr[i], arr[r] = arr[r], arr[i]
	return i
}
