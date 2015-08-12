package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/protorpc"
)

var (
	logicRpcClient *protorpc.Client
	logicRpcQuit   = make(chan struct{}, 1)

	logicService           = "RPC"
	logicServiceConnect    = "RPC.Connect"
	logicServiceDisconnect = "RPC.Disconnect"
)

func InitLogicRpc(network, addr string) (err error) {
	logicRpcClient, err = protorpc.Dial(network, addr)
	if err != nil {
		log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", network, addr, err)
		return
	}
	go protorpc.Reconnect(&logicRpcClient, logicRpcQuit, "tcp", addr)
	log.Debug("logic rpc addr %s:%s connected", network, addr)
	return
}
