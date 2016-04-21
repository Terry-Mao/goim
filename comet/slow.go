package main

import (
	"goim/libs/proto"
	"log"
	"os"
	"time"
)

const globalTimeDelay = 100 * time.Millisecond

var (
	globalNowTime time.Time
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
	globalNowTime = time.Now()
	go globalTimeProc()
}

// globalTimeProc update nowTime per globalTideDelay time.
func globalTimeProc() {
	for {
		globalNowTime = time.Now()
		time.Sleep(globalTimeDelay)
	}
}

func logSlow(key string, p *proto.Proto) {
	// slow log
	userTime := globalNowTime.Sub(p.Time).Seconds()
	if userTime >= Conf.SlowTime.Seconds() {
		slowLog.Printf("key:%s proto:%s userTime:%fs slowTime:%fs\n", key, p.String(), userTime, Conf.SlowTime.Seconds())
	}
}
