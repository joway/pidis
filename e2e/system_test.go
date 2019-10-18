package e2e

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystem_Echo(t *testing.T) {
	setup(t)

	cli := getRedisClient(t)
	result, err := cli.Echo("hi").Result()
	assert.NoError(t, err)
	assert.Equal(t, "hi", result)
}
