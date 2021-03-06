package lock

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Flyingon/go-common/lib/redis"
	"github.com/Flyingon/go-common/util"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/siddontang/go-log/log"
	"math/big"
	"time"
)

const (
	redisKeyLock = "lock"
	//一般过期时间
	commonExpire          = 3
	TransactionLockExpire = 15
)

var GetCmpDel *redigo.Script

func init() {
	GetCmpDel = redigo.NewScript(1, LuaScriptGetCmpDel)
}

// GetLockWithClt 获得锁，返回lockValue 释放时需要此值
func GetLockWithClt(redisClt *redis.Pool, lockKey string, expireSeconds int) (string, error) {
	key := fmt.Sprintf("%s:%s", redisKeyLock, lockKey)
	var err error
	defer func() {
		log.Infof("GetLock[%s]: %v", key, err)
	}()

	randNum, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		util.ReportMonitor("获取锁生成随机数失败-异常", 1, 0)
		return "", fmt.Errorf("GetLock gen range failed, err: %v", err)
	}
	lockValue := util.InterfaceToString(time.Now().Unix()) + util.InterfaceToString(randNum.Int64())
	if expireSeconds <= 0 {
		expireSeconds = commonExpire
	}

	rs, err := redisClt.SetPxNx(key, lockValue, expireSeconds*1000)
	//数据库返回失败
	if err != nil {
		util.ReportMonitor("设置锁失败-异常", 1, 0)
		return "", fmt.Errorf("set lock failed, err: %v", err)
	}
	//该key已经设置过
	if rs == "" {
		util.ReportMonitor("设置锁重复-正常", 1, 0)
		return "", fmt.Errorf("set lock dup, err: %v", err)
	}

	return lockValue, nil
}

// GetLock 获得锁，返回lockValue 释放时需要此值
func GetLock(redisKey, lockKey string, expireSeconds int) (string, error) {
	redisClt := redis.GetRedisClt(redisKey)
	return GetLockWithClt(redisClt, lockKey, expireSeconds)
}

// ReleaseLockWithClt 释放锁 通过lua脚本查询比对再删除
// lua脚本返回值说明: -1: 查询key不存在 -2: 查询key的值不等于lockValue 0: 删除key不存在 1: 删除成功
func ReleaseLockWithClt(redisClt *redis.Pool, lockKey string, lockValue string) (bool, error) {
	key := fmt.Sprintf("%s:%s", redisKeyLock, lockKey)
	var err error
	defer func() {
		log.Infof("ReleaseLock[%s]: %v", key, err)
	}()
	conn, err := redisClt.Conn(context.Background())
	if err != nil {
		util.ReportMonitor("释放锁锁失败-获取redisclt失败-异常", 1, 0)
		return false, fmt.Errorf("release lock get redis pool failed, err: %v", err)
	}
	defer conn.Close()
	rs, err := redigo.Int(GetCmpDel.Do(conn, key, lockValue))
	log.Debugf("lock release lua script rs: %+v, err: %v", rs, err)
	if err != nil {
		util.ReportMonitor("释放锁失败-redis操作失败-异常", 1, 0)
		return false, fmt.Errorf("release lock failed, err: %v", err)
	}
	if rs < 0 {
		util.ReportMonitor(fmt.Sprintf("释放锁失败-释放锁返回失败(%d)-异常", rs), 1, 0)
		return false, fmt.Errorf("release lock failed, rs: %v", rs)
	}
	return true, nil
}

// ReleaseLock 释放锁 通过lua脚本查询比对再删除
// lua脚本返回值说明: -1: 查询key不存在 -2: 查询key的值不等于lockValue 0: 删除key不存在 1: 删除成功
func ReleaseLock(redisKey, lockKey string, lockValue string) (bool, error) {
	redisClt := redis.GetRedisClt(redisKey)
	return ReleaseLockWithClt(redisClt, lockKey, lockValue)
}
