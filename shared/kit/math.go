package kit

import (
	"cmp"
	"math"

	"golang.org/x/exp/rand"
)

// Max 返回两个可比较类型值中的较大值
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Min 返回两个可比较类型值中的较小值
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// RoundF64 四舍五入保留n位小数
func RoundF64(f float64, n int) float64 {
	pow := math.Pow(10, float64(n))
	return math.Round(float64(f)*pow) / pow
}

// FloorF64 向下取整保留n位小数
func FloorF64(f float64, n int) float64 {
	pow := math.Pow(10, float64(n))
	return math.Floor(float64(f)*pow) / pow
}

// CeilF64 向上取整保留n位小数
func CeilF64(f float64, n int) float64 {
	pow := math.Pow(10, float64(n))
	return math.Ceil(float64(f)*pow) / pow
}

// RandInt 返回一个在min和max之间的随机整数
func RandInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// RandFloat64 返回一个在min和max之间的随机浮点数, 保留n位小数
func RandFloat64(min, max float64, n int) float64 {
	return RoundF64(rand.Float64()*(max-min), n)
}
