package util

import (
	"fmt"
	"time"
)

// ReportMonitor 上报监控
// args: 0: 上报累计的数值,默认是1
func ReportMonitor(msg string, args ...float64) {
	value := float64(1)
	if len(args) > 0 {
		value = args[0]
	}
	fmt.Printf("report monitor, msg: %s, val: %d", msg, value)
	return
}

// ReportTimeDuration 耗时上报
// TODO: 自定义实现
func ReportTimeDuration(msg string, duration time.Duration) {
	fmt.Printf("report time duration, msg: %s, val: %d ms", msg, duration.Microseconds())
}
