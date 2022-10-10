package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

const (
	dftMaxRetryTimes = uint32(3)
)

// State the state of task
type State int32

const (
	STATE_UNKNOW  State = 0
	STATE_INIT    State = 100  // 初始
	STATE_RUNNING State = 200  // 运行
	STATE_FINISH  State = 300  // 完成
	STATE_FAILED  State = 400  // 失败
	STATE_TIMEOUT State = 500  // 超时
	STATE_CANCEL  State = 1000 // 取消
)

// Task 任务节点
type Task struct {
	Channel         string   `json:"-"`                           // 任务channel，对应相应的queue
	TaskId          string   `json:"-"`                           // 任务执行时的id
	TaskType        string   `json:"-"`                           // 任务类型
	NextExecuteTime int64    `json:"next_execute_time,omitempty"` // 任务下次执行时间
	RetryTimes      uint32   `json:"-"`                           // 重试次数
	MaxRetry        uint32   `json:"max_retry,omitempty"`         // 最大重试次数
	RunTimeout      uint32   `json:"run_timeout,omitempty"`       // 运行超时时长，单位:s
	State           State    `json:"-"`                           // 任务状态
	ExpireTime      int64    `json:"expire_time,omitempty"`       // 任务失效时间
	Phases          []string `json:"phases,omitempty"`            // 任务启停重试记录
	CreatedAt       int64    `json:"created_at"`                  // 任务创建时间
}

// Config 任务节点配置
type Config struct {
	MaxRetry   *uint32 // 最大重试次数
	CronSpec   string  // 定时任务, cron: "0 */1 * * * ?"
	RunTimeout uint32  // 运行超时时长，单位:s
	Delay      int64   // 延迟运行的时间
}

// GetCronNextTs 查询cron下次执行时间
func GetCronNextTs(cronSpec string) (int64, error) {
	tz := "Asia/Hong_Kong"
	loc, _ := time.LoadLocation(tz)
	sch, err := cron.ParseStandard(cronSpec)
	if err != nil {
		log.Errorf("parse cron_spec failed, err: %s, cron_spec: %s", err.Error(), cronSpec)
		return 0, err
	}
	return sch.Next(time.Now().In(loc)).Unix(), nil
}

// NewTask 仅仅新建普通任务
func NewTask(ctx context.Context, channel, taskType string, config *Config, nextExecuteTs int64) (*Task, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	maxRetry := dftMaxRetryTimes
	if config.MaxRetry != nil {
		maxRetry = *config.MaxRetry
	}
	expiredSeconds := int64(maxRetry*60 + 1) // TODO 任务过期时间，暂时用下次执行时间+默认重试时间+1s

	theTask := &Task{
		TaskId:          strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1),
		State:           STATE_INIT,
		Channel:         channel,
		TaskType:        taskType,
		MaxRetry:        maxRetry,
		RunTimeout:      config.RunTimeout,
		NextExecuteTime: nextExecuteTs,
		ExpireTime:      nextExecuteTs + expiredSeconds,
		CreatedAt:       time.Now().Unix(),
	}
	err := theTask.Save(ctx)
	if err != nil {
		return nil, err
	}
	return theTask, nil
}

// GetTask 查询任务
func GetTask(ctx context.Context, channel, taskType, taskId string) (*Task, error) {
	t := &Task{
		Channel:  channel,
		TaskId:   taskId,
		TaskType: taskType,
	}
	err := t.Load(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// CheckFinish 检查任务是否完成
func (t *Task) CheckFinish(ctx context.Context) (bool, error) {
	err := t.Load(ctx)
	if err != nil {
		return false, err
	}
	if t.State == STATE_FINISH {
		return true, nil
	}
	return false, nil
}

// SchedulerInfo 任务在调度时需要的信息
type SchedulerInfo struct {
	Channel  string `json:"channel"`   // 任务渠道
	TaskId   string `json:"task_id"`   // 任务执行时的id
	TaskType string `json:"task_type"` // 任务类型
}

// TaskInfo 生成调度中的taskInfo信息
func (t *Task) TaskInfo() string {
	return fmt.Sprintf("%s|%s|%s", t.TaskType, t.TaskId, t.Channel)
}

// ParseTaskInfo 解析调度中taskInfo信息
func ParseTaskInfo(taskInfo string) (schInfo *SchedulerInfo, err error) {
	elemList := strings.Split(taskInfo, "|")
	if len(elemList) != 3 {
		err = fmt.Errorf("taskInfo[%s] format is not valid", taskInfo)
		return
	}
	schInfo = &SchedulerInfo{
		TaskType: elemList[0],
		TaskId:   elemList[1],
		Channel:  elemList[2],
	}
	if schInfo.TaskType == "" {
		err = fmt.Errorf("taskInfo[%s] taskType is not valid", taskInfo)
	}
	if schInfo.Channel == "" {
		err = fmt.Errorf("taskInfo[%s] channel is not valid", taskInfo)
	}
	if schInfo.TaskId == "" {
		err = fmt.Errorf("taskInfo[%s] taskID is not valid", taskInfo)
	}
	return
}
