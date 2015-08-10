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
	log.Info("router[%s] start", VERSION)
	// start prof
	perf.Init(Conf.PprofAddrs)
	// start rpc
	buckets := make([]*Bucket, Conf.Bucket)
	for i := 0; i < Conf.Bucket; i++ {
		buckets[i] = NewBucket(Conf.Session, Conf.Server, Conf.Cleaner)
	}
	if err := InitRPC(buckets); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
