package rqueue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Flyingon/go-common/lib/dockertest"
)

var dummyCtx = context.TODO()
var testQueue *Queue
var testTask1 = ""

func testReadyHandler(ctx context.Context, taskInfo string) error {
	testTask1 = taskInfo
	return nil
}

func TestMain(m *testing.M) {
	dockertest.CreateRedisClient()
	testQueue = NewQueue(
		"test",
		dockertest.GetTestClient(),
		"wait",
		"run",
		100,
		300,
		testReadyHandler,
	)
	testTask1 = ""
	testQueue.Run(dummyCtx)
	code := m.Run()

	testQueue.Close()
	dockertest.CloseRedisClient()
	os.Exit(code)
}

func Test_PushTask(t *testing.T) {
	executeTs := time.Now().Unix() + 1
	err := testQueue.Push(dummyCtx, "aaa|1|test:task", float64(executeTs))
	assert.Nil(t, err)
	assert.Equal(t, testTask1, "")
	<-time.After(1 * time.Second)
	assert.Equal(t, testTask1, "aaa|1|test:task")
}

func Test_GetSamplesOneSlug(t *testing.T) {
}
