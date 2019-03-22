package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedis(t *testing.T) {
	redisAddr := os.Getenv("REDIS_TEST")
	if redisAddr == "" {
		t.Skip("skipping redis test; $REDIS_TEST not set")
	}

	r := NewRedis(redisAddr)
	for _, f := range []func(*testing.T, *Redis){
		testRedisIncr,
		testRedisNumber,
		testRedisSetSettings,
		testRedisProvision,
	} {
		r.client.FlushDB()
		f(t, r)
	}
}

func testRedisNumber(t *testing.T, r *Redis) {
	n, err := r.Number()
	assert.Equal(t, redis.Nil, err)
	assert.Equal(t, 0, n)
}

func testRedisSetSettings(t *testing.T, r *Redis) {
	max := 100
	step := 2
	err := r.SetSettings(max, step)
	assert.NoError(t, err)

	rawMax, err := r.client.Get(REDIS_MAX_KEY).Int()
	assert.Equal(t, max, rawMax)
	assert.NoError(t, err)

	rawStep, err := r.client.Get(REDIS_STEP_KEY).Int()
	assert.Equal(t, step, rawStep)
	assert.NoError(t, err)
}

func testRedisIncr(t *testing.T, r *Redis) {
	for _, tc := range []struct {
		max  int
		step int
	}{
		{10, 1},
		{10, 2},
		{10, 9},
		{10, 10},
		{10, 11},
		{0, 1},
	} {
		t.Run(
			fmt.Sprintf("incr max %v step %v", tc.max, tc.step),
			func(t *testing.T) {
				require.NoError(t, r.client.FlushDB().Err())
				require.NoError(t, r.SetSettings(tc.max, tc.step))
				for i := tc.step; i <= tc.max+tc.step; i = i + tc.step {
					n, err := r.Incr()
					if i <= tc.max {
						assert.Equal(t, i, n)
					} else {
						assert.Equal(t, 0, n)
					}
					assert.NoError(t, err)
					t.Log("counter", i, "real", n)
				}
			})
	}
}

func testRedisProvision(t *testing.T, r *Redis) {
	for _, key := range []string{REDIS_INCR_KEY, REDIS_STEP_KEY, REDIS_MAX_KEY} {
		value, err := r.client.Get(key).Int()
		assert.Equal(t, 0, value)
		assert.Equal(t, redis.Nil, err)
	}

	max := 1000
	step := 3
	value := 4
	require.NoError(t, r.Provision(max, step, value))

	for key, expectedValue := range map[string]int{
		REDIS_INCR_KEY: value,
		REDIS_STEP_KEY: step,
		REDIS_MAX_KEY:  max,
	} {
		value, err := r.client.Get(key).Int()
		assert.Equal(t, expectedValue, value)
		assert.NoError(t, err)
	}
}
