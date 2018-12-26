package cache

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

var redisConn Connection

func init() {
	var fns []func(*Options)
	var err error
	fns = append(fns, func(opt *Options) {
		opt.HostPort = "10.6.26.188:6000"
	})
	fns = append(fns, func(opt *Options) {
		opt.Password = "sldlsflskmg1MAku9O"
	})
	fns = append(fns, func(opt *Options) {
		opt.DBIndex = 0
	})
	fns = append(fns, func(opt *Options) {
		opt.MaxConns = 1000000
	})
	fns = append(fns, func(opt *Options) {
		opt.TimeoutIdleConn = 10
	})
	if redisConn, err = NewRedisConnection(fns...); err != nil {
		panic(err.Error())
	}
}

func TestSetAndGetRedis(t *testing.T) {
	var err error
	defer redisConn.Close()
	if err = redisConn.Set("/tmp/demo/test01", "tmp"); err != nil {
		t.Error(err.Error())
	}
	if err = redisConn.Set("/tmp/demo/test02", "tmp"); err != nil {
		t.Error(err.Error())
	}
	var value interface{}
	var tmp string
	if value, err = redisConn.Get("/tmp/demo/test01"); err != nil {
		t.Error(err.Error())
	}
	if tmp, err = redis.String(value, err); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, "tmp", tmp, "`/tmp/demo/test01` != tmp")
}

func TestDelRedis(t *testing.T) {
	var err error
	defer redisConn.Close()
	if err = redisConn.Del("/tmp/demo/test01"); err != nil {
		t.Error(err.Error())
	}
	if err = redisConn.Del("/tmp/demo/test02"); err != nil {
		t.Error(err.Error())
	}
}
