package rdretry

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/Flyingon/go-common/lib/dockertest"
)

var successTimes = uint32(3)
var retryExecuteTimes = 0
var isSuccess = false

func retryWorker(ctx context.Context, data string, executeTimes uint32) error {
	retryExecuteTimes += 1
	if executeTimes == successTimes {
		isSuccess = true
		return nil
	}
	return fmt.Errorf("excute failed, %d", executeTimes)
}

func Test_RetryWorkSuccess(t *testing.T) {
	retryExecuteTimes = 0
	isSuccess = false
	var retryMgr = NewManager("test", &Config{
		MaxTimes: 7,
		Intervals: []uint32{
			1, 3, 5, 10,
		},
		batchNum:    300,
		qpm:         3000,
		queuePrefix: "rdk:retry",
	}, dockertest.GetTestClient(), retryWorker, nil)
	retryMgr.Run()
	defer retryMgr.Close()

	data := `{"a": "a", "b": 1}`
	err := retryMgr.Push(data)
	assert.Nil(t, err)

	<-time.After(1100 * time.Millisecond)
	assert.False(t, isSuccess)
	assert.Equal(t, 1, retryExecuteTimes)
	<-time.After(3 * time.Second)
	assert.False(t, isSuccess)
	assert.Equal(t, 2, retryExecuteTimes)
	<-time.After(5 * time.Second)
	assert.True(t, isSuccess)
	assert.Equal(t, 3, retryExecuteTimes)
	<-time.After(11 * time.Second)
}

func Test_RetryFinish(t *testing.T) {
	retryExecuteTimes = 0
	isSuccess = false
	var retryMgr = NewManager("test", &Config{
		MaxTimes: 5,
		Intervals: []uint32{
			1, 3,
		},
		batchNum:    300,
		qpm:         3000,
		queuePrefix: "rdk:retry",
	}, dockertest.GetTestClient(), retryWorker, nil)
	retryMgr.Run()
	defer retryMgr.Close()

	data := `{"a": "a", "b": 1}`
	successTimes = 3
	err := retryMgr.Push(data)
	assert.Nil(t, err)
	size, err := retryMgr.queue.Size(dummyCtx)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), size)

	<-time.After(1100 * time.Millisecond)
	assert.False(t, isSuccess)
	assert.Equal(t, 1, retryExecuteTimes)
	<-time.After(3 * time.Second)
	assert.False(t, isSuccess)
	assert.Equal(t, 2, retryExecuteTimes)

	size, err = retryMgr.queue.Size(dummyCtx)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)
}

func Test_RetryWorkMax(t *testing.T) {
	retryExecuteTimes = 0
	isSuccess = false
	var retryMgr = NewManager("test", &Config{
		MaxTimes: 1,
		Intervals: []uint32{
			1, 3, 5, 10,
		},
		batchNum:    300,
		qpm:         3000,
		queuePrefix: "rdk:retry",
	}, dockertest.GetTestClient(), retryWorker, nil)
	retryMgr.Run()
	defer retryMgr.Close()

	data := `{"a": "a", "b": 1}`
	successTimes = 3
	err := retryMgr.Push(data)
	assert.Nil(t, err)

	<-time.After(1100 * time.Millisecond)
	assert.False(t, isSuccess)
	assert.Equal(t, 1, retryExecuteTimes)

	size, err := retryMgr.queue.Size(dummyCtx)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)
}
