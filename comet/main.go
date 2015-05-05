package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"github.com/Terry-Mao/goim/libs/perf"
	"runtime"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	log.Info("comet[%s] start", Ver)
	perf.Init(Conf.PprofBind)
	if err := InitRSA(); err != nil {
		panic(err)
	}
	defaultServer := NewServer()
	if err := InitTCP(defaultServer); err != nil {
		panic(err)
	}
	if err := InitPush(defaultServer); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
	// listen
	// go
	// protocol
	// read & write
}
