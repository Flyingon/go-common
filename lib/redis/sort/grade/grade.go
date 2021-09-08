// 分级排序

package grade

import (
	"fmt"
	"github.com/Flyingon/go-common/lib/redis"
	"github.com/Flyingon/go-common/util"
)

const (
	total         = "total"
	maxLevelCount = 100000
)

// Element 分级元素
type Element struct {
	RedisName      string
	RedisKeyPrefix string
	Level          int64
}

func (e *Element) Check() error {
	if redis.GetRedisClt(e.RedisName) == nil {
		return fmt.Errorf("redis[%s] is not init", e.RedisName)
	}
	return nil
}

// Clean !清理所有排序记录！
func (e *Element) Clean() error {
	if err := e.Check(); err != nil {
		return err
	}
	sortKey := fmt.Sprintf("%s:%s", e.RedisKeyPrefix, "zset")
	infoKey := fmt.Sprintf("%s:%s", e.RedisKeyPrefix, "hset")
	cmdList := []*redis.SingleCmd{
		{
			Cmd:  "DEL",
			Args: []interface{}{sortKey},
		},
		{
			Cmd:  "DEL",
			Args: []interface{}{infoKey},
		},
	}
	err := redis.GetRedisClt(e.RedisName).PipeLine(cmdList)
	if err != nil {
		return err
	}
	return nil
}

// LevelIncr 对应分段incr
func (e *Element) LevelIncr(num int) error {
	if err := e.Check(); err != nil {
		return err
	}
	sortKey := fmt.Sprintf("%s:%s", e.RedisKeyPrefix, "zset")
	infoKey := fmt.Sprintf("%s:%s", e.RedisKeyPrefix, "hset")
	levelNum, err := redis.GetRedisClt(e.RedisName).ZCount(sortKey, e.Level, e.Level)
	if err != nil {
		return err
	}
	levelStr := fmt.Sprint(e.Level)
	if levelNum <= 0 {
		_, e := redis.GetRedisClt(e.RedisName).Do("ZADD", sortKey, e.Level, levelStr)
		if e != nil {
			return err
		}
	}
	cmdList := []*redis.SingleCmd{
		{
			Cmd:  "HINCRBY",
			Args: []interface{}{infoKey, levelStr, num},
		},
		{
			Cmd:  "HINCRBY",
			Args: []interface{}{infoKey, total, num},
		},
	}
	err = redis.GetRedisClt(e.RedisName).PipeLine(cmdList)
	if err != nil {
		return err
	}
	return nil
}

// GetExceedPos 正向排序位置
// includeCurrent 排名是否包含当前得分
func (e *Element) GetExceedPos(includeCurrent bool, decimalNum int) (float64, error) {
	exceedNum, totalNum, err := e.getExceedNum(includeCurrent)
	if err != nil {
		return 0, err
	}
	percent := util.RoundNormal(util.DivideFloat(float64(exceedNum), float64(totalNum)), decimalNum)
	return percent, nil
}

func (e *Element) getExceedNum(includeCurrent bool) (int32, int32, error) {
	sortKey := fmt.Sprintf("%s:%s", e.RedisKeyPrefix, "zset")
	infoKey := fmt.Sprintf("%s:%s", e.RedisKeyPrefix, "hset")

	levelList, err := redis.GetRedisClt(e.RedisName).ZRangeByScore(sortKey, "-inf", e.Level, 0, maxLevelCount)
	if err != nil {
		return 0, 0, err
	}
	levelList = append(levelList, total)
	numMap, err := redis.GetRedisClt(e.RedisName).HMGet(infoKey, levelList)
	totalNum := int32(util.InterfaceToInt(numMap[total]))
	exceedNum := int32(0)
	for level, numStr := range numMap {
		if level == total {
			continue
		}
		levelInt := util.InterfaceToInt64(level)
		numInt := util.InterfaceToInt(numStr)
		if e.Level == levelInt {
			if includeCurrent { // 包含当前的
				exceedNum += int32(numInt)
			}
		} else {
			exceedNum += int32(numInt)
		}
	}
	return exceedNum, totalNum, nil
}
