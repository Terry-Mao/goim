package main

import (
	"time"

	log "code.google.com/p/log4go"

	"github.com/Terry-Mao/goim/protorpc"
)

var (
	logicRpcClient *protorpc.Client

	logicService           = "RPC"
	logicServicePing       = "RPC.Ping"
	logicServiceConnect    = "RPC.Connect"
	logicServiceDisconnect = "RPC.Disconnect"
)

func InitLogicRpc(addr string, retry time.Duration) (err error) {
	logicRpcClient, err = protorpc.Dial("tcp", addr)
	if err != nil {
		log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
		return
	}
	go rpcPing(addr, logicRpcClient, retry)
	log.Debug("logic rpc addr:%s connected", addr)

	return
}

func rpcPing(addr string, c *protorpc.Client, retry time.Duration) {
	var err error
	for {
		if err = c.Call(logicServicePing, nil, nil); err != nil {
			log.Error("c.Call(\"%s\", nil, nil) error(%v), retry after:%ds", logicServicePing, err, retry/time.Second)
			rpcTmp, err := protorpc.Dial("tcp", addr)
			if err != nil {
				log.Error("protorpc.Dial(\"tcp\", %s) error(%v)", addr, err)
				time.Sleep(retry)
				continue
			}
			c = rpcTmp
			time.Sleep(retry)
			continue
		}
		log.Debug("rpc ping:%s ok", addr)
		time.Sleep(retry)
		continue
	}
}
