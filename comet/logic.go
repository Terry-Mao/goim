package main

import (
	"net/rpc"
	"time"

	log "code.google.com/p/log4go"
	inet "github.com/Terry-Mao/goim/libs/net"
	"github.com/Terry-Mao/goim/libs/proto"
)

var (
	logicRpcClient *rpc.Client
	logicRpcQuit   = make(chan struct{}, 1)

	logicService           = "RPC"
	logicServicePing       = "RPC.Ping"
	logicServiceConnect    = "RPC.Connect"
	logicServiceDisconnect = "RPC.Disconnect"
)

func InitLogicRpc(addrs string) (err error) {
	var network, addr string
	if network, addr, err = inet.ParseNetwork(addrs); err != nil {
		log.Error("inet.ParseNetwork() error(%v)", err)
		return
	}
	logicRpcClient, err = rpc.Dial(network, addr)
	if err != nil {
		log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", network, addr, err)
	}
	go Reconnect(&logicRpcClient, logicRpcQuit, network, addr)
	log.Debug("logic rpc addr %s:%s connected", network, addr)
	return
}

// Reconnect for ping rpc server and reconnect with it when it's crash.
func Reconnect(dst **rpc.Client, quit chan struct{}, network, address string) {
	var (
		tmp    *rpc.Client
		err    error
		call   *rpc.Call
		ch     = make(chan *rpc.Call, 1)
		client = *dst
		args   = proto.NoArg{}
		reply  = proto.NoReply{}
	)
	for {
		select {
		case <-quit:
			return
		default:
			if client != nil {
				call = <-client.Go(logicServicePing, &args, &reply, ch).Done
				if call.Error != nil {
					log.Error("rpc ping %s error(%v)", address, call.Error)
				}
			}
			if client == nil || call.Error != nil {
				if tmp, err = rpc.Dial(network, address); err == nil {
					*dst = tmp
					client = tmp
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
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
