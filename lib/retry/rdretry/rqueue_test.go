package rdretry

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Flyingon/go-common/lib/dockertest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	dockertest.CreateRedisClient()
	code := m.Run()

	dockertest.CloseRedisClient()
	os.Exit(code)
}

var testQueue *Queue
var testTask1 = ""

func taskHandler(ctx context.Context, data string) error {
	fmt.Println("taskHandler execute: ", data)
	testTask1 = data
	return nil
}

func Test_PushTask(t *testing.T) {
	testQueue = NewQueue(
		dockertest.GetTestClient(),
		"test_retry_queue",
		300,
		3000,
		taskHandler,
		zap.L(),
	)
	testTask1 = ""
	testQueue.Run(dummyCtx)
	defer testQueue.Close()
	executeTs := time.Now().Unix() + 1
	err := testQueue.Push(dummyCtx, "test1", executeTs)
	assert.Nil(t, err)
	assert.Equal(t, "", testTask1)
	size, err := testQueue.Size(dummyCtx)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), size)
	<-time.After(1500 * time.Millisecond)
	assert.Equal(t, "test1", testTask1)
}
