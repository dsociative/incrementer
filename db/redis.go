package db

import (
	"github.com/go-redis/redis"
)

const (
	REDIS_INCR_KEY = "incrementer_value"
	REDIS_MAX_KEY  = "incrementer_max"
	REDIS_STEP_KEY = "incrementer_step"
)

var (
	incrScript = redis.NewScript(`
		local step = redis.call("GET", KEYS[1])
		local max = redis.call("GET", KEYS[2])
		local value = redis.call("INCRBY", KEYS[3], step)
		if step == 0 then
			error("no initial settings")
		elseif value > tonumber(max) then
			redis.call("SET", KEYS[3], 0)
			return 0
		else
			return value
		end
	`)
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (r *Redis) Number() (int, error) {
	return r.client.Get(REDIS_INCR_KEY).Int()
}

func (r *Redis) setSettingsCommandsPipe(max, step int) redis.Pipeliner {
	pipe := r.client.TxPipeline()
	pipe.Set(REDIS_MAX_KEY, max, 0)
	pipe.Set(REDIS_STEP_KEY, step, 0)
	return pipe
}

func (r *Redis) SetSettings(max, step int) error {
	pipe := r.setSettingsCommandsPipe(max, step)
	_, err := pipe.Exec()
	return err
}

func (r *Redis) Provision(max, step, value int) error {
	err := r.client.Get(REDIS_STEP_KEY).Err()
	if err == redis.Nil {
		pipe := r.setSettingsCommandsPipe(max, step)
		pipe.Set(REDIS_INCR_KEY, value, 0)
		_, err = pipe.Exec()
		return err
	}
	return err
}

func (r *Redis) Incr() (int, error) {
	return incrScript.Run(
		r.client,
		[]string{
			REDIS_STEP_KEY,
			REDIS_MAX_KEY,
			REDIS_INCR_KEY,
		},
	).Int()
}
