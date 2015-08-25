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
	perf.Init(Conf.PprofAddrs)
	// router rpc
	if err := InitRouter(); err != nil {
		log.Warn("router rpc current can't connect, retry")
	}
	if err := InitKafka(Conf.KafkaAddrs); err != nil {
		panic(err)
	}
	// logic rpc
	if err := InitRPC(NewDefaultAuther()); err != nil {
		panic(err)
	}
	if err := InitHTTP(); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
