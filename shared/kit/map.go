package kit

import (
	"golang.org/x/exp/constraints"
)

type Sort string

const (
	SortAsc  Sort = "asc"
	SortDesc Sort = "desc"
)

// InOrderRangeMap 有序遍历map
// 参数1: 需要遍历的map
// 参数2:  SortAsc 升序 SortDesc 降序
// 参数3: 回调函数
func InOrderRangeMap[K constraints.Ordered, V any](m map[K]V, fn func(v V, k K), sort ...Sort) {
	if len(m) <= 1 {
		return
	}
	// 获取所有key
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	QuickSort(keys, 0, len(keys)-1, sort...)

	// 按顺序遍历
	for _, k := range keys {
		fn(m[k], k)
	}
}
