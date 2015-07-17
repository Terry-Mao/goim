package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"github.com/Terry-Mao/goim/libs/perf"
	"runtime"
	"time"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	perf.Init(Conf.PprofBind)
	if Conf.Type == ProtoTCP {
		initTCP()
	} else if Conf.Type == ProtoWebsocket {
		initWebsocket()
	}
	time.Sleep(10 * time.Second)
}
