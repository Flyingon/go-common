package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	goredis "github.com/go-redis/redis/v8"
)

// DoPopByScoreToNew 从zset中基于大小传入的大小分值pop出来，并转存到新zset
func DoPopByScoreToNew(ctx context.Context, redisClient *goredis.Client,
	srcQueue, dstQueue, rdScoreMapKey string, minScore, maxScore, popSize int, newScore float64) ([]*Element, error) {
	results, err := zPopMaxToNew.Run(ctx, redisClient,
		[]string{srcQueue, dstQueue, rdScoreMapKey},
		minScore, maxScore, popSize, newScore).Slice()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return nil, err
	}
	return result2elements(results)
}

// DoZPopLessThanToNew 从zset中基于最大门限pop出来，并转存到新zset
func DoZPopLessThanToNew(ctx context.Context, redisClient *goredis.Client,
	srcQueue, dstQueue, rdScoreMapKey string, popScore int64, popSize int, newScore float64) ([]*Element, error) {
	results, err := zPopByScoreToNew.Run(ctx, redisClient,
		[]string{srcQueue, dstQueue, rdScoreMapKey},
		"-inf", popScore, popSize, newScore).Slice()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return nil, err
	}
	return result2elements(results)
}

// DoZPopMaxToNew 从zset中按数量pop出来
func DoZPopMaxToNew(ctx context.Context, redisClient *goredis.Client,
	srcQueue, dstQueue, rdScoreMapKey string, popSize int, newScore float64) ([]*Element, error) {
	results, err := zPopMaxToNew.Run(ctx, redisClient,
		[]string{srcQueue, dstQueue, rdScoreMapKey},
		popSize, newScore).Slice()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return nil, err
	}
	return result2elements(results)
}

// result2elements zset返回处理
func result2elements(results []interface{}) ([]*Element, error) {
	if len(results) == 0 {
		return nil, nil
	}
	elements := make([]*Element, 0)
	for i := 0; i < len(results); i += 2 {
		member, ok := results[i].(string)
		if !ok {
			return nil, errors.New("zset key.type is not string")
		}
		score, err := toFloat64(results[i+1])
		if err != nil {
			return nil, err
		}
		element := &Element{
			Member: member,
			Score:  int64(score),
		}
		elements = append(elements, element)
	}
	return elements, nil
}

func toFloat64(reply interface{}) (float64, error) {
	switch reply := reply.(type) {
	case string:
		n, err := strconv.ParseFloat(reply, 64)
		return n, err
	case nil:
		return 0, errors.New("nil")
	}
	return 0, fmt.Errorf("goredis: unexpected type for Float64, got type %T", reply)
}
