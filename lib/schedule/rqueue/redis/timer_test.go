package redis

import (
	"context"
	"os"
	"testing"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/Flyingon/go-common/lib/dockertest"
)

var dummyCtx = context.TODO()

func TestMain(m *testing.M) {
	dockertest.CreateRedisClient()
	code := m.Run()
	dockertest.CloseRedisClient()

	os.Exit(code)
}

func Test_DoZPopLessThanToNew(t *testing.T) {
	clt := dockertest.GetTestClient()
	srcQueue := "test:wait"
	dstQueue := "test:run"
	scoreMapKey := "test:score:map"
	curTs := time.Now().Unix()
	task1 := "aaa|1|test:task"
	task2 := "bbb|2|test:task"
	// 设置任务score_map
	res := clt.HSet(dummyCtx, scoreMapKey, "aaa", 85)
	assert.Nil(t, res.Err())
	// 添加两个任务
	res = clt.ZAdd(dummyCtx, srcQueue, &goredis.Z{
		Score:  float64(curTs),
		Member: task1,
	}, &goredis.Z{
		Score:  float64(curTs),
		Member: task2,
	})
	assert.Nil(t, res.Err())
	// 任务pop
	tasks, err := DoZPopLessThanToNew(dummyCtx, clt,
		srcQueue, dstQueue, scoreMapKey, curTs, 10, float64(curTs))
	assert.Nil(t, err)
	assert.Equal(t, len(tasks), 2)
	assert.Equal(t, tasks[0].Member, task1)
	assert.Equal(t, tasks[0].Score, curTs)
	assert.Equal(t, tasks[1].Member, task2)
	assert.Equal(t, tasks[1].Score, curTs)
	// 执行中任务检查
	waits, err := clt.ZRangeWithScores(dummyCtx, dstQueue, 0, 10).Result()
	assert.Nil(t, err)
	assert.Equal(t, len(waits), 2)
	assert.Equal(t, waits[1].Member, task1)
	assert.Equal(t, int64(waits[1].Score), curTs+85)
	assert.Equal(t, waits[0].Member, task2)
	assert.Equal(t, int64(waits[0].Score), curTs+60)
}
