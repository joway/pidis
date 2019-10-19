package e2e

import (
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type KVTestSuite struct {
	suite.Suite

	cli *redis.Client
}

func TestKVTestSuite(t *testing.T) {
	suite.Run(t, new(KVTestSuite))
}

func (suite *KVTestSuite) SetupTest() {
	cli, err := getRedisClient()
	suite.cli = cli
	suite.NoError(err)
}

func (suite *KVTestSuite) TearDownTest() {
	suite.NoError(clearRedis(suite.cli))
}

func (suite *KVTestSuite) TestGetSetDel() {
	result, err := suite.cli.Set("k", "v", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)
	result, err = suite.cli.Set("k1", "", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)

	result, err = suite.cli.Get("k").Result()
	suite.NoError(err)
	suite.Equal("v", result)

	count, err := suite.cli.Del("k").Result()
	suite.NoError(err)
	suite.Equal(int64(1), count)

	result, err = suite.cli.Get("k").Result()
	suite.Equal(redis.Nil, err)
	suite.Equal("", result)

	result, err = suite.cli.Get("k1").Result()
	suite.Equal(nil, err)
	suite.Equal("", result)
}

func (suite *KVTestSuite) TestTTL() {
	result, err := suite.cli.Set("k", "v", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)
	result, err = suite.cli.Set("kx", "v", time.Second).Result()
	suite.NoError(err)
	suite.Equal("OK", result)

	ttl, err := suite.cli.TTL("k").Result()
	suite.NoError(err)
	suite.True(ttl == -1)
	ttl, err = suite.cli.TTL("kn").Result()
	suite.NoError(err)
	suite.True(ttl == -2)
	ttl, err = suite.cli.TTL("kx").Result()
	suite.NoError(err)
	suite.Equal(1, int(ttl.Seconds()))

	time.Sleep(time.Millisecond * 1100)
	ttl, err = suite.cli.TTL("kx").Result()
	suite.NoError(err)
	suite.Equal(0, int(ttl.Seconds()))
}
