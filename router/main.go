package main

import (
	"flag"
	"goim/libs/perf"
	"runtime"

	log "github.com/thinkboy/log4go"
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
	// start monitor
	if Conf.MonitorOpen {
		InitMonitor(Conf.MonitorAddrs)
	}
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
