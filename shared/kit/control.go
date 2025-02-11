package kit

// IfElse 如果条件为真，则返回第一个值，否则返回第二个值
func IfElse[T any](condition bool, a T, b T) T {
	if condition {
		return a
	}
	return b
}

func IfElseFunc[T any](condition bool, a func() T, b func() T) T {
	if condition {
		return a()
	}
	return b()
}
