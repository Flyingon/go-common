package rdretry

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	sep = "|"
)

var dummyCtx = context.Background()

// TaskHandler function for task handler
type TaskHandler func(context.Context, string, uint32) error

type Manager struct {
	queue  *Queue
	Config *Config

	taskHandler TaskHandler
	logger      *zap.Logger
}

// NewManager 新建 retry manager
func NewManager(channel string, cfg *Config,
	redisClient *goredis.Client, taskHandler TaskHandler, logger *zap.Logger) *Manager {
	if len(channel) == 0 {
		panic("channel is necessary")
		return nil
	}
	if cfg == nil {
		cfg = &defaultConfig
	}
	if logger == nil {
		logger = zap.L()
	}
	m := &Manager{
		Config:      cfg,
		taskHandler: taskHandler,
		logger:      logger,
	}
	queueName := fmt.Sprintf("%s:%s", cfg.queuePrefix, channel)
	m.queue = NewQueue(redisClient, queueName, cfg.batchNum, cfg.qpm, m.retryHandler, logger)
	return m
}

func (m *Manager) Run() {
	m.queue.Run(dummyCtx)
}

func (m *Manager) Close() {
	m.queue.Close()
}

// Push 数据推入重试队列
func (m *Manager) Push(data string) error {
	retryTimes := uint32(1) // 第一次重试
	info := constructInfo(data, retryTimes)
	executeTs := m.getNextRetryTs(retryTimes)
	return m.queue.Push(dummyCtx, info, executeTs)
}

func (m *Manager) retryHandler(ctx context.Context, info string) error {
	data, retryTimes, err := parseInfo(info)
	if err != nil {
		return err
	}
	if retryTimes > m.Config.MaxTimes {
		return fmt.Errorf("retry times exceed, times: %d", retryTimes)
	}
	if m.taskHandler == nil {
		return fmt.Errorf("retry task is nil")
	}
	err = m.taskHandler(ctx, data, retryTimes)
	if err == nil {
		return nil
	}
	// 失败继续重试
	retryTimes += 1
	if retryTimes > m.Config.MaxTimes {
		return fmt.Errorf("retry times exceed, times: %d", retryTimes)
	}
	nextTs := m.getNextRetryTs(retryTimes)
	if nextTs == 0 {
		return fmt.Errorf("reach max retry times")
	}
	err = m.queue.Push(ctx, constructInfo(data, retryTimes), nextTs)
	if err != nil {
		return fmt.Errorf("retry task reset failed, err: %s", err.Error())
	}
	return nil
}

func parseInfo(info string) (data string, retry uint32, err error) {
	infoList := strings.Split(info, sep)
	if len(infoList) != 2 {
		err = fmt.Errorf("retry info is not valid, info: %s", info)
		return
	}
	data = infoList[0]
	retryTimes, errRetry := strconv.ParseUint(infoList[1], 10, 64)
	if errRetry != nil {
		err = fmt.Errorf("retry times is not valid, info: %s, err: %s", info, errRetry)
		return
	}
	retry = uint32(retryTimes)
	return
}

func constructInfo(data string, retry uint32) string {
	return fmt.Sprintf("%s%s%d", data, sep, retry)
}

func (m *Manager) getNextRetryTs(retryTimes uint32) int64 {
	curTs := time.Now().UTC().Unix()
	intervalIndex := int(retryTimes - 1)
	if len(m.Config.Intervals) > intervalIndex {
		return int64(m.Config.Intervals[intervalIndex]) + curTs
	}
	return 0
}
