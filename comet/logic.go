package main

import (
	log "code.google.com/p/log4go"
	proto "github.com/Terry-Mao/goim/proto/logic"
	"github.com/Terry-Mao/protorpc"
	"time"
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
	}
	go protorpc.Reconnect(&logicRpcClient, logicRpcQuit, network, addr)
	log.Debug("logic rpc addr %s:%s connected", network, addr)
	return
}

func connect(p *Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	if logicRpcClient == nil {
		err = ErrLogic
		return
	}
	arg := &proto.ConnArg{Token: string(p.Body), Server: Conf.ServerId}
	reply := &proto.ConnReply{}
	if err = logicRpcClient.Call(logicServiceConnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	key = reply.Key
	rid = reply.RoomId
	heartbeat = 5 * 60 * time.Second
	return
}

func disconnect(key string, roomId int32) (has bool, err error) {
	if logicRpcClient == nil {
		err = ErrLogic
		return
	}
	arg := &proto.DisconnArg{Key: key, RoomId: roomId}
	reply := &proto.DisconnReply{}
	if err = logicRpcClient.Call(logicServiceDisconnect, arg, reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	has = reply.Has
	return
}
