package db

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/microsvs/base/cmd/discovery"
	"github.com/microsvs/base/pkg/log"
	"github.com/microsvs/base/pkg/rpc"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

type dbStore struct {
	sync.Mutex
	master sqlbuilder.Database
	slaves []sqlbuilder.Database
}

const (
	// mysql 最大空闲连接数
	MYSQL__MAX_OPEN_CONNECTIONS = 100 * 1000
	// mysql 最多同时打开的连接数
	MYSQL__MAX_IDLE_CONNECTIONS = 1 * 2000
	// mysql超时时间
	MYSQL__MAX_CONN_TIMEOUT = 20
)

var (
	dbmap            = make(map[rpc.FGService]*dbStore)
	masterDBInitDone = make(chan struct{})
	slaveDBInitDone  = make(chan struct{})
)

func InitDB(service rpc.FGService) {
	dbmap[service] = &dbStore{}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		go maintainMasterDB(service)
		<-masterDBInitDone
		wg.Done()
	}()
	go func() {
		go maintainSlaveDB(service)
		<-slaveDBInitDone
		wg.Done()
	}()
	wg.Wait()
}

func maintainMasterDB(service rpc.FGService) {
	var (
		err      error
		settings mysql.ConnectionURL
		db       sqlbuilder.Database
	)
	defer func() {
		masterDBInitDone <- struct{}{}
	}()
	// "admin:123456@tcp(127.0.0.1:3306)/database?charset=utf8&parseTime=True&loc=Local")
	url := discovery.KVRead(fmt.Sprintf("db/%s/%s", service, "master"), "")
	if len(url) <= 0 {
		log.ErrorRaw("[maintainMasterDB] db mysql config libkv empty")
		return
	}

	if settings, err = mysql.ParseURL(url); err != nil {
		log.ErrorRaw("[maintainMasterDB] parse mysql config failed. err=%s", err.Error())
		return
	}
	if db, err = mysql.Open(settings); err != nil {
		log.ErrorRaw("[maintainMasterDB] create mysql connection failed. err=%s", err.Error())
		return
	}
	// init mysql connection setting
	{
		db.SetMaxOpenConns(MYSQL__MAX_OPEN_CONNECTIONS)
		db.SetMaxIdleConns(MYSQL__MAX_IDLE_CONNECTIONS)
		db.SetConnMaxLifetime(MYSQL__MAX_CONN_TIMEOUT * time.Second)
		db.SetLogging(true)
	}
	// store mysql connection
	{
		dbmap[service].Mutex.Lock()
		dbmap[service].master = db
		dbmap[service].Mutex.Unlock()
	}
	return
}

func maintainSlaveDB(service rpc.FGService) {
	var (
		err      error
		settings mysql.ConnectionURL
		dbs      []sqlbuilder.Database
	)
	defer func() {
		slaveDBInitDone <- struct{}{}
	}()
	cfg := discovery.KVRead(fmt.Sprintf("db/%s/%s", service, "slave"), "")
	if len(cfg) <= 0 {
		log.ErrorRaw("[maintainSlaveDB] slave mysql config libzk empty")
		return
	}
	dbcfgs := strings.Split(cfg, ",")
	dbs = make([]sqlbuilder.Database, len(dbcfgs))
	for idx, config := range dbcfgs {
		if settings, err = mysql.ParseURL(config); err != nil {
			log.ErrorRaw("[maintainSlaveDB] parse mysql config failed. err=%s", err.Error())
			return
		}
		if dbs[idx], err = mysql.Open(settings); err != nil {
			log.ErrorRaw("[maintainSlaveDB] create mysql connection failed. err=%s", err.Error())
			return
		}
		// init mysql connection setting
		{
			dbs[idx].SetMaxOpenConns(MYSQL__MAX_OPEN_CONNECTIONS)
			dbs[idx].SetMaxIdleConns(MYSQL__MAX_IDLE_CONNECTIONS)
			dbs[idx].SetLogging(true)
		}
	}
	// store mysql connection
	{
		dbmap[service].Mutex.Lock()
		dbmap[service].slaves = dbs
		dbmap[service].Mutex.Unlock()
	}
	return
}

func MasterDB(service rpc.FGService) sqlbuilder.Database {
	var db sqlbuilder.Database
	if _, ok := dbmap[service]; !ok {
		log.ErrorRaw("[MasterDB] service `%s` uninitialize. please call initDB.", service)
		return nil
	}
	dbmap[service].Mutex.Lock()
	db = dbmap[service].master
	dbmap[service].Mutex.Unlock()
	return db
}

func SlaveDB(service rpc.FGService) sqlbuilder.Database {
	if _, ok := dbmap[service]; !ok {
		log.ErrorRaw("[SlaveDB] service `%s` uninitialize. please call initDB.", service)
		return nil
	}
	dbmap[service].Mutex.Lock()
	defer dbmap[service].Mutex.Unlock()
	if len(dbmap[service].slaves) > 0 {
		return dbmap[service].slaves[rand.Intn(len(dbmap[service].slaves))]
	}
	return dbmap[service].master
}
