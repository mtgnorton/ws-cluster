package kit

import (
	"testing"
	"time"
)

func TestSampling(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		ch := make(chan int)
		expect := [10]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
		actual := [10]int{}
		go func() {
			for i := 1; i < 101; i++ {
				time.Sleep(10 * time.Millisecond)
				ch <- i
			}
			close(ch)
		}()
		i := 0
		Sampling(ch, time.Millisecond*100, 0, func(item int) {
			actual[i] = item
			i++
		})
		for i := 0; i < 10; i++ {
			if expect[i] != actual[i] {
				t.Errorf("expect %v, actual %v", expect, actual)
			}
		}
	})
}
