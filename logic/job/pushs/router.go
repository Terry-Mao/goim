package main

import (
	log "code.google.com/p/log4go"
	inet "github.com/Terry-Mao/goim/libs/net"
	proto "github.com/Terry-Mao/goim/proto/router"
	"github.com/Terry-Mao/gopush-cluster/ketama"
	"github.com/Terry-Mao/protorpc"
	rpc "github.com/Terry-Mao/protorpc"
	"strconv"
)

const (
	routerService            = "RouterRPC"
	routerServicePing        = "RouterRPC.Ping"
	routerServiceConnect     = "RouterRPC.Connect"
	routerServiceDisconnect  = "RouterRPC.Disconnect"
	routerServiceGet         = "RouterRPC.Get"
	routerServiceMGet        = "RouterRPC.MGet"
	routerServiceGetSeqCount = "RouterRPC.GetSeqCount"
)

var (
	routerServiceMap = map[string]**protorpc.Client{}
	routerRing       *ketama.HashRing
	routerQuit       = make(chan struct{}, 1)
)

func InitRouter() (err error) {
	var (
		network, addr string
	)
	routerRing = ketama.NewRing(ketama.Base)
	for serverId, addrs := range Conf.RouterRPCAddrs {
		// WARN r must every recycle changed for reconnect
		var (
			r          *rpc.Client
			routerQuit = make(chan struct{}, 1)
		)
		if network, addr, err = inet.ParseNetwork(addrs); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		r, err = rpc.Dial(network, addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", network, addr, err)
		}
		go rpc.Reconnect(&r, routerQuit, network, addr)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[serverId] = &r
		routerRing.AddNode(serverId, 1)
	}
	routerRing.Bake()
	return
}

func getRouterClient(userID int64) (*protorpc.Client, error) {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	if client, ok := routerServiceMap[node]; !ok || *client == nil {
		return nil, ErrRouter
	} else {
		return *client, nil
	}
}

// 获取在线数量
func getOnlineCount(userID int64) (count int32, err error) {
	c, err := getRouterClient(userID)
	if err != nil {
		log.Error("getRouterClient(\"%d\") error(%v)", userID, err)
		return
	}
	arg := &proto.GetSeqCountArg{UserId: userID}
	reply := &proto.GetSeqCountReply{}
	if err = c.Call(routerServiceGetSeqCount, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceGetSeqCount, *arg, err)
		return
	}
	return reply.Count, nil
}
