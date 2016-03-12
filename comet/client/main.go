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
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	if Conf.Type == ProtoTCP {
		initTCP()
	} else if Conf.Type == ProtoWebsocket {
		initWebsocket()
	} else if Conf.Type == ProtoWebsocketTLS {
		initWebsocketTLS()
	}
}
