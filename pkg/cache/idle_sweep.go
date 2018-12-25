package cache

import (
	"sync"
	"time"

	"github.com/uber-go/atomic"
)

type ConnectionState struct {
	lastActivity atomic.Int64
}

type GenericConnection struct {
	conn  Connection
	state ConnectionState
}

type idleSweep struct {
	conns             []GenericConnection
	maxIdleTime       time.Duration
	idleCheckInterval time.Duration
	stopCh            chan struct{}
	started           bool
	mutable           sync.RWMutex
	shipedConnCh      chan Connection
}

func startIdleSweep(opts *Options) *idleSweep {
	is := &idleSweep{
		maxIdleTime:       opts.TimeoutIdleConn,
		idleCheckInterval: time.Duration(100) * time.Millisecond,
		shipedConnCh:      make(chan Connection, 10),
	}
	is.start()
	return is
}

func (is *idleSweep) start() {
	if is.started || is.idleCheckInterval <= 0 {
		return
	}

	is.started = true
	is.stopCh = make(chan struct{})
	go is.pollerLoop()
}

func (is *idleSweep) pollerLoop() {
	ticker := time.NewTimer(is.idleCheckInterval)

	for {
		select {
		case <-ticker.C:
			is.checkIdleConnections()
		case <-is.stopCh:
			ticker.Stop()
			return
		}
	}
}

func (is *idleSweep) checkIdleConnections() {
	now := time.Now()
	idleConnections := make([]Connection, 0, 100)
	is.mutable.RLock()
	for _, conn := range is.conns {
		lastTime := time.Unix(0, conn.state.lastActivity.Load())
		if idleTime := now.Sub(lastTime); idleTime >= is.maxIdleTime {
			idleConnections = append(idleConnections, conn.conn)
		}
	}
	is.mutable.RUnlock()

	for _, conn := range idleConnections {
		is.shipedConnCh <- conn
		conn.Close()
	}
}

func (is *idleSweep) AddConnection(conn Connection) (err error) {
	is.conns = append(is.conns, GenericConnection{
		conn: conn,
		state: ConnectionState{
			lastActivity: *atomic.NewInt64(time.Now().Unix()),
		},
	})
	return
}
