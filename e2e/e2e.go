package e2e

import (
	"github.com/go-redis/redis/v7"
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/util"
	"os"
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

func clearRedis(cli *redis.Client) error {
	//delete all keys
	keys, err := cli.Keys("*").Result()
	cli.Del(keys...)
	return err
}

func getRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	_, err := client.Ping().Result()
	return client, err
}
