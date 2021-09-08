package grade

import (
	"fmt"
	"github.com/Flyingon/go-common/lib/redis"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	redisName = "redis1"
)

// TestLevelIncr ...
func TestLevelIncr(t *testing.T) {
	elm := Element{
		RedisName:      redisName,
		RedisKeyPrefix: "grade:sort",
		Level:          31,
	}
	errInit := elm.LevelIncr(1)
	fmt.Println(errInit)
	redis.SetRedisClt(redisName, redis.NewPool("127.0.0.1:6379", "", 0))
	errClean := elm.Clean()
	assert.Nil(t, errClean)

	for k, v := range map[int64]int{
		3:  5,
		15: 3,
		4:  5,
		20: 1,
		22: 3,
	} {
		elm.Level = k
		errSet := elm.LevelIncr(v)
		assert.Nil(t, errSet)
	}
	elm.Level = 20
	pos1, err1 := elm.GetExceedPos(true, 4)
	fmt.Println(pos1, err1)
	pos2, err2 := elm.GetExceedPos(false, 4)
	fmt.Println(pos2, err2)
}
