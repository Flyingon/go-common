package rqueue

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/Flyingon/go-common/util"
	goredis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/Flyingon/go-common/lib/schedule/rqueue/redis"
)

const (
	redisKeyQueuePrefix     = "rq"
	redisKeyScoreIncrPrefix = "rq:si"
)

// panicBufLen panic调用栈日志buffer大小，默认1024
const panicBufLen = 1024

// QueueType queue type
type QueueType string

const (
	QueueTypeWait QueueType = "wait"
	QueueTypeRun  QueueType = "run"
)

// TaskHandler function for task handler
type TaskHandler func(context.Context, string) error

// Queue 任务队列
type Queue struct {
	channel        string
	redisClient    *goredis.Client
	queueType      QueueType
	dstQueueType   QueueType
	consumeLimiter *rate.Limiter
	batchNum       int
	taskChan       chan string
	taskHandler    TaskHandler // 任务处理函数
	shutdown       chan struct{}
}

// NewQueue 创建新队列
func NewQueue(channel string, redisClient *goredis.Client,
	queueType, dstQueueType QueueType, batchNum int, ConsumeQpm int,
	taskHandler TaskHandler) *Queue {
	limit := rate.Every(time.Minute / time.Duration(ConsumeQpm))
	limiter := rate.NewLimiter(limit, 1)
	return &Queue{
		channel:        channel,
		redisClient:    redisClient,
		queueType:      queueType,
		dstQueueType:   dstQueueType,
		consumeLimiter: limiter,
		batchNum:       batchNum,
		taskChan:       make(chan string, batchNum*2),
		taskHandler:    taskHandler,
		shutdown:       make(chan struct{}),
	}
}

// QueueName 队列名称
func (q *Queue) QueueName() string {
	return fmt.Sprintf("%s:{%s}:%s", redisKeyQueuePrefix, q.channel, q.queueType)
}

func (q *Queue) dstQueueName() string {
	return fmt.Sprintf("%s:{%s}:%s", redisKeyQueuePrefix, q.channel, q.dstQueueType)
}

func (q *Queue) scoreIncrKey() string {
	return fmt.Sprintf("%s:{%s}:%s", redisKeyScoreIncrPrefix, q.channel, q.dstQueueType)
}

func (q *Queue) taskDftHandler(taskInfo string) {
	log.Infof("task[%s] is not implement, do nothing", taskInfo)
}

// Run rqueue begin to run
func (q *Queue) Run(ctx context.Context) {
	//log.Infof("channel[%s].queue[%s] begin, keep monitor on queue[%s]", q.channel, q.queueType, q.QueueName())
	go q.consumer(ctx)
	go q.handler(ctx)
}

// Close rqueue close to run
// 先停止 cosumer, 再停止 handler
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
			log.Warnf("channel[%s].queue[%s].consumer closed", q.channel, q.queueType)
			return
		default:
			err := q.consumeLimiter.Wait(ctx)
			if err != nil {
				log.Errorf(err.Error())
				util.ReportMonitor(fmt.Sprintf("队列[%s]获取令牌桶失败", q.queueType))
			}
			switch q.queueType {
			case QueueTypeWait, QueueTypeRun:
				q.popByNow(ctx) // 这里串行，不需要并发pop
			default:
				log.Errorf("queue[%s].type[%v] is not valid", q.queueType, q.queueType)
				util.ReportMonitor(fmt.Sprintf("队列[%s]类型[%v]无效", q.queueType, q.queueType))
			}
		}
	}
}

// handler 处理消费的内容
func (q *Queue) handler(ctx context.Context) {
	for {
		taskInfo, ok := <-q.taskChan
		if !ok {
			log.Warnf("channel[%s].queue[%s].handler closed", q.channel, q.queueType)
			return
		}
		if q.taskHandler != nil {
			go func(ctx context.Context) { // TODO 协程池 if need
				defer func() {
					// handler recover
					if r := recover(); r != nil {
						buf := make([]byte, panicBufLen)
						buf = buf[:runtime.Stack(buf, false)]
						err := fmt.Errorf("channel[%s].task[%s] recover with panic %v, buf: %s",
							q.channel, taskInfo, r, buf)
						log.Errorf(err.Error())
					}
				}()
				_ = q.taskHandler(ctx, taskInfo)
			}(ctx)
		} else {
			q.taskDftHandler(taskInfo)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (q *Queue) popByNow(ctx context.Context) {
	beginTime := time.Now()
	curTs := beginTime.Unix()
	curTsFloat := float64(curTs)
	elemList, err := redis.DoZPopLessThanToNew(ctx, q.redisClient, // TODO 未兼容redis重启，lua需要重新加载
		q.QueueName(), q.dstQueueName(), q.scoreIncrKey(),
		curTs, q.batchNum, curTsFloat)
	costTime := time.Since(beginTime).Milliseconds()
	if err != nil {
		log.Errorf("queue(%s->%s) failed, err:%v", q.queueType, q.dstQueueType, err)
		util.ReportMonitor(fmt.Sprintf("任务队列(%s->%s)失败", q.queueType, q.dstQueueType))
		// TODO 严重异常
	}
	taskIDList := elemToStrList(elemList)
	if len(taskIDList) > 0 {
		log.Debugf("queue(%s->%s) succ, cost: %d, task_num:%v",
			q.queueType, q.dstQueueType, costTime, len(taskIDList))
		for _, taskName := range taskIDList {
			q.taskChan <- taskName
		}
	}
}

func elemToStrList(elements []*redis.Element) []string {
	var taskIds []string
	for _, e := range elements {
		taskIds = append(taskIds, e.Member)
	}
	return taskIds
}

// Push key push to zset
func (q *Queue) Push(ctx context.Context, elemKey string, executeTs float64) error {
	err := q.redisClient.ZAdd(ctx, q.QueueName(), &goredis.Z{Score: executeTs, Member: elemKey}).Err()
	//redis返回失败
	if err != nil {
		log.Errorf("elem push to queue failed, elem: %s, queue: %s, err:%v", elemKey, q.queueType, err)
		return err
	}
	return nil
}

// Remove 任务移除
func Remove(ctx context.Context, redisClient *goredis.Client,
	queueList []string, elemKey string) error {
	pipeline := redisClient.Pipeline()
	defer func() {
		_ = pipeline.Close()
	}()
	for _, queueName := range queueList {
		_ = pipeline.ZRem(ctx, queueName, elemKey)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// DoZCard 队列执行zCard命令，获取队列元素数量
func (q *Queue) DoZCard(ctx context.Context) (int64, error) {
	num, err := q.redisClient.ZCard(ctx, q.QueueName()).Result()
	if err != nil {
		log.Errorf(err.Error())
		return 0, err
	}
	return num, err
}

// DoZRange 队列执行zRange命令，获取队列中元素
func (q *Queue) DoZRange(ctx context.Context, start, stop int64) ([]*redis.Element, error) {
	results, err := q.redisClient.ZRangeWithScores(ctx, q.QueueName(), start, stop).Result()
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	elements := zSlice2elements(results)
	return elements, nil
}

// DoZRangeByScore 队列执行 zrangebyscore 命令，获取队列中元素
func (q *Queue) DoZRangeByScore(ctx context.Context, min, max string) ([]string, error) {
	results, err := q.redisClient.ZRangeByScore(ctx, q.QueueName(), &goredis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	return results, nil
}

// zSlice2elements zSet返回处理
func zSlice2elements(results []goredis.Z) []*redis.Element {
	if len(results) == 0 {
		return nil
	}
	elements := make([]*redis.Element, len(results))
	for index, result := range results {
		member, _ := result.Member.(string)
		elements[index] = &redis.Element{
			Member: member,
			Score:  int64(result.Score),
		}
	}
	return elements
}

// DoSetIncrScore 设置元素被pop后，进入新队列的等待时间
func (q *Queue) DoSetIncrScore(ctx context.Context, taskType string, incrScore uint32) error {
	_, err := q.redisClient.HSet(ctx, q.scoreIncrKey(), taskType, incrScore).Result()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	return nil
}
