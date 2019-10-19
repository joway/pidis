package e2e

import (
	"fmt"
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

func (suite *KVTestSuite) TestSetNX() {
	result, err := suite.cli.Set("k1", "v", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)

	isCreated, err := suite.cli.SetNX("k1", "v", 0).Result()
	suite.NoError(err)
	suite.False(isCreated)

	isCreated, err = suite.cli.SetNX("k2", "v", 0).Result()
	suite.NoError(err)
	fmt.Println("isCreated", isCreated)
	suite.True(isCreated)
}

func (suite *KVTestSuite) TestSetXX() {
	result, err := suite.cli.Set("k1", "v", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)

	isCreated, err := suite.cli.SetXX("k1", "v", time.Second).Result()
	suite.NoError(err)
	suite.True(isCreated)

	isCreated, err = suite.cli.SetXX("k2", "v", time.Second).Result()
	suite.NoError(err)
	suite.False(isCreated)
}

func (suite *KVTestSuite) TestExists() {
	result, err := suite.cli.Set("k1", "v", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)
	result, err = suite.cli.Set("k2", "v", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)

	count, err := suite.cli.Exists("k1", "k2", "kx").Result()
	suite.NoError(err)
	suite.Equal(int64(2), count)
}

func (suite *KVTestSuite) TestIncr() {
	result, err := suite.cli.Set("k1", "10", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)
	result, err = suite.cli.Set("k2", "x10", 0).Result()
	suite.NoError(err)
	suite.Equal("OK", result)

	num, err := suite.cli.Incr("k1").Result()
	suite.NoError(err)
	suite.Equal(int64(11), num)

	num, err = suite.cli.Incr("k2").Result()
	suite.Error(err)
	suite.Equal(int64(0), num)

	num, err = suite.cli.Incr("k3").Result()
	suite.NoError(err)
	suite.Equal(int64(1), num)
}
