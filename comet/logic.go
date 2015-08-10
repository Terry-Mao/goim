package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/protorpc"
)

var (
	logicRpcClient *protorpc.Client

	logicService           = "RPC"
	logicServicePing       = "RPC.Ping"
	logicServiceConnect    = "RPC.Connect"
	logicServiceDisconnect = "RPC.Disconnect"
)

func InitLogicRpc(addr string) (err error) {
	logicRpcClient, err = protorpc.Dial("tcp", addr)
	if err != nil {
		log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
		return
	}
	var quit chan struct{}
	go protorpc.Reconnect(&logicRpcClient, quit, "tcp", addr)
	log.Debug("logic rpc addr:%s connected", addr)

	return
}
