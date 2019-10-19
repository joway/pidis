package e2e

import (
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SystemTestSuite struct {
	suite.Suite

	cli *redis.Client
}

func TestSystemTestSuite(t *testing.T) {
	suite.Run(t, new(SystemTestSuite))
}

func (suite *SystemTestSuite) SetupTest() {
	cli, err := getRedisClient()
	suite.cli = cli
	suite.NoError(err)
}

func (suite *SystemTestSuite) TearDownTest() {
	suite.NoError(clearRedis(suite.cli))
}

func (suite *SystemTestSuite) TestEcho() {
	result, err := suite.cli.Echo("hi").Result()
	suite.NoError(err)
	suite.Equal("hi", result)
}
