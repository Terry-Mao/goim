package main

import (
	"goim/libs/proto"
	"log"
	"os"
	"time"
)

const globalTimeDelay = 100 * time.Millisecond

const (
	SlowLogTypeSend    = "send"
	SlowLogTypeReceive = "receive"
	SlowLogTypeFinish  = "finish"
)

var (
	globalNowTime *time.Time
	slowLog       *log.Logger
)

func initSlowLog(file string) (err error) {
	var fd *os.File
	fd, err = os.Open(file)
	if err != nil {
		return
	}
	slowLog = log.New(fd, "", log.LstdFlags)

	startGlobalTime()
	return
}

func startGlobalTime() {
	now := time.Now()
	globalNowTime = &now
	go globalTimeProc()
}

// globalTimeProc update nowTime per globalTideDelay time.
func globalTimeProc() {
	for {
		now := time.Now()
		globalNowTime = &now
		time.Sleep(globalTimeDelay)
	}
}

func LogSlow(logType string, key string, p *proto.Proto) {
	// slow log
	userTime := globalNowTime.Sub(p.Time).Seconds()
	if userTime >= Conf.SlowTime.Seconds() {
		slowLog.Printf("logType:%s key:%s userTime:%fs slowtime:%fs msg:%s\n", logType, key, userTime, Conf.SlowTime.Seconds(), string(p.Body))
	}
}
