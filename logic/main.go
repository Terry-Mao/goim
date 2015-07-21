package main

import (
	"flag"
	"runtime"

	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/perf"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	log.Info("logic[%s] start", Ver)
	perf.Init(Conf.PprofBind)
	// start rpc
	if err := InitRouterRpc(Conf.RouterPPCAddrs, Conf.RouterRPCRetry); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
