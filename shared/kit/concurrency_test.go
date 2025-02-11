package kit

import (
	"fmt"
	"testing"
	"time"
)

func TestAsync(t *testing.T) {
	rsCh, cancel := Async([]int{1, 2, 3}, func(i int) (int, error) {
		// 模拟一个耗时操作
		time.Sleep(time.Second * time.Duration(RandInt(1, 5)))
		return i * 2, nil
	}, 1)
	count := 0
	for rs := range rsCh {
		fmt.Println(rs)
		if count == 2 {
			cancel()
		}
		count++
	}
	fmt.Println("done")
}

func TestAsyncOne(t *testing.T) {

	for i := 0; i < 1000; i++ {
		rs := AsyncOne([]int{1, 2, 3, 4, 5, 6, 7}, func(i int) (int, error) {
			// 模拟一个耗时操作
			time.Sleep(time.Millisecond * time.Duration(RandInt(100, 300)))
			return i * 2, nil

		}, 1)
		fmt.Println(rs.Item)
	}

}
