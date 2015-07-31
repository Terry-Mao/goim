package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	rpc "github.com/Terry-Mao/protorpc"
	"strconv"
)

var (
	//routerRPC *RandLB
	routerServiceMap = map[string]*rpc.Client{}
	routerRing       *ketama.HashRing

	routerService           = "RouterRPC"
	routerServicePing       = "RouterRPC.Ping"
	routerServiceConnect    = "RouterRPC.Connect"
	routerServiceDisconnect = "RouterRPC.Disconnect"
)

func InitRouterRpc(addrs []string) (err error) {
	var r *rpc.Client
	routerRing = ketama.NewRing(ketama.Base)
	for _, addr := range addrs {
		r, err = rpc.Dial("tcp", addr)
		if err != nil {
			log.Error("rpc.Dial(\"%s\") error(%s)", addr, err)
			return
		}
		go r.Ping(&r)
		log.Debug("router rpc addr:%s connect", addr)
		routerServiceMap[addr] = r
		routerRing.AddNode(addr, 1)
	}
	routerRing.Bake()

	return
}

func getRouterClient(userID int64) *rpc.Client {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	return routerServiceMap[node]
}
