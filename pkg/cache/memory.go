package cache

import (
	"sync"
	"time"
	"unsafe"

	"github.com/microsvs/base/pkg/utils"
)

var (
	_                Connection = (*Memory)(nil)
	MAX_MEM__RECORDS            = 1000000          // 内存存储的最大记录数
	MAX_MEM__SIZE               = 30 * 1024 * 1024 // 30M
	mem              *Memory
)

// unsafe.Sizeof(map)
// memory pool
type Memory struct {
	option        *Options
	imap          map[string]*item
	maxMemRecords int
	maxMemSize    int
	mutex         sync.RWMutex
	stopCh        chan struct{}
}

type item struct {
	elem   interface{}
	expire time.Time
}

func NewMemoryConnection(fn ...func(*Options)) (*Memory, error) {
	var option = new(Options)
	for _, f := range fn {
		f(option)
	}
	if mem != nil {
		if utils.CompareAnyValues(*option, *mem.option) {
			return mem, nil
		}
	}
	mem = &Memory{
		option:        option,
		imap:          map[string]*item{},
		maxMemRecords: MAX_MEM__RECORDS,
		maxMemSize:    MAX_MEM__SIZE,
		stopCh:        make(chan struct{}),
	}

	go mem.startCheckLoop()
	return mem, nil
}

func (m *Memory) Set(key string, value interface{}) error {
	size := int(unsafe.Sizeof(m.imap))
	if size > m.maxMemSize || len(m.imap) > m.maxMemRecords {
		return MemStackOverflow
	}
	m.mutex.Lock()
	m.imap[key] = &item{
		elem: value,
	}
	m.mutex.Unlock()
	return nil
}

func (m *Memory) Get(key string) (interface{}, error) {
	m.mutex.RLock()
	value, ok := m.imap[key]
	if !ok {
		return nil, KeyNotExist
	}
	m.mutex.RUnlock()
	return value.elem, nil
}

func (m *Memory) Del(key string) error {
	m.mutex.RLock()
	if _, ok := m.imap[key]; !ok {
		return nil
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	if _, ok := m.imap[key]; !ok {
		return nil
	}
	delete(m.imap, key)
	m.mutex.Unlock()
	return nil
}

func (m *Memory) Expire(key string, sec int) error {
	m.mutex.RLock()
	if _, ok := m.imap[key]; !ok {
		return KeyNotExist
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	value, ok := m.imap[key]
	if !ok {
		return KeyNotExist
	}
	value.expire = time.Now().Add(time.Duration(sec))
	m.mutex.Unlock()
	return nil
}

func (m *Memory) Exist(key string) (bool, error) {
	m.mutex.RLock()
	if _, ok := m.imap[key]; !ok {
		return false, nil
	}
	m.mutex.RUnlock()
	return true, nil
}

func (m *Memory) TTL(key string) (time.Duration, error) {
	var sec time.Duration
	m.mutex.RLock()
	if value, ok := m.imap[key]; !ok {
		return 0, KeyNotExist
	} else {
		sec = value.expire.Sub(time.Now())
	}
	m.mutex.RUnlock()
	return sec, nil
}

func (m *Memory) ComplexCmd(cmd string, values ...interface{}) (interface{}, error) {
	// ::TODO
	return nil, MemoryUnsupportedThisOpError
}

// warning: Close will delete all memory
func (m *Memory) Close() error {
	m.stopCh <- struct{}{}
	return nil
}

// *******************check key expire**********************
func (m *Memory) startCheckLoop() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			m.CheckExpireItems()
		case <-m.stopCh:
			m.clear()
		}
	}
	ticker.Stop()
	return
}

func (m *Memory) CheckExpireItems() {
	var keys []string
	curr := time.Now()
	m.mutex.RLock()
	for key, imap := range m.imap {
		if imap.expire.Before(curr) && !imap.expire.IsZero() {
			keys = append(keys, key)
		}
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	for _, key := range keys {
		delete(m.imap, key)
	}
	m.mutex.Unlock()
	return
}

func (m *Memory) clear() {
	m.mutex.Lock()
	m.imap = map[string]*item{}
	m.mutex.Unlock()

	close(m.stopCh)
	return
}
