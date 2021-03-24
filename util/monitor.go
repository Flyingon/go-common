package util

import (
	"github.com/siddontang/go-log/log"
	"time"
)

// ReportMonitor 上报监控
// args: 0: 上报累计的数值,默认是1
func ReportMonitor(msg string, args ...float64) {
	value := float64(1)
	if len(args) > 0 {
		value = args[0]
	}
	log.Debugf("report monitor, msg: %s, val: %d", msg, value)
	return
}

// ReportTimeDuration 耗时上报
func ReportTimeDuration(msg string, duration time.Duration) {
	log.Debugf("report time duration, msg: %s, val: %d ms", msg, duration.Microseconds())
}
