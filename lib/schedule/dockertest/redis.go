package dockertest

import (
	"context"
	"log"
	"net"
	"net/url"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest"
)

var dockerRedisPool *dockertest.Pool
var dockerRedisResource *dockertest.Resource
var testClient *redis.Client

// GetTestClient 单元测试client
func GetTestClient() *redis.Client {
	return testClient
}

func CreateRedisClient() {
	var err error
	dockerRedisPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dockerRedisResource, err = dockerRedisPool.Run("redis", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// if run with docker-machine the hostname needs to be set
	u, err := url.Parse(dockerRedisPool.Client.Endpoint())
	if err != nil {
		log.Fatalf("Could not parse endpoint: %s", dockerRedisPool.Client.Endpoint())
	}

	if err = dockerRedisPool.Retry(func() error {
		testClient = redis.NewClient(&redis.Options{
			Addr:     net.JoinHostPort(u.Hostname(), dockerRedisResource.GetPort("6379/tcp")),
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		ping := testClient.Ping(context.Background())
		return ping.Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
}

func CloseRedisClient() {
	_ = testClient.Close()
	err := dockerRedisPool.Purge(dockerRedisResource)
	if err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}
