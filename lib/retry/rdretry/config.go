package rdretry

type Config struct {
	MaxTimes  uint32   // 最大重试次数
	Intervals []uint32 // 重试时间间隔

	batchNum    uint32
	qpm         uint32
	queuePrefix string
}

var defaultConfig = Config{
	MaxTimes: 7,
	Intervals: []uint32{
		3, 5, 10, 30, 60, 300, 600,
	},
	batchNum:    300,
	qpm:         3000,
	queuePrefix: "rdk:retry",
}
