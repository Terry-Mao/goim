package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/protorpc"
)

var (
	cometServiceMap = make(map[int32]*protorpc.Client)
)

const (
	CometService          = "PushRPC"
	CometServicePing      = "PushRPC.Ping"
	CometServicePushMsg   = "PushRPC.PushMsg"
	CometServicePushMsgs  = "PushRPC.PushMsgs"
	CometServiceMPushMsg  = "PushRPC.MPushMsg"
	CometServiceMPushMsgs = "PushRPC.MPushMsgs"
)

func InitCometRpc(addrs map[int32]string) (err error) {
	for serverID, addr := range addrs {
		var rpcClient *protorpc.Client
		rpcClient, err = protorpc.Dial("tcp", addr)
		if err != nil {
			log.Error("protorpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}
		var quit chan struct{}
		go protorpc.Reconnect(&rpcClient, quit, "tcp", addr)
		log.Info("rpc addr:%s connected", addr)

		cometServiceMap[serverID] = rpcClient
	}

	return
}

// 通过serverID获取机器client
func getClient(serverID int32) *protorpc.Client {
	return cometServiceMap[serverID]
}
