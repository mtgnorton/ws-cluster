package kit

import (
	"sync/atomic"
	"time"
)

// AllowByInterval 控制高频场景下的日志输出频率。
func AllowByInterval(last *atomic.Int64, interval time.Duration) bool {
	now := time.Now().UnixNano()
	limit := interval.Nanoseconds()

	for {
		prev := last.Load()
		if prev != 0 && now-prev < limit {
			return false
		}
		if last.CompareAndSwap(prev, now) {
			return true
		}
	}
}
