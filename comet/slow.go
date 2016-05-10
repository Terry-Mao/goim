package main

import (
	"goim/libs/proto"
	"log"
	"os"
	"sync/atomic"
	"time"
)

const globalTimeDelay = 100 * time.Millisecond

const (
	SlowLogTypeSend    = "send"
	SlowLogTypeReceive = "receive"
	SlowLogTypeFinish  = "finish"
)

var (
	globalNowTime int64
	slowLog       *log.Logger
)

func initSlowLog(file string) (err error) {
	var fd *os.File
	fd, err = os.Open(file)
	if err != nil {
		return err
	}
	slowLog = log.New(fd, "", log.LstdFlags)
	go globalTimeProcess()
	return
}

// globalTimeProcess update nowTime per globalTimeDelay time.
func globalTimeProcess() {
	atomic.StoreInt64(&globalNowTime, time.Now().UnixNano())
	for {
		atomic.StoreInt64(&globalNowTime, time.Now().UnixNano())
		time.Sleep(globalTimeDelay)
	}
}

func globalNow() time.Time {
	return time.Unix(0, atomic.LoadInt64(&globalNowTime))
}

func LogSlow(logType string, key string, p *proto.Proto) {
	if p == nil || p.Time.IsZero() {
		return
	}
	// slow log
	userTime := atomic.LoadInt64(&globalNowTime) - p.Time.UnixNano()
	if userTime >= int64(Conf.SlowTime) {
		slowLog.Printf("logType:%s key:%s userTime:%fs slowtime:%fs msg:%s\n", logType, key, time.Duration(userTime).Seconds(), Conf.SlowTime.Seconds(), string(p.Body))
	}
}
