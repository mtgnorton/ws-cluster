package kit

// SliceRemoveIndex 删除切片中的元素,按照索引删除
// 如果keepOrder为true,则按照顺序删除,时间复杂度为O(n)
// 如果keepOrder为false,则将删除的元素与最后一个元素交换,然后删除最后一个元素,时间复杂度为O(1)
func SliceRemoveIndex[T any](slice []T, index int, keepOrder ...bool) []T {
	if len(slice) == 0 {
		return slice
	}
	if len(keepOrder) > 0 && keepOrder[0] {

		return append(slice[:index], slice[index+1:]...)
	}
	slice[index] = slice[len(slice)-1]
	var zero T
	slice[len(slice)-1] = zero // 显式置零
	return slice[:len(slice)-1]
}

// SliceRemoveFirstElement 删除切片中的元素,按照元素删除
// 如果keepOrder为true,则按照顺序删除,时间复杂度为O(n)
// 如果keepOrder为false,则将删除的元素与最后一个元素交换,然后删除最后一个元素,时间复杂度为O(1)
func SliceRemoveFirstElement[T comparable](slice []T, element T, keepOrder ...bool) []T {
	if len(slice) == 0 {
		return slice
	}
	if len(keepOrder) > 0 && keepOrder[0] {
		for i, v := range slice {
			if v == element {
				return append(slice[:i], slice[i+1:]...)
			}
		}
	}
	for i, v := range slice {
		if v == element {
			slice[i] = slice[len(slice)-1]
			var zero T
			slice[len(slice)-1] = zero // 显式置零
			return slice[:len(slice)-1]
		}
	}
	return slice
}

// SliceRangeRemoveElements 删除切片中满足条件的元素,按照元素删除
//
// 特性	乱序模式	有序模式
// 时间复杂度	O(n)	O(n)
// 元素顺序	可能改变	保持原序
// 内存操作	无新分配，直接交换	原地覆盖，尾部置零
// 适用场景	批量删除且不关心顺序	需严格保持元素顺序
// 内存泄漏防护	非指针类型自动清理	显式置空残留引用
func SliceRangeRemoveElements[T any](slice []T, condition func(T) bool, keepOrder ...bool) []T {
	if len(keepOrder) > 0 && keepOrder[0] {
		// 保持顺序模式 (写指针法)
		writeIdx := 0
		for _, v := range slice {
			if !condition(v) { // 不满足条件,则保留
				slice[writeIdx] = v
				writeIdx++
			}
		}
		// 清理被删除元素的引用 (防止内存泄漏)
		tail := slice[writeIdx:]
		for i := range tail {
			var zero T
			tail[i] = zero // 显式置零
		}
		return slice[:writeIdx]
	}

	// 乱序高效模式 (倒序交换法)
	for i := len(slice) - 1; i >= 0; i-- {
		if condition(slice[i]) {
			slice[i] = slice[len(slice)-1] // 与最后一个元素交换
			slice = slice[:len(slice)-1]   // 截断
		}
	}
	return slice

}

// Intersection 交集：返回在a和b中都存在的唯一元素，按a中首次出现的顺序
func SliceIntersection[T comparable](a, b []T) []T {
	seenB := make(map[T]struct{}, len(b))
	for _, v := range b {
		seenB[v] = struct{}{}
	}

	result := make([]T, 0)
	seenResult := make(map[T]struct{})
	for _, v := range a {
		if _, inB := seenB[v]; inB {
			if _, added := seenResult[v]; !added {
				seenResult[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// Union 并集：返回a和b所有唯一元素，先按a的顺序，再按b中新增元素的顺序
func SliceUnion[T comparable](a, b []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0)

	for _, v := range a {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	for _, v := range b {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Difference 差集：返回在a中存在但b中不存在的唯一元素，按a中顺序
func SliceDifference[T comparable](a, b []T) []T {
	seenB := make(map[T]struct{}, len(b))
	for _, v := range b {
		seenB[v] = struct{}{}
	}

	result := make([]T, 0)
	seenResult := make(map[T]struct{})
	for _, v := range a {
		if _, inB := seenB[v]; !inB {
			if _, added := seenResult[v]; !added {
				seenResult[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}
