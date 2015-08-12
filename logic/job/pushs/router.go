package main

import (
	log "code.google.com/p/log4go"
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
	routerServiceMap = map[string]*protorpc.Client{}
	routerRing       *ketama.HashRing
	routerQuit       = make(chan struct{}, 1)
)

func InitRouter() (err error) {
	var (
		r *rpc.Client
		i = 0
	)
	routerRing = ketama.NewRing(ketama.Base)
	for serverId, addr := range Conf.RouterRPCAddrs {
		r, err = rpc.Dial(Conf.RouterRPCNetworks[i], addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", Conf.RouterRPCNetworks[i], addr, err)
			return
		}
		go rpc.Reconnect(&r, routerQuit, Conf.RouterRPCNetworks[i], addr)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[serverId] = r
		routerRing.AddNode(serverId, 1)
		i++
	}
	routerRing.Bake()
	return
}

func getRouterClient(userID int64) *protorpc.Client {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	return routerServiceMap[node]
}

// 获取在线数量
func getOnlineCount(userID int64) (count int32, err error) {
	c := getRouterClient(userID)
	arg := &proto.GetSeqCountArg{UserId: userID}
	reply := &proto.GetSeqCountReply{}
	if err = c.Call(routerServiceGetSeqCount, arg, reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%s)", routerServiceGetSeqCount, *arg, err)
		return
	}
	return reply.Count, nil
}
