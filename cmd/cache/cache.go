package cache

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/microsvs/base/cmd/discovery"
	pcache "github.com/microsvs/base/pkg/cache"
	"github.com/microsvs/base/pkg/log"
	"github.com/microsvs/base/pkg/rpc"
)

var (
	masterCacheInitDone        = make(chan struct{})
	slaveCacheInitDone         = make(chan struct{})
	cachemap                   = make(map[rpc.FGService]*cacheStore)
	redisHeader         string = "redis://"
)

type cacheStore struct {
	sync.Mutex
	master          pcache.Connection
	masterOptionFn  func(*pcache.Options)
	slaves          []pcache.Connection
	slavesOptionFns []func(*pcache.Options)
}

func InitCache(service rpc.FGService) {
	cachemap[service] = &cacheStore{}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		go maintainMasterCache(service)
		<-masterCacheInitDone
		wg.Done()
	}()
	go func() {
		go maintainSlaveCache(service)
		<-slaveCacheInitDone
		wg.Done()
	}()
	wg.Wait()
}

func maintainMasterCache(service rpc.FGService) {
	var (
		options *pcache.Options
		err     error
		conn    pcache.Connection
	)
	defer func() {
		masterCacheInitDone <- struct{}{}
	}()
	// example: "redis://sldlsflskmg1MAku9O@10.6.26.188:6000"
	cfg := discovery.KVRead(fmt.Sprintf("cache/%s/master", service), "")
	if options, err = parseCacheConfig(pcache.REDIS__CACHE, cfg); err != nil {
		log.ErrorRaw("[maintainMasterCache] parse redis config failed. err=%s", err.Error())
		return
	}
	cachemap[service].masterOptionFn = func(opt *pcache.Options) {
		opt.HostPort = options.HostPort
		opt.Password = options.Password
		opt.TimeoutIdleConn = 120 * time.Second
	}
	if conn, err = pcache.NewConnection(pcache.REDIS__CACHE,
		cachemap[service].masterOptionFn); err != nil {
		log.ErrorRaw("[maintainMasterCache] create new connection from redis pool failed. err=%s", err.Error())
		return
	}
	// store redis connection
	{
		cachemap[service].Lock()
		cachemap[service].master = conn
		cachemap[service].Unlock()
	}
	return
}

func maintainSlaveCache(service rpc.FGService) {
	var (
		options *pcache.Options
		err     error
		conns   []pcache.Connection
		cfgs    []string
	)
	defer func() {
		slaveCacheInitDone <- struct{}{}
	}()
	// example: "redis://sldlsflskmg1MAku9O@10.6.26.188:6000"
	cfg := discovery.KVRead(fmt.Sprintf("cache/%s/slave", service), "")
	if len(cfg) > 0 {
		cfgs = strings.Split(cfg, ",")
	}
	conns = make([]pcache.Connection, len(cfgs))
	cachemap[service].slavesOptionFns = make([]func(*pcache.Options), len(cfgs))
	for idx, config := range cfgs {
		if options, err = parseCacheConfig(pcache.REDIS__CACHE, config); err != nil {
			log.ErrorRaw("[maintainSlaveCache] parse redis config failed. err=%s", err.Error())
			return
		}
		cachemap[service].slavesOptionFns[idx] = func(opt *pcache.Options) {
			opt.HostPort = options.HostPort
			opt.Password = options.Password
			opt.TimeoutIdleConn = 120 * time.Second
		}
		if conns[idx], err = pcache.NewConnection(pcache.REDIS__CACHE,
			cachemap[service].slavesOptionFns[idx]); err != nil {
			log.ErrorRaw("[maintainMasterCache] create new connection from redis pool failed. err=%s", err.Error())
			return
		}
	}
	// store redis connection
	{
		cachemap[service].Lock()
		cachemap[service].slaves = conns
		cachemap[service].Unlock()
	}
	return
}

func parseCacheConfig(typ pcache.TYPE__CACHE, cfg string) (options *pcache.Options, err error) {
	switch typ {
	case pcache.REDIS__CACHE:
		cfg = strings.TrimPrefix(cfg, redisHeader)
		fields := strings.Split(cfg, "@")
		if len(fields) < 2 {
			return &pcache.Options{
				CacheType: pcache.REDIS__CACHE,
				HostPort:  fields[0],
			}, nil
		}
		options = &pcache.Options{
			CacheType: pcache.REDIS__CACHE,
			Password:  fields[0],
			HostPort:  fields[1],
		}
	case pcache.MEM_CACHE:
	default:
	}
	return
}

func MasterCache(service rpc.FGService) pcache.Connection {
	var (
		conn pcache.Connection
		err  error
	)
	if _, ok := cachemap[service]; !ok {
		log.ErrorRaw("[MasterCache] cache uninitialize %s. please call InitCache.", service)
		return nil
	}
	cachemap[service].Lock()
	conn = cachemap[service].master
	cachemap[service].Unlock()

	// check if conn is invalid, create new cache connection
	{
		if err = checkConnPing(conn); err != nil {
			log.ErrorRaw("[MasterCache] ping connection failed. err=%s", err.Error())
			conn.Close()
			cachemap[service].master, err = pcache.NewConnection(pcache.REDIS__CACHE,
				cachemap[service].masterOptionFn)
			if err != nil {
				log.ErrorRaw("[MasterCache] create cache connection failed. err=%s", err.Error())
				return nil
			}
			cachemap[service].Lock()
			conn = cachemap[service].master
			cachemap[service].Unlock()
		}
	}
	return conn
}

func SlaveCache(service rpc.FGService) pcache.Connection {
	var (
		conn pcache.Connection
		idx  int = -1
		err  error
	)
	if _, ok := cachemap[service]; !ok {
		log.ErrorRaw("[SlaveCache] cache uninitialize %s, please call InitCache.", service)
		return nil
	}
	cachemap[service].Lock()
	if len(cachemap[service].slaves) > 0 {
		idx = rand.Intn(len(cachemap[service].slaves))
		conn = cachemap[service].slaves[idx]
	} else {
		conn = cachemap[service].master
	}
	cachemap[service].Unlock()

	// check if conn is invalid, create new cache connection
	{
		if err = checkConnPing(conn); err != nil {
			log.ErrorRaw("[SlaveCache] ping cache connection failed. err=%s", err.Error())
			conn.Close()
			if idx >= 0 {
				cachemap[service].slaves[idx], err = pcache.NewConnection(pcache.REDIS__CACHE,
					cachemap[service].slavesOptionFns[idx])
				if err != nil {
					log.ErrorRaw("[SlaveCache] create cache connection failed. err=%s", err.Error())
					return nil
				}
				cachemap[service].Lock()
				conn = cachemap[service].slaves[idx]
				cachemap[service].Unlock()
			}
		}
	}
	return conn
}

func checkConnPing(conn pcache.Connection) error {
	_, err := conn.ComplexCmd("PING")
	return err
}
