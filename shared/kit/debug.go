package kit

import (
	"fmt"
	"time"
)

// Sampling 抽样数据,通过ch通道传入数据,如果时间间隔达到duration或者数据量达到amount,则执行exec函数
// 如果amount或duration为0,则对应的维度不生效,如果不再记录数据,则关闭ch通道
// 该函数需要异步执行,不会阻塞
// 场景: 如在for循环中输出大量数据,如果全部打印,会导致大量日志,此时可以抽样一部分数据进行打印
func Sampling[T any](ch <-chan T, duration time.Duration, amount int, exec func(T)) {
	periodAmount := 0
	periodStartTime := time.Now()
	for item := range ch {
		periodAmount++
		if (amount > 0 && periodAmount >= amount) || (duration > 0 && time.Since(periodStartTime) >= duration) {
			go exec(item) // 异步执行,避免阻塞
			periodAmount = 0
			periodStartTime = time.Now()
		}
	}
}

// ConsumeTimeStaistics 消费时间统计
func ConsumeTimeStaistics(name string) func(label string) string {

	startTime := time.Now()
	times := []time.Time{}
	return func(label string) string {
		s := fmt.Sprintf("[%s] %s: Relative start time: %s", name, label, time.Since(startTime))
		if len(times) > 0 {
			s += fmt.Sprintf(" Relative previous time: %s", time.Since(times[len(times)-1]))
		}
		times = append(times, time.Now())
		return s
	}
}
