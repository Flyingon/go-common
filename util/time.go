package util

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

var parser cron.Parser
var newParserOnce sync.Once

func Parser() cron.Parser {
	newParserOnce.Do(func() {
		parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	})
	return parser
}

func NowFormat() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Next(spec string, srcTime time.Time) (next time.Time, err error) {
	// 如果是每月生效的，由于该crontab库不支持L. 先用字符串替换
	if strings.Contains(spec, "L") {
		year, month := srcTime.Year(), int(srcTime.Month())
		var days int = 31 // 默认是大月
		switch {
		case month == 2 && year%4 == 0: // 闰2月
			days = 29
		case month == 2 && year%4 != 0: // 平2月
			days = 28
		case month == 4 || month == 6 || month == 9 || month == 11: // 小月 4,6,9,11
			days = 30
		}
		spec = strings.Replace(spec, "L", fmt.Sprint(days), 1)
	}

	sched, err := Parser().Parse(spec)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "spec(%s)", spec)
	}
	return sched.Next(srcTime), nil
}

func Today() string {
	return time.Now().Format("2006-01-02")
}

func ShortToday() string {
	return time.Now().Format("20060102")
}

func WeekStartDate() string {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	return weekStartDate.Format("2006-01-02")
}

func MonthStartDate() string {
	d := time.Now()
	d = d.AddDate(0, 0, -d.Day()+1)
	return d.Format("2006-01-02")
}

func StrToUnix(format string, timeStr string) (int64, error) {
	var outputArg int64 = 0

	loc, err1 := time.LoadLocation("Local")
	if err1 != nil {
		return 0, err1
	}

	tm, err2 := time.ParseInLocation(format, timeStr, loc)
	if err2 != nil {
		return 0, err2
	}

	outputArg = tm.Unix()
	return outputArg, nil
}
