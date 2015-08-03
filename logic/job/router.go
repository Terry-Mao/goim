package main

import (
	log "code.google.com/p/log4go"
	proto "github.com/Terry-Mao/goim/proto/router"
	"github.com/Terry-Mao/gopush-cluster/ketama"
	"github.com/Terry-Mao/protorpc"
	"strconv"
)

var (
	routerServiceMap = map[string]*protorpc.Client{}
	routerRing       *ketama.HashRing

	routerService            = "RouterRPC"
	routerServicePing        = "RouterRPC.Ping"
	routerServiceConnect     = "RouterRPC.Connect"
	routerServiceDisconnect  = "RouterRPC.Disconnect"
	routerServiceGet         = "RouterRPC.Get"
	routerServiceMGet        = "RouterRPC.MGet"
	routerServiceGetSeqCount = "RouterRPC.GetSeqCount"
)

func InitRouterRpc(addrs []string) (err error) {
	var r *protorpc.Client
	routerRing = ketama.NewRing(ketama.Base)
	for _, addr := range addrs {
		r, err = protorpc.Dial("tcp", addr)
		if err != nil {
			log.Error("protorpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}
		go r.Ping(&r)
		log.Debug("router protorpc addr:%s connect", addr)
		routerServiceMap[addr] = r
		routerRing.AddNode(addr, 1)
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
