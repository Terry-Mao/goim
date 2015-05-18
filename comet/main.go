package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"github.com/Terry-Mao/goim/libs/perf"
	"runtime"
)

var (
	DefaultServer *Server
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
	DefaultServer = NewServer()
	if err := InitTCP(); err != nil {
		panic(err)
	}
	if err := InitHttpPush(); err != nil {
		panic(err)
	}
	// start rpc
	if err := InitRPCPush(); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
