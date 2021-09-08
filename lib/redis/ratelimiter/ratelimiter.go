package ratelimiter

import (
	"fmt"
	"github.com/Flyingon/go-common/lib/redis"
	"github.com/Flyingon/go-common/util"
	"github.com/siddontang/go-log/log"
)

const (
	redisKeyPrefix = "rate:limiter"
)

type RateLimiter struct {
	LimitNum      int `json:"limit_num"`      // 限制次数
	LimitInterval int `json:"limit_interval"` // 间隔时间, s
}

// CheckRate 简单频率控制
func (r *RateLimiter) CheckRate(redisKey, limitKey string) error {
	if r.LimitNum <= 0 || r.LimitInterval <= 0 {
		log.Warnf("rate limiter config (%+v) is not valid", r)
		return nil
	}
	key := fmt.Sprintf("%s:%s", redisKeyPrefix, limitKey)
	count, err := redis.GetRedisClt(redisKey).IncrBy(key, 1)
	if err != nil {
		util.ReportMonitor(fmt.Sprintf("限频key(%s)查询失败", limitKey), 1, 0)
		err = fmt.Errorf("ratelimit key(%s) get failed, err: %v", limitKey, err)
		log.Error(err.Error())
		return nil // 这里可能是限频的redis出问题了，降级为保证业务请求
	}
	if count == 1 {
		err = redis.GetRedisClt(redisKey).Expire(key, r.LimitInterval)
		if err != nil {
			util.ReportMonitor(fmt.Sprintf("%s设置限频失败", limitKey), 1, 0)
			return err
		}
	}
	if count > r.LimitNum {
		if ttlRes, e := redis.GetRedisClt(redisKey).Ttl(key); e == nil {
			if ttlRes == -1 {
				redis.GetRedisClt(redisKey).Del([]string{key})
			}
		} else {
			util.ReportMonitor(fmt.Sprintf("%s限频有效期检查失败", limitKey), 1, 0)
		}
		util.ReportMonitor(fmt.Sprintf("%s触发限频", limitKey), 1, 0)
		err = fmt.Errorf("%s rate is exceed limit", limitKey)
		return err
	}
	return nil
}
