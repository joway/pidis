package e2e

import (
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKV_GetSetDel(t *testing.T) {
	setup(t)

	cli := getRedisClient(t)
	result, err := cli.Set("k", "v", 0).Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", result)
	result, err = cli.Set("k1", "", 0).Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", result)

	result, err = cli.Get("k").Result()
	assert.NoError(t, err)
	assert.Equal(t, "v", result)

	count, err := cli.Del("k").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	result, err = cli.Get("k").Result()
	assert.Equal(t, redis.Nil, err)
	assert.Equal(t, "", result)

	result, err = cli.Get("k1").Result()
	assert.Equal(t, nil, err)
	assert.Equal(t, "", result)
}
