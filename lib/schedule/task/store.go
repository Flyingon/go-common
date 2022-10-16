package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	goredis "github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
)

const (
	redisKeyTaskPrefix = "t"
	fieldInfo          = "info"
	fieldState         = "state"
	fieldRetryTimes    = "retry_times"
)

var redisClient *goredis.Client

// SetRedisClient 设置redis实例，必须先执行
func SetRedisClient(rdClt *goredis.Client) {
	redisClient = rdClt
}

func (t *Task) storeRedisKey() string {
	return fmt.Sprintf("%s:%s:%s:%s", redisKeyTaskPrefix, t.Channel, t.TaskType, t.TaskId)
}

func (t *Task) packInfo() (string, error) {
	info, err := jsoniter.MarshalToString(t)
	if err != nil {
		return "", err
	}
	return info, err
}

// Save 任务信息写入redis
func (t *Task) Save(ctx context.Context) error {
	info, err := t.packInfo()
	if err != nil {
		return err
	}
	storeData := []interface{}{
		fieldInfo, info, fieldState, uint32(t.State), fieldRetryTimes, t.RetryTimes,
	}
	_, err = redisClient.HMSet(ctx, t.storeRedisKey(), storeData...).Result()
	if err != nil {
		return err
	}
	return nil
}

// Load 从redis读取任务信息
func (t *Task) Load(ctx context.Context) error {
	valList, err := redisClient.HMGet(ctx, t.storeRedisKey(), fieldInfo, fieldState, fieldRetryTimes).Result()
	if err != nil {
		return err
	}
	if len(valList) != 3 {
		return errors.New("load task failed, res len is not matching")
	}
	if valList[1] == nil {
		return fmt.Errorf("task is not exist, id: %s", t.TaskId)
	}
	if valStr, ok := valList[0].(string); ok && len(valStr) > 0 {
		e := jsoniter.UnmarshalFromString(valStr, t)
		if e != nil {
			return e
		}
	}
	if valStr, ok := valList[1].(string); ok && len(valStr) > 0 {
		valInt, e := strconv.ParseInt(valStr, 10, 64)
		if e != nil {
			return e
		}
		t.State = State(valInt)
	}
	if valStr, ok := valList[2].(string); ok && len(valStr) > 0 {
		valUint, e := strconv.ParseUint(valStr, 10, 64)
		if e != nil {
			return e
		}
		t.RetryTimes = uint32(valUint)
	}
	return nil
}

// Del 删除redis任务信息
func (t *Task) Del(ctx context.Context) error {
	_, err := redisClient.Del(ctx, t.storeRedisKey()).Result()
	if err != nil {
		return err
	}
	return nil
}
