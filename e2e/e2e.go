package e2e

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var endpoint = util.EnvGet("E2E_ENDPOINT", "0.0.0.0:10001")
var isE2ERedis = os.Getenv("E2E_REDIS_ENABLE") != ""

func init() {
	//e2e tests original redis server
	if isE2ERedis {
		return
	}

	dir := "/tmp/pikv/e2e"
	database, _ := db.New(db.Options{DBDir: dir})
	db.ListenRedisProtoServer(database, endpoint)
}

func setup(t *testing.T) {
	if isE2ERedis {
		return
	}
	//delete all keys
	cli := getRedisClient(t)
	keys, err := cli.Keys("*").Result()
	fmt.Println("keys", keys)
	assert.NoError(t, err)
	cli.Del(keys...)
}

func getRedisClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	pong, err := client.Ping().Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)

	return client
}
