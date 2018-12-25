package mq

import (
	"fmt"
	"sync"
	"time"

	"github.com/microsvs/base/cmd/discovery"
	"github.com/microsvs/base/pkg/log"
	"github.com/microsvs/base/pkg/rpc"
	"github.com/microsvs/libkv/store"
	"github.com/streadway/amqp"
)

var (
	mqInitDone = make(chan struct{})
)

type channelStore struct {
	sync.Mutex
	channel *amqp.Channel
}

var channelmap = make(map[rpc.FGService]*channelStore)

func ReconnectMQ(service rpc.FGService, cfg string) (initialSucc bool, conn *amqp.Connection, err error) {
	if conn, err = amqp.Dial(cfg); err != nil {
		log.ErrorRaw("初始化mq失败:%s %s %s", service, cfg, err.Error())
		return false, conn, err
	}
	{
		if chs, err := conn.Channel(); err != nil {
			log.ErrorRaw("初始化mq成功,但获取channel失败:%s %s %s\n", service, cfg, err.Error())
			return false, conn, err
		} else {
			channelmap[service] = &channelStore{}
			channelmap[service].Mutex.Lock()
			if channelmap[service].channel != nil {
				channelmap[service].channel.Close()
			}
			channelmap[service].channel = chs
			channelmap[service].Mutex.Unlock()
		}
	}
	return true, conn, nil
}

func getCfgMQ(service rpc.FGService) (<-chan *store.KVPair, error) {
	return discovery.Watch(fmt.Sprintf("mq/%s/%s", service, "master"))
}

func maintainChannel(service rpc.FGService) {
	var (
		onceLock    sync.Once
		initialSucc bool = true
		err         error
	)
	var ch <-chan *store.KVPair
	if ch, err = getCfgMQ(service); err != nil {
		panic(err)
	}
	for {
		kvpair := <-ch
		if initialSucc, _, err = ReconnectMQ(service, string(kvpair.Value)); err != nil {
			panic(err)
		} else if !initialSucc {
			time.Sleep(time.Second)
			continue // 重试
		}

		onceLock.Do(func() {
			//首次初始化
			mqInitDone <- struct{}{}
		})
	}
}

//InitMQ 初始化
func InitMQ(service rpc.FGService) {
	channelmap[service] = &channelStore{}
	go maintainChannel(service)
	<-mqInitDone
}

//ConsumeDefault 返回默认参数的queue
func ConsumeDefault(service rpc.FGService, queue string) <-chan amqp.Delivery {
	return Consume(service, queue, service.String(), true, false, false, false, nil)
}

//Consume 返回一个某个Service的queue
func Consume(service rpc.FGService, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) <-chan amqp.Delivery {
	if _, ok := channelmap[service]; !ok {
		log.ErrorRaw("获取mq入口但是没有初始化channel %s,请先调用 InitMQ", service)
		return nil
	}
	channelmap[service].Mutex.Lock()
	defer channelmap[service].Mutex.Unlock()
	if err := channelmap[service].channel.Qos(100, 0, false); err != nil {
		log.ErrorRaw("[Consume] setting qos failed in rabbitmq, err=%s", err.Error())
	}
	q, err := channelmap[service].channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	if err != nil {
		log.ErrorRaw("获取mq入口但是获取queue失败 %s,请检查队列名称是否正确, err=%s", service, err.Error())
	}
	return q
}
