package kit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
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

// IsNil 判断是否为nil
func IsNil(a any) bool {
	if a == nil {
		return true
	}
	var rv reflect.Value
	if v, ok := a.(reflect.Value); ok {
		rv = v
	} else {
		rv = reflect.ValueOf(a)
	}
	switch rv.Kind() {
	case reflect.Chan,
		reflect.Map,
		reflect.Slice,
		reflect.Func,
		reflect.Interface,
		reflect.UnsafePointer,
		reflect.Ptr:
		return !rv.IsValid() || rv.IsNil()
	default:
		return false
	}
}

// String 将any类型转换为string类型
func String(a any) string {
	if a == nil {
		return ""
	}
	switch value := a.(type) {
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	case time.Time:
		if value.IsZero() {
			return ""
		}
		return value.String()
	case *time.Time:
		if value == nil {
			return ""
		}
		return value.String()
	default:
		// Empty checks.
		if value == nil {
			return ""
		}
		// Reflect checks.
		var (
			rv   = reflect.ValueOf(value)
			kind = rv.Kind()
		)
		switch kind {
		case reflect.Chan,
			reflect.Map,
			reflect.Slice,
			reflect.Func,
			reflect.Ptr,
			reflect.Interface,
			reflect.UnsafePointer:
			if rv.IsNil() {
				return ""
			}
		case reflect.String:
			return rv.String()
		}
		if kind == reflect.Ptr {
			return String(rv.Elem().Interface())
		}
		if jsonContent, err := json.Marshal(value); err != nil {
			return fmt.Sprint(value)
		} else {
			return string(jsonContent)
		}
	}
}

// ConsumeTimeStaistics 消费时间统计
func ConsumeTimeStaistics(name string) func(label string) string {

	startTime := time.Now()
	times := []time.Time{}
	return func(label string) string {
		s := fmt.Sprintf("[%s] %s: It took %s relative to the start time", name, label, time.Since(startTime))
		if len(times) > 0 {
			s += fmt.Sprintf(" and %s relative to the previous record", time.Since(times[len(times)-1]))
		}
		times = append(times, time.Now())
		return s
	}
}
