package main

import (
	"flag"
	"runtime"

	log "github.com/thinkboy/log4go"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	log.LoadConfiguration(Conf.Log)
	runtime.GOMAXPROCS(runtime.NumCPU())
	//comet
	if err := InitComet(Conf.Comets); err != nil {
		log.Warn("comet rpc current can't connect, retry")
	}
	//round
	round := NewRound(RoundOptions{
		Timer:     Conf.Timer,
		TimerSize: Conf.TimerSize,
	})
	//room
	InitRoomBucket(round,
		RoomOptions{
			BatchNum:   Conf.RoomBatch,
			SignalTime: Conf.RoomSignal,
		})
	//room info
	MergeRoomServers()
	go SyncRoomServers()
	InitPush()
	if err := InitKafka(); err != nil {
		panic(err)
	}
	// block until a signal is received.
	InitSignal()
}
