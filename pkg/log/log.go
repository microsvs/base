package log

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/microsvs/base/pkg/env"
	ierrors "github.com/microsvs/base/pkg/errors"
	"github.com/microsvs/base/pkg/rpc"
	"github.com/microsvs/base/pkg/timer"
	"github.com/microsvs/base/pkg/types"
	"github.com/microsvs/base/pkg/utils"
	"github.com/ncw/directio"
)

const (
	LOG__CHAN_LENGHT          = 10
	LOG__CHAN_AND_CACHE__RATE = 0.7
	// 目录和时间维度支持自定义
	LOG__FILE__DEM = "20060102"  // 日志时间维度, 天
	LOG__DIR__RULE = "200601/02" // 日志目录结构, 结构：月份/天
)

var (
	logInstance  *Log
	appname      string
	buildLog     bytes.Buffer
	logCacheChan = make(chan []byte, LOG__CHAN_LENGHT)
	logStruPool  = sync.Pool{
		New: func() interface{} {
			return new(LogStru)
		},
	}
	cachePool = sync.Pool{
		New: func() interface{} {
			return make([][]byte, 0, int(LOG__CHAN_LENGHT*LOG__CHAN_AND_CACHE__RATE))
		},
	}
	cacheLog = cachePool.Get().([][]byte)
)

//output to console,concurrent limit. Get os ulimit num and console max concurrent num
var consoleLimit = make(chan int, 100)

func init() {
	var (
		appname       string
		logPathPrefix string
		err           error
	)
	pathes := strings.Split(os.Args[0], "/")
	appname = strings.Split(pathes[len(pathes)-1], ".")[0]
	if logPathPrefix, err = env.Get(env.ServiceLog); err != nil {
		panic(err)
	}

	if logInstance, err = LogInstance(
		logPathPrefix,
		appname,
		timer.Now.Format(LOG__FILE__DEM),
		LOG__DIR__RULE,
	); err != nil {
		panic("init FreeGoLog failed, err: " + err.Error())
	}
	go receiveCacheLogLoop()
}

//ConsoleInfo 日志包含字段
type ConsoleInfo struct {
	Mobile string `json:"mobile"`
	UserID string `json:"userid"`
	Client string `json:"client"`
}

//LogStru used inside the whole freego log system
type LogStru struct {
	Module      string          `json:"module"`
	Level       string          `json:"level"`
	Client      string          `json:"client"`
	Token       string          `json:"token"`
	TraceID     string          `json:"traceid"`
	TraceRPCID  string          `json:"rpcid"`
	TraceTag    TraceServiceTag `json:"tracetag"`
	FromService rpc.FGService   `json:"servicefrom"`
	ToService   rpc.FGService   `json:"serviceto"`
	FromIP      string          `json:"ipfrom"`
	ToIP        string          `json:"ipto"`
	Time        time.Time       `json:"timestamp"`
	Interval    int             `json:"interval"`
	File        string          `json:"file"`
	LineNo      int             `json:"line"`
	Cnt         []byte          `json:"cnt"`
	ConsoleInfo ConsoleInfo     `json:"consoleinfo"`
}

//String format log to readable format
func (l *LogStru) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s %d %s %s %s %s %s %d %s", l.Module, l.Level, l.Client, l.Token, l.TraceID, l.TraceRPCID, l.TraceTag, l.FromService, l.ToService, l.FromIP, l.ToIP, l.File, l.LineNo, l.Cnt)
}

func logRaw(level string, format string, v ...interface{}) {
	logC(nil, level, rpc.FGSIgnore, TSTIgnore, format, v...)
}

func logC(ctx context.Context, level string, service rpc.FGService, tag TraceServiceTag, format string, v ...interface{}) {

	request := rpc.GetContextFromKey(ctx, rpc.KeyRawRequest, &http.Request{}).(*http.Request)
	user := rpc.GetContextFromKey(ctx, rpc.KeyUser, &types.User{}).(*types.User)

	logStru := logStruPool.Get().(*LogStru)
	logStru.Module = appname
	logStru.Level = level
	logStru.Time = timer.Now
	logStru.File, logStru.LineNo = WhereAmI()
	logStru.ToService = service
	logStru.FromService = rpc.GetContextFromKey(ctx, rpc.KeyService, rpc.FGSIgnore).(rpc.FGService)
	logStru.TraceID = rpc.GetContextFromKey(ctx, rpc.KeyTraceID, "-").(string)
	logStru.TraceRPCID = rpc.GetContextFromKey(ctx, rpc.KeyRPCID, "0").(string)
	logStru.FromIP = utils.GetClientIPAdress(request)
	fmt.Fprintf(&buildLog, format, v...)
	logStru.Cnt = buildLog.Bytes()
	logStru.ConsoleInfo = ConsoleInfo{
		Mobile: user.Mobile,
		Client: strings.Split(logStru.TraceID, ":")[0],
		UserID: user.ID,
	}
	//print to screen and save to logfile has different format
	cnt, _ := jsoniter.Marshal(logStru)
	logStruPool.Put(logStru)
	writeCacheLog(cnt)
	buildLog.Reset()
	return
}

//DebugRaw print debug information
func DebugRaw(format string, v ...interface{}) {
	logRaw("debug", format, v...)
}

//Debug debug with context
func Debug(ctx context.Context, format string, v ...interface{}) {
	logC(ctx, "debug", rpc.FGSIgnore, TSTIgnore, format, v...)
}

//InfoRaw print debug information
func InfoRaw(format string, v ...interface{}) {
	logRaw("info", format, v...)
}

//Info debug with context
func Info(ctx context.Context, format string, v ...interface{}) {
	logC(ctx, "info", rpc.FGSIgnore, TSTIgnore, format, v...)
}

//ErrorRaw print debug information
func ErrorRaw(format string, v ...interface{}) {
	logRaw("error", format, v...)
}

//Error debug with context
func Error(ctx context.Context, format interface{}, v ...interface{}) (interface{}, error) {
	if fmt, ok := format.(string); ok {
		logC(ctx, "error", rpc.FGSIgnore, TSTIgnore, fmt, v...)
	} else {
		err := format.(error)
		file, line := WhereAmI()
		logC(ctx, "error", rpc.FGSIgnore, TSTIgnore, "%s %d %s", file, line, err.Error())
	}
	return nil, ierrors.FGEInternalError
}

//TraceServiceTag 发起服务调用还是结束服务调用
type TraceServiceTag int

/*
TraceServiceTag 发起服务调用还是结束服务调用
*/
const (
	TSTIgnore TraceServiceTag = iota
	TSTStart
	TSTEnd
)

//Trace print debug information
func Trace(ctx context.Context, service rpc.FGService, tag TraceServiceTag, format string, v ...interface{}) {
	logC(ctx, "trace", service, tag, format, v...)
}

//WhereAmI need to know last call
func WhereAmI() (string, int) {
	depth := 0
	for {
		_, file, line, ok := runtime.Caller(depth)
		if !ok {
			return "unknown", 0
		}
		if strings.HasSuffix(file, "/log.go") {
			depth++
			continue
		}
		cols := strings.Split(file, "/")
		return cols[len(cols)-1], line
	}
}

// return the source filename after the last slash
func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	}
	return original[i+1:]
}

//FormatErrorCode genurate errorcode Desc
func FormatErrorCode(v ...ierrors.FGErrorCode) string {
	s := ""
	for _, e := range v {
		s += fmt.Sprintf("\n%d:%s", int(10000+e), e.String())
	}
	return s
}

// ********************log goroutine*******************
// 总体思路：
// 1. 微服务起一个日志收集的goroutine。使用channel队列往写日志goroutine写数据。
// 2. 使用sync.Pool内存池缓冲日志记录，缓冲池满后一次性io，比如：设置缓冲池size为1kb
type Log struct {
	sync.Mutex
	ServiceName string
	BaseDir     string   // log base dir
	TimeLevel   string   // LOG__FILE__DEM 时间维度生成文件
	DirRule     string   // LOG__DIR__RULE 时间目录结构
	Fd          *os.File // 文件句柄
}

func LogInstance(baseDir, appname, timeLevel, dirRule string) (*Log, error) {
	if logInstance != nil {
		return logInstance, nil
	}
	logInstance = &Log{
		BaseDir:     baseDir,
		ServiceName: appname,
		TimeLevel:   timeLevel,
		DirRule:     dirRule,
	}
	if logInstance != nil {
		if err := logInstance.openFile(); err != nil {
			return nil, err
		}
	}
	return logInstance, nil
}

func (f *Log) openFile() (err error) {
	if f == nil {
		return fmt.Errorf("firstly you must be init log params")
	}
	now := timer.Now
	dirPath := fmt.Sprintf("%s/%s/%s", f.BaseDir, f.ServiceName, now.Format(f.DirRule))
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return
	}
	filePath := fmt.Sprintf("%s/%s_%s.log", dirPath, f.ServiceName, now.Format(LOG__FILE__DEM))
	f.Fd, err = directio.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	return
}

func (f *Log) closeFile() (err error) {
	if f == nil {
		return fmt.Errorf("firstly you must be init log params")
	}
	if f.Fd == nil {
		return
	}
	if err = f.Fd.Close(); err != nil {
		return
	}
	return
}

func (f *Log) updateTimeLevel(currTimeLevel string) {
	f.TimeLevel = currTimeLevel
	return
}

func (f *Log) write(now time.Time, buff *bytes.Buffer) (err error) {
	currTimeLevel := now.Format(LOG__FILE__DEM)
	f.Lock()
	defer f.Unlock()
	if f.TimeLevel < currTimeLevel {
		if err = f.closeFile(); err != nil {
			return
		}
		if err = f.openFile(); err != nil {
			return
		}
		f.updateTimeLevel(currTimeLevel)
	}
	tmpBts := buff.Bytes()
	for times := 0; times < len(tmpBts)/directio.BlockSize; times++ {
		f.Fd.Write(tmpBts[times*directio.BlockSize : (times+1)*directio.BlockSize])
	}
	return nil
}

func writeCacheLog(logRecord []byte) {
	logCacheChan <- logRecord
}

func receiveCacheLogLoop() {
	var (
		logRecord []byte
		ok        bool
		buffer    = new(bytes.Buffer)
		err       error
	)
	for {
		if logRecord, ok = <-logCacheChan; !ok {
			return
		}
		cacheLog = append(cacheLog, logRecord)
		if len(cacheLog) == cap(cacheLog) {
			for _, logRecordPtr := range cacheLog {
				buffer.Write(logRecordPtr)
				buffer.WriteByte('\n')
			}
			if err = logInstance.write(timer.Now, buffer); err != nil {
				panic(err.Error())
			}
			buffer.Reset()
			cacheLog = cacheLog[:0]
		}
	}
	return
}
