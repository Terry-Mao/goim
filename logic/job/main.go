package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"runtime"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	log.LoadConfiguration(Conf.Log)
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := InitComet(Conf.Comets); err != nil {
		panic(err)
	}
	MergeRoomServers()
	go SyncRoomServers()
	InitPush()
	if err := InitKafka(); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
