package main

import (
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

func globalTimeProc() {
	for {
		globalNowTime = time.Now()
		time.Sleep(globalTimeDelay)
	}
}
