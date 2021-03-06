package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/siddontang/go-log/log"
	"io/ioutil"
	"os"
	"time"
)

const dftTimeOut = 3000 // 毫秒

// NewPool args: {maxIdle} {MaxActive} {IdleTimeout}
func NewPool(server, password string, db int, args ...int) *redis.Pool {
	maxIdle := 3000
	maxActive := 3000
	idleTimeout := 240
	log.Infof("maxIdle: %d, MaxActive: %d, idleTimeout: %d", maxIdle, maxActive, idleTimeout)
	if len(args) >= 3 {
		maxIdle = args[0]
		maxActive = args[1]
		idleTimeout = args[2]
	}
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp",
				server,
				redis.DialConnectTimeout(dftTimeOut*time.Millisecond),
				redis.DialReadTimeout(dftTimeOut*time.Millisecond),
				redis.DialWriteTimeout(dftTimeOut*time.Millisecond),
			)
			if err != nil {
				return nil, err
			}
			// 验证密码，如果有密码
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			// 选择db
			c.Do("SELECT", db)
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func InitScripts() {
}

func NewScripts(fileName string) *redis.Script {
	script, err := CreateScriptFromFile(fileName, 1)
	if err != nil {
		log.Error("load scripts from %v fail:%v", fileName, err)
		panic(err)
	}
	return script
}

func CreateScriptFromFile(filename string, keyCount int) (*redis.Script, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file err:%v", err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file err:%v", err)
	}
	luaStr := string(b)
	return redis.NewScript(keyCount, luaStr), nil
}
