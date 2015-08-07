package main

import (
	log "code.google.com/p/log4go"
	"github.com/Terry-Mao/goim/libs/hash/ketama"
	rpc "github.com/Terry-Mao/protorpc"
	"strconv"
)

var (
	routerServiceMap = map[string]*rpc.Client{}
	routerQuit       = make(chan struct{}, 1)
	routerRing       *ketama.HashRing

	routerService           = "RouterRPC"
	routerServiceConnect    = "RouterRPC.Connect"
	routerServiceDisconnect = "RouterRPC.Disconnect"
)

func InitRouter() (err error) {
	var r *rpc.Client
	routerRing = ketama.NewRing(ketama.Base)
	for i := 0; i < len(Conf.RouterRPCAddrs); i++ {
		r, err = rpc.Dial(Conf.RouterRPCNetworks[i], Conf.RouterRPCAddrs[i])
		if err != nil {
			log.Error("rpc.Dial(\"%s\", \"%s\") error(%s)", Conf.RouterRPCNetworks[i], Conf.RouterRPCAddrs[i], err)
			return
		}
		go rpc.Reconnect(&r, routerQuit, Conf.RouterRPCNetworks[i], Conf.RouterRPCAddrs[i])
		log.Debug("router rpc addr:%s connect", Conf.RouterRPCAddrs[i])
		routerServiceMap[Conf.RouterRPCAddrs[i]] = r
		routerRing.AddNode(Conf.RouterRPCAddrs[i], 1)
	}
	routerRing.Bake()
	return
}

func RouterClient(userID int64) *rpc.Client {
	node := routerRing.Hash(strconv.FormatInt(userID, 10))
	return routerServiceMap[node]
}
