package cache

import (
	"testing"

	"github.com/microsvs/base/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var memConn Connection

func init() {
	var fns []func(*Options)
	var err error
	fns = append(fns, func(opt *Options) {
		opt.DBIndex = 0
	})
	fns = append(fns, func(opt *Options) {
		opt.MaxConns = 1000000
	})
	fns = append(fns, func(opt *Options) {
		opt.TimeoutIdleConn = 10
	})
	if memConn, err = NewMemoryConnection(fns...); err != nil {
		panic(err.Error())
	}
}

func TestSetAndGet(t *testing.T) {
	var err error
	if err = memConn.Set("/tmp/b2b/test01", "tmp"); err != nil {
		t.Error(err.Error())
	}
	if err = memConn.Set("/tmp/b2b/test02", "tmp"); err != nil {
		t.Error(err.Error())
	}
	var value interface{}
	var tmp string
	if value, err = memConn.Get("/tmp/b2b/test01"); err != nil {
		t.Error(err.Error())
	}
	if err = utils.GenericTypeConvert(value, &tmp); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, "tmp", tmp, "`/tmp/b2b/test01` != tmp")
}

func TestDel(t *testing.T) {
	var err error
	if err = memConn.Del("/tmp/b2b/test01"); err != nil {
		t.Error(err.Error())
	}
	if err = memConn.Del("/tmp/b2b/test02"); err != nil {
		t.Error(err.Error())
	}
}
