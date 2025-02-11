package kit

import (
	"sync"
	"testing"
	"time"
)

func TestNode_GenerateUniqueId(t *testing.T) {
	// 多个协程
	wg := sync.WaitGroup{}

	ch := make(chan int64, 100)
	m := make(map[int64]bool)

	go func() {
		for v := range ch {
			if _, ok := m[v]; ok {
				t.Error("repeat")
			} else {
				m[v] = true
			}
		}
	}()

	wg.Add(100)
	// 100 个协程同时生成100次
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				id, err := GenerateUniqueId()
				if err != nil {
					t.Error(err)
				}
				ch <- id
			}
		}()
	}
	wg.Wait()
	time.Sleep(time.Second)
	close(ch)
}
