package kit

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var layout = "2006-01-02 15:04:05"

var ErrBeginTimeAfterEndTime = errors.New("begin time after end time")

// TimeGetBeginEndTs
// 获取一段时间或者一天的开始时间和结束时间戳
// 如果 len(days) = 0, 则获取当天的开始时间和结束时间戳
// 如果 len(days) = 1, 则获取那天的开始时间和结束时间戳
// 如果 len(days) >= 2, 则获取前两个日期之间的开始时间和结束时间戳
func TimeGetBeginEndTs(dhs ...string) (beginTs, endTs int64, err error) {
	var (
		beginDay, endDay string
	)
	if len(dhs) == 0 {
		beginDay = time.Now().Format("2006-01-02 15:04:05")
		endDay = beginDay
	}
	if len(dhs) == 1 {
		beginDay = dhs[0]
		endDay = beginDay
	}
	if len(dhs) >= 2 {
		beginDay = dhs[0]
		endDay = dhs[1]
	}

	// 如果匹配2024-9-5，则补全为2024-09-05 00:00:00
	// 创建正则表达式，匹配年、月、日
	re := regexp.MustCompile(`^(\d{4})-(\d{1,2})-(\d{1,2})$`)

	// 使用 ReplaceAllStringFunc 对匹配到的部分进行处理
	reviseFun := func(m string) string {
		return re.ReplaceAllStringFunc(m, func(m string) string {
			var year, month, day string
			ss := strings.Split(m, "-")
			year = ss[0]
			month = ss[1]
			day = ss[2]
			// 补全月和日
			if len(month) == 1 {
				month = "0" + month
			}
			if len(day) == 1 {
				day = "0" + day
			}
			return fmt.Sprintf("%s-%s-%s 00:00:00", year, month, day)
		})
	}
	beginDay = reviseFun(beginDay)
	endDay = reviseFun(endDay)
	fmt.Println(beginDay, endDay)
	beginTime, err := time.Parse(layout, beginDay)
	if err != nil {
		return
	}
	y, m, d := beginTime.Date()
	// 获取当天的开始时间
	beginTs = time.Date(y, m, d, 0, 0, 0, 0, time.Local).Unix()

	endTime, err := time.Parse(layout, endDay)
	if err != nil {
		return
	}
	if beginTime.After(endTime) {
		return 0, 0, ErrBeginTimeAfterEndTime
	}
	y, m, d = endTime.Date()

	// 获取当天的结束时间
	endTs = time.Date(y, m, d, 23, 59, 59, 0, time.Local).Unix()
	return
}

// TimeGetDateByTs 获取时间戳对应的日期
// 如 1609434000(2021-01-01 01:00:00) 对应的日期是 2021-01-01
func TimeGetDateByTs(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02")
}

// TimeGetDateByDateHMS 由Y-m-d H:i:s 格式转换为 Y-m-d 格式
func TimeGetDateByDateHMS(dhs string) (string, error) {
	beginTime, err := time.Parse(layout, dhs)
	if err != nil {
		return "", err
	}
	return TimeGetDateByTs(beginTime.Unix()), nil
}

// TimeGetBeforeDateHMS 获取指定日期之前的日期
func TimeGetBeforeDateHMS(dhs string) (string, error) {
	t, err := time.Parse(layout, dhs)
	if err != nil {
		return "", err
	}
	return TimeGetDateByTs(t.AddDate(0, 0, -1).Unix()), nil
}

// IsOneDate 判断两个日期是否是同一天
func IsOneDate(beginDHS, endDHS string) (bool, error) {
	beginDate, err := TimeGetDateByDateHMS(beginDHS)
	if err != nil {
		return false, err
	}
	endDate, err := TimeGetDateByDateHMS(endDHS)
	if err != nil {
		return false, err
	}
	return beginDate == endDate, nil
}

// TimeDurationDates 获取时间区间内的所有日期
// 比如 参数 2021-01-01 01:00:00 到 2021-01-05 01:00:00, 则返回 2021-01-01, 2021-01-02, 2021-01-03, 2021-01-04, 2021-01-05
func TimeDurationDates(beginDay, endDay string) (dates []string, err error) {
	beginTs, endTs, err := TimeGetBeginEndTs(beginDay, endDay)
	if err != nil {
		return nil, err
	}

	for endTs > beginTs {
		dates = append(dates, TimeGetDateByTs(endTs))
		endTs -= 86400
	}
	return
}
