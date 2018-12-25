package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	CacheTypeUnsupportedError     = errors.New("cache type unsupported error.")
	CacheUninitialConnectionError = errors.New("uninitial cache connection error.")
	MemoryUnsupportedThisOpError  = errors.New("memory unsupport this operation error.")
	KeyNotExist                   = errors.New("record not found.")
	MemStackOverflow              = errors.New("mem stack overflow.")
)

type Connection interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Expire(key string, sec int) error
	Exist(key string) (bool, error)
	TTL(key string) (time.Duration, error)
	Close() error
	ComplexCmd(cmd string, values ...interface{}) (interface{}, error)
}

type TYPE__CACHE string

const (
	REDIS__CACHE TYPE__CACHE = "redis"
	MEM_CACHE    TYPE__CACHE = "memory"
)

// Cache options, include connection , cache itself
type Options struct {
	CacheType TYPE__CACHE
	// max connection num
	MaxConns int
	// idle connection timeout
	TimeoutIdleConn time.Duration
	// cache server HostPort
	HostPort string
	// password
	Password string
	// write timeout
	WriteTimeout time.Duration
	// read timeout
	ReadTimeout time.Duration
	// create & free conn to sync.Pool
	sync.Pool
	// selected db index
	DBIndex int
}

func NewConnection(cacheType TYPE__CACHE, fn ...func(*Options)) (Connection, error) {
	var (
		conn Connection
		err  error
	)
	switch cacheType {
	case REDIS__CACHE:
		conn, err = NewRedisConnection(fn...)
	case MEM_CACHE:
		conn, err = NewMemoryConnection(fn...)
	default:
		return nil, CacheTypeUnsupportedError
	}
	return conn, err
}
