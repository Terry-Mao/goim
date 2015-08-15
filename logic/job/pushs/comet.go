package main

import (
	log "code.google.com/p/log4go"
	inet "github.com/Terry-Mao/goim/libs/net"
	"github.com/Terry-Mao/protorpc"
)

var (
	cometServiceMap = make(map[int32]**protorpc.Client)
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
	for serverID, addrs := range addrs {
		var (
			rpcClient     *protorpc.Client
			quit          chan struct{}
			network, addr string
		)
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		rpcClient, err = protorpc.Dial(network, addr)
		if err != nil {
			log.Error("protorpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}

		go protorpc.Reconnect(&rpcClient, quit, network, addr)
		log.Info("rpc addr:%s connected", addr)

		cometServiceMap[serverID] = &rpcClient
	}

	return
}

// 通过serverID获取机器client
func getClient(serverID int32) (*protorpc.Client, error) {
	if client, ok := cometServiceMap[serverID]; !ok || *client == nil {
		return nil, ErrComet
	} else {
		return *client, nil
	}
}
