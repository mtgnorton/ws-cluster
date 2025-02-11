package kit

import (
	"sync"
	"time"
)

// DefaultTimeoutDetection 全局超时检测
var DefaultTimeoutDetection = &timeoutDetection{callIDs: make(map[int64]struct{})}

type timeoutDetection struct {
	callIDs map[int64]struct{}
	sync.Mutex
}

func DoWithTimeout(duration time.Duration, timeoutOperate func()) (end func()) {
	return DefaultTimeoutDetection.Do(duration, timeoutOperate)
}

// Do 超时检测
// 如果在指定的duration时间内超时，执行timeoutOperate函数。
// 在timeoutOperate函数中，你可以执行一些操作，比如记录日志。
// 在函数结束时，你应该调用end()函数来释放资源。
func (t *timeoutDetection) Do(duration time.Duration, timeoutOperate func()) (end func()) {
	callID, _ := GenerateUniqueId()

	t.Lock()
	t.callIDs[callID] = struct{}{}
	t.Unlock()
	go func() {
		time.AfterFunc(duration, func() {
			t.Lock()
			defer t.Unlock()
			if _, ok := t.callIDs[callID]; ok {
				timeoutOperate()
				delete(t.callIDs, callID)
			}
		})
	}()

	end = func() {
		t.Lock()
		delete(t.callIDs, callID)
		t.Unlock()
	}
	return
}
