package executor_test

import (
	"context"
	"github.com/akutz/memconn"
	"github.com/go-redis/redis/v7"
	"github.com/joway/loki"
	"github.com/joway/pidis/db"
	"github.com/joway/pidis/util"
	"github.com/tidwall/redcon"
	"net"
	"os"
)

var e2eEndpoint = util.EnvGet("E2E_ENDPOINT", "0.0.0.0:10001")
var isE2ERedis = os.Getenv("E2E_REDIS_ENABLE") != ""
var e2eListener net.Listener

func init() {
	//e2e tests original redis server
	if isE2ERedis {
		return
	}

	e2eListener, _ = memconn.Listen("memu", "mem")
	dir := "/tmp/pidis/e2e"
	database, _ := db.New(db.Options{DBDir: dir})
	redisServer := redcon.NewServer(
		"",
		db.GetRedisCmdHandler(database),
		nil,
		nil,
	)
	go func() {
		if err := redisServer.Serve(e2eListener); err != nil {
			loki.Fatal("%v", err)
		}
	}()
}

func e2eClearRedis(cli *redis.Client) error {
	//delete all keys
	keys, err := cli.Keys("*").Result()
	cli.Del(keys...)
	return err
}

func e2eGetRedisClient() (*redis.Client, error) {
	var client *redis.Client
	if isE2ERedis {
		client = redis.NewClient(&redis.Options{
			Addr: e2eEndpoint,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return memconn.Dial("memu", "mem")
			},
		})
	}

	_, err := client.Ping().Result()
	return client, err
}
