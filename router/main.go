package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"github.com/Terry-Mao/goim/libs/perf"
	"runtime"
)

const (
	VERSION = "0.1"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	// init buckets
	InitBuckets()
	log.Info("router[%s] start", VERSION)
	// start prof
	perf.Init(Conf.PprofBind)
	// start http
	if err := InitHttp(); err != nil {
		panic(err)
	}
	// start rpc
	if err := InitRPC(); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
