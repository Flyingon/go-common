/*
	基于 redis zset 实现简单队列
*/

package rdretry

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Queue 任务队列
type Queue struct {
	queueName string
	batchNum  uint32
	taskChan  chan string
	shutdown  chan struct{}
	popSHA    string

	redisClient    *goredis.Client
	consumeLimiter *rate.Limiter
	taskHandler    func(context.Context, string) error // 任务处理函数
	logger         *zap.Logger
}

// NewQueue 创建新队列
func NewQueue(redisClient *goredis.Client, queueName string, batchNum, consumeQpm uint32,
	taskHandler func(context.Context, string) error, logger *zap.Logger) *Queue {
	limit := rate.Every(time.Minute / time.Duration(consumeQpm))
	limiter := rate.NewLimiter(limit, 1)
	return &Queue{
		queueName: queueName,
		batchNum:  batchNum,
		taskChan:  make(chan string, batchNum*2),
		shutdown:  make(chan struct{}),

		redisClient:    redisClient,
		consumeLimiter: limiter,
		taskHandler:    taskHandler,
		logger:         logger,
	}
}

// Run begin to run
func (q *Queue) Run(ctx context.Context) {
	go q.consumer(ctx)
	go q.handler(ctx)
}

// Close queue close to run
// 先停止 consumer, 再停止 handler
func (q *Queue) Close() {
	close(q.shutdown)
	time.Sleep(500 * time.Millisecond)
	close(q.taskChan)
}

// consumer 消费,从队列中pop处理
func (q *Queue) consumer(ctx context.Context) {
	for {
		select {
		case <-q.shutdown:
			return
		default:
			err := q.consumeLimiter.Wait(ctx)
			if err != nil {
				q.logger.Error("consume limiter failed", zap.Error(err))
				// TODO 上报打点
			}
			taskList := q.popByNow(ctx)
			for _, task := range taskList {
				q.taskChan <- task
			}
		}
	}
}

// handler 处理消费的内容
func (q *Queue) handler(ctx context.Context) {
	for {
		taskInfo, ok := <-q.taskChan
		if !ok {
			q.logger.Warn("queue closed", zap.Any("name", q.queueName))
			return
		}
		if q.taskHandler == nil {
			time.Sleep(time.Second)
			continue
		}
		go func(ctx context.Context) { // TODO 协程池 if need
			defer func() {
				// handler recover
				if r := recover(); r != nil {
					q.logger.Error("Catch panic",
						zap.Any("err", r),
						zap.Any("stack", string(debug.Stack())),
					)
				}
			}()
			err := q.taskHandler(ctx, taskInfo)
			if err != nil {
				q.logger.Error("Task handler failed",
					zap.Error(err),
				)
			}
		}(ctx)
		time.Sleep(100 * time.Millisecond)
	}
}

// Push key push to zset
func (q *Queue) Push(ctx context.Context, data string, executeTs int64) error {
	err := q.redisClient.ZAdd(ctx, q.queueName, &goredis.Z{Score: float64(executeTs), Member: data}).Err()
	//redis返回失败
	if err != nil {
		q.logger.Error("Push to queue failed", zap.Error(err),
			zap.Any("queue", q.queueName),
			zap.Any("info", data),
			zap.Any("execute_ts", executeTs),
		)
		return err
	}
	return nil
}

func (q *Queue) popByNow(ctx context.Context) []string {
	beginTime := time.Now()
	curTs := beginTime.Unix()
	val, err := q.redisClient.EvalSha(ctx, q.popSHA, []string{q.queueName}, "-inf", curTs, q.batchNum).Slice()
	if err != nil {
		if isLuaScriptGone(err) { // when redis restart, the script needs to be uploaded again
			sha, errReload := q.redisClient.ScriptLoad(ctx, luaZPopByScore).Result()
			if errReload != nil {
				q.logger.Error("Failed to reload script", zap.Error(errReload))
				time.Sleep(time.Second)
				return nil
			}
			q.popSHA = sha
		}
		q.logger.Error("Failed to pop by now", zap.Error(err))
		time.Sleep(time.Second)
		return nil
	}
	elemsList, err := result2elements(val)
	if err != nil {
		q.logger.Error("Failed to parse zset result", zap.Error(err), zap.Any("val", val))
		return nil
	}
	return elemsToTaskList(elemsList)
}

func (q *Queue) Size(ctx context.Context) (size int64, err error) {
	return q.redisClient.ZCard(ctx, q.queueName).Result()
}

func isLuaScriptGone(err error) bool {
	return strings.HasPrefix(err.Error(), "NOSCRIPT")
}

// Element zset 返回结构定义
type Element struct {
	Member string `json:"member,omitempty"`
	Score  int64  `json:"score,omitempty"`
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

func elemsToTaskList(elements []*Element) []string {
	taskList := make([]string, len(elements))
	for i, e := range elements {
		taskList[i] = e.Member
	}
	return taskList
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
