package timer

/*
该文件产生的原因：各个微服务调用time.Now()系统调用过于频繁
为了减少系统调用次数，优化调用性能，这里采用定时任务，1s更新时间
这样，服务端通过base.Now就可以获取最新的时间
*/

import "time"

const (
	REFRESH__TIME__INTERVAL = 1 // 1s
)

var Now = time.Now()

func refreshTime() {
	var timestamp int64
	timer := time.NewTicker(1 * time.Second)
	for {
		<-timer.C
		timestamp = Now.Unix() + 1
		Now = time.Unix(timestamp, 0)
	}
}

func init() {
	go refreshTime()
}
