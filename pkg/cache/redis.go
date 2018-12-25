package cache

import (
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/microsvs/base/pkg/utils"
)

var _ Connection = (*Redis)(nil)

type Redis struct {
	option  *Options
	pool    *redis.Pool
	mutable sync.RWMutex
}

var redisInstance *Redis

func NewRedisConnection(fn ...func(*Options)) (Connection, error) {
	var option = new(Options)
	var err error
	for _, f := range fn {
		f(option)
	}
	if redisInstance != nil {
		if utils.CompareAnyValues(redisInstance.option, option) {
			return redisInstance, nil
		}
		if err = redisInstance.Close(); err != nil {
			return nil, err
		}
	}
	redisInstance = &Redis{
		option: option,
	}
	if err = initRedisPool(redisInstance); err != nil {
		return nil, err
	}
	return redisInstance, nil
}

// initial redis conn pool
func initRedisPool(instance *Redis) (err error) {
	var fnOptions []redis.DialOption
	if instance.option.DBIndex >= 0 {
		fnOptions = append(fnOptions, redis.DialDatabase(instance.option.DBIndex))
	}
	if len(instance.option.Password) > 0 {
		fnOptions = append(fnOptions, redis.DialPassword(instance.option.Password))
	}
	instance.pool = &redis.Pool{
		IdleTimeout: instance.option.TimeoutIdleConn,
		MaxIdle:     instance.option.MaxConns,
		MaxActive:   instance.option.MaxConns,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			var (
				conn redis.Conn
				err  error
			)
			conn, err = redis.Dial("tcp", instance.option.HostPort, fnOptions...)
			if err != nil {
				return conn, err
			}
			return conn, nil
		},
	}
	return
}

func (r *Redis) Set(key string, value interface{}) (err error) {
	if r == nil {
		return CacheUninitialConnectionError
	}
	_, err = r.pool.Get().Do("SET", key, value)
	return err
}

func (r *Redis) Get(key string) (value interface{}, err error) {
	if r == nil {
		return nil, CacheUninitialConnectionError
	}
	return r.pool.Get().Do("GET", key)
}

func (r *Redis) Expire(key string, sec int) (err error) {
	if r == nil {
		return CacheUninitialConnectionError
	}
	_, err = r.pool.Get().Do("EXPIRE", key, sec)
	return
}

func (r *Redis) Del(key string) (err error) {
	if r == nil {
		return CacheUninitialConnectionError
	}
	if _, err = r.pool.Get().Do("DEL", key); err != nil {
		return
	}
	return
}

func (r *Redis) Exist(key string) (exist bool, err error) {
	var reply interface{}
	if r == nil {
		return false, CacheUninitialConnectionError
	}
	if reply, err = r.pool.Get().Do("EXISTS", key); err != nil {
		return
	}
	if err = utils.GenericTypeConvert(reply, &exist); err != nil {
		return
	}
	return
}

func (r *Redis) TTL(key string) (ttl time.Duration, err error) {
	var reply interface{}
	if r == nil {
		return 0, CacheUninitialConnectionError
	}
	if reply, err = r.pool.Get().Do("TTL", key); err != nil {
		return
	}
	if err = utils.GenericTypeConvert(reply, &ttl); err != nil {
		return
	}
	return
}

func (r *Redis) ComplexCmd(cmd string, values ...interface{}) (interface{}, error) {
	if r == nil {
		return nil, CacheUninitialConnectionError
	}
	return r.pool.Get().Do(cmd, values...)
}

func (r *Redis) Close() (err error) {
	if r == nil {
		return CacheUninitialConnectionError
	}
	// ::TODO
	return
}
