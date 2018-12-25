package discovery

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/microsvs/base/pkg/env"
	"github.com/microsvs/base/pkg/log"
	"github.com/microsvs/libkv"
	"github.com/microsvs/libkv/store"
	"github.com/microsvs/libkv/store/zookeeper"
)

//ZKConn 所有的配置入口
var (
	zkInitDone = make(chan struct{})
	kv         store.Store
	kvType     store.Backend = store.ZK // default zk config
)

var memAtomic sync.Map

//Read 统一的配置中心维护机制，适用于简单变量
func KVRead(path string, def string) string {
	var (
		ret    string = def
		kvpair *store.KVPair
		err    error
	)
	if path[:1] != "/" {
		serviceEnv, _ := env.Get(env.ServiceENV)
		serviceName, _ := env.Get(env.ServiceName)
		serviceVer, _ := env.Get(env.ServiceVer)
		path = fmt.Sprintf("/%s/%s/%s/%s", serviceName, serviceVer, serviceEnv, path)
	}
	if value, ok := memAtomic.Load(path); !ok {
		if kvpair, err = kv.Get(path); err != nil {
			return def
		}
		ret = string(kvpair.Value)
		memAtomic.Store(path, ret)
		go watch(path) // watch key-value change
	} else {
		ret = value.(string)
	}
	return ret
}

func watch(key string) {
	var (
		stopCh   = make(chan struct{})
		err      error
		kvpairCh <-chan *store.KVPair
	)
	if kvpairCh, err = kv.Watch(key, stopCh); err != nil {
		close(stopCh) // stop watch interval goroutine
		memAtomic.Delete(key)
		log.ErrorRaw("watch %s failed. err=%s", key, err.Error())
		return
	}
	for {
		kvpair := <-kvpairCh
		memAtomic.Store(key, string(kvpair.Value))
	}
	return
}

func Watch(path string) (kvpairCh <-chan *store.KVPair, err error) {
	var stopCh = make(chan struct{})
	if path[:1] != "/" {
		serviceEnv, _ := env.Get(env.ServiceENV)
		serviceName, _ := env.Get(env.ServiceName)
		serviceVer, _ := env.Get(env.ServiceVer)
		path = fmt.Sprintf("/%s/%s/%s/%s", serviceName, serviceVer, serviceEnv, path)
	}
	if kvpairCh, err = kv.Watch(path, stopCh); err != nil {
		close(stopCh) // stop watch interval goroutine
		log.ErrorRaw("watch %s failed. err=%s", path, err.Error())
		return
	}
	return kvpairCh, nil
}

func maintainKV() {
	var (
		err      error
		waiting  = make(chan struct{})
		stop     = make(chan struct{})
		retryMax = 3
		curTimes int
		interval time.Duration
	)
	endpointStr, _ := env.Get(env.KVConfig)
	endpoints := strings.Split(endpointStr, ",")
	if kv, err = libkv.NewStore(kvType, endpoints, &store.Config{} /*nil*/); err != nil {
		panic(err)
	}

	//首次初始化
	zkInitDone <- struct{}{}
	for {
		kv.WaitingConnCloseState(waiting, stop)
		select {
		case <-stop:
			close(stop)
			close(waiting)
			return
		case <-waiting:
			// reconnect
			for curTimes <= retryMax {
				kv, err = libkv.NewStore(kvType, endpoints, &store.Config{} /*nil*/)
				if err != nil {
					interval += 3 * time.Second
					time.Sleep(interval)
					curTimes += 1
				} else {
					break
				}
			}
		}
	}
	return
}

func init() {
	zookeeper.Register()
	go maintainKV() // 初始化ZK连接，如果断开或者超时自动重连
	<-zkInitDone
}
